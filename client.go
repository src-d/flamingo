package flamingo

import "io"

// Client is an abstract interface of a platforms-specific client.
// A client can only run for one platform. If you need to handle
// more than one platform you will have to start several clients
// for different platforms.
type Client interface {
	// SetLogOutput will write the logs to the given io.Writer.
	SetLogOutput(io.Writer)

	// AddController adds a new Controller to the Client.
	AddController(Controller)

	// AddActionHandler adds an ActionHandler for the given ID.
	AddActionHandler(string, ActionHandler)

	// AddBot adds a new bot with an ID and a token.
	AddBot(id string, token string)

	// SetIntroHandler sets the IntroHandler for the client.
	SetIntroHandler(IntroHandler)

	// Run starts the client.
	Run() error

	// Stop stops the client.
	Stop() error
}

// ClientType tells us what type of client is.
type ClientType uint

const (
	// SlackClient is a client for Slack.
	SlackClient ClientType = 1 << iota
)
