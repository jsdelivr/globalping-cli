package storage

import (
	"testing"

	"github.com/jsdelivr/globalping-cli/utils"
)

func createDefaultTestStorage(t *testing.T, utils utils.Utils) *LocalStorage {
	s := NewLocalStorage(utils)
	err := s.Init("globalping-cli_" + t.Name())
	if err != nil {
		panic(err)
	}
	t.Cleanup(func() {
		s.Remove()
	})
	return s
}
