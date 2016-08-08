package slack

import (
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mvader/flamingo"
	"github.com/nlopes/slack"
)

type ClientOptions struct {
	EnableWebhook bool
	WebhookAddr   string
}

type slackClient struct {
	sync.RWMutex
	webhook        *WebhookService
	options        ClientOptions
	token          string
	controllers    []flamingo.Controller
	actionHandlers map[string]flamingo.ActionHandler
	bots           map[string]*botClient
}

func NewClient(token string, options ClientOptions) flamingo.Client {
	return &slackClient{
		options:        options,
		token:          token,
		webhook:        NewWebhookService(token),
		actionHandlers: make(map[string]flamingo.ActionHandler),
		bots:           make(map[string]*botClient),
	}
}

func (c *slackClient) SetLogOutput(w io.Writer) {
}

func (c *slackClient) AddController(ctrl flamingo.Controller) {
	c.controllers = append(c.controllers, ctrl)
}

func (c *slackClient) AddActionHandler(id string, handler flamingo.ActionHandler) {
	c.actionHandlers[id] = handler
}

func (c *slackClient) ControllerFor(msg flamingo.Message) (flamingo.Controller, bool) {
	c.Lock()
	defer c.Unlock()

	for _, ctrl := range c.controllers {
		if ctrl.CanHandle(msg) {
			return ctrl, true
		}
	}

	return nil, false
}

func (c *slackClient) ActionHandler(id string) (flamingo.ActionHandler, bool) {
	c.Lock()
	defer c.Unlock()

	handler, ok := c.actionHandlers[id]
	return handler, ok
}

func (c *slackClient) runWebhook() {
	srv := http.Server{
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 3 * time.Second,
		Addr:         c.options.WebhookAddr,
		Handler:      c.webhook,
	}

	log.Fatal(srv.ListenAndServe())
}

func (c *slackClient) Run() error {
	if c.options.EnableWebhook {
		go c.runWebhook()
	}

	actions := c.webhook.Consume()
	for {
		select {
		case action := <-actions:
			go c.handleActionCallback(action)

		case <-time.After(200 * time.Millisecond):
		}
	}
}

func (c *slackClient) handleActionCallback(action slack.AttachmentActionCallback) {
	c.Lock()
	defer c.Unlock()

	parts := strings.Split(action.CallbackID, "::")
	if len(parts) < 3 {
		log.Printf("invalid action with callback %q", action.CallbackID)
		return
	}

	bot, channel, id := parts[0], parts[1], parts[2]
	b, ok := c.bots[bot]
	if !ok {
		log.Printf("bot with id %q not found", bot)
		return
	}

	action.CallbackID = id
	b.handleAction(channel, action)
}
