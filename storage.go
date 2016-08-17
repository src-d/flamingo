package flamingo

import "time"

// StoredBot is the minimal data snapshot to start a previously
// running bot instance. It is meant to be stored.
type StoredBot struct {
	// ID is the ID of the bot, an unique identifier.
	ID string
	// Token is the token needed to access the bot API by the client.
	Token string
	// CreatedAt is the time it was first saved.
	CreatedAt time.Time
	// Extra allows clients to store client-specific data needed.
	Extra interface{}
}

// StoredConversation is the minimal data snapshot to start a previously
// running bot conversation instance. It is meant to be stored.
type StoredConversation struct {
	// ID is the unique identifier of the conversation or channel.
	ID string
	// BotID is the unique identifier of the bot this conversation belongs to.
	BotID string
	// CreatedAt is the time it was first saved.
	CreatedAt time.Time
	// Extra allows clients to store client-specific data needed.
	Extra interface{}
}

// Storage is a service to store and retrieve conversations and bots stored.
type Storage interface {
	// StoreBot saves the given bot.
	StoreBot(StoredBot) error
	// StoreConversation saves the given conversation.
	StoreConversation(StoredConversation) error
	// LoadBots retrieves all stored bots.
	LoadBots() ([]StoredBot, error)
	// LoadConversations retrieves all stored conversations for a single bot.
	LoadConversations(StoredBot) ([]StoredConversation, error)
	// BotExists checks if the bot is already stored.
	BotExists(StoredBot) (bool, error)
	// ConversationExists checks if the conversation is already stored.
	ConversationExists(StoredConversation) (bool, error)
}
