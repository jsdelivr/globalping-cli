package view

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/jsdelivr/globalping-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Output_Latency_Ping(t *testing.T) {
	measurement := &globalping.Measurement{
		Results: []globalping.ProbeMeasurement{
			{
				Probe: globalping.ProbeDetails{
					Continent: "Continent",
					Country:   "Country",
					State:     pointerTo("State"),
					City:      "City",
					ASN:       12345,
					Network:   "Network",
					Tags:      []string{"tag-1"},
				},
				Result: globalping.ProbeResult{
					Status:   globalping.TestStatusFinished,
					StatsRaw: json.RawMessage(`{"min":8,"avg":12,"max":20}`),
				},
			},
			{
				Probe: globalping.ProbeDetails{
					Continent: "Continent B",
					Country:   "Country B",
					State:     pointerTo("State B"),
					City:      "City B",
					ASN:       12349,
					Network:   "Network B",
					Tags:      []string{"tag B"},
				},
				Result: globalping.ProbeResult{
					Status:   globalping.TestStatusFinished,
					StatsRaw: json.RawMessage(`{"min":9,"avg":15,"max":22}`),
				},
			},
		},
	}

	w := new(bytes.Buffer)
	errW := new(bytes.Buffer)
	viewer := NewViewer(
		&Context{
			Cmd:       "ping",
			ToLatency: true,
		},
		NewPrinter(nil, w, errW),
		nil,
	)

	err := viewer.OutputLatency(measurementID1, measurement)
	assert.NoError(t, err)

	assert.Equal(t, "\033[1;38;5;43m> City (State), Country, Continent, Network (AS12345) (tag-1)\033[0m\n"+
		"\033[1;38;5;43m> City B (State B), Country B, Continent B, Network B (AS12349)\033[0m\n", errW.String())
	assert.Equal(t, "\033[1mMin: \033[0m8.00 ms\n"+
		"\033[1mMax: \033[0m20.00 ms\n"+
		"\033[1mAvg: \033[0m12.00 ms\n\n"+
		"\033[1mMin: \033[0m9.00 ms\n"+
		"\033[1mMax: \033[0m22.00 ms\n"+
		"\033[1mAvg: \033[0m15.00 ms\n\n", w.String())
}

func Test_Output_Latency_Ping_StylingDisabled(t *testing.T) {
	measurement := &globalping.Measurement{
		Results: []globalping.ProbeMeasurement{
			{
				Probe: globalping.ProbeDetails{
					Continent: "Continent",
					Country:   "Country",
					State:     pointerTo("State"),
					City:      "City",
					ASN:       12345,
					Network:   "Network",
					Tags:      []string{"tag"},
				},
				Result: globalping.ProbeResult{
					Status:   globalping.TestStatusFinished,
					StatsRaw: json.RawMessage(`{"min":8,"avg":12,"max":20}`),
				},
			},
		},
	}

	w := new(bytes.Buffer)
	printer := NewPrinter(nil, w, w)
	printer.DisableStyling()
	viewer := NewViewer(
		&Context{
			Cmd:       "ping",
			ToLatency: true,
		},
		printer,
		nil,
	)

	err := viewer.OutputLatency(measurementID1, measurement)
	assert.NoError(t, err)

	assert.Equal(t, `> City (State), Country, Continent, Network (AS12345)
Min: 8.00 ms
Max: 20.00 ms
Avg: 12.00 ms

`, w.String())
}

func Test_Output_Latency_DNS(t *testing.T) {
	measurement := &globalping.Measurement{
		Results: []globalping.ProbeMeasurement{
			{
				Probe: globalping.ProbeDetails{
					Continent: "Continent",
					Country:   "Country",
					State:     pointerTo("State"),
					City:      "City",
					ASN:       12345,
					Network:   "Network",
					Tags:      []string{"tag"},
				},
				Result: globalping.ProbeResult{
					Status:     globalping.TestStatusFinished,
					TimingsRaw: []byte(`{"total": 44}`),
				},
			},
		},
	}

	w := new(bytes.Buffer)
	viewer := NewViewer(
		&Context{
			Cmd:       "dns",
			ToLatency: true,
		},
		NewPrinter(nil, w, w),
		nil,
	)

	err := viewer.OutputLatency(measurementID1, measurement)
	assert.NoError(t, err)

	assert.Equal(t, "\033[1;38;5;43m> City (State), Country, Continent, Network (AS12345)\033[0m\n"+
		"\033[1mTotal: \033[0m44 ms\n\n", w.String())
}

func Test_Output_Latency_DNS_StylingDisabled(t *testing.T) {
	measurement := &globalping.Measurement{
		Results: []globalping.ProbeMeasurement{
			{
				Probe: globalping.ProbeDetails{
					Continent: "Continent",
					Country:   "Country",
					State:     pointerTo("State"),
					City:      "City",
					ASN:       12345,
					Network:   "Network",
					Tags:      []string{"tag"},
				},
				Result: globalping.ProbeResult{
					Status:     globalping.TestStatusFinished,
					TimingsRaw: []byte(`{"total": 44}`),
				},
			},
		},
	}

	w := new(bytes.Buffer)
	printer := NewPrinter(nil, w, w)
	printer.DisableStyling()
	viewer := NewViewer(
		&Context{
			Cmd:       "dns",
			ToLatency: true,
		},
		printer,
		nil,
	)

	err := viewer.OutputLatency(measurementID1, measurement)
	assert.NoError(t, err)

	assert.Equal(t, `> City (State), Country, Continent, Network (AS12345)
Total: 44 ms

`, w.String())
}

