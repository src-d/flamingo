package storage

import (
	"testing"

	"github.com/src-d/flamingo"
	"github.com/stretchr/testify/assert"
)

func RunStorageTest(storage flamingo.Storage, t *testing.T) {
	assert := assert.New(t)
	ok, err := storage.BotExists(flamingo.StoredBot{ID: "1"})
	assert.Nil(err)
	assert.False(ok)

	ok, err = storage.ConversationExists(flamingo.StoredConversation{
		ID: "2", BotID: "1",
	})
	assert.Nil(err)
	assert.False(ok)

	assert.Nil(storage.StoreBot(flamingo.StoredBot{ID: "1"}))
	assert.Nil(storage.StoreConversation(flamingo.StoredConversation{
		ID:    "2",
		BotID: "1",
	}))

	ok, err = storage.BotExists(flamingo.StoredBot{ID: "1"})
	assert.Nil(err)
	assert.True(ok)

	ok, err = storage.ConversationExists(flamingo.StoredConversation{
		ID: "2", BotID: "1",
	})
	assert.Nil(err)
	assert.True(ok)

	assert.Nil(storage.StoreBot(flamingo.StoredBot{ID: "4"}))
	assert.Nil(storage.StoreConversation(flamingo.StoredConversation{
		ID:    "3",
		BotID: "1",
	}))

	bots, err := storage.LoadBots()
	assert.Nil(err)
	assert.Equal(2, len(bots))

	convs, err := storage.LoadConversations(flamingo.StoredBot{ID: "1"})
	assert.Nil(err)
	assert.Equal(2, len(convs))

	convs, err = storage.LoadConversations(flamingo.StoredBot{ID: "2"})
	assert.Nil(err)
	assert.Equal(0, len(convs))
}
