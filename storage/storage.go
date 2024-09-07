package storage

import (
	"encoding/json"
	"os"
	"path"

	"github.com/jsdelivr/globalping-cli/globalping"
)

type LocalStorage struct {
	name       string
	configName string
	config     *Config
}

func NewLocalStorage(name string) *LocalStorage {
	return &LocalStorage{
		name:       name,
		configName: "config.json",
	}
}

func (s *LocalStorage) Init() error {
	homeDir, err := s.joinHomeDir("")
	if err != nil {
		return err
	}
	err = os.MkdirAll(homeDir, 0755)
	if err != nil {
		return err
	}
	_, err = s.LoadConfig()
	if err != nil {
		if os.IsNotExist(err) {
			s.config = &Config{
				Profile:  "default",
				Profiles: make(map[string]*Profile),
			}
			s.SaveConfig()
		}
	}
	return nil
}

type Profile struct {
	Token *globalping.Token `json:"token"`
}

type Config struct {
	Profile  string              `json:"profile"`
	Profiles map[string]*Profile `json:"profiles"`
}

func (s *LocalStorage) LoadConfig() (*Config, error) {
	if s.config != nil {
		return s.config, nil
	}
	path, err := s.joinHomeDir(s.configName)
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
	path, err := s.joinHomeDir(s.configName)
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

func (s *LocalStorage) joinHomeDir(name string) (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(dir, s.name, name), nil
}
