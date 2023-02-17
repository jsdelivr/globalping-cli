package model

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

// https://stackoverflow.com/questions/50092462/how-to-unmarshal-an-inconsistent-json-field-that-can-be-a-string-or-an-array-o
type ResultData struct {
	Status           string                   `json:"status"`
	RawOutput        string                   `json:"rawOutput"`
	ResolvedAddress  string                   `json:"resolvedAddress"`
	ResolvedHostname string                   `json:"resolvedHostname"`
	Timings          []map[string]interface{} `json:"timings,omitempty"`
	Stats            map[string]interface{}   `json:"stats,omitempty"`
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
