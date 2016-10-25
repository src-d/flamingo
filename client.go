package flamingo

import (
	"io"
	"time"
)

// Client is an abstract interface of a platforms-specific client.
// A client can only run for one platform. If you need to handle
// more than one platform you will have to start several clients
// for different platforms.
type Client interface {
	// SetLogOutput will write the logs to the given io.Writer.
	SetLogOutput(io.Writer)

	// Use adds one or more middlewares. All middlewares will be executed in the
	// order they were added. Middlewares are only executed for controllers and
	// are executed for all of them.
	Use(...Middleware)

	// AddController adds a new Controller to the Client.
	AddController(Controller)

	// AddActionHandler adds an ActionHandler for the given ID.
	AddActionHandler(string, ActionHandler)

	// AddBot adds a new bot with an ID and a token.
	AddBot(id string, token string, extra interface{})

	// SetIntroHandler sets the IntroHandler for the client.
	SetIntroHandler(IntroHandler)

	// SetErrorHandler sets the error handler of the client. The error handler
	// will receive the result of recover() after a panic has been caught.
	// All running instances of bots are restarted after a panic.
	SetErrorHandler(ErrorHandler)

	// SetStorage sets the storage to be used to store conversations and bots.
	// In this package clients, if the Storage is added *before* calling the
	// Run method bots and conversations will be loaded from there.
	SetStorage(Storage)

	// AddScheduledJob will run the given Job forever after the given
	// duration from the last execution.
	AddScheduledJob(ScheduleTime, Job)

	// Run starts the client.
	Run() error

	// Stop stops the client.
	Stop() error
}

// Broadcaster defines a client that can send messages to the registered conversations
// following certain condition
type Broadcaster interface {
	// Broadcast sends message, and returns the number of
	// processed bots, conversations, errors and error occurred
	Broadcast(interface{}, Condition) (uint64, uint64, uint64, error)
}

// ErrorHandler will handle an error after a panic. The parameter it receives is the
// result of recover()
type ErrorHandler func(interface{})

// HandlerFunc is a function that receives a bot and a message and does something with them.
type HandlerFunc func(Bot, Message) error

// Middleware is a function that receives a bot and a message and the next handler to be called after it.
type Middleware func(Bot, Message, HandlerFunc) error

// ClientType tells us what type of client is.
type ClientType uint

const (
	// SlackClient is a client for Slack.
	SlackClient ClientType = 1 << iota
)

// Job is a function that will execute like a cron job after a
// certain amount of time to perform some kind of task.
type Job func(Bot, Channel) error

// ScheduleTime gives the next time to run a job.
type ScheduleTime interface {
	// Next time right after the given time a job should be run.
	Next(time.Time) time.Time
}

type intervalSchedule struct {
	every time.Duration
}

// NewIntervalSchedule creates a ScheduleTime that runs in intervals
// of the given duration.
func NewIntervalSchedule(every time.Duration) ScheduleTime {
	return &intervalSchedule{
		every,
	}
}

func (s *intervalSchedule) Next(now time.Time) time.Time {
	return now.Add(s.every)
}

type dateSchedule struct {
	hour, minutes, seconds int
}

// NewDateSchedule creates a ScheduleTime that runs once a day at a given
// hour, minutes and seconds.
func NewDateSchedule(hour, minutes, seconds int) ScheduleTime {
	return &dateSchedule{
		hour, minutes, seconds,
	}
}

func (s *dateSchedule) Next(now time.Time) time.Time {
	d := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		s.hour,
		s.minutes,
		s.seconds,
		0,
		now.Location(),
	)

	if d.Before(now) {
		d = d.Add(24 * time.Hour)
	}

	return d
}

//Condition modelates a group of conditions
type Condition struct {
	IsValidChannel ValidChannel
	IsValidBot     ValidBot
}

//ValidChannel returns true if the passed Channel and msg are valid
type ValidChannel func(channel Channel, msg interface{}) bool

//ValidChannel returns true if the passed botId and msg are valid
type ValidBot func(botId string, msg interface{}) bool
