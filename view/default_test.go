package view

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jsdelivr/globalping-go"
	"github.com/stretchr/testify/assert"
)

func Test_Output_Default_HTTP_Get(t *testing.T) {
	measurement := &globalping.Measurement{
		Results: []globalping.ProbeMeasurement{
			{
				Probe: globalping.ProbeDetails{
					Continent: "EU",
					Country:   "DE",
					City:      "Berlin",
					ASN:       123,
					Network:   "Network 1",
				},
				Result: globalping.ProbeResult{
					RawOutput:  "Headers 1\nBody 1",
					RawHeaders: "Headers 1",
					RawBody:    pointerTo("Body 1"),
				},
			},

			{
				Probe: globalping.ProbeDetails{
					Continent: "NA",
					Country:   "US",
					City:      "New York",
					State:     pointerTo("NY"),
					ASN:       567,
					Network:   "Network 2",
				},
				Result: globalping.ProbeResult{
					RawOutput:  "Headers 2\nBody 2",
					RawHeaders: "Headers 2",
					RawBody:    pointerTo("Body 2"),
				},
			},
		},
	}

	opts := &globalping.MeasurementCreate{
		Options: &globalping.MeasurementOptions{
			Request: &globalping.RequestOptions{
				Method: "GET",
			},
		},
	}

	w := new(bytes.Buffer)
	printer := NewPrinter(nil, w, w)
	printer.DisableStyling()
	viewer := NewViewer(&Context{
		Cmd:    "http",
		CIMode: true,
	}, printer, nil)

	viewer.OutputDefault(measurementID1, measurement, opts)

	assert.Equal(t, `> Berlin, DE, EU, Network 1 (AS123)
Body 1

> New York (NY), US, NA, Network 2 (AS567)
Body 2
`, w.String())
}

func Test_Output_Default_HTTP_Get_Share(t *testing.T) {
	measurement := &globalping.Measurement{
		Results: []globalping.ProbeMeasurement{
			{
				Probe: globalping.ProbeDetails{
					Continent: "EU",
					Country:   "DE",
					City:      "Berlin",
					ASN:       123,
					Network:   "Network 1",
				},
				Result: globalping.ProbeResult{
					RawOutput:  "Headers 1\nBody 1",
					RawHeaders: "Headers 1",
					RawBody:    pointerTo("Body 1"),
				},
			},

			{
				Probe: globalping.ProbeDetails{
					Continent: "NA",
					Country:   "US",
					City:      "New York",
					State:     pointerTo("NY"),
					ASN:       567,
					Network:   "Network 2",
				},
				Result: globalping.ProbeResult{
					RawOutput:  "Headers 2\nBody 2",
					RawHeaders: "Headers 2",
					RawBody:    pointerTo("Body 2"),
				},
			},
		},
	}

	opts := &globalping.MeasurementCreate{
		Options: &globalping.MeasurementOptions{
			Request: &globalping.RequestOptions{
				Method: "GET",
			},
		},
	}
	w := new(bytes.Buffer)
	errW := new(bytes.Buffer)
	printer := NewPrinter(nil, w, errW)
	printer.DisableStyling()
	viewer := NewViewer(&Context{
		Cmd:    "http",
		CIMode: true,
		Share:  true,
	}, printer, nil)

	viewer.OutputDefault(measurementID1, measurement, opts)

	assert.Equal(t, fmt.Sprintf(`> Berlin, DE, EU, Network 1 (AS123)
> New York (NY), US, NA, Network 2 (AS567)
> View the results online: https://globalping.io?measurement=%s
`, measurementID1), errW.String())

	assert.Equal(t, `Body 1

Body 2
`, w.String())
}

