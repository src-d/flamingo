package storage

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/src-d/flamingo"
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

func TestFileStorageNewFileOpenFail(t *testing.T) {
	_, err := NewFile("/")
	assert.NotNil(t, err)
}

func TestFileStorageNewFileUnmarshalFail(t *testing.T) {
	assert := assert.New(t)
	f, err := ioutil.TempFile("", "unmarshal_error")
	assert.Nil(err)
	_, err = f.WriteString("some_garbage")
	assert.Nil(err)
	_, err = NewFile(f.Name())
	assert.NotNil(err)

	assert.Nil(os.Remove(f.Name()))
}

func TestFileStorageSaveMarshalFail(t *testing.T) {
	assert := assert.New(t)
	storage, err := NewFile("./foo.json")
	assert.Nil(err)

	bot := flamingo.StoredBot{Extra: func() {}}
	assert.NotNil(storage.StoreBot(bot))
}

func TestFileStorageSaveRemoveFileFail(t *testing.T) {
	storage := fileStorage{file: "/etc/passwd", data: newBotStorage()}

	bot := flamingo.StoredBot{ID: "1"}
	assert.NotNil(t, storage.StoreBot(bot))
}
