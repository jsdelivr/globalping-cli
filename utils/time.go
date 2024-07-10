package utils

import (
	"math"
	_time "time"
)

type Time interface {
	Now() _time.Time
}

type time struct{}

func NewTime() Time {
	return &time{}
}

func (d *time) Now() _time.Time {
	return _time.Now()
}

func FormatSeconds(seconds int64) string {
	if seconds < 60 {
		return Pluralize(seconds, "second")
	}
	if seconds < 3600 {
		return Pluralize(int64(math.Round(float64(seconds)/60)), "minute")
	}
	if seconds < 86400 {
		return Pluralize(int64(math.Round(float64(seconds)/3600)), "hour")
	}
	return Pluralize(int64(math.Round(float64(seconds)/86400)), "day")
}
