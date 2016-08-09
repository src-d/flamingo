package slack

import (
	"errors"
	"testing"
	"time"

	"github.com/mvader/flamingo"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
)

type postMessageArgs struct {
	channel string
	text    string
	params  slack.PostMessageParameters
}

type posterMock struct {
	msgs     []postMessageArgs
	callback func(postMessageArgs) bool
}

func (m *posterMock) PostMessage(channel, text string, params slack.PostMessageParameters) (string, string, error) {
	args := postMessageArgs{channel, text, params}
	m.msgs = append(m.msgs, args)
	if m.callback != nil {
		if !m.callback(args) {
			return "", "", errors.New("error")
		}
	}

	return "", "", nil
}

func (m *posterMock) GetUserInfo(id string) (*slack.User, error) {
	return &slack.User{
		ID:       id,
		Name:     "user",
		RealName: "real name",
	}, nil
}

func (m *posterMock) GetChannelInfo(id string) (*slack.Channel, error) {
	ch := &slack.Channel{}
	ch.ID = id
	ch.Name = "channel"
	return ch, nil
}

func newPosterMock(callback func(postMessageArgs) bool) *posterMock {
	return &posterMock{
		callback: callback,
	}
}

func TestSay(t *testing.T) {
	assert := assert.New(t)
	mock := newPosterMock(nil)
	bot := &bot{
		poster: mock,
		channel: flamingo.Channel{
			ID: "foo",
		},
	}

	assert.Nil(bot.Say(flamingo.NewOutgoingMessage("hi there")))
	assert.Nil(bot.Say(flamingo.OutgoingMessage{
		ChannelID: "bar",
		Text:      "hi there you too",
	}))

	assert.Equal(len(mock.msgs), 2)
	assert.Equal(mock.msgs[0].channel, "foo")
	assert.Equal(mock.msgs[0].text, "hi there")

	assert.Equal(mock.msgs[1].channel, "bar")
	assert.Equal(mock.msgs[1].text, "hi there you too")
}

func TestReply(t *testing.T) {
	assert := assert.New(t)
	mock := newPosterMock(nil)
	bot := &bot{
		poster: mock,
		channel: flamingo.Channel{
			ID: "foo",
		},
	}

	assert.Nil(bot.Reply(
		flamingo.Message{
			User: flamingo.User{Username: "baz"},
		},
		flamingo.NewOutgoingMessage("hi there"),
	))

	assert.Equal(len(mock.msgs), 1)
	assert.Equal(mock.msgs[0].channel, "foo")
	assert.Equal(mock.msgs[0].text, "@baz: hi there")
}

func TestAsk(t *testing.T) {
	assert := assert.New(t)
	mock := newPosterMock(nil)
	ch := make(chan *slack.MessageEvent, 1)
	bot := &bot{
		poster: mock,
		channel: flamingo.Channel{
			ID: "foo",
		},
		msgs: ch,
	}

	ch <- &slack.MessageEvent{
		Msg: slack.Msg{
			Text:        "fine, thanks",
			Attachments: []slack.Attachment{slack.Attachment{}},
		},
	}

	msg, err := bot.Ask(flamingo.NewOutgoingMessage("how are you?"))
	assert.Nil(err)
	assert.Equal(msg.Text, "fine, thanks")
}

func TestConversation(t *testing.T) {
	assert := assert.New(t)
	mock := newPosterMock(nil)
	ch := make(chan *slack.MessageEvent, 2)
	bot := &bot{
		poster: mock,
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

	msgs, err := bot.Conversation(flamingo.Conversation{
		flamingo.NewOutgoingMessage("hi, how are you?"),
		flamingo.NewOutgoingMessage("fine, too"),
	})
	assert.Nil(err)
	assert.Equal(len(msgs), 2)

	assert.Equal(msgs[0].Text, "fine, thanks. And you?")
	assert.Equal(msgs[1].Text, "cool")
}

func TestWaitForActionIgnorePolicy(t *testing.T) {
	assert := assert.New(t)
	mock := newPosterMock(nil)
	ch := make(chan *slack.MessageEvent, 1)
	actions := make(chan slack.AttachmentActionCallback, 1)
	bot := &bot{
		poster: mock,
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
	assert.Nil(err)
	assert.Equal(action.Extra.(slack.AttachmentActionCallback).CallbackID, "foo")
	assert.Equal(len(mock.msgs), 0)
}

func TestWaitForActionReplyPolicy(t *testing.T) {
	assert := assert.New(t)
	mock := newPosterMock(nil)
	ch := make(chan *slack.MessageEvent, 1)
	actions := make(chan slack.AttachmentActionCallback, 1)
	bot := &bot{
		poster: mock,
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
				slack.AttachmentAction{
					Name:  "foo",
					Value: "foo-1",
				},
			},
		}
	}()

	action, err := bot.WaitForAction("foo", flamingo.ReplyPolicy("wait, what?"))
	assert.Nil(err)
	assert.Equal(action.UserAction.Value, "foo-1")
	assert.Equal(action.Extra.(slack.AttachmentActionCallback).CallbackID, "foo")
	assert.Equal(len(mock.msgs), 2)
}

func TestForm(t *testing.T) {
	assert := assert.New(t)
	mock := newPosterMock(nil)
	bot := &bot{
		id:     "bar",
		poster: mock,
		channel: flamingo.Channel{
			ID: "foo",
		},
	}

	assert.Nil(bot.Form(flamingo.Form{
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
	}))

	assert.Equal(len(mock.msgs), 1)
	assert.Equal(mock.msgs[0].channel, "foo")
	assert.Equal(mock.msgs[0].text, " ")
	assert.Equal(len(mock.msgs[0].params.Attachments), 1)
}
