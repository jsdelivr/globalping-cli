package cmd

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/jsdelivr/globalping-cli/mocks"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_Execute_Traceroute_Default(t *testing.T) {
	t.Cleanup(sessionCleanup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("traceroute")
	expectedOpts.Limit = 2
	expectedOpts.Options.Protocol = "tcp"
	expectedOpts.Options.Port = 99

	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(expectedOpts).Times(1).Return(expectedResponse, false, nil)

	viewerMock := mocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().Output(measurementID1, expectedOpts).Times(1).Return(nil)

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("traceroute")
	root := NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)
	os.Args = []string{"globalping", "traceroute", "jsdelivr.com",
		"from", "Berlin",
		"--limit", "2",
		"--protocol", "tcp",
		"--port", "99",
	}
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("traceroute")
	expectedCtx.Limit = 2
	expectedCtx.Protocol = "tcp"
	expectedCtx.Port = 99
	assert.Equal(t, expectedCtx, ctx)

	b, err := os.ReadFile(getMeasurementsPath())
	assert.NoError(t, err)
	expectedHistory := measurementID1 + "\n"
	assert.Equal(t, expectedHistory, string(b))

	b, err = os.ReadFile(getHistoryPath())
	assert.NoError(t, err)
	expectedHistory = createDefaultExpectedHistoryLogItem(
		measurementID1,
		"traceroute jsdelivr.com from Berlin --limit 2 --protocol tcp --port 99",
	)
	assert.Equal(t, expectedHistory, string(b))
}
