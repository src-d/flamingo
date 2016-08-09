package slack

import (
	"log"
	"time"

	"github.com/mvader/flamingo"
	"github.com/nlopes/slack"
)

type botConversation struct {
	bot      string
	channel  flamingo.Channel
	rtm      slackRTM
	actions  chan slack.AttachmentActionCallback
	messages chan *slack.MessageEvent
	delegate handlerDelegate
}

func newBotConversation(bot, channel string, rtm slackRTM, delegate handlerDelegate) (*botConversation, error) {
	ch, err := rtm.GetChannelInfo(channel)
	if err != nil {
		return nil, err
	}

	return &botConversation{
		rtm: rtm,
		bot: bot,
		channel: flamingo.Channel{
			ID:    ch.ID,
			Name:  ch.Name,
			Type:  flamingo.SlackClient,
			IsDM:  !ch.IsChannel,
			Extra: ch,
		},
		actions:  make(chan slack.AttachmentActionCallback),
		messages: make(chan *slack.MessageEvent),
		delegate: delegate,
	}, nil
}

func (c *botConversation) run() {
	for {
		select {
		case msg := <-c.messages:
			message, err := c.convertMessage(msg)
			if err != nil {
				log.Printf("error converting message: %q", err.Error())
				continue
			}

			ctrl, ok := c.delegate.ControllerFor(message)
			if !ok {
				log.Printf("no controller for message %q", message.Text)
				continue
			}

			if err := ctrl.Handle(c.createBot(), message); err != nil {
				log.Printf("error handling message: %q", err.Error())
			}

		case action := <-c.actions:
			handler, ok := c.delegate.ActionHandler(action.CallbackID)
			if !ok {
				log.Printf("no handler for callback %q", action.CallbackID)
				continue
			}

			handler(c.createBot(), convertAction(action))
		case <-time.After(50 * time.Millisecond):
		}
	}
}

func (c *botConversation) createBot() flamingo.Bot {
	return &bot{
		id:      c.bot,
		channel: c.channel,
		api:     c.rtm,
		msgs:    c.messages,
		actions: c.actions,
	}
}

func (c *botConversation) convertMessage(src *slack.MessageEvent) (flamingo.Message, error) {
	var userID = src.Msg.User
	if userID == "" {
		userID = src.Msg.BotID
	}

	user, err := c.rtm.GetUserInfo(userID)
	if err != nil {
		return flamingo.Message{}, err
	}

	return newMessage(flamingo.User{
		ID:       userID,
		Username: user.Name,
		Name:     user.RealName,
		IsBot:    user.IsBot,
		Type:     flamingo.SlackClient,
		Extra:    user,
	}, c.channel, src.Msg), nil
}
