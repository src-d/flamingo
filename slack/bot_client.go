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

type slackRTMWrapper struct {
	*slack.RTM
}

func (s *slackRTMWrapper) IncomingEvents() chan slack.RTMEvent {
	return s.RTM.IncomingEvents
}

type handlerDelegate interface {
	ControllerFor(flamingo.Message) (flamingo.Controller, bool)
	ActionHandler(string) (flamingo.ActionHandler, bool)
}

type botClient struct {
	id string
	sync.RWMutex
	rtm           slackRTM
	conversations map[string]*botConversation
	delegate      handlerDelegate
}

func newBotClient(id, token string, delegate handlerDelegate) *botClient {
	client := slack.New(token)
	client.SetDebug(false)

	cli := &botClient{
		id:            id,
		rtm:           &slackRTMWrapper{client.NewRTM()},
		conversations: make(map[string]*botConversation),
		delegate:      delegate,
	}
	go cli.runRTM()

	return cli
}

func (c *botClient) runRTM() {
	for {
		select {
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

	switch evt := e.Data.(type) {
	case *slack.MessageEvent:
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
