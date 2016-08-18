package storage

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"

	"github.com/src-d/flamingo"
)

type botStorage struct {
	Bots                  map[string]flamingo.StoredBot
	Conversations         map[string][]flamingo.StoredConversation
	ExistingConversations map[string]bool
}

type fileStorage struct {
	sync.RWMutex
	file string
	data botStorage
}

// NewFile creates a new storage that will be saved to a disk file.
// This storage truncates the file every time it saves. Be sure to not
// use multiple instances of a file storage on the same file. Instead, reuse
// the same Storage across multiple clients.
// Read operations are very cheap, since they are done in-memory and kept up to
// date with the writes. Write operations are slower, though, because every
// time a save operation is performed, it will truncate the file and write all
// again.
func NewFile(file string) (flamingo.Storage, error) {
	storage := &fileStorage{file: file}
	if err := storage.load(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *fileStorage) load() error {
	s.Lock()
	defer s.Unlock()
	bytes, err := ioutil.ReadFile(s.file)
	if os.IsNotExist(err) {
		s.data = botStorage{
			Bots:                  make(map[string]flamingo.StoredBot),
			Conversations:         make(map[string][]flamingo.StoredConversation),
			ExistingConversations: make(map[string]bool),
		}
		return nil
	} else if err != nil {
		return err
	}

	var storage botStorage
	if err := json.Unmarshal(bytes, &storage); err != nil {
		return err
	}

	s.data = storage
	return nil
}

func (s *fileStorage) save() error {
	bytes, err := json.Marshal(s.data)
	if err != nil {
		return err
	}

	if err := os.Remove(s.file); err != nil && !os.IsNotExist(err) {
		return err
	}

	return ioutil.WriteFile(s.file, bytes, 0777)
}

func (s *fileStorage) StoreBot(bot flamingo.StoredBot) error {
	s.Lock()
	defer s.Unlock()
	s.data.Bots[bot.ID] = bot
	return s.save()
}

func (s *fileStorage) StoreConversation(conv flamingo.StoredConversation) error {
	s.Lock()
	defer s.Unlock()
	s.data.Conversations[conv.BotID] = append(s.data.Conversations[conv.BotID], conv)
	s.data.ExistingConversations[conv.ID] = true
	return s.save()
}

func (s *fileStorage) LoadBots() ([]flamingo.StoredBot, error) {
	s.Lock()
	defer s.Unlock()
	var bots []flamingo.StoredBot
	for _, b := range s.data.Bots {
		bots = append(bots, b)
	}
	return bots, nil
}

func (s *fileStorage) LoadConversations(bot flamingo.StoredBot) ([]flamingo.StoredConversation, error) {
	s.Lock()
	defer s.Unlock()
	return s.data.Conversations[bot.ID], nil
}

func (s *fileStorage) BotExists(bot flamingo.StoredBot) (bool, error) {
	s.Lock()
	defer s.Unlock()
	_, ok := s.data.Bots[bot.ID]
	return ok, nil
}

func (s *fileStorage) ConversationExists(conv flamingo.StoredConversation) (bool, error) {
	s.Lock()
	defer s.Unlock()
	_, ok := s.data.ExistingConversations[conv.ID]
	return ok, nil
}
