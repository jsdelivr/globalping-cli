package utils

import (
	"errors"
	"math"
	"os/exec"
	"runtime"
	_time "time"
)

type Utils interface {
	Now() _time.Time
	OpenBrowser(url string) error
}

type utils struct{}

func NewUtils() Utils {
	return &utils{}
}

func (u *utils) Now() _time.Time {
	return _time.Now()
}

func (u *utils) OpenBrowser(url string) error {
	switch runtime.GOOS {
	case "linux":
		// WSL workaround
		err := exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", url).Start()
		if err != nil {
			return exec.Command("xdg-open", url).Start()
		}
		return nil
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return errors.New("unsupported platform")
	}
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
