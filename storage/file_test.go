package storage

import (
	"os"
	"testing"

	"github.com/mvader/flamingo"
	"github.com/stretchr/testify/assert"
)

func TestFileStorage(t *testing.T) {
	assert := assert.New(t)
	storage, err := NewFile("./foo.json")
	assert.Nil(err)
	RunStorageTest(storage, t)

	storage, err = NewFile("./foo.json")
	ok, err := storage.BotExists(flamingo.StoredBot{ID: "1"})
	assert.Nil(err)
	assert.True(ok)

	ok, err = storage.ConversationExists(flamingo.StoredConversation{
		ID: "2", BotID: "1",
	})
	assert.Nil(err)
	assert.True(ok)

	bots, err := storage.LoadBots()
	assert.Nil(err)
	assert.Equal(2, len(bots))

	convs, err := storage.LoadConversations(flamingo.StoredBot{ID: "1"})
	assert.Nil(err)
	assert.Equal(2, len(convs))

	convs, err = storage.LoadConversations(flamingo.StoredBot{ID: "2"})
	assert.Nil(err)
	assert.Equal(0, len(convs))

	assert.Nil(os.Remove("./foo.json"))
}
