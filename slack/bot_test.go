package slack

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/mvader/slack"
	"github.com/src-d/flamingo"
	"github.com/stretchr/testify/require"
)

type postMessageArgs struct {
	channel string
	text    string
	params  slack.PostMessageParameters
}

type updateMessageArgs struct {
	channel string
	id      string
	text    string
	params  slack.UpdateMessageParameters
}

type apiMock struct {
	users    map[string]*slack.User
	msgs     []postMessageArgs
	updates  []updateMessageArgs
	callback func(postMessageArgs) bool
}

func (m *apiMock) setUser(user *slack.User) {
	m.users[user.Name] = user
}

func (m *apiMock) GetUserByUsername(username string) (*slack.User, error) {
	u, ok := m.users[username]
	if !ok {
		return nil, errors.New("not_found")
	}

	return u, nil
}

func (m *apiMock) PostMessage(channel, text string, params slack.PostMessageParameters) (string, string, error) {
	args := postMessageArgs{channel, text, params}
	m.msgs = append(m.msgs, args)
	if m.callback != nil {
		if !m.callback(args) {
			return "", "", errors.New("error")
		}
	}

	return "", "", nil
}

func (m *apiMock) UpdateMessage(channel, id, text string, params slack.UpdateMessageParameters) (string, string, string, error) {
	args := updateMessageArgs{channel, id, text, params}
	m.updates = append(m.updates, args)
	return "", "", "", nil
}

func (m *apiMock) GetUserInfo(id string) (*slack.User, error) {
	return &slack.User{
		ID:       id,
		Name:     "user",
		RealName: "real name",
	}, nil
}

func (m *apiMock) GetChannelInfo(id string) (*slack.Channel, error) {
	ch := &slack.Channel{}
	ch.ID = id
	ch.Name = "channel"
	return ch, nil
}

func (m *apiMock) OpenIMChannel(ch string) (bool, bool, string, error) {
	return true, true, ch, nil
}

func newapiMock(callback func(postMessageArgs) bool) *apiMock {
	return &apiMock{
		callback: callback,
		users:    make(map[string]*slack.User),
	}
}

func ignoreID(id string, err error) error {
	return err
}

func ignoreIDs(_, _ string, err error) error {
	return err
}

func TestSay(t *testing.T) {
	require := require.New(t)
	mock := newapiMock(nil)
	bot := &bot{
		id:  "bar",
		api: mock,
		channel: flamingo.Channel{
			ID: "foo",
		},
	}

	require.Nil(ignoreID(bot.Say(flamingo.NewOutgoingMessage("hi there"))))
	require.Nil(ignoreID(bot.Say(flamingo.OutgoingMessage{
		ChannelID: "bar",
		Text:      "hi there you too",
	})))

	require.Equal(len(mock.msgs), 2)
	require.Equal(mock.msgs[0].channel, "foo")
	require.Equal(mock.msgs[0].text, "hi there")

	require.Equal(mock.msgs[1].channel, "bar")
	require.Equal(mock.msgs[1].text, "hi there you too")
}

func TestSayTo(t *testing.T) {
	require := require.New(t)
	mock := newapiMock(nil)
	mock.setUser(&slack.User{
		ID:   "destination",
		Name: "fooo",
	})
	bot := &bot{
		id:  "bar",
		api: mock,
		channel: flamingo.Channel{
			ID: "foo",
		},
	}

	require.Nil(ignoreIDs(bot.SayTo("fooo", flamingo.NewOutgoingMessage("hi there"))))

	require.Equal(len(mock.msgs), 1)
	require.Equal(mock.msgs[0].channel, "destination")
	require.Equal(mock.msgs[0].text, "hi there")
}

