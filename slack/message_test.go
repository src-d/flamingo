package slack

import (
	"testing"
	"time"

	"github.com/mvader/flamingo"
	"github.com/stretchr/testify/assert"
)

func TestCreatePostParams(t *testing.T) {
	assert := assert.New(t)

	msg := flamingo.NewOutgoingMessage("foo")
	msg.Sender = &flamingo.MessageSender{
		Username: "foo",
		IconURL:  "bar",
	}

	params := createPostParams(msg)
	assert.Equal(params.IconURL, "bar")
	assert.Equal(params.Username, "foo")
}

func TestParseTimestamp(t *testing.T) {
	cases := []struct {
		ts                                  string
		year                                int
		month                               time.Month
		day, hour, minutes, seconds, millis int
	}{
		{"1458170917.164398", 2016, time.March, 17, 0, 28, 37, 164398},
		{"1458170917", 2016, time.March, 17, 0, 28, 37, 0},
	}

	for _, c := range cases {
		ts := parseTimestamp(c.ts)
		expected := time.Date(c.year, c.month, c.day, c.hour, c.minutes, c.seconds, c.millis, ts.Location())
		assert.Equal(t, ts, expected)
	}
}
