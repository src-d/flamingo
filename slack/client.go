package slack

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/mvader/flamingo"
	"github.com/nlopes/slack"
)

type ClientOptions struct {
	EnableWebhook bool
	WebhookAddr   string
	Debug         bool
}

type clientBot interface {
	handleAction(string, slack.AttachmentActionCallback)
	stop()
}

type slackRTMWrapper struct {
	*slack.RTM
}

func (s *slackRTMWrapper) IncomingEvents() chan slack.RTMEvent {
	return s.RTM.IncomingEvents
}

type slackClient struct {
	sync.RWMutex
	webhook         *WebhookService
	options         ClientOptions
	token           string
	controllers     []flamingo.Controller
	actionHandlers  map[string]flamingo.ActionHandler
	bots            map[string]clientBot
	shutdown        chan struct{}
	shutdownWebhook chan struct{}
}

func NewClient(token string, options ClientOptions) flamingo.Client {
	cli := &slackClient{
		options:         options,
		token:           token,
		webhook:         NewWebhookService(token),
		actionHandlers:  make(map[string]flamingo.ActionHandler),
		bots:            make(map[string]clientBot),
		shutdown:        make(chan struct{}, 1),
		shutdownWebhook: make(chan struct{}, 1),
	}

	cli.SetLogOutput(nil)
	return cli
}

func (c *slackClient) SetLogOutput(w io.Writer) {
	var nilWriter io.Writer

	var format = log15.LogfmtFormat()
	if w == nilWriter || w == nil {
		w = os.Stdout
		format = log15.TerminalFormat()
	}

	var maxLvl = log15.LvlInfo
	if c.options.Debug {
		maxLvl = log15.LvlDebug
	}

	log15.Root().SetHandler(log15.LvlFilterHandler(maxLvl, log15.StreamHandler(w, format)))
}

func (c *slackClient) AddController(ctrl flamingo.Controller) {
	c.Lock()
	defer c.Unlock()
	c.controllers = append(c.controllers, ctrl)
}

func (c *slackClient) AddActionHandler(id string, handler flamingo.ActionHandler) {
	c.Lock()
	defer c.Unlock()
	log15.Debug("added action handler", "id", id)
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

func (c *slackClient) AddBot(id, token string) {
	c.Lock()
	defer c.Unlock()

	client := slack.New(token)
	client.SetDebug(false)
	c.bots[id] = newBotClient(id, &slackRTMWrapper{client.NewRTM()}, c)
}

func (c *slackClient) Stop() error {
	for id, bot := range c.bots {
		log15.Debug("shutting down bot", "id", id)
		bot.stop()
		log15.Debug("shut down bot", "id", id)
	}

	c.shutdown <- struct{}{}
	c.shutdownWebhook <- struct{}{}
	return nil
}

func (c *slackClient) runWebhook() {
	srv := http.Server{
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 3 * time.Second,
		Addr:         c.options.WebhookAddr,
		Handler:      c.webhook,
	}

	l, err := net.Listen("tcp", c.options.WebhookAddr)
	if err != nil {
		log15.Crit("error creating listener", "err", err)
	}
	defer l.Close()

	go func() {
		<-c.shutdownWebhook
		l.Close()
	}()

	if err := srv.Serve(l); err != nil {
		log15.Crit("error running webhook", "err", err.Error())
	}
}

func (c *slackClient) Run() error {
	log15.Info("Starting flamingo slack client")
	if c.options.EnableWebhook {
		log15.Info("Starting webhook server endpoint")
		go c.runWebhook()
	}

	actions := c.webhook.Consume()
	for {
		select {
		case action := <-actions:
			log15.Debug("action received", "callback", action.CallbackID)
			go c.handleActionCallback(action)

		case <-c.shutdown:
			return nil

		case <-time.After(50 * time.Millisecond):
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
