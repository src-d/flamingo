package slack

import (
	"fmt"
	"time"

	"github.com/mvader/flamingo"
	"github.com/nlopes/slack"
)

type messagePoster interface {
	PostMessage(string, string, slack.PostMessageParameters) (string, string, error)
	GetUserInfo(string) (*slack.User, error)
}

type bot struct {
	id      string
	channel flamingo.Channel
	poster  messagePoster
	msgs    <-chan *slack.MessageEvent
	actions <-chan slack.AttachmentActionCallback
}

func (b *bot) ID() string {
	return b.id
}

func (b *bot) Reply(replyTo flamingo.Message, msg flamingo.OutgoingMessage) error {
	msg.Text = fmt.Sprintf("@%s: %s", replyTo.User.Username, msg.Text)
	return b.Say(msg)
}

func (b *bot) Ask(msg flamingo.OutgoingMessage) (flamingo.Message, error) {
	if err := b.Say(msg); err != nil {
		return flamingo.Message{}, err
	}

	return b.waitForMessage()
}

func (b *bot) waitForMessage() (flamingo.Message, error) {
	msg := <-b.msgs
	return b.convertMessage(msg)
}

func (b *bot) Conversation(convo flamingo.Conversation) ([]flamingo.Message, error) {
	var messages = make([]flamingo.Message, 0, len(convo))
	for _, m := range convo {
		if err := b.Say(m); err != nil {
			return nil, err
		}

		msg, err := b.waitForMessage()
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (b *bot) Say(msg flamingo.OutgoingMessage) error {
	channel := b.channel.ID
	if msg.ChannelID != "" {
		channel = msg.ChannelID
	}

	params, err := createPostParams(msg)
	if err != nil {
		return err
	}

	_, _, err = b.poster.PostMessage(channel, msg.Text, params)
	return err
}

func (b *bot) WaitForAction(id string, policy flamingo.ActionWaitingPolicy) (flamingo.Action, error) {
	for {
		select {
		case action := <-b.actions:
			if action.CallbackID == id {
				return convertAction(action), nil
			} else if policy.Reply {
				err := b.Say(flamingo.NewOutgoingMessage(policy.Message))
				if err != nil {
					return flamingo.Action{}, err
				}
			}
		case <-time.After(300 * time.Millisecond):
		}
	}
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

	return newMessage(user, b.channel, src), nil
}

func (b *bot) findUser(id string) (flamingo.User, error) {
	user, err := b.poster.GetUserInfo(id)
	if err != nil {
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
