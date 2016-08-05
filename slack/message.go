package slack

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mvader/flamingo"
	"github.com/nlopes/slack"
)

type Attachment struct {
	Color         string
	AuthorName    string
	AuthorSubname string
	AuthorLink    string
	AuthorIcon    string

	Title     string
	TitleLink string
	Pretext   string
	Text      string

	ImageURL string
	ThumbURL string

	Fields     []Field
	ID         string
	Actions    []Action
	MarkdownIn []string

	Footer     string
	FooterIcon string
}

type Field struct {
	Title string
	Value string
	Short bool
}

type Action struct {
	Name    string
	Text    string
	Value   string
	Style   string
	Confirm []Confirmation
}

type Confirmation struct {
	Title       string
	Text        string
	OkText      string
	DismissText string
}

func newMessage(user flamingo.User, channel flamingo.Channel, src *slack.MessageEvent) flamingo.Message {
	var attachments = make([]flamingo.Attachment, 0, len(src.Msg.Attachments))
	for _, a := range src.Msg.Attachments {
		attachments = append(attachments, convertAttachment(a))
	}

	return flamingo.Message{
		User:        user,
		Type:        flamingo.SlackClient,
		Channel:     channel,
		Text:        src.Msg.Text,
		Time:        parseTimestamp(src.Msg.Timestamp),
		Attachments: attachments,
		Extra:       src.Msg,
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

func createPostParams(msg flamingo.OutgoingMessage) (slack.PostMessageParameters, error) {
	params := slack.PostMessageParameters{
		LinkNames: 1,
		Markdown:  true,
		AsUser:    true,
	}

	if msg.Sender != nil {
		params.IconURL = msg.Sender.IconURL
		params.Username = msg.Sender.Username
	}

	for _, a := range msg.Attachments {
		if att, ok := a.ForClient(flamingo.SlackClient); ok {
			slackAtt, ok := att.(Attachment)
			if !ok {
				return params, fmt.Errorf("attachment %#v is not of type slack.Attachment", att)
			}

			params.Attachments = append(params.Attachments, toSlackAttachment(slackAtt))
		}
	}

	return params, nil
}

func toSlackAttachment(attachment Attachment) slack.Attachment {
	return slack.Attachment{}
}

func convertAttachment(attachment slack.Attachment) flamingo.Attachment {
	a := Attachment{
		Color:         attachment.Color,
		AuthorName:    attachment.AuthorName,
		AuthorSubname: attachment.AuthorSubname,
		AuthorLink:    attachment.AuthorLink,
		AuthorIcon:    attachment.AuthorIcon,
		Title:         attachment.Title,
		TitleLink:     attachment.TitleLink,
		Pretext:       attachment.Pretext,
		Text:          attachment.Text,
		ImageURL:      attachment.ImageURL,
		ThumbURL:      attachment.ThumbURL,
		Footer:        attachment.Footer,
		FooterIcon:    attachment.FooterIcon,
		MarkdownIn:    attachment.MarkdownIn,
	}

	for _, f := range attachment.Fields {
		a.Fields = append(a.Fields, Field{
			Title: f.Title,
			Value: f.Value,
			Short: f.Short,
		})
	}

	for _, action := range attachment.Actions {
		act := Action{
			Name:  action.Name,
			Text:  action.Text,
			Value: action.Value,
			Style: action.Style,
		}

		for _, c := range action.Confirm {
			act.Confirm = append(act.Confirm, Confirmation{
				Title:       c.Title,
				Text:        c.Text,
				OkText:      c.OkText,
				DismissText: c.DismissText,
			})
		}
	}

	return flamingo.NewAttachment().Add(flamingo.SlackClient, a)
}
