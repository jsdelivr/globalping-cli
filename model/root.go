package model

import (
	"math"
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

	Area            *pterm.AreaPrinter
	Hostname        string
	CompletedStats  []MeasurementStats
	InProgressStats []MeasurementStats
	CallCount       int      // Number of measurements created
	MaxHistory      int      // Maximum number of measurements to keep in history
	History         *Rbuffer // History of measurements
}

type MeasurementStats struct {
	Sent int     // Number of packets sent
	Rcv  int     // Number of packets received
	Lost int     // Number of packets lost
	Loss float64 // Percentage of packets lost
	Last float64 // Last RTT
	Min  float64 // Minimum RTT
	Avg  float64 // Average RTT
	Max  float64 // Maximum RTT
	Mdev float64 // Mean deviation of RTT
	Time float64 // Total time
}

func NewMeasurementStats() MeasurementStats {
	return MeasurementStats{Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1}
}

type Rbuffer struct {
	Index int
	Slice []string
}

func (q *Rbuffer) Push(id string) {
	q.Slice[q.Index] = id
	q.Index = (q.Index + 1) % len(q.Slice)
}

func (q *Rbuffer) ToString(sep string) string {
	s := ""
	i := q.Index
	isFirst := true
	for {
		if q.Slice[i] != "" {
			if isFirst {
				isFirst = false
				s += q.Slice[i]
			} else {
				s += sep + q.Slice[i]
			}
		}
		i = (i + 1) % len(q.Slice)
		if i == q.Index {
			break
		}
	}
	return s
}

func NewRbuffer(size int) *Rbuffer {
	return &Rbuffer{
		Index: 0,
		Slice: make([]string, size),
	}
}
