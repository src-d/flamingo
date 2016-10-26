package slack

import (
	"bytes"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/mvader/slack"
	"github.com/src-d/flamingo"
	"github.com/src-d/flamingo/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type helloCtrl struct {
	sync.RWMutex
	msgs        []flamingo.Message
	calledIntro int
}

func (*helloCtrl) CanHandle(msg flamingo.Message) bool {
	return msg.Text == "hello"
}

func (c *helloCtrl) Handle(bot flamingo.Bot, msg flamingo.Message) error {
	c.Lock()
	defer c.Unlock()
	c.msgs = append(c.msgs, msg)
	return nil
}

func (c *helloCtrl) HandleIntro(b flamingo.Bot, channel flamingo.Channel) error {
	c.Lock()
	defer c.Unlock()
	c.calledIntro++
	return nil
}

func TestControllerFor(t *testing.T) {
	assert := assert.New(t)
	cli := newClient("", ClientOptions{})
	var ctrl flamingo.Controller = &helloCtrl{}
	cli.AddController(ctrl)

	result, ok := cli.ControllerFor(flamingo.Message{Text: "hello"})
	assert.True(ok)
	assert.Equal(reflect.ValueOf(ctrl.Handle).Pointer(), reflect.ValueOf(result).Pointer())

	result, ok = cli.ControllerFor(flamingo.Message{Text: "goodbye"})
	assert.False(ok)
	assert.Nil(result)
}

func TestActionHandler(t *testing.T) {
	assert := assert.New(t)
	cli := newClient("", ClientOptions{})
	var handler flamingo.ActionHandler = func(b flamingo.Bot, a flamingo.Action) {
	}
	cli.AddActionHandler("foo", handler)

	result, ok := cli.ActionHandler("foo")
	assert.True(ok)
	assert.Equal(reflect.ValueOf(handler).Pointer(), reflect.ValueOf(result).Pointer())

	result, ok = cli.ActionHandler("bar")
	assert.False(ok)
	assert.Nil(result)
}

func TestRunAndStopWebhook(t *testing.T) {
	assert := assert.New(t)
	cli := newClient("xAB3yVzGS4BQ3O9FACTa8Ho4", ClientOptions{
		Webhook: WebhookOptions{Addr: "127.0.0.1:8989"},
	})
	go cli.runWebhook()
	<-time.After(50 * time.Millisecond)

	data := url.Values{}
	data.Set("payload", testCallback)
	resp, err := http.Post("http://127.0.0.1:8989", "application/x-www-form-urlencoded", bytes.NewBufferString(data.Encode()))
	assert.Nil(err)
	assert.Equal(resp.StatusCode, http.StatusOK)

	cli.shutdownWebhook <- struct{}{}
	<-time.After(50 * time.Millisecond)

	client := http.Client{
		Timeout: 50 * time.Millisecond,
	}
	resp, err = client.Post("http://127.0.0.1:8989", "application/x-www-form-urlencoded", bytes.NewBufferString(data.Encode()))
	assert.NotNil(err)
}

type clientBotMock struct {
	sync.RWMutex
	stopped       bool
	actions       []slack.AttachmentActionCallback
	channels      []string
	handledJobs   int
	conversations []string
}

func (b *clientBotMock) stop() {
	b.Lock()
	defer b.Unlock()
	b.stopped = true
}

func (b *clientBotMock) handleAction(channel string, action slack.AttachmentActionCallback) {
	b.Lock()
	defer b.Unlock()
	b.channels = append(b.channels, channel)
	b.actions = append(b.actions, action)
}

func (b *clientBotMock) handleJob(job flamingo.Job) {
	b.Lock()
	defer b.Unlock()
	b.handledJobs++
}

func (b *clientBotMock) addConversation(id string) error {
	b.Lock()
	defer b.Unlock()
	b.conversations = append(b.conversations, id)
	return nil
}

