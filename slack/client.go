package slack

import (
	"io"

	"github.com/mvader/flamingo"
)

type ClientOptions struct {
}

type slackClient struct {
	token          string
	controllers    []flamingo.Controller
	actionHandlers map[string]flamingo.ActionHandler
}

func NewClient(token string, options ClientOptions) flamingo.Client {
	return &slackClient{
		token:          token,
		actionHandlers: make(map[string]flamingo.ActionHandler),
	}
}

func (c *slackClient) SetLogOutput(w io.Writer) {
}

func (c *slackClient) AddController(ctrl flamingo.Controller) {
	c.controllers = append(c.controllers, ctrl)
}

func (c *slackClient) AddActionHandler(id string, handler flamingo.ActionHandler) {
	c.actionHandlers[id] = handler
}

func (c *slackClient) Run() error {
	return nil
}
