package slack

import "github.com/mvader/flamingo"

type bot struct {
	ID string
}

func newBot() flamingo.Bot {
	return &bot{}
}

func (b *bot) ID() string {
	return b.ID
}

func (b *bot) Reply(msg flamingo.OutgoingMessage) error {
	return nil
}

func (b *bot) Ask(msg flamingo.OutgoingMessage) (flamingo.Message, error) {
	return nil, nil
}

func (b *bot) Conversation(convo flamingo.Conversation) ([]flamingo.Message, error) {
	return nil, nil
}

func (b *bot) Say(msg flamingo.OutgoingMessage) error {
	return nil
}
