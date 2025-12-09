package api

import (
	"testing"
	"time"

	"github.com/jsdelivr/globalping-cli/storage"
	"github.com/jsdelivr/globalping-cli/utils"
)

var (
	defaultCurrentTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
)

func getTokenJSON() []byte {
	return []byte(`{
"access_token":"token",
"token_type":"bearer",
"refresh_token":"refresh",
"expires_in": 3600
}`)
}

func createDefaultTestStorage(t *testing.T, utils utils.Utils) *storage.LocalStorage {
	s := storage.NewLocalStorage(utils)
	err := s.Init("globalping-cli_" + t.Name())
	if err != nil {
		panic(err)
	}
	t.Cleanup(func() {
		s.Remove()
	})
	return s
}
