package slack

import (
	"strconv"
	"strings"
	"time"

	"github.com/mvader/flamingo"
	"github.com/nlopes/slack"
)

func newMessage(user flamingo.User, channel flamingo.Channel, src slack.Msg) flamingo.Message {
	return flamingo.Message{
		User:    user,
		Type:    flamingo.SlackClient,
		Channel: channel,
		Text:    src.Text,
		Time:    parseTimestamp(src.Timestamp),
		Extra:   src,
	}
}

func parseTimestamp(timestamp string) time.Time {
	parts := strings.Split(timestamp, ".")
	sec, _ := strconv.ParseInt(parts[0], 10, 64)
	var nsec int64
	if len(parts) > 1 {
		nsec, _ = strconv.ParseInt(parts[1], 10, 64)
	}
	return time.Unix(sec, nsec)
}

func createPostParams(msg flamingo.OutgoingMessage) slack.PostMessageParameters {
	params := slack.PostMessageParameters{
		LinkNames: 1,
		Markdown:  true,
		AsUser:    true,
	}

	if msg.Sender != nil {
		params.IconURL = msg.Sender.IconURL
		params.Username = msg.Sender.Username
	}

	return params
}
