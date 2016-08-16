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
		ts       string
		expected time.Time
	}{
		{"1458170917.164398", time.Unix(1458170917, 164398)},
		{"1458170917", time.Unix(1458170917, 0)},
	}

	for _, c := range cases {
		assert.Equal(t, parseTimestamp(c.ts), c.expected)
	}
}
