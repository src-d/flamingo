package flamingo

import (
	"regexp"
	"strings"
	"time"
)

// OutgoingMessage is a message that is going to be sent to the user.
type OutgoingMessage struct {
	// ChannelID if is different from the channel in which the current handler
	// or controller is executing.
	ChannelID string
	// Text of the message.
	Text string
	// Sender, if provided, changes the username and icon of the message.
	Sender *MessageSender
}

// NewOutgoingMessage creates a simple OutgoingMessage with just text.
func NewOutgoingMessage(text string) OutgoingMessage {
	return OutgoingMessage{Text: text}
}

// MessageSender define the properties of the sender of a message.
type MessageSender struct {
	// Username is the name of the user.
	Username string
	// IconURL for the message poster picture.
	IconURL string
}

// Message is the data of a message received.
type Message struct {
	// ID of the message.
	ID string
	// Type of client this message comes from.
	Type ClientType
	// User that posted the message.
	User User
	// Channel the message was posted in.
	Channel Channel
	// Time of the message.
	Time time.Time
	// Text of the message.
	Text string
	// Extra contains extra data given by the specific content.
	Extra interface{}
}

// MatchString reports if the message text matches the given text lowercased
// with leading and trailing spaces removed.
func (m Message) MatchString(str string) bool {
	return normalize(m.Text) == normalize(str)
}

// MatchStringCase reports if the message text matches the given text with case sensitivity
// with leading and trailing spaces removed.
func (m Message) MatchStringCase(str string) bool {
	return strings.TrimSpace(m.Text) == strings.TrimSpace(str)
}

// MatchRegex reports if the message text matches the given regexp
// with leading and trailing spaces removed.
func (m Message) MatchRegex(regex *regexp.Regexp) bool {
	return regex.MatchString(strings.TrimSpace(m.Text))
}

func normalize(str string) string {
	return strings.ToLower(strings.TrimSpace(str))
}

// Conversation is a collection of OutgoingMessages.
type Conversation []OutgoingMessage

// User is the representation of an user.
type User struct {
	// ID of the user.
	ID string
	// Username is the handle of the user.
	Username string
	// Name is the real name, if any.
	Name string
	// Email is the email of the user, if any.
	Email string
	// IsBot will be true if the user is a bot.
	IsBot bool
	// Type is the specific client this user comes from.
	Type ClientType
	// Extra contains extra data given by the specific content.
	Extra interface{}
}

// Channel represents a group, channel, direct message or conversation, all
// in one, depending on the specific client you are using.
type Channel struct {
	// ID of the channel
	ID string
	// Name if any
	Name string
	// IsDM will be true if is a direct message. That is, between user and bot.
	IsDM bool
	// Users is the list of users in the channel. Some implementations may not
	// provide this.
	Users []User
	// Type is the type of client this channel comes from.
	Type ClientType
	// Extra contains extra data given by the specific client.
	Extra interface{}
}
