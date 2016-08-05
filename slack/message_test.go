package slack

import (
	"testing"
	"time"

	"github.com/mvader/flamingo"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
)

func TestCreatePostParamsWithCallback(t *testing.T) {
	assert := assert.New(t)

	msg := flamingo.NewOutgoingMessage("foo")
	msg.AddAttachment(flamingo.SlackClient, Attachment{
		ID: "fooo",
	})
	msg.Sender = &flamingo.MessageSender{
		Username: "foo",
		IconURL:  "bar",
	}

	params, err := createPostParams(
		"aaaa",
		"bbbb",
		msg,
	)
	assert.Nil(err)
	assert.Equal(len(params.Attachments), 1)
	assert.Equal(params.Attachments[0].CallbackID, "aaaa::bbbb::fooo")
	assert.Equal(params.IconURL, "bar")
	assert.Equal(params.Username, "foo")
}

func TestParseTimestamp(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		ts       string
		expected time.Time
	}{
		{"1458170917.164398", time.Date(2016, time.March, 17, 0, 28, 37, 164398, time.Local)},
		{"1458170917", time.Date(2016, time.March, 17, 0, 28, 37, 0, time.Local)},
	}

	for _, c := range cases {
		assert.Equal(parseTimestamp(c.ts), c.expected)
	}
}

func TestConvertAttachment(t *testing.T) {
	assert := assert.New(t)
	att := convertAttachment(slack.Attachment{
		Color:         "color",
		AuthorName:    "authorname",
		AuthorSubname: "authorsubname",
		AuthorLink:    "authorlink",
		AuthorIcon:    "authoricon",
		CallbackID:    "id",
		Title:         "title",
		TitleLink:     "titlelink",
		Pretext:       "pretext",
		Text:          "text",
		ImageURL:      "imageurl",
		ThumbURL:      "thumburl",
		Footer:        "footer",
		FooterIcon:    "footericon",
		MarkdownIn:    []string{"a", "b"},
		Fields: []slack.AttachmentField{
			slack.AttachmentField{
				Title: "field1",
				Value: "field1val",
				Short: true,
			},
			slack.AttachmentField{
				Title: "field2",
				Value: "field2val",
				Short: false,
			},
		},
		Actions: []slack.AttachmentAction{
			slack.AttachmentAction{
				Name:  "1name",
				Text:  "1text",
				Value: "1val",
				Style: "1style",
				Type:  "button",
				Confirm: []slack.ConfirmationField{
					slack.ConfirmationField{
						Title:       "title1",
						Text:        "text1",
						OkText:      "oktext1",
						DismissText: "dismisstext1",
					},
					slack.ConfirmationField{
						Title:       "title2",
						Text:        "text2",
						OkText:      "oktext2",
						DismissText: "dismisstext2",
					},
				},
			},
			slack.AttachmentAction{
				Name:  "2name",
				Text:  "2text",
				Value: "2val",
				Style: "2style",
				Type:  "button",
			},
		},
	})

	_att, ok := att.ForClient(flamingo.SlackClient)
	assert.True(ok)

	slackAtt, ok := _att.(Attachment)
	assert.True(ok)

	assert.Equal(slackAtt.ID, "id")
	assert.Equal(slackAtt.Color, "color")
	assert.Equal(slackAtt.AuthorName, "authorname")
	assert.Equal(slackAtt.AuthorSubname, "authorsubname")
	assert.Equal(slackAtt.AuthorLink, "authorlink")
	assert.Equal(slackAtt.AuthorIcon, "authoricon")
	assert.Equal(slackAtt.Title, "title")
	assert.Equal(slackAtt.TitleLink, "titlelink")
	assert.Equal(slackAtt.Pretext, "pretext")
	assert.Equal(slackAtt.Text, "text")
	assert.Equal(slackAtt.ImageURL, "imageurl")
	assert.Equal(slackAtt.ThumbURL, "thumburl")
	assert.Equal(slackAtt.Footer, "footer")
	assert.Equal(slackAtt.FooterIcon, "footericon")
	assert.Equal(slackAtt.MarkdownIn, []string{"a", "b"})
	assert.Equal(len(slackAtt.Fields), 2)
	assert.Equal(slackAtt.Fields[0].Short, true)
	assert.Equal(slackAtt.Fields[0].Value, "field1val")
	assert.Equal(slackAtt.Fields[0].Title, "field1")
	assert.Equal(slackAtt.Fields[1].Short, false)
	assert.Equal(slackAtt.Fields[1].Value, "field2val")
	assert.Equal(slackAtt.Fields[1].Title, "field2")
	assert.Equal(len(slackAtt.Actions), 2)
	assert.Equal(slackAtt.Actions[0].Name, "1name")
	assert.Equal(slackAtt.Actions[0].Style, "1style")
	assert.Equal(slackAtt.Actions[0].Value, "1val")
	assert.Equal(slackAtt.Actions[0].Text, "1text")
	assert.Equal(len(slackAtt.Actions[0].Confirm), 2)
	assert.Equal(slackAtt.Actions[0].Confirm[0].Title, "title1")
	assert.Equal(slackAtt.Actions[0].Confirm[0].Text, "text1")
	assert.Equal(slackAtt.Actions[0].Confirm[0].OkText, "oktext1")
	assert.Equal(slackAtt.Actions[0].Confirm[0].DismissText, "dismisstext1")
	assert.Equal(slackAtt.Actions[0].Confirm[1].Title, "title2")
	assert.Equal(slackAtt.Actions[0].Confirm[1].Text, "text2")
	assert.Equal(slackAtt.Actions[0].Confirm[1].OkText, "oktext2")
	assert.Equal(slackAtt.Actions[0].Confirm[1].DismissText, "dismisstext2")
	assert.Equal(slackAtt.Actions[1].Name, "2name")
	assert.Equal(slackAtt.Actions[1].Style, "2style")
	assert.Equal(slackAtt.Actions[1].Value, "2val")
	assert.Equal(slackAtt.Actions[1].Text, "2text")
}

