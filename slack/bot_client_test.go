package slack

import (
	"testing"
	"time"

	"github.com/mvader/slack"
	"github.com/stretchr/testify/assert"
)

type slackRTMMock struct {
	events chan slack.RTMEvent
	apiMock
}

func (m *slackRTMMock) IncomingEvents() chan slack.RTMEvent {
	return m.events
}

func TestHandleAction(t *testing.T) {
	assert := assert.New(t)

	client := newBotClient(
		"aaaa",
		&slackRTMMock{},
		NewClient("", ClientOptions{Debug: true}).(*slackClient),
	)
	defer client.stop()

	convo := &botConversation{
		actions:  make(chan slack.AttachmentActionCallback, 1),
		shutdown: make(chan struct{}, 1),
		closed:   make(chan struct{}, 1),
		messages: make(chan *slack.MessageEvent, 1),
	}
	client.conversations["bbbb"] = convo
	go convo.run()

	client.handleAction("bbbb", slack.AttachmentActionCallback{
		CallbackID: "foo",
	})

	select {
	case action := <-convo.actions:
		assert.Equal("foo", action.CallbackID)
	case <-time.After(50 * time.Millisecond):
		assert.FailNow("action was not received by conversation")
	}

	client.handleAction("cccc", slack.AttachmentActionCallback{
		CallbackID: "bar",
	})

	select {
	case <-convo.actions:
		assert.FailNow("action should not have been received by conversation")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestHandleRTMEvent(t *testing.T) {
	assert := assert.New(t)
	mock := &slackRTMMock{
		events: make(chan slack.RTMEvent),
	}

	client := newBotClient(
		"aaaa",
		mock,
		NewClient("", ClientOptions{Debug: true}).(*slackClient),
	)
	defer client.stop()

	convo := &botConversation{
		actions:  make(chan slack.AttachmentActionCallback, 1),
		shutdown: make(chan struct{}, 1),
		closed:   make(chan struct{}, 1),
		messages: make(chan *slack.MessageEvent, 1),
	}
	client.conversations["bbbb"] = convo
	convo.closed <- struct{}{}

	events := []interface{}{
		&slack.LatencyReport{},
		&slack.RTMError{},
		&slack.InvalidAuthEvent{},
		&slack.MessageEvent{
			Msg: slack.Msg{
				Channel: "bbbb",
				Text:    "text",
			},
		},
	}

	for _, e := range events {
		mock.events <- slack.RTMEvent{Data: e}
	}

	select {
	case msg := <-convo.messages:
		assert.Equal("text", msg.Text)
	case <-time.After(100 * time.Millisecond):
		assert.FailNow("didn't get the message")
	}
}

func TestHandleRTMEventOpenConvo(t *testing.T) {
	assert := assert.New(t)
	mock := &slackRTMMock{
		events: make(chan slack.RTMEvent),
	}

	client := newBotClient(
		"aaaa",
		mock,
		NewClient("", ClientOptions{Debug: true}).(*slackClient),
	)
	defer client.stop()

	convo := &botConversation{
		actions:  make(chan slack.AttachmentActionCallback, 1),
		shutdown: make(chan struct{}, 1),
		closed:   make(chan struct{}, 1),
		messages: make(chan *slack.MessageEvent, 1),
	}
	client.conversations["bbbb"] = convo
	convo.closed <- struct{}{}

	mock.events <- slack.RTMEvent{
		Data: &slack.MessageEvent{
			Msg: slack.Msg{
				Channel: "aaaa",
				Text:    "text",
			},
		},
	}

	<-time.After(50 * time.Millisecond)
	assert.Equal(2, len(client.conversations))
}

func TestHandleIMCreatedEvent(t *testing.T) {
	assert := assert.New(t)
	mock := &slackRTMMock{
		events: make(chan slack.RTMEvent),
	}

	ctrl := &helloCtrl{}
	cli := NewClient("", ClientOptions{Debug: true}).(*slackClient)
	cli.SetIntroHandler(ctrl)

	client := newBotClient(
		"aaaa",
		mock,
		cli,
	)
	defer client.stop()

	mock.events <- slack.RTMEvent{
		Data: &slack.IMCreatedEvent{
			Channel: slack.ChannelCreatedInfo{
				ID: "D345345",
			},
		},
	}

	<-time.After(50 * time.Millisecond)
	assert.Equal(1, len(client.conversations))
	assert.Equal(1, ctrl.calledIntro)
}

func TestHandleGroupJoinedEvent(t *testing.T) {
	assert := assert.New(t)
	mock := &slackRTMMock{
		events: make(chan slack.RTMEvent),
	}

	ctrl := &helloCtrl{}
	cli := NewClient("", ClientOptions{Debug: true}).(*slackClient)
	cli.SetIntroHandler(ctrl)

	client := newBotClient(
		"aaaa",
		mock,
		cli,
	)
	defer client.stop()

	ev := slack.RTMEvent{Data: &slack.GroupJoinedEvent{}}
	ev.Data.(*slack.GroupJoinedEvent).Channel.ID = "G394820"
	mock.events <- ev

	<-time.After(50 * time.Millisecond)
	assert.Equal(1, len(client.conversations))
	assert.Equal(1, ctrl.calledIntro)
}
