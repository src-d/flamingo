package slack

import (
	"log"
	"sync"
	"time"

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

func newBotClient(id string, client *slack.Client, delegate handlerDelegate) *botClient {
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
		log.Printf("conversation with id %q not found in bot %q", channel, c.id)
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
			log.Printf("conversation %q does not exist for bot %q, creating", evt.Channel, c.id)
			conv, err := newBotConversation(c.id, evt.Channel, c.rtm, c.delegate)
			if err != nil {
				log.Printf("unable to create conversation %q for bot %q: %q", evt.Channel, c.id, err.Error())
				return
			}

			c.conversations[evt.Channel] = conv
			go conv.run()
		}

		log.Printf("message for channel %q: %s", evt.Channel, evt.Text)
		conv.messages <- evt
	case *slack.LatencyReport:
		log.Printf("Current latency: %v", evt.Value)

	case *slack.RTMError:
		log.Printf("Real Time Error: %q", evt.Error())

	case *slack.InvalidAuthEvent:
		log.Fatalf("Invalid credentials for bot %q", c.id)
	}
}
