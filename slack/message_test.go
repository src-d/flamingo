package slack

import (
	"testing"
	"time"

	"github.com/src-d/flamingo"
	"github.com/stretchr/testify/require"
)

func TestCreatePostParams(t *testing.T) {
	require := require.New(t)

	msg := flamingo.NewOutgoingMessage("foo")
	msg.Sender = &flamingo.MessageSender{
		Username: "foo",
		IconURL:  "bar",
	}

	params := createPostParams(msg)
	require.Equal(params.IconURL, "bar")
	require.Equal(params.Username, "foo")
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
		require.Equal(t, parseTimestamp(c.ts), c.expected)
	}
}

func TestImageToMessage(t *testing.T) {
	require := require.New(t)

	params := imageToMessage(flamingo.Image{
		ThumbnailURL: "foo",
		URL:          "bar",
		Text:         "baz",
	})

	require.Equal(1, len(params.Attachments))
	require.Equal("bar", params.Attachments[0].ImageURL)
	require.Equal("baz", params.Attachments[0].Title)
	require.Equal("bar", params.Attachments[0].TitleLink)
	require.Equal("foo", params.Attachments[0].ThumbURL)
}
