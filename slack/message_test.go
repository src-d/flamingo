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

	params, err := createPostParams(msg)
	assert.Nil(err)
	assert.Equal(params.IconURL, "bar")
	assert.Equal(params.Username, "foo")
}

func TestParseTimestamp(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		ts       string
		expected time.Time
	}{
		{"1458170917.164398", time.Date(2016, time.March, 17, 0, 28, 37, 164398, time.Local)},
		{"1458170917", time.Date(2016, time.March, 17, 0, 28, 37, 0, time.Local)},
	}

	for _, c := range cases {
		assert.Equal(parseTimestamp(c.ts), c.expected)
	}
}
