package slack

import (
	"io"

	"github.com/mvader/flamingo"
)

type ClientOptions struct {
}

type slackClient struct {
}

func NewClient(options ClientOptions) flamingo.Client {
	return &slackClient{}
}

func (c *slackClient) SetLogOutput(w io.Writer) {
}

func (c *slackClient) Run() error {
	return nil
}
