package cmd

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/mocks"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_Execute_MTR_Default(t *testing.T) {
	t.Cleanup(sessionCleanup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := &globalping.MeasurementCreate{
		Type:   "mtr",
		Target: "jsdelivr.com",
		Limit:  2,
		Options: &globalping.MeasurementOptions{
			Protocol: "tcp",
			Port:     99,
			Packets:  16,
		},
		Locations: []globalping.Locations{
			{Magic: "Berlin"},
		},
	}
	expectedResponse := getMeasurementCreateResponse(measurementID1)

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(expectedOpts).Times(1).Return(expectedResponse, false, nil)

	viewerMock := mocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().Output(measurementID1, expectedOpts).Times(1).Return(nil)

	ctx := &view.Context{
		MaxHistory: 1,
	}
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	defer r.Close()
	defer w.Close()

	printer := view.NewPrinter(nil, w, w)
	root := NewRoot(printer, ctx, viewerMock, nil, gbMock, nil)
	os.Args = []string{"globalping", "mtr", "jsdelivr.com",
		"from", "Berlin",
		"--limit", "2",
		"--protocol", "tcp",
		"--port", "99",
		"--packets", "16",
	}
	err = root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)
	w.Close()

	output, err := io.ReadAll(r)
	assert.NoError(t, err)
	assert.Equal(t, "", string(output))

	expectedCtx := &view.Context{
		Cmd:        "mtr",
		Target:     "jsdelivr.com",
		From:       "Berlin",
		Limit:      2,
		Protocol:   "tcp",
		Port:       99,
		Packets:    16,
		CIMode:     true,
		MaxHistory: 1,
	}
	assert.Equal(t, expectedCtx, ctx)

	b, err := os.ReadFile(getMeasurementsPath())
	assert.NoError(t, err)
	expectedHistory := []byte(measurementID1 + "\n")
	assert.Equal(t, expectedHistory, b)
}
