package globalping

import "encoding/json"

// Docs: https://www.jsdelivr.com/docs/api.globalping.io

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

type IPVersion int

const (
	IPVersion4 IPVersion = 4
	IPVersion6 IPVersion = 6
)

type MeasurementOptions struct {
	Query     *QueryOptions   `json:"query,omitempty"`
	Request   *RequestOptions `json:"request,omitempty"`
	Protocol  string          `json:"protocol,omitempty"`
	Port      int             `json:"port,omitempty"`
	Resolver  string          `json:"resolver,omitempty"`
	Trace     bool            `json:"trace,omitempty"`
	Packets   int             `json:"packets,omitempty"`
	IPVersion IPVersion       `json:"ipVersion,omitempty"`
}

type MeasurementCreate struct {
	Limit             int                 `json:"limit"`
	Locations         []Locations         `json:"locations"`
	Type              string              `json:"type"`
	Target            string              `json:"target"`
	InProgressUpdates bool                `json:"inProgressUpdates"`
	Options           *MeasurementOptions `json:"measurementOptions,omitempty"`
}

type MeasurementError struct {
	Code    int
	Message string
}

func (e *MeasurementError) Error() string {
	return e.Message
}

type MeasurementCreateResponse struct {
	ID          string `json:"id"`
	ProbesCount int    `json:"probesCount"`
}

type MeasurementCreateError struct {
	Error struct {
		Message string                 `json:"message"`
		Type    string                 `json:"type"`
		Params  map[string]interface{} `json:"params,omitempty"`
	} `json:"error"`
}

type ProbeDetails struct {
	Continent string   `json:"continent"`
	Region    string   `json:"region"`
	Country   string   `json:"country"`
	City      string   `json:"city"`
	State     string   `json:"state,omitempty"`
	ASN       int      `json:"asn"`
	Network   string   `json:"network,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}

type MeasurementStatus string

const (
	StatusInProgress MeasurementStatus = "in-progress"
	StatusFailed     MeasurementStatus = "failed"
	StatusOffline    MeasurementStatus = "offline"
	StatusFinished   MeasurementStatus = "finished"
)

type ProbeResult struct {
	Status           MeasurementStatus `json:"status"`
	RawOutput        string            `json:"rawOutput"`
	RawHeaders       string            `json:"rawHeaders"`
	RawBody          string            `json:"rawBody"`
	ResolvedAddress  string            `json:"resolvedAddress"`
	ResolvedHostname string            `json:"resolvedHostname"`
	StatsRaw         json.RawMessage   `json:"stats,omitempty"`
	TimingsRaw       json.RawMessage   `json:"timings,omitempty"`
}

type PingStats struct {
	Min   float64 `json:"min"`   // The lowest rtt value.
	Avg   float64 `json:"avg"`   // The average rtt value.
	Max   float64 `json:"max"`   // The highest rtt value.
	Total int     `json:"total"` // The number of sent packets.
	Rcv   int     `json:"rcv"`   // The number of received packets.
	Drop  int     `json:"drop"`  // The number of dropped packets (total - rcv).
	Loss  float64 `json:"loss"`  // The percentage of dropped packets.
	Mdev  float64 `json:"mdev"`  // The mean deviation of the rtt values.
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

type ProbeMeasurement struct {
	Probe  ProbeDetails `json:"probe"`
	Result ProbeResult  `json:"result"`
}

type Measurement struct {
	ID          string             `json:"id"`
	Type        string             `json:"type"`
	Status      MeasurementStatus  `json:"status"`
	CreatedAt   string             `json:"createdAt"`
	UpdatedAt   string             `json:"updatedAt"`
	Target      string             `json:"target"`
	ProbesCount int                `json:"probesCount"`
	Results     []ProbeMeasurement `json:"results"`
}
