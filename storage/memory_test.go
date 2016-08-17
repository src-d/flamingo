package storage

import "testing"

func TestMemoryStorage(t *testing.T) {
	RunStorageTest(NewMemory(), t)
}
