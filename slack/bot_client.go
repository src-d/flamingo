package slack

import (
	"sync"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/mvader/slack"
	"github.com/src-d/flamingo"
)

type slackRTM interface {
	IncomingEvents() chan slack.RTMEvent
	slackAPI
}

type handlerDelegate interface {
	ControllerFor(flamingo.Message) (flamingo.Controller, bool)
	ActionHandler(string) (flamingo.ActionHandler, bool)
	HandleIntro(flamingo.Bot, flamingo.Channel)
	Storage() flamingo.Storage
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
	c.RLock()
	conv, ok := c.conversations[channel]
	c.RUnlock()
	if !ok {
		var err error
		conv, err = c.newConversation(channel)
		if err != nil {
			log15.Error("unable to create conversation for bot", "channel", channel, "bot", c.id, "error", err.Error())
			return
		}
	}

	conv.actions <- action
}

func (c *botClient) handleJob(job flamingo.Job) {
	c.Lock()
	var wg sync.WaitGroup
	for _, conv := range c.conversations {
		wg.Add(1)
		go func(conv *botConversation) {
			conv.handleJob(job)
			wg.Done()
		}(conv)
	}
	c.Unlock()
	wg.Wait()
}

func (c *botClient) handleRTMEvent(e slack.RTMEvent) {
	log15.Debug("received event of type", "type", e.Type)

	switch evt := e.Data.(type) {
	case *slack.MessageEvent:
		// For now, ignore all messages that are not new messages
		switch evt.SubType {
		case "", "me_message", "bot_message":
			c.handleMessageEvent(evt)
		}

	case *slack.LatencyReport:
		log15.Info("Current latency", "latency", evt.Value)

	case *slack.RTMError:
		log15.Error("Real Time Error", "error", evt.Error())

	case *slack.IMCreatedEvent:
		c.handleNewConversation(evt.Channel.ID)

	case *slack.GroupJoinedEvent:
		c.handleNewConversation(evt.Channel.ID)

	case *slack.ChannelJoinedEvent:
		c.handleNewConversation(evt.Channel.ID)

	case *slack.InvalidAuthEvent:
		log15.Crit("Invalid credentials for bot", "bot", c.id)
	}
}

func (c *botClient) handleMessageEvent(evt *slack.MessageEvent) {
	if evt.BotID == c.id || evt.User == c.id {
		log15.Debug("got message from self, ignoring")
		return
	}

	c.RLock()
	conv, ok := c.conversations[evt.Channel]
	c.RUnlock()
	if !ok {
		var err error
		conv, err = c.newConversation(evt.Channel)
		if err != nil {
			log15.Error("unable to create conversation for bot", "channel", evt.Channel, "bot", c.id, "error", err.Error())
			return
		}
	}

	log15.Debug("message for channel", "channel", evt.Channel, "text", evt.Text)
	conv.messages <- evt
}

func (c *botClient) handleNewConversation(channelID string) {
	conv, err := c.newConversation(channelID)
	if err != nil {
		log15.Error("unable to create conversation for bot", "channel", channelID, "bot", c.id, "error", err.Error())
		return
	}

	conv.handleIntro()
}

func (c *botClient) newConversation(channel string) (*botConversation, error) {
	c.Lock()
	defer c.Unlock()
	log15.Debug("conversation does not exist for bot, creating", "channel", channel, "bot", c.id)
	conv, err := newBotConversation(c.id, channel, c.rtm, c.delegate)
	if err != nil {
		return nil, err
	}

	storage := c.delegate.Storage()
	conversation := flamingo.StoredConversation{
		ID:        channel,
		BotID:     c.id,
		CreatedAt: time.Now(),
	}
	ok, err := storage.ConversationExists(conversation)
	if err != nil {
		return nil, err
	}

	if !ok {
		if err := storage.StoreConversation(conversation); err != nil {
			return nil, err
		}
	}

	c.conversations[channel] = conv
	go conv.run()
	return conv, nil
}

func (c *botClient) addConversation(id string) error {
	_, err := c.newConversation(id)
	return err
}

func (c *botClient) stop() {
	c.shutdown <- struct{}{}
	close(c.shutdown)
	<-c.closed
	close(c.closed)
}
