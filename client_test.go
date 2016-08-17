package flamingo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIntervalSchedule(t *testing.T) {
	s := NewIntervalSchedule(4 * time.Second)
	now := time.Now()

	assert.Equal(t, 4*time.Second, s.Next(now).Sub(now))
}

func TestDateSchedule(t *testing.T) {
	s := NewDateSchedule(15, 30, 25)
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	d := time.Date(now.Year(), now.Month(), now.Day(), 15, 30, 25, 0, now.Location())

	assert.Equal(t, d, s.Next(now))

	now = time.Date(now.Year(), now.Month(), now.Day(), 16, 0, 0, 0, now.Location())
	d = time.Date(now.Year(), now.Month(), now.Day()+1, 15, 30, 25, 0, now.Location())

	assert.Equal(t, d, s.Next(now))
}
