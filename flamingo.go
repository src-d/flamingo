package flamingo

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/nlopes/slack"
)

// Options will be used to setup the Engine.
type Options struct {
	// Debug will print debug logs if enabled.
	Debug bool
	// DebugOutput will print the logs of the internal API.
	DebugOutput io.Writer
	// StdOutput will print the human-readable log and handler logs.
	StdOutput io.Writer
}

// Engine will run forever fetching events from Slack on realtime
// and performing actions depending on these events.
type Engine interface {
	// AddHandler adds a new Handler.
	AddHandler(Handler)
	// Run starts the Engine.
	Run()
}

type engine struct {
	name     string
	handlers []Handler
	rtm      *slack.RTM
	logger   log15.Logger
}

// New creates a new Engine.
func New(name, token string, opts Options) Engine {
	var nilWriter io.Writer
	api := slack.New(token)

	if opts.DebugOutput == nil || opts.DebugOutput == nilWriter {
		opts.DebugOutput = os.Stdout
	}
	slack.SetLogger(log.New(opts.DebugOutput, name, log.Lshortfile|log.LstdFlags))
	api.SetDebug(opts.Debug)

	var format = log15.LogfmtFormat()
	if opts.StdOutput == nil || opts.StdOutput == nilWriter {
		opts.StdOutput = os.Stdout
		format = log15.TerminalFormat()
	}

	var maxLvl = log15.LvlInfo
	if opts.Debug {
		maxLvl = log15.LvlDebug
	}

	logger := log15.New(name, "msg")
	logger.SetHandler(log15.LvlFilterHandler(maxLvl, log15.StreamHandler(opts.StdOutput, format)))
	rtm := api.NewRTM()

	return &engine{
		name:   name,
		rtm:    rtm,
		logger: logger,
	}
}

func (e *engine) AddHandler(handler Handler) {
	e.handlers = append(e.handlers, handler)
}

func (e *engine) Run() {
	go e.rtm.ManageConnection()
	var shutdown = make(chan struct{}, 1)
	for {
		select {
		case <-shutdown:
			e.logger.Info("Shutting down engine...")
			if err := e.rtm.Disconnect(); err != nil {
				e.logger.Error("error disconnecting RTM", "err", err)
			}
			return
		case evt := <-e.rtm.IncomingEvents:
			e.handleEvent(evt, shutdown)

		case <-time.After(1 * time.Second):
		}
	}
}

func (e *engine) guard() {
	if r := recover(); r != nil {
		switch t := r.(type) {
		case error:
			e.logger.Error("recovered from error", "err", t)
		default:
			e.logger.Error("unexpected error occurred", "recovery", t)
		}
	}
}

func (e *engine) handleEvent(evt slack.RTMEvent, shutdown chan struct{}) {
	defer e.guard()

	e.logger.Debug("Received event", "type", evt.Type)
	switch ev := evt.Data.(type) {
	case *slack.MessageEvent:
		e.handleMessageEvent(ev)
	case *slack.AckErrorEvent:
		e.logger.Error("API returned error", "err", ev.Error())

	case *slack.RTMError:
		e.logger.Error("Error occurred", "err", ev.Error())

	case *slack.InvalidAuthEvent:
		e.logger.Crit("Unable to access, invalid credentials")
		shutdown <- struct{}{}

	default:
		e.logger.Debug("Event ignored", "type", evt.Type)
	}
}

func (e *engine) handleMessageEvent(evt *slack.MessageEvent) {
	if e.isHelp(evt.Text) {
		e.logger.Debug("printing help")
		e.SendMsg(evt.Channel, e.help())
		return
	}

	user, err := e.GetUser(evt.User)
	if err != nil {
		e.logger.Error("error retrieving user info", "err", err, "user", evt.User)
		return
	}

	channel := e.GetChannel(evt.Channel)

	log15.Debug(
		"Message received",
		"user", user.Name,
		"channel", channel.Name(),
		"text", evt.Text,
	)

	for _, h := range e.handlers {
		if h.IsMatch(channel.Name(), evt.Text) {
			h.Handle(Message{
				Channel:  channel,
				Text:     evt.Text,
				User:     user,
				Delegate: e,
			})
		}
	}

	e.logger.Debug(
		"Message ignored, no handler could handle it",
		"channel", channel.Name(),
		"user", user.Name,
		"text", evt.Text,
	)
}

func (e *engine) GetUser(ID string) (*slack.User, error) {
	return e.rtm.GetUserInfo(ID)
}

func (e *engine) GetChannel(ID string) Channel {
	channel, _ := e.rtm.GetChannelInfo(ID)
	return Channel{Channel: channel, ID: ID}
}

func (e *engine) Debug(msg string, args ...interface{}) {
	e.logger.Debug(msg, args...)
}

func (e *engine) Info(msg string, args ...interface{}) {
	e.logger.Info(msg, args...)
}

func (e *engine) Warn(msg string, args ...interface{}) {
	e.logger.Warn(msg, args...)
}

func (e *engine) Error(msg string, args ...interface{}) {
	e.logger.Error(msg, args...)
}

func (e *engine) Crit(msg string, args ...interface{}) {
	e.logger.Crit(msg, args...)
}

func (e *engine) SendCustomMsg(channel, text string, params slack.PostMessageParameters) error {
	_, _, err := e.rtm.PostMessage(channel, text, params)
	return err
}

func (e *engine) SendMsg(channel, text string) error {
	return e.SendCustomMsg(channel, text, slack.PostMessageParameters{
		LinkNames: 1,
		AsUser:    true,
		Markdown:  true,
	})
}

func (e *engine) isHelp(text string) bool {
	helpText := fmt.Sprintf("%s help", e.name)
	return normalizeString(helpText) == normalizeString(text)
}

func (e *engine) help() string {
	var buf bytes.Buffer

	buf.WriteString("*Help*\n\n")
	for _, h := range e.handlers {
		buf.WriteString(fmt.Sprintf("*%s:* %s\n\n", h.Name(), h.Help()))
	}

	return buf.String()
}

func normalizeString(str string) string {
	return strings.ToLower(strings.TrimSpace(str))
}
