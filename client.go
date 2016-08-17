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

	// AddController adds a new Controller to the Client.
	AddController(Controller)

	// AddActionHandler adds an ActionHandler for the given ID.
	AddActionHandler(string, ActionHandler)

	// AddBot adds a new bot with an ID and a token.
	AddBot(id string, token string)

	// SetIntroHandler sets the IntroHandler for the client.
	SetIntroHandler(IntroHandler)

	// AddScheduledJob will run the given Job forever after the given
	// duration from the last execution.
	AddScheduledJob(ScheduleTime, Job)

	// Run starts the client.
	Run() error

	// Stop stops the client.
	Stop() error
}

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
