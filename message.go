package flamingo

import (
	"fmt"

	"github.com/nlopes/slack"
)

// Delegate will interact with the Engine exponsing some of its
// operations to be used in the handlers.
type Delegate interface {
	// Debug will print a debug log
	Debug(string, ...interface{})
	// Info will print an info log
	Info(string, ...interface{})
	// Crit will print a crit log
	Crit(string, ...interface{})
	// Error will print an error log
	Error(string, ...interface{})
	// Warn will print a warn log
	Warn(string, ...interface{})
	// SendCustomMsg will send a message with custom parameters
	SendCustomMsg(string, string, slack.PostMessageParameters) error
	// SendMessage sends a message to a channel with the default params
	SendMsg(string, string) error
	// GetChannel retrieves a channel info by its ID
	GetChannel(string) Channel
	// GetUser retrieves a channel user by its ID
	GetUser(string) (*slack.User, error)
}

// Channel is a wrapper around the API channel.
type Channel struct {
	*slack.Channel
	ID string
}

// Name returns the name of the channel or empty if it does not exist (for example, is a private message).
func (c Channel) Name() string {
	if c.Channel == nil {
		return ""
	}

	return c.Channel.Name
}

// Message contains the channel, the user and the text of the message
// along with the delegate, which will make it possible to do actions
// with a message.
type Message struct {
	Channel Channel
	Text    string
	User    *slack.User
	Delegate
}

// Reply will send a reply to the user of the message with the given text.
func (m Message) Reply(text string) error {
	m.Delegate.Debug("sending message", "channel", m.Channel.Name, "to", m.User.Name, "text", text)
	return m.Delegate.SendMsg(m.Channel.ID, fmt.Sprintf("@%s: %s", m.User.Name, text))
}

// CustomReply will send a reply to the user of the message with
// the given text and custom parameters.
func (m Message) CustomReply(text string, params slack.PostMessageParameters) error {
	m.Delegate.Debug("sending custom message", "channel", m.Channel.Name, "to", m.User.Name, "text", text)
	return m.Delegate.SendCustomMsg(m.Channel.ID, fmt.Sprintf("@%s: %s", m.User.Name, text), params)
}