func TestImage(t *testing.T) {
	require := require.New(t)
	mock := newapiMock(nil)
	bot := &bot{
		id:  "bar",
		api: mock,
		channel: flamingo.Channel{
			ID: "foo",
		},
	}

	require.Nil(ignoreID(bot.Image(flamingo.Image{URL: "foo"})))

	require.Equal(len(mock.msgs), 1)
	require.Equal(len(mock.msgs[0].params.Attachments), 1)
	require.Equal(mock.msgs[0].params.Attachments[0].ImageURL, "foo")
}

func TestReply(t *testing.T) {
	require := require.New(t)
	mock := newapiMock(nil)
	bot := &bot{
		id:  "bar",
		api: mock,
		channel: flamingo.Channel{
			ID: "foo",
		},
	}

	require.Nil(ignoreID(bot.Reply(
		flamingo.Message{
			User: flamingo.User{Username: "baz"},
		},
		flamingo.NewOutgoingMessage("hi there"),
	)))

	require.Equal(len(mock.msgs), 1)
	require.Equal(mock.msgs[0].channel, "foo")
	require.Equal(mock.msgs[0].text, "@baz: hi there")
}

func TestAsk(t *testing.T) {
	require := require.New(t)
	mock := newapiMock(nil)
	ch := make(chan *slack.MessageEvent, 1)
	bot := &bot{
		id:  "bar",
		api: mock,
		channel: flamingo.Channel{
			ID: "foo",
		},
		msgs: ch,
	}

	ch <- &slack.MessageEvent{
		Msg: slack.Msg{
			Text:        "fine, thanks",
			Attachments: []slack.Attachment{{}},
		},
	}

	_, msg, err := bot.Ask(flamingo.NewOutgoingMessage("how are you?"))
	require.Nil(err)
	require.Equal(msg.Text, "fine, thanks")
}

func TestAskUntil(t *testing.T) {
	require := require.New(t)
	mock := newapiMock(nil)
	ch := make(chan *slack.MessageEvent, 2)
	bot := &bot{
		id:  "bar",
		api: mock,
		channel: flamingo.Channel{
			ID: "foo",
		},
		msgs: ch,
	}

	for i := 1; i <= 2; i++ {
		ch <- &slack.MessageEvent{
			Msg: slack.Msg{
				Text: fmt.Sprint(i),
			},
		}
	}

	_, msg, err := bot.AskUntil(flamingo.NewOutgoingMessage("how many eyes does a human have?"), func(msg flamingo.Message) *flamingo.OutgoingMessage {
		if msg.Text == "2" {
			return nil
		}

		return &flamingo.OutgoingMessage{Text: "nope"}
	})
	require.Nil(err)
	require.Equal(2, len(mock.msgs))
	require.Equal(msg.Text, "2")
}

func TestConversation(t *testing.T) {
	require := require.New(t)
	mock := newapiMock(nil)
	ch := make(chan *slack.MessageEvent, 2)
	bot := &bot{
		id:  "bar",
		api: mock,
		channel: flamingo.Channel{
			ID: "foo",
		},
		msgs: ch,
	}

	ch <- &slack.MessageEvent{
		Msg: slack.Msg{
			Text: "fine, thanks. And you?",
		},
	}

	ch <- &slack.MessageEvent{
		Msg: slack.Msg{
			Text: "cool",
		},
	}

	_, msgs, err := bot.Conversation(flamingo.Conversation{
		flamingo.NewOutgoingMessage("hi, how are you?"),
		flamingo.NewOutgoingMessage("fine, too"),
	})
	require.Nil(err)
	require.Equal(len(msgs), 2)

	require.Equal(msgs[0].Text, "fine, thanks. And you?")
	require.Equal(msgs[1].Text, "cool")
}

