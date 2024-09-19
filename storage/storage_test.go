package storage

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/stretchr/testify/assert"
)

func Test_Config(t *testing.T) {
	_storage := NewLocalStorage(".test_globalping-cli")
	defer _storage.Remove()
	err := _storage.Init()
	if err != nil {
		t.Fatal(err)
	}
	config, err := _storage.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, &Config{
		Profile:  "default",
		Profiles: make(map[string]*Profile),
	}, config)

	profile := _storage.GetProfile()
	profile.Token = &globalping.Token{
		AccessToken:  "token",
		RefreshToken: "refresh",
		TokenType:    "bearer",
		Expiry:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	err = _storage.SaveConfig()
	if err != nil {
		t.Fatal(err)
	}
	path, err := _storage.joinHomeDir(_storage.configName)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	c := &Config{}
	err = json.Unmarshal(b, c)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, &Config{
		Profile: "default",
		Profiles: map[string]*Profile{
			"default": {Token: profile.Token},
		},
	}, c)
}