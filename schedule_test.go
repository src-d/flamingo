package flamingo

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIntervalSchedule(t *testing.T) {
	s := NewIntervalSchedule(4 * time.Second)
	now := time.Now()

	require.Equal(t, 4*time.Second, s.Next(now).Sub(now))
}

func TestTimeSchedule(t *testing.T) {
	s := NewTimeSchedule(15, 30, 25)
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	d := time.Date(now.Year(), now.Month(), now.Day(), 15, 30, 25, 0, now.Location())

	require.Equal(t, d, s.Next(now))

	now = time.Date(now.Year(), now.Month(), now.Day(), 16, 0, 0, 0, now.Location())
	d = time.Date(now.Year(), now.Month(), now.Day()+1, 15, 30, 25, 0, now.Location())

	require.Equal(t, d, s.Next(now))
}

func TestDayTimeSchedule(t *testing.T) {
	s := NewDayTimeSchedule([]time.Weekday{
		time.Wednesday,
		time.Thursday,
		time.Friday,
	}, 15, 0, 0)

	cases := []struct {
		now     time.Time
		nextDay time.Weekday
	}{
		{
			time.Date(2016, time.October, 3, 0, 0, 0, 0, time.Local), // monday
			time.Wednesday,
		},
		{
			time.Date(2016, time.October, 4, 0, 0, 0, 0, time.Local),
			time.Wednesday,
		},
		{
			time.Date(2016, time.October, 5, 0, 0, 0, 0, time.Local),
			time.Wednesday,
		},
		{
			time.Date(2016, time.October, 6, 0, 0, 0, 0, time.Local),
			time.Thursday,
		},
		{
			time.Date(2016, time.October, 7, 0, 0, 0, 0, time.Local),
			time.Friday,
		},
		{
			time.Date(2016, time.October, 8, 0, 0, 0, 0, time.Local),
			time.Wednesday,
		},
		{
			time.Date(2016, time.October, 9, 0, 0, 0, 0, time.Local),
			time.Wednesday,
		},
	}

	for _, c := range cases {
		nextTime := s.Next(c.now)
		require.Equal(t, c.nextDay, nextTime.Weekday())
		require.Equal(t, 15, nextTime.Hour())
		require.Equal(t, 0, nextTime.Minute())
		require.Equal(t, 0, nextTime.Second())
	}
}

func TestDayTimeScheduleAfter(t *testing.T) {
	s := NewDayTimeSchedule([]time.Weekday{
		time.Wednesday,
		time.Thursday,
		time.Friday,
	}, 15, 0, 0)

	cases := []struct {
		now     time.Time
		nextDay time.Weekday
	}{
		{
			time.Date(2016, time.October, 3, 16, 0, 0, 0, time.Local), // monday
			time.Wednesday,
		},
		{
			time.Date(2016, time.October, 4, 16, 0, 0, 0, time.Local),
			time.Wednesday,
		},
		{
			time.Date(2016, time.October, 5, 16, 0, 0, 0, time.Local),
			time.Wednesday,
		},
		{
			time.Date(2016, time.October, 6, 16, 0, 0, 0, time.Local),
			time.Thursday,
		},
		{
			time.Date(2016, time.October, 7, 16, 0, 0, 0, time.Local),
			time.Friday,
		},
		{
			time.Date(2016, time.October, 8, 16, 0, 0, 0, time.Local),
			time.Wednesday,
		},
		{
			time.Date(2016, time.October, 9, 16, 0, 0, 0, time.Local),
			time.Wednesday,
		},
	}

	for i, c := range cases {
		nextTime := s.Next(c.now)
		require.Equal(t, c.nextDay, nextTime.Weekday(), fmt.Sprint(i))
		require.Equal(t, 15, nextTime.Hour())
		require.Equal(t, 0, nextTime.Minute())
		require.Equal(t, 0, nextTime.Second())
	}
}

func TestDayTimeScheduleNoDays(t *testing.T) {
	s := NewDayTimeSchedule(nil, 15, 0, 0)
	require.Equal(t, time.Time{}, s.Next(time.Now()))
}
