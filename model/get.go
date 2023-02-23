package model

import "encoding/json"

// Modeled from https://github.com/jsdelivr/globalping/blob/master/docs/measurement/get.md

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
	Status           string                 `json:"status"`
	RawOutput        string                 `json:"rawOutput"`
	ResolvedAddress  string                 `json:"resolvedAddress"`
	ResolvedHostname string                 `json:"resolvedHostname"`
	Stats            map[string]interface{} `json:"stats,omitempty"`
	TimingsRaw       json.RawMessage        `json:"timings,omitempty"`
}

type Timings struct {
	Arr       []map[string]interface{}
	Interface map[string]interface{}
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
	ProbesCount int                   `json:"probesCount"`
	Results     []MeasurementResponse `json:"results"`
}
