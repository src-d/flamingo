package main

import (
	"log"
	"os"
	"strings"

	"github.com/mvader/flamingo"
	"github.com/mvader/flamingo/slack"
)

type helloController struct{}

func (c *helloController) CanHandle(msg flamingo.Message) bool {
	return strings.ToLower(strings.TrimSpace(msg.Text)) == "hello"
}

func (c *helloController) Handle(bot flamingo.Bot, msg flamingo.Message) error {
	if err := bot.Say(flamingo.NewOutgoingMessage("hello!")); err != nil {
		return err
	}

	resp, err := bot.Ask(flamingo.NewOutgoingMessage("how are you?"))
	if err != nil {
		return err
	}

	text := strings.ToLower(strings.TrimSpace(resp.Text))
	if text == "good" || text == "fine" {
		if err := bot.Say(flamingo.NewOutgoingMessage("i'm glad!")); err != nil {
			return err
		}
	} else {
		if err := bot.Say(flamingo.NewOutgoingMessage(":(")); err != nil {
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

	client.AddController(&helloController{})
	client.AddBot(id, token)

	log.Fatal(client.Run())
}
