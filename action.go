package flamingo

// Action is a message received by the application to
// notify the user did something.
type Action struct {
	// UserAction is the action performed by the user.
	UserAction UserAction
	// User is the user that performed the action.
	User User
	// Channel is the channel in which the action was performed.
	Channel Channel
	// OriginalMessage is the message the action originated from.
	OriginalMessage Message
	// Extra parameters. The contents of this field depend on the
	// specific client you are using. Usually, it is the full
	// data received.
	Extra interface{}
}

// UserAction is the kind of action performed by the user.
type UserAction struct {
	// Name of the action performed.
	Name string
	// Value of the action performed.
	Value string
}

// ActionWaitingPolicy defines the policy to follow while waiting for
// actions to arrive.
type ActionWaitingPolicy struct {
	// Reply will send a reply every time an action or message arrives and
	// it's not the one we're waiting for.
	Reply bool
	// Message is the text of the reply message. Only necessary if Reply is true.
	Message string
}

// IgnorePolicy returns an ActionWaitingPolicy that ignores all incoming
// messages and actions but the one we're looking for.
func IgnorePolicy() ActionWaitingPolicy {
	return ActionWaitingPolicy{}
}

// ReplyPolicy returns an ActionWaitingPolicy that replies to every message or
// action received but the on we're looking for.
func ReplyPolicy(msg string) ActionWaitingPolicy {
	return ActionWaitingPolicy{true, msg}
}

// ActionHandler is a function that will handle a certain kind of Action.
type ActionHandler func(Bot, Action)
