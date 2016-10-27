package slack

import (
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/dkumor/acmewrapper"
	"github.com/mvader/slack"
	"github.com/src-d/flamingo"
	"github.com/src-d/flamingo/storage"
)

// ClientOptions are the configurable options of the slack client.
type ClientOptions struct {
	// Debug will print extra debug log messages.
	Debug bool
	// Webhook contains the options for the slack webhook.
	Webhook WebhookOptions
}

// WebhookOptions are the configurable options of the slack webhook.
type WebhookOptions struct {
	// Enabled will start the webhook endpoint if true.
	Enabled bool
	// UseLetsEncrypt will setup for you a letsencrypt certificate if true.
	UseLetsEncrypt bool
	// Domain name of the webhook service, used in the letsencrypt certificate process.
	Domain string
	// VerificationToken is the token used to check incoming actions come from slack.
	VerificationToken string
	// Addr is the address on which the webhook will be run. Required
	// if EnableWebhook is true.
	Addr string
	// UseHTTPS will use HTTPS instead of HTTP to serve the webhook.
	// Note that using HTTPS is required by slack for receiving webhooks.
	// If UseLetsEncrypt is true, the value of this property will be ignored.
	UseHTTPS bool
	// CertFile is the path to the SSL certificate. If UseHTTPS is true it is required.
	CertFile string
	// KeyFile is the path to the SSL key. If UseHTTPS is true it is required.
	KeyFile string
	// RegistrationFile is the file to read/write registration info to. The ACME protocol requires remembering some details
	// about a registration. Therefore, the file is saved at the given location.
	// If not given, and PrivateKeyFile is given, then gives an error - if you're saving your private key,
	// you need to save your user registration.
	RegistrationFile string
	// PrivateKeyFile is the file to read/write the private key from/to. If this is not empty, and the file does not exist,
	// then the user is assumed not to be registered, and the file is created. if this is empty, then
	// a new private key is generated and used for all queries. The private key is lost on stopping the program.
	PrivateKeyFile string
}

type clientBot interface {
	handleAction(string, slack.AttachmentActionCallback)
	handleJob(flamingo.Job)
	addConversation(string) error
	stop()
}

type slackRTMWrapper struct {
	*slack.RTM
}

func (s *slackRTMWrapper) IncomingEvents() chan slack.RTMEvent {
	return s.RTM.IncomingEvents
}

func (s *slackRTMWrapper) GetUserByUsername(username string) (*slack.User, error) {
	users, err := s.RTM.GetUsers()
	if err != nil {
		return nil, err
	}

	for _, u := range users {
		if u.Name == username {
			return &u, nil
		}
	}

	return nil, errors.New("not_found")
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
	introHandler    flamingo.IntroHandler
	scheduledJobs   []*scheduledJob
	scheduledWg     *sync.WaitGroup
	storage         flamingo.Storage
	loadedBots      []clientBot
	errorHandler    flamingo.ErrorHandler
	middlewares     []flamingo.Middleware
}

type scheduledJob struct {
	mut      *sync.RWMutex
	job      flamingo.Job
	schedule flamingo.ScheduleTime
	stop     chan struct{}
}

// NewClient creates a new Slack Client with the given token and options.
func NewClient(token string, options ClientOptions) flamingo.Client {
	if options.Webhook.Addr == "" {
		options.Webhook.Addr = ":8080"
	}

	cli := &slackClient{
		options:         options,
		token:           token,
		webhook:         NewWebhookService(options.Webhook.VerificationToken),
		actionHandlers:  make(map[string]flamingo.ActionHandler),
		bots:            make(map[string]clientBot),
		shutdown:        make(chan struct{}, 1),
		shutdownWebhook: make(chan struct{}, 1),
		scheduledWg:     new(sync.WaitGroup),
		storage:         storage.NewMemory(),
	}

	cli.SetLogOutput(nil)
	return cli
}

