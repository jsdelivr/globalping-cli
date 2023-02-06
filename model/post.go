package model

// Modeled from https://github.com/jsdelivr/globalping/blob/master/docs/measurement/post-create.md

// Nested structs
type Locations struct {
	Magic string `json:"magic"`
}

type QueryOptions struct {
	Type string `json:"type,omitempty"`
}

type RequestOptions struct {
	Headers map[string]string `json:"headers,omitempty"`
	Path    string            `json:"path,omitempty"`
	Host    string            `json:"host,omitempty"`
	Query   string            `json:"query,omitempty"`
	Method  string            `json:"method,omitempty"`
}

type MeasurementOptions struct {
	Query    *QueryOptions   `json:"query,omitempty"`
	Request  *RequestOptions `json:"request,omitempty"`
	Protocol string          `json:"protocol,omitempty"`
	Port     int             `json:"port,omitempty"`
	Resolver string          `json:"resolver,omitempty"`
	Trace    bool            `json:"trace,omitempty"`
	Packets  int             `json:"packets,omitempty"`
}

// Main struct
type PostMeasurement struct {
	Limit     int                 `json:"limit"`
	Locations []Locations         `json:"locations"`
	Type      string              `json:"type"`
	Target    string              `json:"target"`
	Options   *MeasurementOptions `json:"measurementOptions,omitempty"`
}

type PostResponse struct {
	ID          string `json:"id"`
	ProbesCount int    `json:"probesCount"`
}

type PostError struct {
	Error struct {
		Message string                 `json:"message"`
		Type    string                 `json:"type"`
		Params  map[string]interface{} `json:"params,omitempty"`
	} `json:"error"`
}
