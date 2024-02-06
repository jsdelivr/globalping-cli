package view

import (
	"io"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_Output_Default_HTTP_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
					RawBody:    "Body 1",
				},
			},

			{
				Probe: globalping.ProbeDetails{
					Continent: "NA",
					Country:   "US",
					City:      "New York",
					State:     "NY",
					ASN:       567,
					Network:   "Network 2",
				},
				Result: globalping.ProbeResult{
					RawOutput:  "Headers 2\nBody 2",
					RawHeaders: "Headers 2",
					RawBody:    "Body 2",
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

	m := &globalping.MeasurementCreate{
		Options: &globalping.MeasurementOptions{
			Request: &globalping.RequestOptions{
				Method: "GET",
			},
		},
	}

	viewer := NewViewer(&Context{
		Cmd: "http",
		CI:  true,
	}, gbMock)

	viewer.Output(measurementID1, m)
	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> EU, DE, Berlin, ASN:123, Network 1\n> NA, US, (NY), New York, ASN:567, Network 2\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "Body 1\n\nBody 2\n", string(outContent))
}

func Test_Output_Default_HTTP_Get_Share(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
					RawBody:    "Body 1",
				},
			},

			{
				Probe: globalping.ProbeDetails{
					Continent: "NA",
					Country:   "US",
					City:      "New York",
					State:     "NY",
					ASN:       567,
					Network:   "Network 2",
				},
				Result: globalping.ProbeResult{
					RawOutput:  "Headers 2\nBody 2",
					RawHeaders: "Headers 2",
					RawBody:    "Body 2",
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

	m := &globalping.MeasurementCreate{
		Options: &globalping.MeasurementOptions{
			Request: &globalping.RequestOptions{
				Method: "GET",
			},
		},
	}

	viewer := NewViewer(&Context{
		Cmd:   "http",
		CI:    true,
		Share: true,
	}, gbMock)

	viewer.Output(measurementID1, m)
	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> EU, DE, Berlin, ASN:123, Network 1\n> NA, US, (NY), New York, ASN:567, Network 2\n> View the results online: https://www.jsdelivr.com/globalping?measurement=nzGzfAGL7sZfUs3c\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "Body 1\n\nBody 2\n", string(outContent))
}

func Test_Output_Default_HTTP_Get_Full(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
					RawBody:    "Body 1",
				},
			},

			{
				Probe: globalping.ProbeDetails{
					Continent: "NA",
					Country:   "US",
					City:      "New York",
					State:     "NY",
					ASN:       567,
					Network:   "Network 2",
				},
				Result: globalping.ProbeResult{
					RawOutput:  "Headers 2\nBody 2",
					RawHeaders: "Headers 2",
					RawBody:    "Body 2",
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

	m := &globalping.MeasurementCreate{
		Options: &globalping.MeasurementOptions{
			Request: &globalping.RequestOptions{
				Method: "GET",
			},
		},
	}

	viewer := NewViewer(&Context{
		Cmd:  "http",
		CI:   true,
		Full: true,
	}, gbMock)

	viewer.Output(measurementID1, m)

	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> EU, DE, Berlin, ASN:123, Network 1\n> NA, US, (NY), New York, ASN:567, Network 2\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "Headers 1\nBody 1\n\nHeaders 2\nBody 2\n", string(outContent))
}

func Test_Output_Default_HTTP_Head(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
					State:     "NY",
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

	m := &globalping.MeasurementCreate{
		Options: &globalping.MeasurementOptions{
			Request: &globalping.RequestOptions{
				Method: "HEAD",
			},
		},
	}

	viewer := NewViewer(&Context{
		Cmd: "http",
		CI:  true,
	}, gbMock)

	viewer.Output(measurementID1, m)

	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> EU, DE, Berlin, ASN:123, Network 1\n> NA, US, (NY), New York, ASN:567, Network 2\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "Headers 1\n\nHeaders 2\n", string(outContent))
}

func Test_Output_Default_Ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
					State:     "NY",
					ASN:       567,
					Network:   "Network 2",
				},
				Result: globalping.ProbeResult{
					RawOutput: "Ping Results 2",
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

	m := &globalping.MeasurementCreate{}

	viewer := NewViewer(&Context{
		Cmd: "ping",
		CI:  true,
	}, gbMock)

	viewer.Output(measurementID1, m)
	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> EU, DE, Berlin, ASN:123, Network 1\n> NA, US, (NY), New York, ASN:567, Network 2\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "Ping Results 1\n\nPing Results 2\n", string(outContent))
}
