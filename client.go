package flamingo

import "io"

type Client interface {
	SetLogOutput(io.Writer)
	Run() error
}

type ClientType uint

const (
	SlackClient ClientType = 1 << iota
)
