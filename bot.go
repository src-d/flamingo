package flamingo

type Bot interface {
	ID() string
	Reply(OutgoingMessage) error
	Ask(OutgoingMessage) (Message, error)
	Conversation(Conversation) ([]Message, error)
	Say(OutgoingMessage) error
}