func TestRunAndStop(t *testing.T) {
	assert := assert.New(t)
	cli := newClient("xAB3yVzGS4BQ3O9FACTa8Ho4", ClientOptions{
		Webhook: WebhookOptions{Addr: "127.0.0.1:8787", Enabled: true},
		Debug:   true,
	})

	bot := &clientBotMock{}
	bot2 := &clientBotMock{}
	cli.bots["bot"] = bot
	cli.bots["bot2"] = bot2

	var stopped = make(chan struct{}, 1)
	go func() {
		cli.Run()
		stopped <- struct{}{}
	}()

	<-time.After(50 * time.Millisecond)
	data := url.Values{}
	data.Set("payload", testCallback)
	resp, err := http.Post("http://127.0.0.1:8787", "application/x-www-form-urlencoded", bytes.NewBufferString(data.Encode()))
	assert.Nil(err)
	assert.Equal(resp.StatusCode, http.StatusOK)

	assert.Nil(cli.Stop())
	<-time.After(50 * time.Millisecond)

	select {
	case <-stopped:
	case <-time.After(50 * time.Millisecond):
		assert.FailNow("did not stop")
	}
	assert.True(bot.stopped)
	assert.Equal(1, len(bot.actions))
	assert.Equal("test_callback", bot.actions[0].CallbackID)
	assert.Equal("channel", bot.channels[0])

	assert.True(bot2.stopped)
	assert.Equal(0, len(bot2.actions))
}

func TestSetIntroHandler(t *testing.T) {
	cli := newClient("", ClientOptions{})
	ctrl := &helloCtrl{}
	cli.SetIntroHandler(ctrl)
	assert.Equal(t, reflect.ValueOf(ctrl).Pointer(), reflect.ValueOf(cli.introHandler).Pointer())
}

func TestHandleIntro(t *testing.T) {
	cli := newClient("", ClientOptions{})
	ctrl := &helloCtrl{}
	cli.HandleIntro(nil, flamingo.Channel{})
	cli.SetIntroHandler(ctrl)
	cli.HandleIntro(nil, flamingo.Channel{})
	assert.Equal(t, 1, ctrl.calledIntro)
}

func TestStartScheduledJobsAndStop(t *testing.T) {
	cli := newClient("", ClientOptions{})
	cli.AddScheduledJob(flamingo.NewIntervalSchedule(1*time.Second), func(_ flamingo.Bot, _ flamingo.Channel) error {
		assert.FailNow(t, "scheduled job was run")
		return nil
	})

	go cli.Run()
	<-time.After(50 * time.Millisecond)
	cli.Stop()
}

func TestScheduledJobs(t *testing.T) {
	cli := newClient("", ClientOptions{})
	cli.AddScheduledJob(flamingo.NewIntervalSchedule(50*time.Millisecond), func(_ flamingo.Bot, _ flamingo.Channel) error {
		return nil
	})
	mock := &clientBotMock{}
	cli.bots["foo"] = mock

	go cli.Run()
	defer cli.Stop()

	<-time.After(90 * time.Millisecond)
	mock.RLock()
	defer mock.RUnlock()
	assert.Equal(t, 1, mock.handledJobs)
}

func TestLoadFromStorage(t *testing.T) {
	cli := newClient("", ClientOptions{})
	storage := storage.NewMemory()
	storage.StoreBot(flamingo.StoredBot{ID: "1", Token: "foo"})
	storage.StoreConversation(flamingo.StoredConversation{ID: "2", BotID: "1"})
	cli.SetStorage(storage)
	assert.Nil(t, cli.loadFromStorage())
	_, ok := cli.bots["1"]
	assert.True(t, ok)

	_, ok = cli.bots["1"].(*botClient).conversations["2"]
	assert.True(t, ok)
}

func TestSave(t *testing.T) {
	cli := newClient("", ClientOptions{})
	storage := storage.NewMemory()
	cli.SetStorage(storage)
	cli.AddBot("1", "foo", nil)
	cli.bots["1"].addConversation("2")

	ok, _ := storage.BotExists(flamingo.StoredBot{ID: "1"})
	assert.True(t, ok)
	ok, _ = storage.ConversationExists(flamingo.StoredConversation{ID: "2"})
	assert.True(t, ok)
}

