package slack

import (
	"testing"
	"time"

	"github.com/src-d/flamingo"
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

func TestImageToMessage(t *testing.T) {
	assert := assert.New(t)

	params := imageToMessage(flamingo.Image{
		ThumbnailURL: "foo",
		URL:          "bar",
		Text:         "baz",
	})

	assert.Equal(1, len(params.Attachments))
	assert.Equal("bar", params.Attachments[0].ImageURL)
	assert.Equal("baz", params.Attachments[0].Title)
	assert.Equal("bar", params.Attachments[0].TitleLink)
	assert.Equal("foo", params.Attachments[0].ThumbURL)
}
