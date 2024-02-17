package view

import (
	"io"
	"os"
	"testing"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_Output_Json(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	b := []byte(`{"fake": "results"}`)

	gbMock := mocks.NewMockClient(ctrl)
	measurement := getPingGetMeasurement(measurementID1)
	gbMock.EXPECT().GetMeasurement(measurementID1).Times(1).Return(measurement, nil)
	gbMock.EXPECT().GetMeasurementRaw(measurementID1).Times(1).Return(b, nil)

	r, w, err := os.Pipe()
	assert.NoError(t, err)
	defer r.Close()
	defer w.Close()

	viewer := NewViewer(
		&Context{
			ToJSON: true,
			Share:  true,
		},
		NewPrinter(nil, w, w),
		nil,
		gbMock,
	)

	m := &globalping.MeasurementCreate{}
	err = viewer.Output(measurementID1, m)
	assert.NoError(t, err)
	w.Close()

	outContent, err := io.ReadAll(r)
	assert.NoError(t, err)
	assert.Equal(t, `{"fake": "results"}
> View the results online: https://www.jsdelivr.com/globalping?measurement=nzGzfAGL7sZfUs3c

`, string(outContent))
}
