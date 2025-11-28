package utils

import (
	"os"
	_time "time"
)

var ShareURL = "https://globalping.io?measurement="

type Config struct {
	GlobalpingToken            string
	GlobalpingAuthClientID     string
	GlobalpingAuthClientSecret string
	GlobalpingAPIInterval      _time.Duration
}

func NewConfig() *Config {
	return &Config{
		GlobalpingAuthClientID:     "be231712-03f4-45bf-9f15-023506ce0b72",
		GlobalpingAuthClientSecret: "public",
		GlobalpingAPIInterval:      500 * _time.Millisecond,
	}
}

func (c *Config) Load() {
	c.GlobalpingToken = os.Getenv("GLOBALPING_TOKEN")
}
