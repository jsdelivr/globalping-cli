package model

// Modeled from https://github.com/jsdelivr/globalping/blob/master/docs/measurement/get.md

// Nested structs
type MeasurementResponse struct {
	Result struct {
		RawOutput string `json:"rawOutput"`
	} `json:"result"`
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