func Test_Output_Default_HTTP_Get_Full(t *testing.T) {
	now := time.Now()

	measurement := &globalping.Measurement{
		Results: []globalping.ProbeMeasurement{
			{
				Probe: globalping.ProbeDetails{
					Continent: "EU",
					Country:   "DE",
					City:      "Berlin",
					ASN:       123,
					Network:   "Network 1",
				},
				Result: globalping.ProbeResult{
					TLS: &globalping.HTTPTLSCertificate{
						Authorized: true,
						Protocol:   "TLSv1.3",
						CipherName: "TLS_AES_256_GCM_SHA384",
						Subject: globalping.TLSCertificateSubject{
							CommonName:      "Sub CN",
							AlternativeName: "Sub alt",
						},
						Issuer: globalping.TLSCertificateIssuer{
							CommonName:   "Iss CN",
							Organization: "Iss O",
							Country:      "Iss C",
						},
						CreatedAt:      now,
						ExpiresAt:      now.AddDate(1, 0, 0),
						SerialNumber:   "03:DD",
						Fingerprint256: "79:BD",
						KeyType:        pointerTo("EC"),
						KeyBits:        pointerTo(256),
					},
					RawOutput:       "HTTP/1.1 301\nHeaders 1\nBody 1",
					RawHeaders:      "Headers 1",
					RawBody:         pointerTo("Body 1"),
					ResolvedAddress: pointerTo("1.1.1.1"),
				},
			},
			{
				Probe: globalping.ProbeDetails{
					Continent: "NA",
					Country:   "US",
					City:      "New York",
					State:     pointerTo("NY"),
					ASN:       567,
					Network:   "Network 2",
				},
				Result: globalping.ProbeResult{
					TLS: &globalping.HTTPTLSCertificate{
						Authorized: false,
						Error:      "TLS Error",
						Protocol:   "TLSv1.3",
						CipherName: "TLS_AES_256_GCM_SHA384",
						Subject: globalping.TLSCertificateSubject{
							CommonName:      "Sub CN",
							AlternativeName: "Sub alt",
						},
						Issuer: globalping.TLSCertificateIssuer{
							CommonName:   "Iss CN",
							Organization: "Iss O",
							Country:      "Iss C",
						},
						CreatedAt:      now,
						ExpiresAt:      now.AddDate(1, 0, 0),
						SerialNumber:   "03:DD",
						Fingerprint256: "79:BD",
						KeyType:        pointerTo("EC"),
						KeyBits:        pointerTo(256),
					},
					RawOutput:  "HTTP/1.1 301\nHeaders 2\nBody 2",
					RawHeaders: "Headers 2",
					RawBody:    pointerTo("Body 2"),
				},
			},
		},
	}

	opts := &globalping.MeasurementCreate{
		Options: &globalping.MeasurementOptions{
			Request: &globalping.RequestOptions{
				Method: "GET",
			},
		},
	}
	w := new(bytes.Buffer)
	errW := new(bytes.Buffer)
	printer := NewPrinter(nil, w, errW)
	printer.DisableStyling()
	viewer := NewViewer(&Context{
		Cmd:    "http",
		CIMode: true,
		Full:   true,
	}, printer, nil)

	viewer.OutputDefault(measurementID1, measurement, opts)

	validity := fmt.Sprintf("%s; %s", now.Format(time.RFC3339), now.AddDate(1, 0, 0).Format(time.RFC3339))

	assert.Equal(t, `> Berlin, DE, EU, Network 1 (AS123)
Resolved address: 1.1.1.1

TLSv1.3/TLS_AES_256_GCM_SHA384
Subject: Sub CN; Sub alt
Issuer: Iss CN; Iss O; Iss C
Validity: `+validity+`
Serial number: 03:DD
Fingerprint: 79:BD
Key type: EC256

HTTP/1.1 301
Headers 1

> New York (NY), US, NA, Network 2 (AS567)
TLSv1.3/TLS_AES_256_GCM_SHA384
Error: TLS Error
Subject: Sub CN; Sub alt
Issuer: Iss CN; Iss O; Iss C
Validity: `+validity+`
Serial number: 03:DD
Fingerprint: 79:BD
Key type: EC256

HTTP/1.1 301
Headers 2

`, errW.String())
	assert.Equal(t, `Body 1

Body 2
`, w.String())
}

func Test_Output_Default_HTTP_Head(t *testing.T) {
	measurement := &globalping.Measurement{
		Results: []globalping.ProbeMeasurement{
			{
				Probe: globalping.ProbeDetails{
					Continent: "EU",
					Country:   "DE",
					City:      "Berlin",
					ASN:       123,
					Network:   "Network 1",
				},
				Result: globalping.ProbeResult{
					RawOutput:  "Headers 1",
					RawHeaders: "Headers 1",
				},
			},

			{
				Probe: globalping.ProbeDetails{
					Continent: "NA",
					Country:   "US",
					City:      "New York",
					State:     pointerTo("NY"),
					ASN:       567,
					Network:   "Network 2",
				},
				Result: globalping.ProbeResult{
					RawOutput:  "Headers 2",
					RawHeaders: "Headers 2",
				},
			},
		},
	}

	opts := &globalping.MeasurementCreate{
		Options: &globalping.MeasurementOptions{
			Request: &globalping.RequestOptions{
				Method: "HEAD",
			},
		},
	}
	w := new(bytes.Buffer)
	errW := new(bytes.Buffer)
	printer := NewPrinter(nil, w, errW)
	printer.DisableStyling()
	viewer := NewViewer(&Context{
		Cmd:    "http",
		CIMode: true,
	}, printer, nil)

	viewer.OutputDefault(measurementID1, measurement, opts)

	assert.Equal(t, `> Berlin, DE, EU, Network 1 (AS123)
> New York (NY), US, NA, Network 2 (AS567)
`, errW.String())
	assert.Equal(t, `Headers 1

Headers 2
`, w.String())
}

