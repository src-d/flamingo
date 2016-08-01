package flamingo

// Handler is a plugin for the engine. If the handler finds a match
// in the text and channel of the message it will be handled.
type Handler interface {
	// Handle will perform operations given a message.
	Handle(Message)
	// IsMatch indicates if this handler should handle the message with the given channel and text.
	IsMatch(string, string) bool
	// Name returns the name of the handler.
	Name() string
	// Help returns a help message for the user, which can contain markdown.
	Help() string
}
