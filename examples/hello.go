package main

import (
	"strings"

	"github.com/mvader/flamingo"
)

const (
	token = "YOUR_TOKEN"
	name  = "test-bot"
)

type HelloHandler struct {
}

func (HelloHandler) Name() string { return "hello-handler" }
func (HelloHandler) Help() string { return "i say hello ```say hello```" }

func (HelloHandler) IsMatch(_ string, text string) bool {
	return strings.ToLower(strings.TrimSpace(text)) == "say hello"
}

func (HelloHandler) Handle(msg flamingo.Message) {
	msg.Reply("Hello!")
}

func main() {
	e := flamingo.New(name, token, flamingo.Options{
		Debug: true,
	})
	e.AddHandler(&HelloHandler{})
	e.Run()
}