func TestAddBotOnlyOnce(t *testing.T) {
	assert := assert.New(t)
	cli := newClient("", ClientOptions{})
	storage := storage.NewMemory()
	cli.SetStorage(storage)
	cli.AddBot("1", "foo", nil)
	cli.AddBot("2", "foo", nil)
	cli.AddBot("1", "foo", nil)

	bots, _ := storage.LoadBots()
	assert.Equal(2, len(bots))
	assert.Equal(2, len(cli.loadedBots))
}

func TestWrap(t *testing.T) {
	require := require.New(t)
	cli := newClient("", ClientOptions{})

	var result []string
	cli.Use(func(bot flamingo.Bot, msg flamingo.Message, next flamingo.HandlerFunc) error {
		result = append(result, "1")
		return next(bot, msg)
	})

	cli.Use(func(bot flamingo.Bot, msg flamingo.Message, next flamingo.HandlerFunc) error {
		result = append(result, "2")
		return next(bot, msg)
	})

	handler := cli.wrap(func(_ flamingo.Bot, msg flamingo.Message) error {
		result = append(result, msg.Text)
		return nil
	})

	require.Nil(handler(nil, flamingo.Message{Text: "3"}))
	require.Equal([]string{"1", "2", "3"}, result)
}

func TestBroadcast(t *testing.T) {
	require := require.New(t)
	mock := newSlackRTMMock()

	bot1 := &botClient{
		id:  "bot1",
		rtm: mock,
		conversations: map[string]*botConversation{
			"conv1-1": &botConversation{
				rtm:     mock,
				channel: flamingo.Channel{ID: "1"},
			},
			"conv1-2": &botConversation{
				rtm:     mock,
				channel: flamingo.Channel{ID: "2"},
			},
		},
	}
	bot2 := &botClient{
		id:  "bot2",
		rtm: mock,
		conversations: map[string]*botConversation{
			"conv2-1": &botConversation{
				rtm:     mock,
				channel: flamingo.Channel{ID: "3"},
			},
			"conv2-2": &botConversation{
				rtm:     mock,
				channel: flamingo.Channel{ID: "4"},
			},
		},
	}
	bot3 := &botClient{
		id:  "bot3",
		rtm: mock,
		conversations: map[string]*botConversation{
			"conv3-1": &botConversation{
				rtm:     mock,
				channel: flamingo.Channel{ID: "5"},
			},
			"conv3-2": &botConversation{
				rtm:     mock,
				channel: flamingo.Channel{ID: "6"},
			},
		},
	}
	cli := &slackClient{
		bots: map[string]clientBot{
			"bot1": bot1,
			"bot2": bot2,
			"bot3": bot3,
		},
	}

	filter := func(bot string, channel flamingo.Channel) bool {
		return bot != "bot1" && channel.ID != "3"
	}
	bots, convs, errors, err := cli.Broadcast(flamingo.NewOutgoingMessage("foo"), filter)
	require.Nil(err)
	require.Equal(uint64(0), errors)
	require.Equal(uint64(2), bots)
	require.Equal(uint64(3), convs)
	require.Equal(int(convs), len(mock.msgs))
}

func TestSend(t *testing.T) {
	require := require.New(t)
	mock := newapiMock(nil)
	bot := &bot{
		id:  "bar",
		api: mock,
		channel: flamingo.Channel{
			ID: "foo",
		},
	}

	require.Nil(send(bot, flamingo.NewOutgoingMessage("foo")))
	require.Nil(send(bot, flamingo.Image{URL: "foo"}))
	require.Nil(send(bot, flamingo.Form{Text: "foo"}))
	require.Equal(len(mock.msgs), 3)
}

func newClient(token string, options ClientOptions) *slackClient {
	options.Debug = true
	options.Webhook.VerificationToken = token
	return NewClient(token, options).(*slackClient)
}
