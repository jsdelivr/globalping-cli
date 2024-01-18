package view

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/jsdelivr/globalping-cli/model"
	"github.com/stretchr/testify/assert"
)

func TestOutputLatency_Ping_Not_CI(t *testing.T) {
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

	id := "abc123"
	data := &model.GetMeasurement{
		Results: []model.MeasurementResponse{
			{
				Probe: model.ProbeData{
					Continent: "Continent",
					Country:   "Country",
					State:     "State",
					City:      "City",
					ASN:       12345,
					Network:   "Network",
					Tags:      []string{"tag-1"},
				},
				Result: model.ResultData{
					StatsRaw: json.RawMessage(`{"min":8,"avg":12,"max":20}`),
				},
			},
			{
				Probe: model.ProbeData{
					Continent: "Continent B",
					Country:   "Country B",
					State:     "State B",
					City:      "City B",
					ASN:       12349,
					Network:   "Network B",
					Tags:      []string{"tag B"},
				},
				Result: model.ResultData{
					StatsRaw: json.RawMessage(`{"min":9,"avg":15,"max":22}`),
				},
			},
		},
	}
	ctx := model.Context{
		Cmd: "ping",
	}

	err = OutputLatency(id, data, ctx)
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

func TestOutputLatency_Ping_CI(t *testing.T) {
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

	id := "abc123"
	data := &model.GetMeasurement{
		Results: []model.MeasurementResponse{
			{
				Probe: model.ProbeData{
					Continent: "Continent",
					Country:   "Country",
					State:     "State",
					City:      "City",
					ASN:       12345,
					Network:   "Network",
					Tags:      []string{"tag"},
				},
				Result: model.ResultData{
					StatsRaw: json.RawMessage(`{"min":8,"avg":12,"max":20}`),
				},
			},
		},
	}
	ctx := model.Context{
		Cmd: "ping",
		CI:  true,
	}

	err = OutputLatency(id, data, ctx)
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

func TestOutputLatency_DNS_Not_CI(t *testing.T) {
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

	id := "abc123"
	data := &model.GetMeasurement{
		Results: []model.MeasurementResponse{
			{
				Probe: model.ProbeData{
					Continent: "Continent",
					Country:   "Country",
					State:     "State",
					City:      "City",
					ASN:       12345,
					Network:   "Network",
					Tags:      []string{"tag"},
				},
				Result: model.ResultData{
					TimingsRaw: []byte(`{"total": 44}`),
				},
			},
		},
	}
	ctx := model.Context{
		Cmd: "dns",
	}

	err = OutputLatency(id, data, ctx)
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

func TestOutputLatency_DNS_CI(t *testing.T) {
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

	id := "abc123"
	data := &model.GetMeasurement{
		Results: []model.MeasurementResponse{
			{
				Probe: model.ProbeData{
					Continent: "Continent",
					Country:   "Country",
					State:     "State",
					City:      "City",
					ASN:       12345,
					Network:   "Network",
					Tags:      []string{"tag"},
				},
				Result: model.ResultData{
					TimingsRaw: []byte(`{"total": 44}`),
				},
			},
		},
	}
	ctx := model.Context{
		Cmd: "dns",
		CI:  true,
	}

	err = OutputLatency(id, data, ctx)
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

func TestOutputLatency_Http_Not_CI(t *testing.T) {
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

	id := "abc123"
	data := &model.GetMeasurement{
		Results: []model.MeasurementResponse{
			{
				Probe: model.ProbeData{
					Continent: "Continent",
					Country:   "Country",
					State:     "State",
					City:      "City",
					ASN:       12345,
					Network:   "Network",
					Tags:      []string{"tag"},
				},
				Result: model.ResultData{
					TimingsRaw: []byte(`{"total": 44,"download":11,"firstByte":20,"dns":5,"tls":2,"tcp":4}`),
				},
			},
		},
	}
	ctx := model.Context{
		Cmd: "http",
	}

	err = OutputLatency(id, data, ctx)
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

func TestOutputLatency_Http_CI(t *testing.T) {
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

	id := "abc123"
	data := &model.GetMeasurement{
		Results: []model.MeasurementResponse{
			{
				Probe: model.ProbeData{
					Continent: "Continent",
					Country:   "Country",
					State:     "State",
					City:      "City",
					ASN:       12345,
					Network:   "Network",
					Tags:      []string{"tag"},
				},
				Result: model.ResultData{
					TimingsRaw: []byte(`{"total": 44,"download":11,"firstByte":20,"dns":5,"tls":2,"tcp":4}`),
				},
			},
		},
	}
	ctx := model.Context{
		Cmd: "http",
		CI:  true,
	}

	err = OutputLatency(id, data, ctx)
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
