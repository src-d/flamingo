package slack

import (
	"sync"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/mvader/flamingo"
	"github.com/nlopes/slack"
)

type slackRTM interface {
	IncomingEvents() chan slack.RTMEvent
	slackAPI
}

type handlerDelegate interface {
	ControllerFor(flamingo.Message) (flamingo.Controller, bool)
	ActionHandler(string) (flamingo.ActionHandler, bool)
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
		if evt.BotID == c.id {
			log15.Debug("got message from self, ignoring")
			return
		}

		conv, ok := c.conversations[evt.Channel]
		if !ok {
			log15.Debug("conversation does not exist for bot, creating", "channel", evt.Channel, "bot", c.id)
			var err error
			conv, err = newBotConversation(c.id, evt.Channel, c.rtm, c.delegate)
			if err != nil {
				log15.Error("unable to create conversation for bot", "channel", evt.Channel, "bot", c.id, "error", err.Error())
				return
			}

			c.conversations[evt.Channel] = conv
			go conv.run()
		}

		log15.Debug("message for channel", "channel", evt.Channel, "text", evt.Text)
		conv.messages <- evt
	case *slack.LatencyReport:
		log15.Info("Current latency", "latency", evt.Value)

	case *slack.RTMError:
		log15.Error("Real Time Error", "error", evt.Error())

	case *slack.InvalidAuthEvent:
		log15.Crit("Invalid credentials for bot", "bot", c.id)
	}
}

func (c *botClient) stop() {
	c.shutdown <- struct{}{}
	close(c.shutdown)
	<-c.closed
	close(c.closed)
}
