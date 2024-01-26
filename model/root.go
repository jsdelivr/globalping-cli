package model

import (
	"time"

	"github.com/pterm/pterm"
)

// Used in thc client TUI
type Context struct {
	Cmd      string
	Target   string
	From     string
	Resolver string

	Limit   int
	Packets int // Number of packets to send

	JsonOutput bool // JsonOutput is a flag that determines whether the output should be in JSON format.
	Latency    bool // Latency is a flag that outputs only stats of a measurement
	CI         bool // CI flag is used to determine whether the output should be in a format that is easy to parse by a CI tool
	Full       bool // Full output
	Share      bool // Display share message
	Infinite   bool // Infinite flag

	APIMinInterval time.Duration // Minimum interval between API calls
	Area           *pterm.AreaPrinter
	Stats          []MeasurementStats
}

type MeasurementStats struct {
	Sent int     // Number of packets sent
	Lost int     // Number of packets lost
	Loss float64 // Percentage of packets lost
	Last float64 // Last RTT
	Min  float64 // Minimum RTT
	Avg  float64 // Average RTT
	Max  float64 // Maximum RTT
}