func TestToSlackAttachment(t *testing.T) {
	assert := assert.New(t)
	slackAtt := toSlackAttachment("bot", "chan", Attachment{
		Color:         "color",
		AuthorName:    "authorname",
		AuthorSubname: "authorsubname",
		AuthorLink:    "authorlink",
		AuthorIcon:    "authoricon",
		ID:            "id",
		Title:         "title",
		TitleLink:     "titlelink",
		Pretext:       "pretext",
		Text:          "text",
		ImageURL:      "imageurl",
		ThumbURL:      "thumburl",
		Footer:        "footer",
		FooterIcon:    "footericon",
		MarkdownIn:    []string{"a", "b"},
		Fields: []Field{
			Field{
				Title: "field1",
				Value: "field1val",
				Short: true,
			},
			Field{
				Title: "field2",
				Value: "field2val",
				Short: false,
			},
		},
		Actions: []Action{
			Action{
				Name:  "1name",
				Text:  "1text",
				Value: "1val",
				Style: "1style",
				Confirm: []Confirmation{
					Confirmation{
						Title:       "title1",
						Text:        "text1",
						OkText:      "oktext1",
						DismissText: "dismisstext1",
					},
					Confirmation{
						Title:       "title2",
						Text:        "text2",
						OkText:      "oktext2",
						DismissText: "dismisstext2",
					},
				},
			},
			Action{
				Name:  "2name",
				Text:  "2text",
				Value: "2val",
				Style: "2style",
			},
		},
	})

	assert.Equal(slackAtt.CallbackID, "bot::chan::id")
	assert.Equal(slackAtt.Color, "color")
	assert.Equal(slackAtt.AuthorName, "authorname")
	assert.Equal(slackAtt.AuthorSubname, "authorsubname")
	assert.Equal(slackAtt.AuthorLink, "authorlink")
	assert.Equal(slackAtt.AuthorIcon, "authoricon")
	assert.Equal(slackAtt.Title, "title")
	assert.Equal(slackAtt.TitleLink, "titlelink")
	assert.Equal(slackAtt.Pretext, "pretext")
	assert.Equal(slackAtt.Text, "text")
	assert.Equal(slackAtt.ImageURL, "imageurl")
	assert.Equal(slackAtt.ThumbURL, "thumburl")
	assert.Equal(slackAtt.Footer, "footer")
	assert.Equal(slackAtt.FooterIcon, "footericon")
	assert.Equal(slackAtt.MarkdownIn, []string{"a", "b"})
	assert.Equal(len(slackAtt.Fields), 2)
	assert.Equal(slackAtt.Fields[0].Short, true)
	assert.Equal(slackAtt.Fields[0].Value, "field1val")
	assert.Equal(slackAtt.Fields[0].Title, "field1")
	assert.Equal(slackAtt.Fields[1].Short, false)
	assert.Equal(slackAtt.Fields[1].Value, "field2val")
	assert.Equal(slackAtt.Fields[1].Title, "field2")
	assert.Equal(len(slackAtt.Actions), 2)
	assert.Equal(slackAtt.Actions[0].Name, "1name")
	assert.Equal(slackAtt.Actions[0].Style, "1style")
	assert.Equal(slackAtt.Actions[0].Value, "1val")
	assert.Equal(slackAtt.Actions[0].Text, "1text")
	assert.Equal(slackAtt.Actions[0].Type, "button")
	assert.Equal(len(slackAtt.Actions[0].Confirm), 2)
	assert.Equal(slackAtt.Actions[0].Confirm[0].Title, "title1")
	assert.Equal(slackAtt.Actions[0].Confirm[0].Text, "text1")
	assert.Equal(slackAtt.Actions[0].Confirm[0].OkText, "oktext1")
	assert.Equal(slackAtt.Actions[0].Confirm[0].DismissText, "dismisstext1")
	assert.Equal(slackAtt.Actions[0].Confirm[1].Title, "title2")
	assert.Equal(slackAtt.Actions[0].Confirm[1].Text, "text2")
	assert.Equal(slackAtt.Actions[0].Confirm[1].OkText, "oktext2")
	assert.Equal(slackAtt.Actions[0].Confirm[1].DismissText, "dismisstext2")
	assert.Equal(slackAtt.Actions[1].Name, "2name")
	assert.Equal(slackAtt.Actions[1].Style, "2style")
	assert.Equal(slackAtt.Actions[1].Value, "2val")
	assert.Equal(slackAtt.Actions[1].Text, "2text")
	assert.Equal(slackAtt.Actions[1].Type, "button")
}
