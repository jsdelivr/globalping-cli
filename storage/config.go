package storage

import (
	"encoding/json"
	"os"

	"github.com/jsdelivr/globalping-cli/globalping"
)

type Profile struct {
	Token *globalping.Token `json:"token"`
}

type Config struct {
	Profile       string              `json:"profile"`
	Profiles      map[string]*Profile `json:"profiles"`
	LastMigration int                 `json:"last_migration"`
}

func (s *LocalStorage) LoadConfig() (*Config, error) {
	if s.config != nil {
		return s.config, nil
	}
	path, err := s.joinConfigDir(s.configName)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	s.config = &Config{
		Profile:  "default",
		Profiles: make(map[string]*Profile),
	}
	err = json.Unmarshal(b, s.config)
	if err != nil {
		return nil, err
	}
	return s.config, nil
}

func (s *LocalStorage) SaveConfig() error {
	if s.config == nil {
		return nil
	}
	path, err := s.joinConfigDir(s.configName)
	if err != nil {
		return err
	}
	b, err := json.Marshal(s.config)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

func (s *LocalStorage) GetProfile() *Profile {
	p := s.config.Profiles[s.config.Profile]
	if p == nil {
		p = &Profile{}
		s.config.Profiles[s.config.Profile] = p
	}
	return p
}
