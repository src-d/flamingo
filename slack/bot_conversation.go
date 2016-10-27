package slack

import (
	"strings"
	"sync"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/mvader/slack"
	"github.com/src-d/flamingo"
)

type botConversation struct {
	sync.RWMutex
	working  bool
	bot      string
	channel  flamingo.Channel
	rtm      slackRTM
	actions  chan slack.AttachmentActionCallback
	messages chan *slack.MessageEvent
	shutdown chan struct{}
	closed   chan struct{}
	delegate handlerDelegate
}

func newBotConversation(bot, channelID string, rtm slackRTM, delegate handlerDelegate, members ...string) (*botConversation, error) {
	var channel flamingo.Channel

	var users = make([]flamingo.User, len(members))
	for i, m := range members {
		users[i] = flamingo.User{ID: m}
	}

	// Channel IDs prefixed with C are channels,
	// prefixed with G are groups and prefixed with D are directs
	if strings.HasPrefix(channelID, "C") {
		ch, err := rtm.GetChannelInfo(channelID)
		if err != nil {
			return nil, err
		}

		channel = flamingo.Channel{
			ID:    ch.ID,
			Name:  ch.Name,
			Type:  flamingo.SlackClient,
			IsDM:  false,
			Extra: ch,
			Users: users,
		}
	} else {
		channel = flamingo.Channel{
			ID:    channelID,
			Type:  flamingo.SlackClient,
			IsDM:  strings.HasPrefix(channelID, "D"),
			Users: users,
		}
	}

	return &botConversation{
		rtm:      rtm,
		bot:      bot,
		channel:  channel,
		actions:  make(chan slack.AttachmentActionCallback, 1),
		messages: make(chan *slack.MessageEvent, 1),
		shutdown: make(chan struct{}, 1),
		closed:   make(chan struct{}, 1),
		delegate: delegate,
	}, nil
}

func (c *botConversation) run() {
	defer c.recoverAndRestart()

	for {
		select {
		case <-c.shutdown:
			c.closed <- struct{}{}
			return
		case msg, ok := <-c.messages:
			if !ok {
				continue
			}

			if c.isWorking() {
				go c.requeueMessage(msg)
				<-time.After(50 * time.Millisecond)
				continue
			}

			c.handleMessage(msg)

		case action, ok := <-c.actions:
			if !ok {
				continue
			}

			if c.isWorking() {
				go c.requeueAction(action)
				<-time.After(50 * time.Millisecond)
				continue
			}

			c.handleAction(action)
		case <-time.After(50 * time.Millisecond):
		}
	}
}

func (c *botConversation) requeueMessage(msg *slack.MessageEvent) {
	c.messages <- msg
}

func (c *botConversation) requeueAction(action *slack.AttachmentActionCallback) {
	c.actions <- action
}

func (c *botConversation) isWorking() bool {
	c.Lock()
	defer c.Unlock()
	return c.working
}

func (c *botConversation) handleMessage(msg *slack.MessageEvent) {
	message, err := c.convertMessage(msg)
	if err != nil {
		log15.Error("error converting message", "err", err.Error())
		continue
	}

	handler, ok := c.delegate.ControllerFor(message)
	if !ok {
		log15.Warn("no controller for message", "text", message.Text)
		continue
	}

	go func() {
		defer c.recoverWithLog("panic caught handling msg")

		c.setWorking(true)
		defer c.setWorking(false)
		if err := handler(c.createBot(), message); err != nil {
			log15.Error("error handling message", "error", err.Error())
		}
	}()
}

func (c *botConversation) handleAction(action *slack.AttachmentActionCallback) {
	handler, ok := c.delegate.ActionHandler(action.CallbackID)
	if !ok {
		log15.Warn("no handler for callback", "id", action.CallbackID)
		continue
	}

	act, err := convertAction(action, c.rtm)
	if err != nil {
		log15.Error("error converting action", "err", err.Error())
		continue
	}

	go func() {
		defer c.recoverWithLog("panic caught handling action")

		c.setWorking(true)
		defer c.setWorking(false)
		handler(c.createBot(), act)
	}()
}

func (c *botConversation) recoverWithLog(msg string) {
	if r := recover(); r != nil {
		if err, ok := r.(error); ok {
			log15.Error(msg, "err", err.Error())
		}

		if handler := c.delegate.ErrorHandler(); handler != nil {
			handler(r)
		}
	}
}

func (c *botConversation) recoverAndRestart() {
	func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				log15.Error("panic caught on bot conversation", "err", err.Error())
			}

			if handler := c.delegate.ErrorHandler(); handler != nil {
				handler(r)
			}

			log15.Info("restarting bot conversation")
			go c.run()
		}
	}()
}

func (c *botConversation) setWorking(working bool) {
	c.Lock()
	defer c.Unlock()
	c.working = working
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
		log15.Error("unable to find user", "id", userID)
		return flamingo.Message{}, err
	}

	return newMessage(convertUser(user), c.channel, src.Msg), nil
}

func (c *botConversation) handleIntro() {
	c.delegate.HandleIntro(c.createBot(), c.channel)
}

func (c *botConversation) handleJob(job flamingo.Job) {
	if err := job(c.createBot(), c.channel); err != nil {
		log15.Error("error running job", "bot", c.bot, "channel", c.channel.ID, "err", err.Error())
	}
}

func (c *botConversation) stop() {
	c.shutdown <- struct{}{}
	close(c.shutdown)
	<-c.closed
	close(c.closed)
	close(c.actions)
	close(c.messages)
}
