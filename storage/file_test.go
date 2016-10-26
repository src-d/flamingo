package storage

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/src-d/flamingo"
	"github.com/stretchr/testify/require"
)

func TestFileStorage(t *testing.T) {
	require := require.New(t)
	storage, err := NewFile("./foo.json")
	require.Nil(err)
	RunStorageTest(storage, t)

	storage, err = NewFile("./foo.json")
	ok, err := storage.BotExists(flamingo.StoredBot{ID: "1"})
	require.Nil(err)
	require.True(ok)

	ok, err = storage.ConversationExists(flamingo.StoredConversation{
		ID: "2", BotID: "1",
	})
	require.Nil(err)
	require.True(ok)

	bots, err := storage.LoadBots()
	require.Nil(err)
	require.Equal(2, len(bots))

	convs, err := storage.LoadConversations(flamingo.StoredBot{ID: "1"})
	require.Nil(err)
	require.Equal(2, len(convs))

	convs, err = storage.LoadConversations(flamingo.StoredBot{ID: "2"})
	require.Nil(err)
	require.Equal(0, len(convs))

	require.Nil(os.Remove("./foo.json"))
}

func TestFileStorageNewFileOpenFail(t *testing.T) {
	_, err := NewFile("/")
	require.NotNil(t, err)
}

func TestFileStorageNewFileUnmarshalFail(t *testing.T) {
	require := require.New(t)
	f, err := ioutil.TempFile("", "unmarshal_error")
	require.Nil(err)
	_, err = f.WriteString("some_garbage")
	require.Nil(err)
	_, err = NewFile(f.Name())
	require.NotNil(err)

	require.Nil(os.Remove(f.Name()))
}

func TestFileStorageSaveMarshalFail(t *testing.T) {
	require := require.New(t)
	storage, err := NewFile("./foo.json")
	require.Nil(err)

	bot := flamingo.StoredBot{Extra: func() {}}
	require.NotNil(storage.StoreBot(bot))
}

func TestFileStorageSaveRemoveFileFail(t *testing.T) {
	storage := fileStorage{file: "/etc/passwd", data: newBotStorage()}

	bot := flamingo.StoredBot{ID: "1"}
	require.NotNil(t, storage.StoreBot(bot))
}
