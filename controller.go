package flamingo

// Controller is an object that will handle incoming messages
// if they can be handled by it.
type Controller interface {
	// CanHandle tells if the controller can handle the message.
	CanHandle(Message) bool

	// Handle performs the actual work, here an instance of Bot is
	// provided along with the Message to be able to communicate with
	// the user.
	Handle(Bot, Message) error
}

// IntroHandler is a handler that is triggered whenever someone
// starts a new conversation or group with the bot.
type IntroHandler interface {
	// HandleIntro performs the actual work, here an instance of Bot
	// is provided along with the Channel to be able to communicate
	// with the users in the channel.
	HandleIntro(Bot, Channel) error
}
