package flamingo

type AnswerChecker func(Message) *OutgoingMessage

type Bot interface {
	ID() string
	Reply(Message, OutgoingMessage) (string, error)
	Ask(OutgoingMessage) (string, Message, error)
	Conversation(Conversation) ([]string, []Message, error)
	Say(OutgoingMessage) (string, error)
	Form(Form) (string, error)
	Image(Image) (string, error)
	UpdateMessage(string, string) (string, error)
	UpdateForm(string, Form) (string, error)
	WaitForAction(string, ActionWaitingPolicy) (Action, error)
	AskUntil(OutgoingMessage, AnswerChecker) (string, Message, error)
}
