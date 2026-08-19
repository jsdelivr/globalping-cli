package storage

import (
	"testing"
)

func createDefaultTestStorage(t *testing.T) *LocalStorage {
	s := NewLocalStorage(nil)
	err := s.Init("globalping-cli_" + t.Name())

	if err != nil {
		panic(err)
	}

	t.Cleanup(func() {
		if err := s.Remove(); err != nil {
			t.Error(err)
		}
	})

	return s
}
