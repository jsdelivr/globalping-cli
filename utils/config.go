package utils

import (
	"os"
	_time "time"
)

type Config struct {
	GlobalpingToken       string
	GlobalpingAPIURL      string
	GlobalpingAPIInterval _time.Duration
}

func NewConfig() *Config {
	return &Config{
		GlobalpingAPIURL:      "https://api.globalping.io/v1",
		GlobalpingAPIInterval: 500 * _time.Millisecond,
	}
}

func (c *Config) Load() {
	c.GlobalpingToken = os.Getenv("GLOBALPING_TOKEN")
}
