package flamingo

import "io"

type Client interface {
	SetLogOutput(io.Writer)
	AddController(Controller)
	AddActionHandler(string, ActionHandler)
	AddBot(string, string)
	SetIntroHandler(IntroHandler)
	Run() error
	Stop() error
}

type ClientType uint

const (
	SlackClient ClientType = 1 << iota
)
