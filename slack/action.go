package slack

import (
	"github.com/mvader/flamingo"
	"github.com/mvader/slack"
)

func convertAction(action slack.AttachmentActionCallback) flamingo.Action {
	var userAction flamingo.UserAction
	for _, a := range action.Actions {
		userAction = flamingo.UserAction{
			Name:  a.Name,
			Value: a.Value,
		}
	}

	user := flamingo.User{
		ID:       action.User.ID,
		Username: action.User.Name,
	}
	channel := flamingo.Channel{
		ID:   action.Channel.ID,
		Name: action.Channel.Name,
	}
	return flamingo.Action{
		UserAction:      userAction,
		User:            user,
		Channel:         channel,
		OriginalMessage: newMessage(user, channel, action.OriginalMessage.Msg),
		Extra:           action,
	}
}
