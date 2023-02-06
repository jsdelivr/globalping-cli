package model

// Modeled from https://github.com/jsdelivr/globalping/blob/master/docs/measurement/get.md

type ProbeData struct {
	Continent string `json:"continent"`
	Country   string `json:"country"`
	City      string `json:"city"`
	State     string `json:"state,omitempty"`
	ASN       int    `json:"asn"`
}

type Timings struct {
	TTL int `json:"ttl,omitempty"`
	RTT int `json:"rtt,omitempty"`
}

type Stats struct {
	Min  float32 `json:"min"`
	Max  float32 `json:"max"`
	Avg  float32 `json:"avg"`
	Loss float32 `json:"loss"`
}

type ResultData struct {
	RawOutput string `json:"rawOutput"`
}

// Nested structs
type MeasurementResponse struct {
	Probe  ProbeData  `json:"probe"`
	Result ResultData `json:"result"`
}

// Main struct
type GetMeasurement struct {
	ID        string                `json:"id"`
	Type      string                `json:"type"`
	Status    string                `json:"status"`
	CreatedAt string                `json:"createdAt"`
	UpdatedAt string                `json:"updatedAt"`
	Results   []MeasurementResponse `json:"results"`
}
