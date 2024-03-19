package view

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
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

	m := &globalping.MeasurementCreate{
		Options: &globalping.MeasurementOptions{
			Request: &globalping.RequestOptions{
				Method: "GET",
			},
		},
	}

	w := new(bytes.Buffer)
	viewer := NewViewer(&Context{
		Cmd:    "http",
		CIMode: true,
	}, NewPrinter(nil, w, w), nil, gbMock)

	viewer.Output(measurementID1, m)

	assert.Equal(t, `> Berlin, DE, EU, Network 1 (AS123)
Body 1

> New York (NY), US, NA, Network 2 (AS567)
Body 2
`, w.String())
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

	m := &globalping.MeasurementCreate{
		Options: &globalping.MeasurementOptions{
			Request: &globalping.RequestOptions{
				Method: "GET",
			},
		},
	}
	w := new(bytes.Buffer)
	viewer := NewViewer(&Context{
		Cmd:    "http",
		CIMode: true,
		Share:  true,
	}, NewPrinter(nil, w, w), nil, gbMock)

	viewer.Output(measurementID1, m)

	assert.Equal(t, fmt.Sprintf(`> Berlin, DE, EU, Network 1 (AS123)
Body 1

> New York (NY), US, NA, Network 2 (AS567)
Body 2
> View the results online: https://www.jsdelivr.com/globalping?measurement=%s
`, measurementID1), w.String())
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

	m := &globalping.MeasurementCreate{
		Options: &globalping.MeasurementOptions{
			Request: &globalping.RequestOptions{
				Method: "GET",
			},
		},
	}
	w := new(bytes.Buffer)
	viewer := NewViewer(&Context{
		Cmd:    "http",
		CIMode: true,
		Full:   true,
	}, NewPrinter(nil, w, w), nil, gbMock)

	viewer.Output(measurementID1, m)

	assert.Equal(t, `> Berlin, DE, EU, Network 1 (AS123)
Headers 1
Body 1

> New York (NY), US, NA, Network 2 (AS567)
Headers 2
Body 2
`, w.String())
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

	m := &globalping.MeasurementCreate{
		Options: &globalping.MeasurementOptions{
			Request: &globalping.RequestOptions{
				Method: "HEAD",
			},
		},
	}
	w := new(bytes.Buffer)
	viewer := NewViewer(&Context{
		Cmd:    "http",
		CIMode: true,
	}, NewPrinter(nil, w, w), nil, gbMock)

	viewer.Output(measurementID1, m)

	assert.Equal(t, `> Berlin, DE, EU, Network 1 (AS123)
Headers 1

> New York (NY), US, NA, Network 2 (AS567)
Headers 2
`, w.String())
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

	m := &globalping.MeasurementCreate{}
	w := new(bytes.Buffer)
	viewer := NewViewer(&Context{
		Cmd:    "ping",
		CIMode: true,
	}, NewPrinter(nil, w, w), nil, gbMock)

	viewer.Output(measurementID1, m)

	assert.Equal(t, `> Berlin, DE, EU, Network 1 (AS123)
Ping Results 1

> New York (NY), US, NA, Network 2 (AS567)
Ping Results 2
`, w.String())
}
