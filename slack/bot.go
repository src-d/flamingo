package slack

import (
	"fmt"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/mvader/slack"
	"github.com/src-d/flamingo"
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
	actions chan slack.AttachmentActionCallback
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

	message, err := b.WaitForMessage()
	return ts, message, err
}

func (b *bot) WaitForMessage() (flamingo.Message, error) {
	for {
		msg := <-b.msgs
		if msg.BotID == b.ID() || msg.User == b.ID() {
			log15.Debug("received message from self, ignoring")
			continue
		}
		return b.convertMessage(msg)
	}
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

		msg, err := b.WaitForMessage()
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
	return b.WaitForActions([]string{id}, policy)
}

func (b *bot) WaitForActions(ids []string, policy flamingo.ActionWaitingPolicy) (flamingo.Action, error) {
	for {
		select {
		case action, ok := <-b.actions:
			if !ok {
				continue
			}

			if inSlice(ids, action.CallbackID) {
				return convertAction(action, b.api)
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

func inSlice(slice []string, str string) bool {
	for _, s := range slice {
		if str == s {
			return true
		}
	}
	return false
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

func (b *bot) InvokeAction(id string, user flamingo.User, action flamingo.UserAction) {
	var ch slack.Channel
	ch.Name = b.channel.Name
	ch.ID = b.channel.ID

	b.actions <- slack.AttachmentActionCallback{
		Actions: []slack.AttachmentAction{
			slack.AttachmentAction{
				Name:  action.Name,
				Value: action.Value,
			},
		},
		Channel:    ch,
		CallbackID: id,
		User: slack.User{
			ID:       user.ID,
			Name:     user.Username,
			RealName: user.Name,
			IsBot:    user.IsBot,
			Profile: slack.UserProfile{
				Email: user.Email,
			},
		},
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

	return newMessage(user, b.channel, src.Msg), nil
}

func (b *bot) findUser(id string) (flamingo.User, error) {
	user, err := b.api.GetUserInfo(id)
	if err != nil {
		log15.Error("unable to find user", "id", id)
		return flamingo.User{}, err
	}

	return convertUser(user), nil
}
