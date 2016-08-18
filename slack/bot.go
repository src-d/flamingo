package slack

import (
	"fmt"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/src-d/flamingo"
	"github.com/mvader/slack"
)

type slackAPI interface {
	PostMessage(string, string, slack.PostMessageParameters) (string, string, error)
	UpdateMessage(string, string, string, slack.UpdateMessageParameters) (string, string, string, error)
	GetUserInfo(string) (*slack.User, error)
	GetChannelInfo(string) (*slack.Channel, error)
}

type bot struct {
	id      string
	channel flamingo.Channel
	api     slackAPI
	msgs    <-chan *slack.MessageEvent
	actions <-chan slack.AttachmentActionCallback
}

func (b *bot) ID() string {
	return b.id
}

func (b *bot) Reply(replyTo flamingo.Message, msg flamingo.OutgoingMessage) (string, error) {
	msg.Text = fmt.Sprintf("@%s: %s", replyTo.User.Username, msg.Text)
	return b.Say(msg)
}

func (b *bot) Ask(msg flamingo.OutgoingMessage) (string, flamingo.Message, error) {
	ts, err := b.Say(msg)
	if err != nil {
		return "", flamingo.Message{}, err
	}

	message, err := b.waitForMessage()
	return ts, message, err
}

func (b *bot) waitForMessage() (flamingo.Message, error) {
	msg := <-b.msgs
	return b.convertMessage(msg)
}

func (b *bot) Conversation(convo flamingo.Conversation) ([]string, []flamingo.Message, error) {
	var messages = make([]flamingo.Message, 0, len(convo))
	var timestamps = make([]string, 0, len(convo))
	for _, m := range convo {
		ts, err := b.Say(m)
		if err != nil {
			return nil, nil, err
		}

		timestamps = append(timestamps, ts)

		msg, err := b.waitForMessage()
		if err != nil {
			return nil, nil, err
		}
		messages = append(messages, msg)
	}

	return timestamps, messages, nil
}

func (b *bot) Say(msg flamingo.OutgoingMessage) (string, error) {
	channel := b.channel.ID
	if msg.ChannelID != "" {
		channel = msg.ChannelID
	}

	_, ts, err := b.api.PostMessage(channel, msg.Text, createPostParams(msg))
	if err != nil {
		log15.Error("error posting message to channel", "channel", channel, "error", err.Error(), "text", msg.Text)
	}

	return ts, err
}

func (b *bot) WaitForAction(id string, policy flamingo.ActionWaitingPolicy) (flamingo.Action, error) {
	for {
		select {
		case action := <-b.actions:
			if action.CallbackID == id {
				return convertAction(action), nil
			} else if policy.Reply {
				log15.Debug("received action with another id waiting for action", "id", action.CallbackID)
				_, err := b.Say(flamingo.NewOutgoingMessage(policy.Message))
				if err != nil {
					return flamingo.Action{}, err
				}
			}
		case m := <-b.msgs:
			if policy.Reply {
				log15.Debug("received msg waiting for action, replying default msg", "text", m.Text)
				_, err := b.Say(flamingo.NewOutgoingMessage(policy.Message))
				if err != nil {
					return flamingo.Action{}, err
				}
			}
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func (b *bot) Form(form flamingo.Form) (string, error) {
	params := formToMessage(b.ID(), b.channel.ID, form)
	_, ts, err := b.api.PostMessage(b.channel.ID, " ", params)
	if err != nil {
		log15.Error("error posting form", "err", err.Error())
	}

	return ts, err
}

func (b *bot) Image(img flamingo.Image) (string, error) {
	_, ts, err := b.api.PostMessage(b.channel.ID, " ", imageToMessage(img))
	if err != nil {
		log15.Error("error posting image", "err", err.Error())
	}

	return ts, err
}

func (b *bot) UpdateMessage(id string, replacement string) (string, error) {
	_, ts, _, err := b.api.UpdateMessage(b.channel.ID, id, replacement, slack.NewUpdateMessageParameters())
	if err != nil {
		log15.Error("error updating message", "id", id, "err", err.Error())
	}

	return ts, err
}

func (b *bot) UpdateForm(id string, replacement flamingo.Form) (string, error) {
	msg := formToMessage(b.ID(), b.channel.ID, replacement)
	params := slack.NewUpdateMessageParameters()
	params.Attachments = msg.Attachments
	_, ts, _, err := b.api.UpdateMessage(b.channel.ID, id, " ", params)
	if err != nil {
		log15.Error("error updating form", "id", id, "err", err.Error())
	}

	return ts, err
}

func (b *bot) AskUntil(msg flamingo.OutgoingMessage, check flamingo.AnswerChecker) (string, flamingo.Message, error) {
	var (
		id  string
		m   flamingo.Message
		err error
	)

	for {
		id, m, err = b.Ask(msg)
		if err != nil {
			return "", flamingo.Message{}, nil
		}

		errMsg := check(m)
		if errMsg == nil {
			break
		}
		msg = *errMsg
	}

	return id, m, err
}

func (b *bot) convertMessage(src *slack.MessageEvent) (flamingo.Message, error) {
	var userID = src.Msg.User
	if userID == "" {
		userID = src.Msg.BotID
	}

	user, err := b.findUser(userID)
	if err != nil {
		return flamingo.Message{}, err
	}

	return newMessage(user, b.channel, src.Msg), nil
}

func (b *bot) findUser(id string) (flamingo.User, error) {
	user, err := b.api.GetUserInfo(id)
	if err != nil {
		log15.Error("unable to find user", "id", id)
		return flamingo.User{}, err
	}

	return flamingo.User{
		ID:       id,
		Username: user.Name,
		Name:     user.RealName,
		IsBot:    user.IsBot,
		Type:     flamingo.SlackClient,
		Extra:    user,
	}, nil
}
