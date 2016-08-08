package flamingo

type Bot interface {
	ID() string
	Reply(Message, OutgoingMessage) error
	Ask(OutgoingMessage) (Message, error)
	Conversation(Conversation) ([]Message, error)
	Say(OutgoingMessage) error
	Form(Form) error
	WaitForAction(string, ActionWaitingPolicy) (Action, error)
}
