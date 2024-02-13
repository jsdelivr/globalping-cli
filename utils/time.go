package utils

import _time "time"

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
