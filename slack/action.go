package slack

import (
	"github.com/mvader/slack"
	"github.com/src-d/flamingo"
)

func convertAction(action slack.AttachmentActionCallback, api slackAPI) (flamingo.Action, error) {
	var userAction flamingo.UserAction
	for _, a := range action.Actions {
		userAction = flamingo.UserAction{
			Name:  a.Name,
			Value: a.Value,
		}
	}

	info, err := api.GetUserInfo(action.User.ID)
	if err != nil {
		return flamingo.Action{}, err
	}

	user := convertUser(info)
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
	}, nil
}