func Test_Output_Latency_Http(t *testing.T) {
	measurement := &globalping.Measurement{
		Results: []globalping.ProbeMeasurement{
			{
				Probe: globalping.ProbeDetails{
					Continent: "Continent",
					Country:   "Country",
					State:     pointerTo("State"),
					City:      "City",
					ASN:       12345,
					Network:   "Network",
					Tags:      []string{"tag"},
				},
				Result: globalping.ProbeResult{
					Status:     globalping.TestStatusFinished,
					TimingsRaw: []byte(`{"total": 44,"download":11,"firstByte":20,"dns":5,"tls":2,"tcp":4}`),
				},
			},
		},
	}

	w := new(bytes.Buffer)
	viewer := NewViewer(
		&Context{
			Cmd:       "http",
			ToLatency: true,
		},
		NewPrinter(nil, w, w),
		nil,
	)

	err := viewer.OutputLatency(measurementID1, measurement)
	assert.NoError(t, err)

	assert.Equal(t, "\033[1;38;5;43m> City (State), Country, Continent, Network (AS12345)\033[0m\n"+
		"\033[1mTotal: \033[0m44 ms\n"+
		"\033[1mDownload: \033[0m11 ms\n"+
		"\033[1mFirst byte: \033[0m20 ms\n"+
		"\033[1mDNS: \033[0m5 ms\n"+
		"\033[1mTLS: \033[0m2 ms\n"+
		"\033[1mTCP: \033[0m4 ms\n\n", w.String())
}

func Test_Output_Latency_Http_StylingDisabled(t *testing.T) {
	measurement := &globalping.Measurement{
		Results: []globalping.ProbeMeasurement{
			{
				Probe: globalping.ProbeDetails{
					Continent: "Continent",
					Country:   "Country",
					State:     pointerTo("State"),
					City:      "City",
					ASN:       12345,
					Network:   "Network",
					Tags:      []string{"tag"},
				},
				Result: globalping.ProbeResult{
					Status:     globalping.TestStatusFinished,
					TimingsRaw: []byte(`{"total": 44,"download":11,"firstByte":20,"dns":5,"tls":2,"tcp":4}`),
				},
			},
		},
	}

	w := new(bytes.Buffer)
	errW := new(bytes.Buffer)
	printer := NewPrinter(nil, w, errW)
	printer.DisableStyling()
	viewer := NewViewer(
		&Context{
			Cmd:       "http",
			ToLatency: true,
		},
		printer,
		nil,
	)

	err := viewer.OutputLatency(measurementID1, measurement)
	assert.NoError(t, err)

	assert.Equal(t, `> City (State), Country, Continent, Network (AS12345)
`, errW.String())
	assert.Equal(t, `Total: 44 ms
Download: 11 ms
First byte: 20 ms
DNS: 5 ms
TLS: 2 ms
TCP: 4 ms

`, w.String())
}

func Test_Output_Latency_NullableValues(t *testing.T) {
	for _, test := range []struct {
		name     string
		command  string
		stats    json.RawMessage
		timings  json.RawMessage
		expected string
	}{
		{
			name:     "ping statistics",
			command:  "ping",
			stats:    json.RawMessage(`{"min":null,"avg":null,"max":null}`),
			expected: "Min: -\nMax: -\nAvg: -\n\n",
		},
		{
			name:     "HTTP timings",
			command:  "http",
			timings:  json.RawMessage(`{"total":44,"download":null,"firstByte":20,"dns":null,"tls":null,"tcp":4}`),
			expected: "Total: 44 ms\nDownload: -\nFirst byte: 20 ms\nDNS: -\nTLS: -\nTCP: 4 ms\n\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
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
							Status:     globalping.TestStatusFinished,
							StatsRaw:   test.stats,
							TimingsRaw: test.timings,
						},
					},
				},
			}
			w := new(bytes.Buffer)
			errW := new(bytes.Buffer)
			printer := NewPrinter(nil, w, errW)
			printer.DisableStyling()
			viewer := NewViewer(&Context{Cmd: test.command, ToLatency: true}, printer, nil)

			err := viewer.OutputLatency(measurementID1, measurement)

			require.NoError(t, err)
			assert.Equal(t, "> Berlin, DE, EU, Network (AS123)\n", errW.String())
			assert.Equal(t, test.expected, w.String())
			assert.NotContains(t, w.String(), "%!")
		})
	}
}

func Test_Output_Latency_Offline(t *testing.T) {
	measurement := &globalping.Measurement{
		Results: []globalping.ProbeMeasurement{
			{
				Probe: globalping.ProbeDetails{
					Continent: "Continent",
					Country:   "Country",
					State:     pointerTo("State"),
					City:      "City",
					ASN:       12345,
					Network:   "Network",
				},
				Result: globalping.ProbeResult{
					Status:    globalping.TestStatusOffline,
					RawOutput: "This probe is currently offline. Please try again later.",
				},
			},
		},
	}

	w := new(bytes.Buffer)
	errW := new(bytes.Buffer)
	printer := NewPrinter(nil, w, errW)
	printer.DisableStyling()
	viewer := NewViewer(
		&Context{
			Cmd:       "ping",
			ToLatency: true,
			Share:     true,
		},
		printer,
		nil,
	)

	err := viewer.OutputLatency(measurementID1, measurement)
	assert.NoError(t, err)

	assert.Equal(t, `> City (State), Country, Continent, Network (AS12345)
> View the results online: https://globalping.io?measurement=1zGzfAGL7sZfUs3c
`, errW.String())
	assert.Equal(t, "This probe is currently offline. Please try again later.\n\n", w.String())
}
