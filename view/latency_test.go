package view

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/mocks"
	"github.com/stretchr/testify/assert"
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

	osStdErr := os.Stderr
	osStdOut := os.Stdout

	rStdErr, myStdErr, err := os.Pipe()
	assert.NoError(t, err)
	defer rStdErr.Close()

	rStdOut, myStdOut, err := os.Pipe()
	assert.NoError(t, err)
	defer rStdOut.Close()

	os.Stderr = myStdErr
	os.Stdout = myStdOut

	defer func() {
		os.Stderr = osStdErr
		os.Stdout = osStdOut
	}()

	viewer := NewViewer(&Context{
		Cmd:       "ping",
		ToLatency: true,
	}, gbMock)

	err = viewer.Output(measurementID1, &globalping.MeasurementCreate{})
	assert.NoError(t, err)
	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> Continent, Country, (State), City, ASN:12345, Network (tag-1)\n> Continent B, Country B, (State B), City B, ASN:12349, Network B\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "Min: 8.00 ms\nMax: 20.00 ms\nAvg: 12.00 ms\n\nMin: 9.00 ms\nMax: 22.00 ms\nAvg: 15.00 ms\n\n", string(outContent))
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

	osStdErr := os.Stderr
	osStdOut := os.Stdout

	rStdErr, myStdErr, err := os.Pipe()
	assert.NoError(t, err)
	defer rStdErr.Close()

	rStdOut, myStdOut, err := os.Pipe()
	assert.NoError(t, err)
	defer rStdOut.Close()

	os.Stderr = myStdErr
	os.Stdout = myStdOut

	defer func() {
		os.Stderr = osStdErr
		os.Stdout = osStdOut
	}()

	viewer := NewViewer(&Context{
		Cmd:       "ping",
		ToLatency: true,
		CI:        true,
	}, gbMock)

	err = viewer.Output(measurementID1, &globalping.MeasurementCreate{})
	assert.NoError(t, err)
	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> Continent, Country, (State), City, ASN:12345, Network\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "Min: 8.00 ms\nMax: 20.00 ms\nAvg: 12.00 ms\n\n", string(outContent))
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

	osStdErr := os.Stderr
	osStdOut := os.Stdout

	rStdErr, myStdErr, err := os.Pipe()
	assert.NoError(t, err)
	defer rStdErr.Close()

	rStdOut, myStdOut, err := os.Pipe()
	assert.NoError(t, err)
	defer rStdOut.Close()

	os.Stderr = myStdErr
	os.Stdout = myStdOut

	defer func() {
		os.Stderr = osStdErr
		os.Stdout = osStdOut
	}()

	viewer := NewViewer(&Context{
		Cmd:       "dns",
		ToLatency: true,
	}, gbMock)

	err = viewer.Output(measurementID1, &globalping.MeasurementCreate{})
	assert.NoError(t, err)
	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> Continent, Country, (State), City, ASN:12345, Network\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "Total: 44 ms\n\n", string(outContent))
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

	osStdErr := os.Stderr
	osStdOut := os.Stdout

	rStdErr, myStdErr, err := os.Pipe()
	assert.NoError(t, err)
	defer rStdErr.Close()

	rStdOut, myStdOut, err := os.Pipe()
	assert.NoError(t, err)
	defer rStdOut.Close()

	os.Stderr = myStdErr
	os.Stdout = myStdOut

	defer func() {
		os.Stderr = osStdErr
		os.Stdout = osStdOut
	}()

	viewer := NewViewer(&Context{
		Cmd:       "dns",
		ToLatency: true,
		CI:        true,
	}, gbMock)

	err = viewer.Output(measurementID1, &globalping.MeasurementCreate{})
	assert.NoError(t, err)
	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> Continent, Country, (State), City, ASN:12345, Network\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "Total: 44 ms\n\n", string(outContent))
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

	osStdErr := os.Stderr
	osStdOut := os.Stdout

	rStdErr, myStdErr, err := os.Pipe()
	assert.NoError(t, err)
	defer rStdErr.Close()

	rStdOut, myStdOut, err := os.Pipe()
	assert.NoError(t, err)
	defer rStdOut.Close()

	os.Stderr = myStdErr
	os.Stdout = myStdOut

	defer func() {
		os.Stderr = osStdErr
		os.Stdout = osStdOut
	}()

	viewer := NewViewer(&Context{
		Cmd:       "http",
		ToLatency: true,
	}, gbMock)

	err = viewer.Output(measurementID1, &globalping.MeasurementCreate{})
	assert.NoError(t, err)
	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> Continent, Country, (State), City, ASN:12345, Network\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "Total: 44 ms\nDownload: 11 ms\nFirst byte: 20 ms\nDNS: 5 ms\nTLS: 2 ms\nTCP: 4 ms\n\n", string(outContent))
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

	osStdErr := os.Stderr
	osStdOut := os.Stdout

	rStdErr, myStdErr, err := os.Pipe()
	assert.NoError(t, err)
	defer rStdErr.Close()

	rStdOut, myStdOut, err := os.Pipe()
	assert.NoError(t, err)
	defer rStdOut.Close()

	os.Stderr = myStdErr
	os.Stdout = myStdOut

	defer func() {
		os.Stderr = osStdErr
		os.Stdout = osStdOut
	}()

	viewer := NewViewer(&Context{
		Cmd:       "http",
		ToLatency: true,
		CI:        true,
	}, gbMock)

	err = viewer.Output(measurementID1, &globalping.MeasurementCreate{})
	assert.NoError(t, err)
	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> Continent, Country, (State), City, ASN:12345, Network\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "Total: 44 ms\nDownload: 11 ms\nFirst byte: 20 ms\nDNS: 5 ms\nTLS: 2 ms\nTCP: 4 ms\n\n", string(outContent))
}
