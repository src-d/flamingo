package flamingo

import "time"

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

type timeSchedule struct {
	hour, minutes, seconds int
}

// NewTimeSchedule creates a ScheduleTime that runs once a day at a given
// hour, minutes and seconds.
func NewTimeSchedule(hour, minutes, seconds int) ScheduleTime {
	return &timeSchedule{
		hour, minutes, seconds,
	}
}

func (s *timeSchedule) Next(now time.Time) time.Time {
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

type dayTimeSchedule struct {
	days         map[time.Weekday]struct{}
	timeSchedule ScheduleTime
}

// NewDayTimeSchedule creates a ScheduleTime that runs once a day at a given
// hour, minutes and seconds only on the given set of days.
func NewDayTimeSchedule(days []time.Weekday, hour, minutes, seconds int) ScheduleTime {
	var daySet = make(map[time.Weekday]struct{})
	for _, d := range days {
		daySet[d] = struct{}{}
	}

	return &dayTimeSchedule{
		daySet,
		NewTimeSchedule(hour, minutes, seconds),
	}
}

var zero time.Time

func (s *dayTimeSchedule) Next(now time.Time) time.Time {
	if len(s.days) == 0 {
		return zero
	}

	for {
		if _, ok := s.days[now.Weekday()]; ok {
			return s.timeSchedule.Next(now)
		}

		now = now.Add(24 * time.Hour)
	}
}
