package slack

import (
	"testing"

	"github.com/src-d/flamingo"
	"github.com/stretchr/testify/assert"
)

func TestFormToMessage(t *testing.T) {
	assert := assert.New(t)

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

	assert.Equal(5, len(params.Attachments))
	assert.Equal("foo", params.Attachments[0].AuthorName)
	assert.Equal("bar", params.Attachments[0].AuthorIcon)
	assert.Equal("", params.Attachments[0].Footer)
	assert.Equal("", params.Attachments[1].Footer)
	assert.Equal("", params.Attachments[2].Footer)
	assert.Equal("", params.Attachments[3].Footer)
	assert.Equal("footer", params.Attachments[4].Footer)
	assert.Equal("foo", params.Attachments[3].ImageURL)
	assert.Equal("hello", params.Attachments[4].Text)

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

	assert.Equal(1, len(params.Attachments))
	assert.Equal("footer", params.Attachments[0].Footer)
	assert.Equal("foo", params.Attachments[0].ImageURL)
}

func TestCombinedAttachment(t *testing.T) {
	assert := assert.New(t)

	a := combinedAttachment("aaaa", "bbbb", flamingo.Form{
		Title: "title",
		Text:  "text",
		Color: "color",
		Fields: []flamingo.FieldGroup{
			flamingo.NewButtonGroup("fooo", flamingo.Button{}),
			flamingo.NewTextFieldGroup(flamingo.TextField{}),
		},
	})

	assert.Equal("aaaa::bbbb::fooo", a.CallbackID)
	assert.Equal("title", a.Title)
	assert.Equal("text", a.Text)
	assert.Equal("color", a.Color)
	assert.Equal(1, len(a.Actions))
	assert.Equal(1, len(a.Fields))
}

func TestButtonToAction(t *testing.T) {
	assert := assert.New(t)

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
	assert.Equal("button", a.Type)
	assert.Equal("text", a.Text)
	assert.Equal("name", a.Name)
	assert.Equal("value", a.Value)
	assert.Equal("danger", a.Style)
	assert.Equal(1, len(a.Confirm))
	assert.Equal("title", a.Confirm[0].Title)
	assert.Equal("text", a.Confirm[0].Text)
	assert.Equal("ok", a.Confirm[0].OkText)
	assert.Equal("dismiss", a.Confirm[0].DismissText)
}

func TestTextFieldToField(t *testing.T) {
	assert := assert.New(t)

	a := textFieldToField(flamingo.TextField{
		Title: "title",
		Value: "value",
		Short: true,
	})

	assert.Equal("title", a.Title)
	assert.Equal("value", a.Value)
	assert.True(a.Short)
}

func TestHeaderAttachment(t *testing.T) {
	assert := assert.New(t)

	a := headerAttachment(flamingo.Form{
		Title: "title",
		Text:  "text",
		Color: "ffffff",
	})

	assert.Equal("title", a.Title)
	assert.Equal("text", a.Text)
	assert.Equal("ffffff", a.Color)
}

func TestGroupToAttachment(t *testing.T) {
	assert := assert.New(t)

	a := groupToAttachment("foo", "bar", flamingo.NewTextFieldGroup(
		flamingo.TextField{Title: "title", Value: "value", Short: true},
		flamingo.TextField{Title: "title2", Value: "value2", Short: false},
	))
	assert.Equal("", a.CallbackID)
	assert.Equal(2, len(a.Fields))
	assert.Equal(0, len(a.Actions))
	assert.Equal("title", a.Fields[0].Title)
	assert.Equal("value", a.Fields[0].Value)
	assert.True(a.Fields[0].Short)
	assert.Equal("title2", a.Fields[1].Title)
	assert.Equal("value2", a.Fields[1].Value)
	assert.False(a.Fields[1].Short)

	a = groupToAttachment("foo", "bar", flamingo.NewButtonGroup(
		"",
		flamingo.Button{Name: "name", Value: "value", Text: "text", Type: flamingo.DangerButton},
	))
	assert.Equal("", a.CallbackID)
	assert.Equal(0, len(a.Fields))
	assert.Equal(1, len(a.Actions))
	assert.Equal("button", a.Actions[0].Type)
	assert.Equal("text", a.Actions[0].Text)
	assert.Equal("name", a.Actions[0].Name)
	assert.Equal("value", a.Actions[0].Value)
	assert.Equal("danger", a.Actions[0].Style)

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
	assert.Equal("foo::bar::action", a.CallbackID)
	assert.Equal(0, len(a.Fields))
	assert.Equal(1, len(a.Actions))
	assert.Equal(1, len(a.Actions[0].Confirm))
	assert.Equal("title", a.Actions[0].Confirm[0].Title)
	assert.Equal("text", a.Actions[0].Confirm[0].Text)
	assert.Equal("ok", a.Actions[0].Confirm[0].OkText)
	assert.Equal("dismiss", a.Actions[0].Confirm[0].DismissText)
}
