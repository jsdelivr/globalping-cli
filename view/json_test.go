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

func Test_Output_Json(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	b := []byte(`{"fake": "results"}`)

	gbMock := mocks.NewMockClient(ctrl)
	measurement := getPingGetMeasurement(measurementID1)
	gbMock.EXPECT().GetMeasurement(measurementID1).Times(1).Return(measurement, nil)
	gbMock.EXPECT().GetRawMeasurement(measurementID1).Times(1).Return(b, nil)

	viewer := NewViewer(&Context{
		ToJSON: true,
		Share:  true,
	}, gbMock)

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
	err = viewer.Output(measurementID1, m)
	assert.NoError(t, err)
	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> View the results online: https://www.jsdelivr.com/globalping?measurement=nzGzfAGL7sZfUs3c\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "{\"fake\": \"results\"}\n\n", string(outContent))
}