func (c *slackClient) SetStorage(storage flamingo.Storage) {
	c.Lock()
	defer c.Unlock()
	c.storage = storage
}

func (c *slackClient) SetErrorHandler(handler flamingo.ErrorHandler) {
	c.Lock()
	defer c.Unlock()
	c.errorHandler = handler
}

func (c *slackClient) ErrorHandler() flamingo.ErrorHandler {
	c.Lock()
	defer c.Unlock()
	return c.errorHandler
}

func (c *slackClient) SetLogOutput(w io.Writer) {
	var nilWriter io.Writer

	handler := log15.StdoutHandler
	if w != nilWriter && w != nil {
		handler = log15.MultiHandler(
			log15.StdoutHandler,
			log15.StreamHandler(w, log15.LogfmtFormat()),
		)
	}

	var maxLvl = log15.LvlInfo
	if c.options.Debug {
		maxLvl = log15.LvlDebug
	}

	log15.Root().SetHandler(log15.LvlFilterHandler(maxLvl, handler))
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

func (c *slackClient) ControllerFor(msg flamingo.Message) (flamingo.HandlerFunc, bool) {
	c.Lock()
	defer c.Unlock()

	for _, ctrl := range c.controllers {
		if ctrl.CanHandle(msg) {
			return c.wrap(ctrl.Handle), true
		}
	}

	return nil, false
}

func (c *slackClient) wrap(handler flamingo.HandlerFunc) flamingo.HandlerFunc {
	if len(c.middlewares) == 0 {
		return handler
	}

	var (
		middlewares = c.middlewares[:]
		idx         int
		length      = len(middlewares)
		next        flamingo.HandlerFunc
	)

	next = func(bot flamingo.Bot, msg flamingo.Message) error {
		idx++
		if idx >= length {
			return handler(bot, msg)
		}

		return middlewares[idx](bot, msg, next)
	}

	return func(bot flamingo.Bot, msg flamingo.Message) error {
		return middlewares[0](bot, msg, next)
	}
}

func (c *slackClient) ActionHandler(id string) (flamingo.ActionHandler, bool) {
	c.Lock()
	defer c.Unlock()

	handler, ok := c.actionHandlers[id]
	return handler, ok
}

func (c *slackClient) SetIntroHandler(handler flamingo.IntroHandler) {
	c.Lock()
	defer c.Unlock()
	c.introHandler = handler
}

func (c *slackClient) Use(middlewares ...flamingo.Middleware) {
	c.Lock()
	defer c.Unlock()
	for _, m := range middlewares {
		if m != nil {
			c.middlewares = append(c.middlewares, m)
		}
	}
}

func (c *slackClient) AddBot(id, token string, extra interface{}) {
	c.Lock()
	defer c.Unlock()

	bot := flamingo.StoredBot{
		ID:        id,
		Token:     token,
		CreatedAt: time.Now(),
		Extra:     extra,
	}
	ok, err := c.storage.BotExists(bot)
	if err != nil {
		log15.Error("unable to check if bot exists", "id", id, "err", err.Error())
		return
	}

	if !ok {
		if err := c.storage.StoreBot(bot); err != nil {
			log15.Error("unable to add bot", "id", id, "err", err.Error())
			return
		}
	}

	if _, ok := c.bots[id]; ok {
		return
	}

	client := slack.New(token)
	client.SetDebug(false)
	rtm := client.NewRTM()
	go rtm.ManageConnection()
	c.bots[id] = newBotClient(id, &slackRTMWrapper{rtm}, c)
	c.loadedBots = append(c.loadedBots, c.bots[id])
}

func (c *slackClient) HandleIntro(bot flamingo.Bot, channel flamingo.Channel) {
	if c.introHandler != nil {
		if err := c.introHandler.HandleIntro(bot, channel); err != nil {
			log15.Error("error handling intro", "channel", channel.ID, "err", err.Error())
		}
	} else {
		log15.Warn("there is no intro handler, ignoring")
	}
}

func (c *slackClient) AddScheduledJob(schedule flamingo.ScheduleTime, job flamingo.Job) {
	c.Lock()
	defer c.Unlock()
	c.scheduledJobs = append(c.scheduledJobs, &scheduledJob{
		mut:      new(sync.RWMutex),
		job:      job,
		schedule: schedule,
	})
}

func (c *slackClient) Storage() flamingo.Storage {
	return c.storage
}

func (c *slackClient) Stop() error {
	for id, bot := range c.bots {
		log15.Debug("shutting down bot", "id", id)
		bot.stop()
		log15.Debug("shut down bot", "id", id)
	}

	for _, j := range c.scheduledJobs {
		j.mut.RLock()
		if j.stop != nil {
			j.stop <- struct{}{}
		}
		j.mut.RUnlock()
	}

	c.RLock()
	c.scheduledWg.Wait()
	c.RUnlock()
	c.shutdown <- struct{}{}
	c.shutdownWebhook <- struct{}{}
	return nil
}

func (c *slackClient) runWebhook() error {
	if c.options.Webhook.VerificationToken == "" {
		return errors.New("webhook verification token is empty")
	}

	if (c.options.Webhook.UseHTTPS || c.options.Webhook.UseLetsEncrypt) &&
		(c.options.Webhook.CertFile == "" || c.options.Webhook.KeyFile == "") {
		return errors.New("cert and key files need to be provided if HTTPS or letsencrypt are enabled")
	}

	var (
		tlsconfig *tls.Config
		listener  net.Listener
	)

	if c.options.Webhook.UseHTTPS || c.options.Webhook.UseLetsEncrypt {
		w, err := acmewrapper.New(acmewrapper.Config{
			Domains:          []string{c.options.Webhook.Domain},
			Address:          c.options.Webhook.Addr,
			TLSCertFile:      c.options.Webhook.CertFile,
			TLSKeyFile:       c.options.Webhook.KeyFile,
			RegistrationFile: c.options.Webhook.RegistrationFile,
			PrivateKeyFile:   c.options.Webhook.PrivateKeyFile,
			AcmeDisabled:     !c.options.Webhook.UseLetsEncrypt,
			TOSCallback:      acmewrapper.TOSAgree,
		})
		if err != nil {
			return err
		}

		tlsconfig = w.TLSConfig()
		listener, err = tls.Listen("tcp", c.options.Webhook.Addr, tlsconfig)
		if err != nil {
			return err
		}
	} else {
		var err error
		listener, err = net.Listen("tcp", c.options.Webhook.Addr)
		if err != nil {
			return err
		}
	}

	go func() {
		<-c.shutdownWebhook
		listener.Close()
	}()

	return (&http.Server{
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 3 * time.Second,
		Addr:         c.options.Webhook.Addr,
		Handler:      c.webhook,
		TLSConfig:    tlsconfig,
	}).Serve(listener)
}

func (c *slackClient) runScheduledJobs() {
	c.Lock()
	defer c.Unlock()
	for i, j := range c.scheduledJobs {
		c.scheduledJobs[i].mut.Lock()
		c.scheduledJobs[i].stop = make(chan struct{}, 1)
		c.scheduledJobs[i].mut.Unlock()
		c.scheduledWg.Add(1)
		go c.runScheduledJob(*j)
	}
}

func (c *slackClient) runScheduledJob(j scheduledJob) {
	now := time.Now()
	interval := j.schedule.Next(now).Sub(now)
	for {
		select {
		case <-time.After(interval):
			wg := new(sync.WaitGroup)

			for _, b := range c.bots {
				wg.Add(1)
				go func(b clientBot) {
					defer func() {
						if r := recover(); r != nil {
							if err, ok := r.(error); ok {
								log15.Error("panic caught running scheduled job", "err", err.Error())
							}

							if handler := c.ErrorHandler(); handler != nil {
								handler(r)
							}
						}
					}()

					b.handleJob(j.job)
					wg.Done()
				}(b)
			}

			wg.Wait()
			now := time.Now()
			interval = j.schedule.Next(now).Sub(now)

		case <-j.stop:
			j.mut.Lock()
			defer j.mut.Unlock()
			close(j.stop)
			c.scheduledWg.Done()
			return
		}
	}
}

func (c *slackClient) loadFromStorage() error {
	log15.Info("Loading data from storage...")
	defer log15.Info("Loaded data from storage...")

	bots, err := c.storage.LoadBots()
	if err != nil {
		return err
	}

	for _, b := range bots {
		if _, ok := c.bots[b.ID]; !ok {
			c.AddBot(b.ID, b.Token, nil)
		}

		convs, err := c.storage.LoadConversations(b)
		if err != nil {
			return err
		}

		for _, conv := range convs {
			if err := c.bots[b.ID].addConversation(conv.ID); err != nil {
				log15.Error("error starting conversation", "conversation", conv.ID, "bot", b.ID)
			}
		}
	}

	return nil
}

func (c *slackClient) Broadcast(msg flamingo.Sendable, filter flamingo.BroadcastFilter) (
	bots uint64,
	conversations uint64,
	errors uint64,
	err error,
) {
	for _, b := range c.bots {
		convs, errs := broadCastTo(b.(*botClient), msg, filter)
		errors += errs
		conversations += convs
		if convs > 0 {
			bots++
		}
	}

	if errors == conversations {
		err = flamingo.ErrAllMessagesLost
	} else if errors > 0 {
		err = flamingo.ErrSomeMessagesLost
	}

	return
}

func broadCastTo(bot *botClient, msg flamingo.Sendable, filter flamingo.BroadcastFilter) (uint64, uint64) {
	var conversations, errors uint64
	for _, c := range bot.conversations {
		if filter(bot.id, c.channel) {
			if err := send(c.createBot(), msg); err != nil {
				errors++
			}
			conversations++
		}
	}

	return conversations, errors
}

func send(bot flamingo.Bot, msg flamingo.Sendable) (err error) {
	switch msg.(type) {
	case flamingo.OutgoingMessage:
		_, err = bot.Say(msg.(flamingo.OutgoingMessage))
	case flamingo.Form:
		_, err = bot.Form(msg.(flamingo.Form))
	case flamingo.Image:
		_, err = bot.Image(msg.(flamingo.Image))
	}
	return
}

func (c *slackClient) Run() error {
	log15.Info("Starting flamingo slack client")
	if c.storage != nil {
		if err := c.loadFromStorage(); err != nil {
			return err
		}
	}

	if c.options.Webhook.Enabled {
		log15.Info("Starting webhook server endpoint", "address", c.options.Webhook.Addr)
		go func() {
			if err := c.runWebhook(); err != nil {
				log15.Crit("error running webhook, stopping", "err", err.Error())

				if err := c.Stop(); err != nil {
					log15.Crit("error stopping client", "err", err.Error())
				}
			}
		}()
	}

	if len(c.scheduledJobs) > 0 {
		c.runScheduledJobs()
	}

	actions := c.webhook.Consume()
	for {
		select {
		case action := <-actions:
			log15.Debug("action received", "callback", action.CallbackID)
			c.handleActionCallback(action)

		case <-c.shutdown:
			return nil

		case <-time.After(50 * time.Millisecond):
		}
	}
}

func (c *slackClient) handleActionCallback(action slack.AttachmentActionCallback) {
	parts := strings.Split(action.CallbackID, "::")
	if len(parts) < 3 {
		log15.Error("invalid action", "callback", action.CallbackID)
		return
	}

	bot, channel, id := parts[0], parts[1], parts[2]
	c.RLock()
	b, ok := c.bots[bot]
	c.RUnlock()
	if !ok {
		log15.Warn("bot not found", "id", bot)
		return
	}

	action.CallbackID = id
	b.handleAction(channel, action)
}