func TestWaitForActionIgnorePolicy(t *testing.T) {
	require := require.New(t)
	mock := newapiMock(nil)
	ch := make(chan *slack.MessageEvent, 1)
	actions := make(chan slack.AttachmentActionCallback, 1)
	bot := &bot{
		id:  "bar",
		api: mock,
		channel: flamingo.Channel{
			ID: "foo",
		},
		msgs:    ch,
		actions: actions,
	}

	ch <- &slack.MessageEvent{
		Msg: slack.Msg{
			Text: "some msg",
		},
	}

	go func() {
		<-time.After(50 * time.Millisecond)
		actions <- slack.AttachmentActionCallback{
			CallbackID: "bar",
		}

		actions <- slack.AttachmentActionCallback{
			CallbackID: "foo",
		}
	}()

	action, err := bot.WaitForAction("foo", flamingo.IgnorePolicy())
	require.Nil(err)
	require.Equal(action.Extra.(slack.AttachmentActionCallback).CallbackID, "foo")
	require.Equal(len(mock.msgs), 0)
}

var errUserInfo = errors.New("couldn't get user info")

type userInfoFailingAPI struct {
	slackAPI
}

func (api userInfoFailingAPI) GetUserInfo(_ string) (*slack.User, error) {
	return nil, errUserInfo
}

func TestWaitForActionConversionFail(t *testing.T) {
	actions := make(chan slack.AttachmentActionCallback, 1)
	bot := &bot{
		api:     &userInfoFailingAPI{newapiMock(nil)},
		actions: actions,
	}

	go func() {
		<-time.After(50 * time.Millisecond)
		actions <- slack.AttachmentActionCallback{CallbackID: "foo"}
	}()

	_, err := bot.WaitForAction("foo", flamingo.ActionWaitingPolicy{})
	require.Equal(t, err, errUserInfo)
}

func TestWaitForActionReplyPolicy(t *testing.T) {
	require := require.New(t)
	mock := newapiMock(nil)
	ch := make(chan *slack.MessageEvent, 1)
	actions := make(chan slack.AttachmentActionCallback, 1)
	bot := &bot{
		id:  "bar",
		api: mock,
		channel: flamingo.Channel{
			ID: "foo",
		},
		msgs:    ch,
		actions: actions,
	}

	ch <- &slack.MessageEvent{
		Msg: slack.Msg{
			Text: "some msg",
		},
	}

	go func() {
		<-time.After(150 * time.Millisecond)
		actions <- slack.AttachmentActionCallback{
			CallbackID: "bar",
		}

		actions <- slack.AttachmentActionCallback{
			CallbackID: "foo",
			Actions: []slack.AttachmentAction{
				{
					Name:  "foo",
					Value: "foo-1",
				},
			},
		}
	}()

	action, err := bot.WaitForAction("foo", flamingo.ReplyPolicy("wait, what?"))
	require.Nil(err)
	require.Equal(action.UserAction.Value, "foo-1")
	require.Equal(action.Extra.(slack.AttachmentActionCallback).CallbackID, "foo")
	require.Equal(len(mock.msgs), 2)
}

func TestForm(t *testing.T) {
	require := require.New(t)
	mock := newapiMock(nil)
	bot := &bot{
		id:  "bar",
		api: mock,
		channel: flamingo.Channel{
			ID: "foo",
		},
	}

	require.Nil(ignoreID(bot.Form(flamingo.Form{
		Title:   "title",
		Text:    "text",
		Color:   "color",
		Combine: true,
		Fields: []flamingo.FieldGroup{
			flamingo.NewButtonGroup("baz", flamingo.Button{
				Name:  "yes",
				Value: "yes",
				Text:  "Yes",
				Type:  flamingo.PrimaryButton,
			}),
			flamingo.NewTextFieldGroup(flamingo.TextField{
				Title: "title",
				Value: "value",
			}),
		},
	})))

	require.Equal(len(mock.msgs), 1)
	require.Equal(mock.msgs[0].channel, "foo")
	require.Equal(mock.msgs[0].text, " ")
	require.Equal(len(mock.msgs[0].params.Attachments), 1)
}