func Test_Output_Default_HTTP_Get_NullableFields(t *testing.T) {
	measurement := &globalping.Measurement{
		Results: []globalping.ProbeMeasurement{
			{
				Probe: globalping.ProbeDetails{
					Continent: "EU",
					Country:   "DE",
					City:      "Berlin",
					ASN:       123,
					Network:   "Network",
				},
				Result: globalping.ProbeResult{
					TLS: &globalping.HTTPTLSCertificate{
						Authorized: true,
						Protocol:   "TLSv1.3",
						CipherName: "TLS_AES_256_GCM_SHA384",
						Subject: globalping.TLSCertificateSubject{
							CommonName:      "subject",
							AlternativeName: "alt",
						},
						Issuer: globalping.TLSCertificateIssuer{
							CommonName:   "issuer",
							Organization: "organization",
							Country:      "country",
						},
						SerialNumber:   "serial",
						Fingerprint256: "fingerprint",
					},
					RawOutput:  "HTTP/1.1 204 No Content\n",
					RawHeaders: "HTTP/1.1 204 No Content",
				},
			},
		},
	}
	opts := &globalping.MeasurementCreate{
		Options: &globalping.MeasurementOptions{
			Request: &globalping.RequestOptions{Method: http.MethodGet},
		},
	}
	w := new(bytes.Buffer)
	errW := new(bytes.Buffer)
	printer := NewPrinter(nil, w, errW)
	printer.DisableStyling()
	viewer := NewViewer(&Context{Cmd: "http", Full: true}, printer, nil)

	viewer.OutputDefault(measurementID1, measurement, opts)

	assert.Equal(t, "", w.String())
	assert.Equal(t, `> Berlin, DE, EU, Network (AS123)
TLSv1.3/TLS_AES_256_GCM_SHA384
Subject: subject; alt
Issuer: issuer; organization; country
Validity: 0001-01-01T00:00:00Z; 0001-01-01T00:00:00Z
Serial number: serial
Fingerprint: fingerprint

HTTP/1.1 204 No Content
HTTP/1.1 204 No Content

`, errW.String())
	assert.NotContains(t, errW.String(), "Key type:")
	assert.NotContains(t, errW.String(), "%!")
}

func Test_Output_Default_Ping(t *testing.T) {
	measurement := &globalping.Measurement{
		Results: []globalping.ProbeMeasurement{
			{
				Probe: globalping.ProbeDetails{
					Continent: "EU",
					Country:   "DE",
					City:      "Berlin",
					ASN:       123,
					Network:   "Network 1",
				},
				Result: globalping.ProbeResult{
					RawOutput: "Ping Results 1",
				},
			},

			{
				Probe: globalping.ProbeDetails{
					Continent: "NA",
					Country:   "US",
					City:      "New York",
					State:     pointerTo("NY"),
					ASN:       567,
					Network:   "Network 2",
				},
				Result: globalping.ProbeResult{
					RawOutput: "Ping Results 2",
				},
			},
		},
	}

	opts := &globalping.MeasurementCreate{}
	w := new(bytes.Buffer)
	errW := new(bytes.Buffer)
	printer := NewPrinter(nil, w, errW)
	printer.DisableStyling()
	viewer := NewViewer(&Context{
		Cmd:    "ping",
		CIMode: true,
	}, printer, nil)

	viewer.OutputDefault(measurementID1, measurement, opts)

	assert.Equal(t, `> Berlin, DE, EU, Network 1 (AS123)
> New York (NY), US, NA, Network 2 (AS567)
`, errW.String())
	assert.Equal(t, `Ping Results 1

Ping Results 2
`, w.String())
}
