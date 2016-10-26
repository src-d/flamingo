package slack

import (
	"testing"
	"time"

	"github.com/mvader/slack"
	"github.com/src-d/flamingo"
	"github.com/stretchr/testify/require"
)

func TestBotConversation(t *testing.T) {
	require := require.New(t)

	mock := &slackRTMMock{
		events: make(chan slack.RTMEvent),
	}
	cli := NewClient("", ClientOptions{Debug: true}).(*slackClient)
	convo, err := newBotConversation("aaaa", "Cbbbb", mock, cli)
	require.Nil(err)
	go convo.run()
	defer convo.stop()

	ctrl := &helloCtrl{}
	cli.AddController(ctrl)
	var entered bool
	cli.AddActionHandler("foo", func(b flamingo.Bot, action flamingo.Action) {
		entered = true
	})

	convo.messages <- &slack.MessageEvent{
		Msg: slack.Msg{
			Text: "hello",
		},
	}

	convo.actions <- slack.AttachmentActionCallback{
		CallbackID: "foo",
	}

	<-time.After(100 * time.Millisecond)
	require.True(entered)
	require.Equal(1, len(ctrl.msgs))
}
