package model

// Modeled from https://github.com/jsdelivr/globalping/blob/master/docs/measurement/get.md

type ProbeData struct {
	Continent string   `json:"continent"`
	Country   string   `json:"country"`
	City      string   `json:"city"`
	State     string   `json:"state,omitempty"`
	ASN       int      `json:"asn"`
	Network   string   `json:"network,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}

// Nested structs
type MeasurementResponse struct {
	Probe  ProbeData              `json:"probe"`
	Result map[string]interface{} `json:"result"` // This is too dynamic depending on the type of measurement
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
