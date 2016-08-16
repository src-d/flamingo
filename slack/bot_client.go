package slack

import (
	"fmt"
	"sync"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/mvader/flamingo"
	"github.com/mvader/slack"
)

type slackRTM interface {
	IncomingEvents() chan slack.RTMEvent
	slackAPI
}

type handlerDelegate interface {
	ControllerFor(flamingo.Message) (flamingo.Controller, bool)
	ActionHandler(string) (flamingo.ActionHandler, bool)
	HandleIntro(flamingo.Bot, flamingo.Channel)
}

type botClient struct {
	id string
	sync.RWMutex
	rtm           slackRTM
	shutdown      chan struct{}
	closed        chan struct{}
	conversations map[string]*botConversation
	delegate      handlerDelegate
}

func newBotClient(id string, rtm slackRTM, delegate handlerDelegate) *botClient {
	cli := &botClient{
		id:            id,
		rtm:           rtm,
		shutdown:      make(chan struct{}, 1),
		closed:        make(chan struct{}, 1),
		conversations: make(map[string]*botConversation),
		delegate:      delegate,
	}
	go cli.runRTM()

	return cli
}

func (c *botClient) runRTM() {
	log15.Info("starting real time", "bot", c.id)
	for {
		select {
		case <-c.shutdown:
			for id, convo := range c.conversations {
				log15.Debug("shutting down conversation", "channel", id)
				convo.stop()
				log15.Debug("shut down conversation", "channel", id)
			}
			c.closed <- struct{}{}
			return
		case e := <-c.rtm.IncomingEvents():
			go c.handleRTMEvent(e)
		case <-time.After(50 * time.Millisecond):
		}
	}
}

func (c *botClient) handleAction(channel string, action slack.AttachmentActionCallback) {
	c.Lock()
	defer c.Unlock()

	conv, ok := c.conversations[channel]
	if !ok {
		log15.Warn("conversation not found in bot", "channel", channel, "id", c.id)
		return
	}

	conv.actions <- action
}

func (c *botClient) handleRTMEvent(e slack.RTMEvent) {
	c.Lock()
	defer c.Unlock()

	log15.Debug("received event of type", "type", e.Type)

	switch evt := e.Data.(type) {
	case *slack.MessageEvent:
		c.handleMessageEvent(evt)

	case *slack.LatencyReport:
		log15.Info("Current latency", "latency", evt.Value)

	case *slack.RTMError:
		log15.Error("Real Time Error", "error", evt.Error())

	case *slack.IMCreatedEvent:
		c.handleIMCreatedEvent(evt)

	case *slack.GroupJoinedEvent:
		c.handleGroupJoinedEvent(evt)

	case *slack.InvalidAuthEvent:
		log15.Crit("Invalid credentials for bot", "bot", c.id)
	}
}

func (c *botClient) handleMessageEvent(evt *slack.MessageEvent) {
	if evt.BotID == c.id {
		log15.Debug("got message from self, ignoring")
		return
	}

	conv, ok := c.conversations[evt.Channel]
	if !ok {
		log15.Info("this", "that", fmt.Sprintf("%#v", evt.Channel), "those", evt.SubType)
		if err := c.newConversation(evt.Channel); err != nil {
			log15.Error("unable to create conversation for bot", "channel", evt.Channel, "bot", c.id, "error", err.Error())
			return
		}
		conv = c.conversations[evt.Channel]
	}

	log15.Debug("message for channel", "channel", evt.Channel, "text", evt.Text)
	conv.messages <- evt
}

func (c *botClient) handleIMCreatedEvent(evt *slack.IMCreatedEvent) {
	if err := c.newConversation(evt.Channel.ID); err != nil {
		log15.Error("unable to create IM conversation for bot", "channel", evt.Channel.ID, "bot", c.id, "error", err.Error())
		return
	}

	conv := c.conversations[evt.Channel.ID]
	conv.handleIntro()
}

func (c *botClient) handleGroupJoinedEvent(evt *slack.GroupJoinedEvent) {
	if err := c.newConversation(evt.Channel.ID); err != nil {
		log15.Error("unable to create group conversation for bot", "channel", evt.Channel.ID, "bot", c.id, "error", err.Error())
		return
	}

	conv := c.conversations[evt.Channel.ID]
	conv.handleIntro()
}

func (c *botClient) newConversation(channel string) error {
	log15.Debug("conversation does not exist for bot, creating", "channel", channel, "bot", c.id)
	conv, err := newBotConversation(c.id, channel, c.rtm, c.delegate)
	if err != nil {
		return err
	}

	c.conversations[channel] = conv
	go conv.run()
	return nil
}

func (c *botClient) stop() {
	c.shutdown <- struct{}{}
	close(c.shutdown)
	<-c.closed
	close(c.closed)
}
