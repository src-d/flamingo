package flamingo

import "time"

type OutgoingMessage struct {
	ChannelID string
	Text      string
	Sender    *MessageSender
}

func NewOutgoingMessage(text string) OutgoingMessage {
	return OutgoingMessage{Text: text}
}

type MessageSender struct {
	Username string
	IconURL  string
}

type Message struct {
	ID      string
	Type    ClientType
	User    User
	Channel Channel
	Time    time.Time
	Text    string
	Extra   interface{}
}

type Conversation []OutgoingMessage

type User struct {
	ID       string
	Username string
	Name     string
	IsBot    bool
	Type     ClientType
	Extra    interface{}
}

type Channel struct {
	ID    string
	Name  string
	IsDM  bool
	Users []User
	Type  ClientType
	Extra interface{}
}
