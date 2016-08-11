package flamingo

type Action struct {
	UserAction      UserAction
	User            User
	Channel         Channel
	OriginalMessage Message
	Extra           interface{}
}

type UserAction struct {
	Name  string
	Value string
}

type ActionWaitingPolicy struct {
	Reply   bool
	Message string
}

func IgnorePolicy() ActionWaitingPolicy {
	return ActionWaitingPolicy{}
}

func ReplyPolicy(msg string) ActionWaitingPolicy {
	return ActionWaitingPolicy{true, msg}
}

type ActionHandler func(Bot, Action)
