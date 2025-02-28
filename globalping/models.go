package globalping

import (
	"encoding/json"
	"time"
)

// Docs: https://globalping.io/docs/api.globalping.io

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
	Code    int                    `json:"-"`
	Message string                 `json:"message"`
	Type    string                 `json:"type"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

func (e *MeasurementError) Error() string {
	return e.Message
}

type MeasurementErrorResponse struct {
	Error *MeasurementError `json:"error"`
}

type MeasurementCreateResponse struct {
	ID          string `json:"id"`
	ProbesCount int    `json:"probesCount"`
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
	Status    MeasurementStatus `json:"status"`    // The current measurement status.
	RawOutput string            `json:"rawOutput"` //  The raw output of the test. Can be presented to users but is not meant to be parsed by clients.

	// Common
	ResolvedAddress  string `json:"resolvedAddress"`  // The resolved IP address of the target
	ResolvedHostname string `json:"resolvedHostname"` // The resolved hostname of the target

	// Ping
	StatsRaw json.RawMessage `json:"stats,omitempty"` // Summary rtt and packet loss statistics. All times are in milliseconds.

	// DNS
	StatusCode     int             `json:"statusCode"`        // The HTTP status code.
	StatusCodeName string          `json:"statusCodeName"`    // The HTTP status code name.
	Resolver       string          `json:"resolver"`          // The hostname or IP of the resolver that answered the query.
	AnswersRaw     json.RawMessage `json:"answers,omitempty"` // An array of the received resource records.

	// HTTP
	RawHeaders string              `json:"rawHeaders"`        // The raw HTTP response headers.
	RawBody    string              `json:"rawBody"`           // The raw HTTP response body or null if there was no body in response. Note that only the first 10 kb are returned.
	Truncated  bool                `json:"truncated"`         // Indicates whether the rawBody value was truncated due to being too big.
	HeadersRaw json.RawMessage     `json:"headers,omitempty"` // The HTTP response headers.
	TLS        *HTTPTLSCertificate `json:"tls,omitempty"`     // Information about the TLS certificate or null if no TLS certificate is available.

	// Common
	HopsRaw    json.RawMessage `json:"hops,omitempty"`
	TimingsRaw json.RawMessage `json:"timings,omitempty"`
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

type TracerouteTiming struct {
	RTT float64 `json:"rtt"` // The round-trip time for this packet.
}

type TracerouteHop struct {
	ResolvedAddress  string             `json:"resolvedAddress"`  // The resolved IP address of the target
	ResolvedHostname string             `json:"resolvedHostname"` // The resolved hostname of the target
	Timings          []TracerouteTiming `json:"timings"`          // An array containing details for each packet. All times are in milliseconds.
}

type DNSAnswer struct {
	Name  string `json:"name"`  // The record domain name.
	Type  string `json:"type"`  // The record type.
	TTL   int    `json:"ttl"`   // The record time-to-live value in seconds.
	Class string `json:"class"` // The record class.
	Value string `json:"value"` // The record value.
}

type DNSTimings struct {
	Total float64 `json:"total"` // The total query time in milliseconds.
}

type TraceDNSHop struct {
	Resolver string      `json:"resolver"` // The hostname or IP of the resolver that answered the query.
	Answers  []DNSAnswer `json:"answers"`  // An array of the received resource records.
	Timings  DNSTimings  `json:"timings"`  // Details about the query times. All times are in milliseconds.
}

type MTRStats struct {
	Min   float64 `json:"min"`   // The lowest rtt value.
	Avg   float64 `json:"avg"`   // The average rtt value.
	Max   float64 `json:"max"`   // The highest rtt value.
	StDev float64 `json:"stDev"` // The standard deviation of the rtt values.

	JMin  float64 `json:"jMin"`  // The lowest jitter value.
	JAvg  float64 `json:"jAvg"`  // The average jitter value.
	JMax  float64 `json:"jMax"`  // The highest jitter value.
	Total int     `json:"total"` // The number of sent packets.
	Rcv   int     `json:"rcv"`   // The number of received packets.
	Drop  int     `json:"drop"`  // The number of dropped packets (total - rcv).

	Loss float64 `json:"loss"` // The percentage of dropped packets.
}

type MTRTiming struct {
	RTT float64 `json:"rtt"` // The round-trip time for this packet.
}

type MTRHop struct {
	ResolvedAddress  string      `json:"resolvedAddress"`  // The resolved IP address of the target
	ResolvedHostname string      `json:"resolvedHostname"` // The resolved hostname of the target
	ASN              []int       `json:"asn"`              // An array containing the ASNs assigned to this hop.
	Stats            MTRStats    `json:"stats"`            // Summary rtt and packet loss statistics. All times are in milliseconds.
	Timings          []MTRTiming `json:"timings"`          // An array containing details for each packet. All times are in milliseconds.
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

type TLSCertificateSubject struct {
	CommonName      string `json:"CN"`  // The subject's common name.
	AlternativeName string `json:"alt"` // The subject's alternative name.
}

type TLSCertificateIssuer struct {
	Country      string `json:"C"`  // The issuer's country.
	Organization string `json:"O"`  // The issuer's organization.
	CommonName   string `json:"CN"` // The issuer's common name.
}

type HTTPTLSCertificate struct {
	Protocol       string                `json:"protocol"`       // The negotiated SSL/TLS protocol version.
	ChipherName    string                `json:"cipherName"`     // The OpenSSL name of the cipher suite.
	Authorized     bool                  `json:"authorized"`     // Indicates whether a trusted authority signed the certificate
	Error          string                `json:"error"`          // The reason for rejecting the certificate if authorized is false
	CreatedAt      time.Time             `json:"createdAt"`      // The creation date and time of the certificate
	ExpiresAt      time.Time             `json:"expiresAt"`      // The expiration date and time of the certificate
	Subject        TLSCertificateSubject `json:"subject"`        // Information about the certificate subject.
	Issuer         TLSCertificateIssuer  `json:"issuer"`         // Information about the certificate issuer.
	KeyType        string                `json:"keyType"`        // The type of the used key, or null for unrecognized types.
	KeyBits        int                   `json:"keyBits"`        // The size of the used key, or null for unrecognized types.
	SerialNumber   string                `json:"serialNumber"`   // The certificate serial number as a : separated HEX string
	Fingerprint256 string                `json:"fingerprint256"` // The SHA-256 digest of the DER-encoded certificate as a : separated HEX string
	PublicKey      string                `json:"publicKey"`      // The public key as a : separated HEX string, or null for unrecognized types.
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