func TestSendFormTo(t *testing.T) {
	require := require.New(t)
	mock := newapiMock(nil)
	mock.setUser(&slack.User{
		ID:   "destination",
		Name: "fooo",
	})
	bot := &bot{
		id:  "bar",
		api: mock,
		channel: flamingo.Channel{
			ID: "foo",
		},
	}

	require.Nil(ignoreIDs(bot.SendFormTo("fooo", flamingo.Form{
		Title:   "title",
		Text:    "text",
		Color:   "color",
		Combine: true,
		Fields: []flamingo.FieldGroup{
			flamingo.NewButtonGroup("baz", flamingo.Button{
				Name:  "yes",
				Value: "yes",
				Text:  "Yes",
				Type:  flamingo.PrimaryButton,
			}),
			flamingo.NewTextFieldGroup(flamingo.TextField{
				Title: "title",
				Value: "value",
			}),
		},
	})))

	require.Equal(len(mock.msgs), 1)
	require.Equal(mock.msgs[0].channel, "destination")
	require.Equal(mock.msgs[0].text, " ")
	require.Equal(len(mock.msgs[0].params.Attachments), 1)
}

func TestUpdateForm(t *testing.T) {
	require := require.New(t)
	mock := newapiMock(nil)
	bot := &bot{
		id:  "bar",
		api: mock,
		channel: flamingo.Channel{
			ID: "foo",
		},
	}

	require.Nil(ignoreID(bot.UpdateForm("id", flamingo.Form{
		Title:   "title",
		Text:    "text",
		Color:   "color",
		Combine: true,
		Fields: []flamingo.FieldGroup{
			flamingo.NewButtonGroup("baz", flamingo.Button{
				Name:  "yes",
				Value: "yes",
				Text:  "Yes",
				Type:  flamingo.PrimaryButton,
			}),
			flamingo.NewTextFieldGroup(flamingo.TextField{
				Title: "title",
				Value: "value",
			}),
		},
	})))

	require.Equal(len(mock.msgs), 0)
	require.Equal(len(mock.updates), 1)
	require.Equal(mock.updates[0].channel, "foo")
	require.Equal(mock.updates[0].id, "id")
	require.Equal(mock.updates[0].text, " ")
	require.Equal(len(mock.updates[0].params.Attachments), 1)
}

func TestUpdateMessage(t *testing.T) {
	require := require.New(t)
	mock := newapiMock(nil)
	bot := &bot{
		id:  "bar",
		api: mock,
		channel: flamingo.Channel{
			ID: "foo",
		},
	}

	require.Nil(ignoreID(bot.UpdateMessage("id", "new text")))

	require.Equal(len(mock.msgs), 0)
	require.Equal(len(mock.updates), 1)
	require.Equal(mock.updates[0].channel, "foo")
	require.Equal(mock.updates[0].id, "id")
	require.Equal(mock.updates[0].text, "new text")
	require.Equal(len(mock.updates[0].params.Attachments), 0)
}

func TestInvokeAction(t *testing.T) {
	bot := &bot{
		channel: flamingo.Channel{
			ID: "chan",
		},
		actions: make(chan slack.AttachmentActionCallback, 1),
	}

	bot.InvokeAction(
		"action",
		flamingo.User{
			ID:       "foo",
			Name:     "bar",
			Username: "baz",
			Email:    "qux",
		},
		flamingo.UserAction{
			Name:  "fooo",
			Value: "baar",
		},
	)

	action := <-bot.actions
	require.Equal(t, "action", action.CallbackID)
	require.Equal(t, 1, len(action.Actions))
	require.Equal(t, "fooo", action.Actions[0].Name)
	require.Equal(t, "baar", action.Actions[0].Value)
	require.Equal(t, "foo", action.User.ID)
	require.Equal(t, "bar", action.User.RealName)
	require.Equal(t, "baz", action.User.Name)
	require.Equal(t, "qux", action.User.Profile.Email)
	require.Equal(t, "chan", action.Channel.ID)
}
