package view

import (
	"math"
	"time"
)

type Context struct {
	Cmd       string
	Target    string
	From      string
	Limit     int  // Number of probes to use
	CIMode    bool // Determine whether the output should be in a format that is easy to parse by a CI tool
	ToJSON    bool // Determines whether the output should be in JSON format.
	ToLatency bool // Determines whether the output should be only the stats of a measurement
	Share     bool // Display share message

	Packets   int // Number of packets to send
	Port      int
	Protocol  string
	Resolver  string
	QueryType string
	Host      string
	Path      string
	Query     string
	Method    string
	Headers   []string
	Trace     bool
	Full      bool // Full output
	Infinite  bool // Infinite flag
	Ipv6      bool // IPv6 flag
	Ipv4      bool // IPv4 flag

	Head uint // Number of first measurements to show
	Tail uint // Number of last measurements to show

	APIMinInterval time.Duration // Minimum interval between API calls

	IsLocationFromSession bool // Determine whether the previous location is used
	RecordToSession       bool // Record measurement to session history

	Hostname            string
	IsHeaderPrinted     bool
	AggregatedStats     []*MeasurementStats
	MeasurementsCreated int
	History             *HistoryBuffer // History of measurements
}

type MeasurementStats struct {
	Sent  int     // Number of packets sent
	Rcv   int     // Number of packets received
	Lost  int     // Number of packets lost
	Loss  float64 // Percentage of packets lost
	Last  float64 // Last RTT
	Min   float64 // Minimum RTT
	Avg   float64 // Average RTT
	Max   float64 // Maximum RTT
	Mdev  float64 // Mean deviation of RTT
	Time  float64 // Total time of measurement, in milliseconds
	Tsum  float64 // Total sum of RTT
	Tsum2 float64 // Total sum of RTT squared
}

func NewMeasurementStats() *MeasurementStats {
	return &MeasurementStats{Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1}
}
