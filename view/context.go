package view

import "time"

type Context struct {
	Cmd        string
	Target     string
	Targets    []string
	Comparison bool
	From       string
	Limit      int  // Number of probes to use
	Timeout    int  // Probe-side measurement timeout in seconds
	CIMode     bool // Determine whether the output should be in a format that is easy to parse by a CI tool
	ToJSON     bool // Determines whether the output should be in JSON format.
	ToLatency  bool // Determines whether the output should be only the stats of a measurement
	Table      bool // Display measurement results in a table
	Share      bool // Display share message

	Packets   int // Number of packets to send
	Port      uint16
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

	TableOutputRows     int
	MeasurementsCreated int
	History             *HistoryBuffer // History of measurements
}
