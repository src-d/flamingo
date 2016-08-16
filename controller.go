package flamingo

type Controller interface {
	CanHandle(Message) bool
	Handle(Bot, Message) error
}

type IntroHandler interface {
	HandleIntro(Bot, Channel) error
}
