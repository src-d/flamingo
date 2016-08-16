package main

import (
	"log"
	"os"
	"strings"

	"github.com/mvader/flamingo"
	"github.com/mvader/flamingo/slack"
)

type helloController struct{}

func (c *helloController) HandleIntro(bot flamingo.Bot, channel flamingo.Channel) error {
	_, err := bot.Say(flamingo.NewOutgoingMessage("Hey! I am a bot, try saying `hello` to me"))
	return err
}

func (c *helloController) CanHandle(msg flamingo.Message) bool {
	return strings.ToLower(strings.TrimSpace(msg.Text)) == "hello"
}

func (c *helloController) Handle(bot flamingo.Bot, msg flamingo.Message) error {
	if _, err := bot.Say(flamingo.NewOutgoingMessage("hello!")); err != nil {
		return err
	}

	_, resp, err := bot.Ask(flamingo.NewOutgoingMessage("how are you?"))
	if err != nil {
		return err
	}

	text := strings.ToLower(strings.TrimSpace(resp.Text))
	if text == "good" || text == "fine" {
		if _, err := bot.Say(flamingo.NewOutgoingMessage("i'm glad!")); err != nil {
			return err
		}
	} else {
		if _, err := bot.Say(flamingo.NewOutgoingMessage(":(")); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	token := os.Getenv("SLACK_TOKEN")
	id := os.Getenv("BOT_ID")
	client := slack.NewClient(token, slack.ClientOptions{
		Debug: true,
	})

	ctrl := &helloController{}
	client.AddController(ctrl)
	client.AddBot(id, token)
	client.SetIntroHandler(ctrl)

	log.Fatal(client.Run())
}
