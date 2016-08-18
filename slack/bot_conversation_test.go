package slack

import (
	"testing"
	"time"

	"github.com/src-d/flamingo"
	"github.com/mvader/slack"
	"github.com/stretchr/testify/assert"
)

func TestBotConversation(t *testing.T) {
	assert := assert.New(t)

	mock := &slackRTMMock{
		events: make(chan slack.RTMEvent),
	}
	cli := NewClient("", ClientOptions{Debug: true}).(*slackClient)
	convo, err := newBotConversation("aaaa", "Cbbbb", mock, cli)
	assert.Nil(err)
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
	assert.True(entered)
	assert.Equal(1, len(ctrl.msgs))
}
