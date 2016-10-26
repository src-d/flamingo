package slack

import (
	"testing"

	"github.com/src-d/flamingo"
	"github.com/stretchr/testify/require"
)

func TestFormToMessage(t *testing.T) {
	require := require.New(t)

	params := formToMessage("aaaa", "bbbb", flamingo.Form{
		AuthorName:    "foo",
		AuthorIconURL: "bar",
		Title:         "title",
		Text:          "text",
		Color:         "color",
		Footer:        "footer",
		Fields: []flamingo.FieldGroup{
			flamingo.NewButtonGroup("fooo", flamingo.Button{}),
			flamingo.NewTextFieldGroup(flamingo.TextField{}),
			flamingo.Image{URL: "foo"},
			flamingo.Text("hello"),
		},
	})

	require.Equal(5, len(params.Attachments))
	require.Equal("foo", params.Username)
	require.Equal("bar", params.IconURL)
	require.Equal("", params.Attachments[0].Footer)
	require.Equal("", params.Attachments[1].Footer)
	require.Equal("", params.Attachments[2].Footer)
	require.Equal("", params.Attachments[3].Footer)
	require.Equal("footer", params.Attachments[4].Footer)
	require.Equal("foo", params.Attachments[3].ImageURL)
	require.Equal("hello", params.Attachments[4].Text)

	params = formToMessage("aaaa", "bbbb", flamingo.Form{
		Title:   "title",
		Text:    "text",
		Color:   "color",
		Footer:  "footer",
		Combine: true,
		Fields: []flamingo.FieldGroup{
			flamingo.NewButtonGroup("fooo", flamingo.Button{}),
			flamingo.NewTextFieldGroup(flamingo.TextField{}),
			flamingo.Image{URL: "foo"},
		},
	})

	require.Equal(1, len(params.Attachments))
	require.Equal("footer", params.Attachments[0].Footer)
	require.Equal("foo", params.Attachments[0].ImageURL)
}

func TestCombinedAttachment(t *testing.T) {
	require := require.New(t)

	a := combinedAttachment("aaaa", "bbbb", flamingo.Form{
		Title: "title",
		Text:  "text",
		Color: "color",
		Fields: []flamingo.FieldGroup{
			flamingo.NewButtonGroup("fooo", flamingo.Button{}),
			flamingo.NewTextFieldGroup(flamingo.TextField{}),
		},
	})

	require.Equal("aaaa::bbbb::fooo", a.CallbackID)
	require.Equal("title", a.Title)
	require.Equal("text", a.Text)
	require.Equal("color", a.Color)
	require.Equal(1, len(a.Actions))
	require.Equal(1, len(a.Fields))
}

func TestButtonToAction(t *testing.T) {
	require := require.New(t)

	a := buttonToAction(flamingo.Button{
		Name:  "name",
		Value: "value",
		Text:  "text",
		Type:  flamingo.DangerButton,
		Confirmation: &flamingo.Confirmation{
			Title:   "title",
			Text:    "text",
			Ok:      "ok",
			Dismiss: "dismiss",
		},
	})
	require.Equal("button", a.Type)
	require.Equal("text", a.Text)
	require.Equal("name", a.Name)
	require.Equal("value", a.Value)
	require.Equal("danger", a.Style)
	require.Equal(1, len(a.Confirm))
	require.Equal("title", a.Confirm[0].Title)
	require.Equal("text", a.Confirm[0].Text)
	require.Equal("ok", a.Confirm[0].OkText)
	require.Equal("dismiss", a.Confirm[0].DismissText)
}

func TestTextFieldToField(t *testing.T) {
	require := require.New(t)

	a := textFieldToField(flamingo.TextField{
		Title: "title",
		Value: "value",
		Short: true,
	})

	require.Equal("title", a.Title)
	require.Equal("value", a.Value)
	require.True(a.Short)
}

func TestHeaderAttachment(t *testing.T) {
	require := require.New(t)

	a := headerAttachment(flamingo.Form{
		Title: "title",
		Text:  "text",
		Color: "ffffff",
	})

	require.Equal("title", a.Title)
	require.Equal("text", a.Text)
	require.Equal("ffffff", a.Color)
}

func TestGroupToAttachment(t *testing.T) {
	require := require.New(t)

	a := groupToAttachment("foo", "bar", flamingo.NewTextFieldGroup(
		flamingo.TextField{Title: "title", Value: "value", Short: true},
		flamingo.TextField{Title: "title2", Value: "value2", Short: false},
	))
	require.Equal("", a.CallbackID)
	require.Equal(2, len(a.Fields))
	require.Equal(0, len(a.Actions))
	require.Equal("title", a.Fields[0].Title)
	require.Equal("value", a.Fields[0].Value)
	require.True(a.Fields[0].Short)
	require.Equal("title2", a.Fields[1].Title)
	require.Equal("value2", a.Fields[1].Value)
	require.False(a.Fields[1].Short)

	a = groupToAttachment("foo", "bar", flamingo.NewButtonGroup(
		"",
		flamingo.Button{Name: "name", Value: "value", Text: "text", Type: flamingo.DangerButton},
	))
	require.Equal("", a.CallbackID)
	require.Equal(0, len(a.Fields))
	require.Equal(1, len(a.Actions))
	require.Equal("button", a.Actions[0].Type)
	require.Equal("text", a.Actions[0].Text)
	require.Equal("name", a.Actions[0].Name)
	require.Equal("value", a.Actions[0].Value)
	require.Equal("danger", a.Actions[0].Style)

	a = groupToAttachment("foo", "bar", flamingo.NewButtonGroup(
		"action",
		flamingo.Button{
			Name:  "name",
			Value: "value",
			Text:  "text",
			Type:  flamingo.DangerButton,
			Confirmation: &flamingo.Confirmation{
				Title:   "title",
				Text:    "text",
				Ok:      "ok",
				Dismiss: "dismiss",
			},
		},
	))
	require.Equal("foo::bar::action", a.CallbackID)
	require.Equal(0, len(a.Fields))
	require.Equal(1, len(a.Actions))
	require.Equal(1, len(a.Actions[0].Confirm))
	require.Equal("title", a.Actions[0].Confirm[0].Title)
	require.Equal("text", a.Actions[0].Confirm[0].Text)
	require.Equal("ok", a.Actions[0].Confirm[0].OkText)
	require.Equal("dismiss", a.Actions[0].Confirm[0].DismissText)
}
