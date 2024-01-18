package view

import (
	"io"
	"os"
	"testing"

	"github.com/jsdelivr/globalping-cli/model"
	"github.com/stretchr/testify/assert"
)

func TestOutputDefaultHTTPGet(t *testing.T) {
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

	ctx := model.Context{
		Cmd: "http",
		CI:  true,
	}

	m := model.PostMeasurement{
		Options: &model.MeasurementOptions{
			Request: &model.RequestOptions{
				Method: "GET",
			},
		},
	}

	id := "123abc"

	data := &model.GetMeasurement{
		Results: []model.MeasurementResponse{
			{
				Probe: model.ProbeData{
					Continent: "EU",
					Country:   "DE",
					City:      "Berlin",
					ASN:       123,
					Network:   "Network 1",
				},
				Result: model.ResultData{
					RawOutput:  "Headers 1\nBody 1",
					RawHeaders: "Headers 1",
					RawBody:    "Body 1",
				},
			},

			{
				Probe: model.ProbeData{
					Continent: "NA",
					Country:   "US",
					City:      "New York",
					State:     "NY",
					ASN:       567,
					Network:   "Network 2",
				},
				Result: model.ResultData{
					RawOutput:  "Headers 2\nBody 2",
					RawHeaders: "Headers 2",
					RawBody:    "Body 2",
				},
			},
		},
	}

	OutputDefault(id, data, ctx, m)
	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> EU, DE, Berlin, ASN:123, Network 1\n> NA, US, (NY), New York, ASN:567, Network 2\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "Body 1\n\nBody 2\n", string(outContent))
}

func TestOutputDefaultHTTPGetShare(t *testing.T) {
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

	ctx := model.Context{
		Cmd:   "http",
		CI:    true,
		Share: true,
	}

	m := model.PostMeasurement{
		Options: &model.MeasurementOptions{
			Request: &model.RequestOptions{
				Method: "GET",
			},
		},
	}

	id := "123abc"

	data := &model.GetMeasurement{
		Results: []model.MeasurementResponse{
			{
				Probe: model.ProbeData{
					Continent: "EU",
					Country:   "DE",
					City:      "Berlin",
					ASN:       123,
					Network:   "Network 1",
				},
				Result: model.ResultData{
					RawOutput:  "Headers 1\nBody 1",
					RawHeaders: "Headers 1",
					RawBody:    "Body 1",
				},
			},

			{
				Probe: model.ProbeData{
					Continent: "NA",
					Country:   "US",
					City:      "New York",
					State:     "NY",
					ASN:       567,
					Network:   "Network 2",
				},
				Result: model.ResultData{
					RawOutput:  "Headers 2\nBody 2",
					RawHeaders: "Headers 2",
					RawBody:    "Body 2",
				},
			},
		},
	}

	OutputDefault(id, data, ctx, m)
	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> EU, DE, Berlin, ASN:123, Network 1\n> NA, US, (NY), New York, ASN:567, Network 2\n> View the results online: https://www.jsdelivr.com/globalping?measurement=123abc\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "Body 1\n\nBody 2\n", string(outContent))
}

func TestOutputDefaultHTTPGetFull(t *testing.T) {
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

	ctx := model.Context{
		Cmd:  "http",
		CI:   true,
		Full: true,
	}

	m := model.PostMeasurement{
		Options: &model.MeasurementOptions{
			Request: &model.RequestOptions{
				Method: "GET",
			},
		},
	}

	id := "123abc"

	data := &model.GetMeasurement{
		Results: []model.MeasurementResponse{
			{
				Probe: model.ProbeData{
					Continent: "EU",
					Country:   "DE",
					City:      "Berlin",
					ASN:       123,
					Network:   "Network 1",
				},
				Result: model.ResultData{
					RawOutput:  "Headers 1\nBody 1",
					RawHeaders: "Headers 1",
					RawBody:    "Body 1",
				},
			},

			{
				Probe: model.ProbeData{
					Continent: "NA",
					Country:   "US",
					City:      "New York",
					State:     "NY",
					ASN:       567,
					Network:   "Network 2",
				},
				Result: model.ResultData{
					RawOutput:  "Headers 2\nBody 2",
					RawHeaders: "Headers 2",
					RawBody:    "Body 2",
				},
			},
		},
	}

	OutputDefault(id, data, ctx, m)
	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> EU, DE, Berlin, ASN:123, Network 1\n> NA, US, (NY), New York, ASN:567, Network 2\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "Headers 1\nBody 1\n\nHeaders 2\nBody 2\n", string(outContent))
}

func TestOutputDefaultHTTPHead(t *testing.T) {
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

	ctx := model.Context{
		Cmd: "http",
		CI:  true,
	}

	m := model.PostMeasurement{
		Options: &model.MeasurementOptions{
			Request: &model.RequestOptions{
				Method: "HEAD",
			},
		},
	}

	id := "123abc"

	data := &model.GetMeasurement{
		Results: []model.MeasurementResponse{
			{
				Probe: model.ProbeData{
					Continent: "EU",
					Country:   "DE",
					City:      "Berlin",
					ASN:       123,
					Network:   "Network 1",
				},
				Result: model.ResultData{
					RawOutput:  "Headers 1",
					RawHeaders: "Headers 1",
				},
			},

			{
				Probe: model.ProbeData{
					Continent: "NA",
					Country:   "US",
					City:      "New York",
					State:     "NY",
					ASN:       567,
					Network:   "Network 2",
				},
				Result: model.ResultData{
					RawOutput:  "Headers 2",
					RawHeaders: "Headers 2",
				},
			},
		},
	}

	OutputDefault(id, data, ctx, m)
	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> EU, DE, Berlin, ASN:123, Network 1\n> NA, US, (NY), New York, ASN:567, Network 2\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "Headers 1\n\nHeaders 2\n", string(outContent))
}

func TestOutputDefaultPing(t *testing.T) {
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

	ctx := model.Context{
		Cmd: "ping",
		CI:  true,
	}

	m := model.PostMeasurement{}

	id := "123abc"

	data := &model.GetMeasurement{
		Results: []model.MeasurementResponse{
			{
				Probe: model.ProbeData{
					Continent: "EU",
					Country:   "DE",
					City:      "Berlin",
					ASN:       123,
					Network:   "Network 1",
				},
				Result: model.ResultData{
					RawOutput: "Ping Results 1",
				},
			},

			{
				Probe: model.ProbeData{
					Continent: "NA",
					Country:   "US",
					City:      "New York",
					State:     "NY",
					ASN:       567,
					Network:   "Network 2",
				},
				Result: model.ResultData{
					RawOutput: "Ping Results 2",
				},
			},
		},
	}

	OutputDefault(id, data, ctx, m)
	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> EU, DE, Berlin, ASN:123, Network 1\n> NA, US, (NY), New York, ASN:567, Network 2\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "Ping Results 1\n\nPing Results 2\n", string(outContent))
}
