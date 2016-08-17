package storage

import (
	"sync"

	"github.com/mvader/flamingo"
)

type memoryStorage struct {
	sync.RWMutex
	bots          map[string]*flamingo.StoredBot
	convs         map[string][]*flamingo.StoredConversation
	existingConvs map[string]struct{}
}

// NewMemory creates a new in-memory storage for bots and conversations.
func NewMemory() flamingo.Storage {
	return &memoryStorage{
		bots:          make(map[string]*flamingo.StoredBot),
		convs:         make(map[string][]*flamingo.StoredConversation),
		existingConvs: make(map[string]struct{}),
	}
}

func (s *memoryStorage) StoreBot(bot flamingo.StoredBot) error {
	s.Lock()
	defer s.Unlock()
	s.bots[bot.ID] = &bot
	return nil
}

func (s *memoryStorage) StoreConversation(conv flamingo.StoredConversation) error {
	s.Lock()
	defer s.Unlock()
	s.convs[conv.BotID] = append(s.convs[conv.BotID], &conv)
	s.existingConvs[conv.ID] = struct{}{}
	return nil
}

func (s *memoryStorage) LoadBots() ([]flamingo.StoredBot, error) {
	s.Lock()
	defer s.Unlock()
	var bots []flamingo.StoredBot
	for _, b := range s.bots {
		bots = append(bots, *b)
	}
	return bots, nil
}

func (s *memoryStorage) LoadConversations(bot flamingo.StoredBot) ([]flamingo.StoredConversation, error) {
	s.Lock()
	defer s.Unlock()
	var convs []flamingo.StoredConversation
	for _, c := range s.convs[bot.ID] {
		convs = append(convs, *c)
	}
	return convs, nil
}

func (s *memoryStorage) BotExists(bot flamingo.StoredBot) (bool, error) {
	s.Lock()
	defer s.Unlock()
	_, ok := s.bots[bot.ID]
	return ok, nil
}

func (s *memoryStorage) ConversationExists(conv flamingo.StoredConversation) (bool, error) {
	s.Lock()
	defer s.Unlock()
	_, ok := s.existingConvs[conv.ID]
	return ok, nil
}
