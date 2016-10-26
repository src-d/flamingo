package storage

import (
	"testing"

	"github.com/src-d/flamingo"
	"github.com/stretchr/testify/require"
)

func RunStorageTest(storage flamingo.Storage, t *testing.T) {
	require := require.New(t)
	ok, err := storage.BotExists(flamingo.StoredBot{ID: "1"})
	require.Nil(err)
	require.False(ok)

	ok, err = storage.ConversationExists(flamingo.StoredConversation{
		ID: "2", BotID: "1",
	})
	require.Nil(err)
	require.False(ok)

	require.Nil(storage.StoreBot(flamingo.StoredBot{ID: "1"}))
	require.Nil(storage.StoreConversation(flamingo.StoredConversation{
		ID:    "2",
		BotID: "1",
	}))

	ok, err = storage.BotExists(flamingo.StoredBot{ID: "1"})
	require.Nil(err)
	require.True(ok)

	ok, err = storage.ConversationExists(flamingo.StoredConversation{
		ID: "2", BotID: "1",
	})
	require.Nil(err)
	require.True(ok)

	require.Nil(storage.StoreBot(flamingo.StoredBot{ID: "4"}))
	require.Nil(storage.StoreConversation(flamingo.StoredConversation{
		ID:    "3",
		BotID: "1",
	}))

	bots, err := storage.LoadBots()
	require.Nil(err)
	require.Equal(2, len(bots))

	convs, err := storage.LoadConversations(flamingo.StoredBot{ID: "1"})
	require.Nil(err)
	require.Equal(2, len(convs))

	convs, err = storage.LoadConversations(flamingo.StoredBot{ID: "2"})
	require.Nil(err)
	require.Equal(0, len(convs))
}
