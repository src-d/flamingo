package slack

import (
	"fmt"

	"github.com/mvader/slack"
	"github.com/src-d/flamingo"
)

type btnStyle flamingo.ButtonType

func (b btnStyle) String() string {
	switch b {
	case btnStyle(flamingo.PrimaryButton):
		return "primary"
	case btnStyle(flamingo.DangerButton):
		return "danger"
	default:
		return ""
	}
}

func formToMessage(bot, channel string, form flamingo.Form) slack.PostMessageParameters {
	params := slack.PostMessageParameters{
		LinkNames: 1,
		Markdown:  true,
		AsUser:    true,
	}

	if form.Combine {
		params.Attachments = append(params.Attachments, combinedAttachment(bot, channel, form))
	} else {
		params.Attachments = append(params.Attachments, headerAttachment(form))
		for _, g := range form.Fields {
			att := groupToAttachment(bot, channel, g)
			att.Color = form.Color
			params.Attachments = append(params.Attachments, att)
		}
	}

	if len(params.Attachments) > 0 {
		params.Attachments[len(params.Attachments)-1].Footer = form.Footer
	}

	return params
}

func combinedAttachment(bot, channel string, form flamingo.Form) slack.Attachment {
	a := headerAttachment(form)
	for _, g := range form.Fields {
		addGroupToAttachment(&a, bot, channel, g)
	}

	return a
}

func buttonToAction(f flamingo.Button) slack.AttachmentAction {
	action := slack.AttachmentAction{
		Type:  "button",
		Text:  f.Text,
		Name:  f.Name,
		Value: f.Value,
		Style: btnStyle(f.Type).String(),
	}

	if f.Confirmation != nil {
		action.Confirm = append(action.Confirm, slack.ConfirmationField{
			Title:       f.Confirmation.Title,
			Text:        f.Confirmation.Text,
			OkText:      f.Confirmation.Ok,
			DismissText: f.Confirmation.Dismiss,
		})
	}

	return action
}

func textFieldToField(f flamingo.TextField) slack.AttachmentField {
	return slack.AttachmentField{
		Title: f.Title,
		Value: f.Value,
		Short: f.Short,
	}
}

func headerAttachment(form flamingo.Form) slack.Attachment {
	return slack.Attachment{
		Title:      form.Title,
		Text:       form.Text,
		Color:      form.Color,
		MarkdownIn: []string{"text", "pretext", "title", "values", "footer"},
	}
}

func groupToAttachment(bot, channel string, group flamingo.FieldGroup) slack.Attachment {
	a := slack.Attachment{}
	addGroupToAttachment(&a, bot, channel, group)
	return a
}

func addGroupToAttachment(a *slack.Attachment, bot, channel string, group flamingo.FieldGroup) {
	if group.Type() == flamingo.ButtonGroup && group.ID() != "" {
		a.CallbackID = fmt.Sprintf("%s::%s::%s", bot, channel, group.ID())
	}

	for _, i := range group.Items() {
		switch f := i.(type) {
		case flamingo.Button:
			a.Actions = append(a.Actions, buttonToAction(f))
		case flamingo.TextField:
			a.Fields = append(a.Fields, textFieldToField(f))
		case flamingo.Image:
			a.ImageURL = f.URL
			a.ThumbURL = f.ThumbnailURL
			a.Title = f.Text
			a.TitleLink = f.URL
		}
	}
}
