package slack

import (
	"testing"
	"time"

	"github.com/mvader/flamingo"
	"github.com/stretchr/testify/assert"
)

func TestCreatePostParamsWithCallback(t *testing.T) {
	assert := assert.New(t)

	msg := flamingo.NewOutgoingMessage("foo")
	msg.AddAttachment(flamingo.SlackClient, Attachment{
		ID: "fooo",
	})

	params, err := createPostParams(
		"aaaa",
		"bbbb",
		msg,
	)
	assert.Nil(err)
	assert.Equal(len(params.Attachments), 1)
	assert.Equal(params.Attachments[0].CallbackID, "aaaa::bbbb::fooo")
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
