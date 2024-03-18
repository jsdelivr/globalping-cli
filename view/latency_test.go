package view

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_Output_Latency_Ping_Not_CI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	measurement := &globalping.Measurement{
		Results: []globalping.ProbeMeasurement{
			{
				Probe: globalping.ProbeDetails{
					Continent: "Continent",
					Country:   "Country",
					State:     "State",
					City:      "City",
					ASN:       12345,
					Network:   "Network",
					Tags:      []string{"tag-1"},
				},
				Result: globalping.ProbeResult{
					StatsRaw: json.RawMessage(`{"min":8,"avg":12,"max":20}`),
				},
			},
			{
				Probe: globalping.ProbeDetails{
					Continent: "Continent B",
					Country:   "Country B",
					State:     "State B",
					City:      "City B",
					ASN:       12349,
					Network:   "Network B",
					Tags:      []string{"tag B"},
				},
				Result: globalping.ProbeResult{
					StatsRaw: json.RawMessage(`{"min":9,"avg":15,"max":22}`),
				},
			},
		},
	}

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().GetMeasurement(measurementID1).Times(1).Return(measurement, nil)

	w := new(bytes.Buffer)
	viewer := NewViewer(
		&Context{
			Cmd:       "ping",
			ToLatency: true,
		},
		NewPrinter(nil, w, w),
		nil,
		gbMock,
	)

	err := viewer.Output(measurementID1, &globalping.MeasurementCreate{})
	assert.NoError(t, err)

	assert.Equal(t, `> Continent, Country, (State), City, ASN:12345, Network (tag-1)
Min: 8.00 ms
Max: 20.00 ms
Avg: 12.00 ms

> Continent B, Country B, (State B), City B, ASN:12349, Network B
Min: 9.00 ms
Max: 22.00 ms
Avg: 15.00 ms

`, w.String())
}

func Test_Output_Latency_Ping_CI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	measurement := &globalping.Measurement{
		Results: []globalping.ProbeMeasurement{
			{
				Probe: globalping.ProbeDetails{
					Continent: "Continent",
					Country:   "Country",
					State:     "State",
					City:      "City",
					ASN:       12345,
					Network:   "Network",
					Tags:      []string{"tag"},
				},
				Result: globalping.ProbeResult{
					StatsRaw: json.RawMessage(`{"min":8,"avg":12,"max":20}`),
				},
			},
		},
	}

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().GetMeasurement(measurementID1).Times(1).Return(measurement, nil)

	w := new(bytes.Buffer)
	viewer := NewViewer(
		&Context{
			Cmd:       "ping",
			ToLatency: true,
			CIMode:    true,
		},
		NewPrinter(nil, w, w),
		nil,
		gbMock,
	)

	err := viewer.Output(measurementID1, &globalping.MeasurementCreate{})
	assert.NoError(t, err)

	assert.Equal(t, `> Continent, Country, (State), City, ASN:12345, Network
Min: 8.00 ms
Max: 20.00 ms
Avg: 12.00 ms

`, w.String())
}

func Test_Output_Latency_DNS_Not_CI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	measurement := &globalping.Measurement{
		Results: []globalping.ProbeMeasurement{
			{
				Probe: globalping.ProbeDetails{
					Continent: "Continent",
					Country:   "Country",
					State:     "State",
					City:      "City",
					ASN:       12345,
					Network:   "Network",
					Tags:      []string{"tag"},
				},
				Result: globalping.ProbeResult{
					TimingsRaw: []byte(`{"total": 44}`),
				},
			},
		},
	}

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().GetMeasurement(measurementID1).Times(1).Return(measurement, nil)

	w := new(bytes.Buffer)
	viewer := NewViewer(
		&Context{
			Cmd:       "dns",
			ToLatency: true,
		},
		NewPrinter(nil, w, w),
		nil,
		gbMock,
	)

	err := viewer.Output(measurementID1, &globalping.MeasurementCreate{})
	assert.NoError(t, err)

	assert.Equal(t, `> Continent, Country, (State), City, ASN:12345, Network
Total: 44 ms

`, w.String())
}

func Test_Output_Latency_DNS_CI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	measurement := &globalping.Measurement{
		Results: []globalping.ProbeMeasurement{
			{
				Probe: globalping.ProbeDetails{
					Continent: "Continent",
					Country:   "Country",
					State:     "State",
					City:      "City",
					ASN:       12345,
					Network:   "Network",
					Tags:      []string{"tag"},
				},
				Result: globalping.ProbeResult{
					TimingsRaw: []byte(`{"total": 44}`),
				},
			},
		},
	}

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().GetMeasurement(measurementID1).Times(1).Return(measurement, nil)

	w := new(bytes.Buffer)
	viewer := NewViewer(
		&Context{
			Cmd:       "dns",
			ToLatency: true,
			CIMode:    true,
		},
		NewPrinter(nil, w, w),
		nil,
		gbMock,
	)

	err := viewer.Output(measurementID1, &globalping.MeasurementCreate{})
	assert.NoError(t, err)

	assert.Equal(t, `> Continent, Country, (State), City, ASN:12345, Network
Total: 44 ms

`, w.String())
}

func Test_Output_Latency_Http_Not_CI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	measurement := &globalping.Measurement{
		Results: []globalping.ProbeMeasurement{
			{
				Probe: globalping.ProbeDetails{
					Continent: "Continent",
					Country:   "Country",
					State:     "State",
					City:      "City",
					ASN:       12345,
					Network:   "Network",
					Tags:      []string{"tag"},
				},
				Result: globalping.ProbeResult{
					TimingsRaw: []byte(`{"total": 44,"download":11,"firstByte":20,"dns":5,"tls":2,"tcp":4}`),
				},
			},
		},
	}

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().GetMeasurement(measurementID1).Times(1).Return(measurement, nil)

	w := new(bytes.Buffer)
	viewer := NewViewer(
		&Context{
			Cmd:       "http",
			ToLatency: true,
		},
		NewPrinter(nil, w, w),
		nil,
		gbMock,
	)

	err := viewer.Output(measurementID1, &globalping.MeasurementCreate{})
	assert.NoError(t, err)

	assert.Equal(t, `> Continent, Country, (State), City, ASN:12345, Network
Total: 44 ms
Download: 11 ms
First byte: 20 ms
DNS: 5 ms
TLS: 2 ms
TCP: 4 ms

`, w.String())
}

func Test_Output_Latency_Http_CI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	measurement := &globalping.Measurement{
		Results: []globalping.ProbeMeasurement{
			{
				Probe: globalping.ProbeDetails{
					Continent: "Continent",
					Country:   "Country",
					State:     "State",
					City:      "City",
					ASN:       12345,
					Network:   "Network",
					Tags:      []string{"tag"},
				},
				Result: globalping.ProbeResult{
					TimingsRaw: []byte(`{"total": 44,"download":11,"firstByte":20,"dns":5,"tls":2,"tcp":4}`),
				},
			},
		},
	}

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().GetMeasurement(measurementID1).Times(1).Return(measurement, nil)

	w := new(bytes.Buffer)
	viewer := NewViewer(
		&Context{
			Cmd:       "http",
			ToLatency: true,
			CIMode:    true,
		},
		NewPrinter(nil, w, w),
		nil,
		gbMock,
	)

	err := viewer.Output(measurementID1, &globalping.MeasurementCreate{})
	assert.NoError(t, err)

	assert.Equal(t, `> Continent, Country, (State), City, ASN:12345, Network
Total: 44 ms
Download: 11 ms
First byte: 20 ms
DNS: 5 ms
TLS: 2 ms
TCP: 4 ms

`, w.String())
}
