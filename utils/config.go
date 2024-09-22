package utils

import (
	"os"
	_time "time"
)

type Config struct {
	GlobalpingToken            string
	GlobalpingAPIURL           string
	GlobalpingAuthURL          string
	GlobalpingDashboardURL     string
	GlobalpingAuthClientID     string
	GlobalpingAuthClientSecret string
	GlobalpingAPIInterval      _time.Duration
}

func NewConfig() *Config {
	return &Config{
		GlobalpingAPIURL:           "https://api.globalping.io/v1",
		GlobalpingAuthURL:          "https://auth.globalping.io",
		GlobalpingDashboardURL:     "https://dash.globalping.io",
		GlobalpingAuthClientID:     "be231712-03f4-45bf-9f15-023506ce0b72",
		GlobalpingAuthClientSecret: "public",
		GlobalpingAPIInterval:      500 * _time.Millisecond,
	}
}

func (c *Config) Load() {
	c.GlobalpingToken = os.Getenv("GLOBALPING_TOKEN")
}
