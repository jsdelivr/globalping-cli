package model

import "encoding/json"

// Modeled from https://www.jsdelivr.com/docs/api.globalping.io

type ProbeData struct {
	Continent string   `json:"continent"`
	Region    string   `json:"region"`
	Country   string   `json:"country"`
	City      string   `json:"city"`
	State     string   `json:"state,omitempty"`
	ASN       int      `json:"asn"`
	Network   string   `json:"network,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}

type ResultData struct {
	Status           string          `json:"status"`
	RawOutput        string          `json:"rawOutput"`
	RawHeaders       string          `json:"rawHeaders"`
	RawBody          string          `json:"rawBody"`
	ResolvedAddress  string          `json:"resolvedAddress"`
	ResolvedHostname string          `json:"resolvedHostname"`
	StatsRaw         json.RawMessage `json:"stats,omitempty"`
	TimingsRaw       json.RawMessage `json:"timings,omitempty"`
}

type PingStats struct {
	Min   float64 `json:"min"`   // The lowest rtt value.
	Avg   float64 `json:"avg"`   // The average rtt value.
	Max   float64 `json:"max"`   // The highest rtt value.
	Total int     `json:"total"` // The number of sent packets.
	Rcv   int     `json:"rcv"`   // The number of received packets.
	Drop  int     `json:"drop"`  // The number of dropped packets (total - rcv).
	Loss  float64 `json:"loss"`  // The percentage of dropped packets.
}

type PingTiming struct {
	RTT float64 `json:"rtt"` // The round-trip time for this packet.
	TTL int     `json:"ttl"` // The packet time-to-live value.
}

type DNSTimings struct {
	Total float64 `json:"total"` // The total query time in milliseconds.
}

type HTTPTimings struct {
	Total     int `json:"total"`     // The total HTTP request time
	DNS       int `json:"dns"`       // The time required to perform the DNS lookup.
	TCP       int `json:"tcp"`       // The time from performing the DNS lookup to establishing the TCP connection.
	TLS       int `json:"tls"`       // The time from establishing the TCP connection to establishing the TLS session.
	FirstByte int `json:"firstByte"` // The time from establishing the TCP/TLS connection to the first response byte.
	Download  int `json:"download"`  // The time from the first byte to downloading the whole response.
}

// Nested structs
type MeasurementResponse struct {
	Probe  ProbeData  `json:"probe"`
	Result ResultData `json:"result"`
}

// Main struct
type GetMeasurement struct {
	ID          string                `json:"id"`
	Type        string                `json:"type"`
	Status      string                `json:"status"`
	CreatedAt   string                `json:"createdAt"`
	UpdatedAt   string                `json:"updatedAt"`
	Target      string                `json:"target"`
	ProbesCount int                   `json:"probesCount"`
	Results     []MeasurementResponse `json:"results"`
}
