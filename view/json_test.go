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

func Test_Output_Json(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	b := []byte(`{"fake": "results"}`)

	gbMock := mocks.NewMockClient(ctrl)
	measurement := createPingMeasurement(measurementID1)
	gbMock.EXPECT().GetMeasurement(measurementID1).Times(1).Return(measurement, nil)
	gbMock.EXPECT().GetMeasurementRaw(measurementID1).Times(1).Return(b, nil)

	w := new(bytes.Buffer)
	viewer := NewViewer(
		&Context{
			ToJSON: true,
			Share:  true,
			CIMode: true,
		},
		NewPrinter(nil, w, w),
		nil,
		gbMock,
	)

	m := &globalping.MeasurementCreate{}
	err := viewer.Output(measurementID1, m)
	assert.NoError(t, err)

	assert.Equal(t, fmt.Sprintf(`{"fake": "results"}
> View the results online: https://www.jsdelivr.com/globalping?measurement=%s

`, measurementID1), w.String())
}
