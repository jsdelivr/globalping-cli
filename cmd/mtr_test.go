package cmd

import (
	"bytes"
	"context"
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

	expectedOpts := createDefaultMeasurementCreate("mtr")
	expectedOpts.Limit = 2
	expectedOpts.Options.Protocol = "tcp"
	expectedOpts.Options.Port = 99
	expectedOpts.Options.Packets = 16

	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(expectedOpts).Times(1).Return(expectedResponse, nil)

	viewerMock := mocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().Output(measurementID1, expectedOpts).Times(1).Return(nil)

	utilsMock := mocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("mtr")
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, nil)
	os.Args = []string{"globalping", "mtr", "jsdelivr.com",
		"from", "Berlin",
		"--limit", "2",
		"--protocol", "tcp",
		"--port", "99",
		"--packets", "16",
	}
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("mtr")
	expectedCtx.Limit = 2
	expectedCtx.Protocol = "tcp"
	expectedCtx.Port = 99
	expectedCtx.Packets = 16

	assert.Equal(t, expectedCtx, ctx)

	b, err := os.ReadFile(getMeasurementsPath())
	assert.NoError(t, err)
	expectedHistory := measurementID1 + "\n"
	assert.Equal(t, expectedHistory, string(b))

	b, err = os.ReadFile(getHistoryPath())
	assert.NoError(t, err)
	expectedHistory = createDefaultExpectedHistoryLogItem(
		"1",
		measurementID1,
		"mtr jsdelivr.com from Berlin --limit 2 --protocol tcp --port 99 --packets 16",
	)
	assert.Equal(t, expectedHistory, string(b))
}

func Test_Execute_MTR_IPv4(t *testing.T) {
	t.Cleanup(sessionCleanup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("mtr")
	expectedOpts.Options.IPVersion = globalping.IPVersion4

	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(expectedOpts).Times(1).Return(expectedResponse, nil)

	viewerMock := mocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().Output(measurementID1, expectedOpts).Times(1).Return(nil)

	utilsMock := mocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("mtr")
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, nil)
	os.Args = []string{"globalping", "mtr", "jsdelivr.com",
		"from", "Berlin",
		"--ipv4",
	}
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("mtr")
	expectedCtx.Ipv4 = true

	assert.Equal(t, expectedCtx, ctx)
}

func Test_Execute_MTR_IPv6(t *testing.T) {
	t.Cleanup(sessionCleanup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("mtr")
	expectedOpts.Options.IPVersion = globalping.IPVersion6

	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(expectedOpts).Times(1).Return(expectedResponse, nil)

	viewerMock := mocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().Output(measurementID1, expectedOpts).Times(1).Return(nil)

	utilsMock := mocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("mtr")
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, nil)
	os.Args = []string{"globalping", "mtr", "jsdelivr.com",
		"from", "Berlin",
		"--ipv6",
	}
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("mtr")
	expectedCtx.Ipv6 = true

	assert.Equal(t, expectedCtx, ctx)
}
