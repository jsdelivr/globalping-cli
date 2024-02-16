package cmd

import (
	"context"
	"errors"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/mocks"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_Execute_Ping_Default(t *testing.T) {
	t.Cleanup(sessionCleanup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := getPingMeasurementCreate()
	expectedResponse := getMeasurementCreateResponse(measurementID1)

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(expectedOpts).Times(1).Return(expectedResponse, false, nil)

	viewerMock := mocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().Output(measurementID1, expectedOpts).Times(1).Return(nil)

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime)

	ctx := &view.Context{
		MaxHistory: 1,
	}
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	defer r.Close()
	defer w.Close()

	printer := view.NewPrinter(nil, w, w)
	root := NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)
	os.Args = []string{"globalping", "ping", "jsdelivr.com"}
	err = root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)
	w.Close()

	output, err := io.ReadAll(r)
	assert.NoError(t, err)
	assert.Equal(t, "", string(output))

	expectedCtx := getExpectedPingViewContext()
	assert.Equal(t, expectedCtx, ctx)

	b, err := os.ReadFile(getMeasurementsPath())
	assert.NoError(t, err)
	expectedHistory := []byte(measurementID1 + "\n")
	assert.Equal(t, expectedHistory, b)
}

func Test_Execute_Ping_Infinite(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on windows") // Signal(syscall.SIGINT) is not supported on Windows
	}
	t.Cleanup(sessionCleanup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts1 := getPingMeasurementCreate()
	expectedOpts1.Options.Packets = 16
	expectedOpts2 := getPingMeasurementCreate()
	expectedOpts2.Options.Packets = 16
	expectedOpts2.Locations[0].Magic = measurementID1
	expectedOpts3 := getPingMeasurementCreate()
	expectedOpts3.Options.Packets = 16
	expectedOpts3.Locations[0].Magic = measurementID2

	expectedResponse1 := getMeasurementCreateResponse(measurementID1)
	expectedResponse2 := getMeasurementCreateResponse(measurementID2)
	expectedResponse3 := getMeasurementCreateResponse(measurementID3)

	gbMock := mocks.NewMockClient(ctrl)
	call1 := gbMock.EXPECT().CreateMeasurement(expectedOpts1).Return(expectedResponse1, false, nil)
	call2 := gbMock.EXPECT().CreateMeasurement(expectedOpts2).Return(expectedResponse2, false, nil).After(call1)
	gbMock.EXPECT().CreateMeasurement(expectedOpts3).Return(expectedResponse3, false, nil).After(call2)

	viewerMock := mocks.NewMockViewer(ctrl)
	outputCall1 := viewerMock.EXPECT().OutputInfinite(measurementID1).DoAndReturn(func(id string) error {
		time.Sleep(2 * time.Millisecond)
		return nil
	})
	outputCall2 := viewerMock.EXPECT().OutputInfinite(measurementID2).DoAndReturn(func(id string) error {
		time.Sleep(2 * time.Millisecond)
		return nil
	}).After(outputCall1)
	viewerMock.EXPECT().OutputInfinite(measurementID3).DoAndReturn(func(id string) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}).After(outputCall2)
	viewerMock.EXPECT().OutputSummary().Times(1)

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).Times(3)

	ctx := &view.Context{}
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	defer r.Close()
	defer w.Close()

	printer := view.NewPrinter(nil, w, w)
	root := NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "--infinite"}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)
	go func() {
		time.Sleep(7 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		p.Signal(syscall.SIGINT)
	}()
	err = root.Cmd.ExecuteContext(context.TODO())
	<-sig

	assert.NoError(t, err)
	w.Close()

	output, err := io.ReadAll(r)
	assert.NoError(t, err)
	assert.Equal(t, "", string(output))

	expectedCtx := &view.Context{
		Cmd:        "ping",
		Target:     "jsdelivr.com",
		Limit:      1,
		CIMode:     true,
		Infinite:   true,
		CallCount:  3,
		From:       measurementID2,
		Packets:    16,
		MStartedAt: defaultCurrentTime,
	}
	assert.Equal(t, expectedCtx, ctx)

	b, err := os.ReadFile(getMeasurementsPath())
	assert.NoError(t, err)
	expectedHistory := []byte(measurementID1 + "\n")
	assert.Equal(t, expectedHistory, b)
}

func Test_Execute_Ping_Infinite_Output_Error(t *testing.T) {
	t.Cleanup(sessionCleanup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts1 := getPingMeasurementCreate()
	expectedOpts1.Options.Packets = 16

	expectedResponse1 := getMeasurementCreateResponse(measurementID1)

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(expectedOpts1).Return(expectedResponse1, false, nil)

	viewerMock := mocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().OutputInfinite(measurementID1).Return(errors.New("error message"))
	viewerMock.EXPECT().OutputSummary().Times(0)

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime)

	ctx := &view.Context{}
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	defer r.Close()
	defer w.Close()

	printer := view.NewPrinter(nil, w, w)
	root := NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "--infinite"}
	err = root.Cmd.ExecuteContext(context.TODO())
	assert.Equal(t, "error message", err.Error())
	w.Close()

	output, err := io.ReadAll(r)
	assert.NoError(t, err)
	assert.Equal(t, "Error: error message\n", string(output))

	expectedCtx := &view.Context{
		Cmd:        "ping",
		Target:     "jsdelivr.com",
		Limit:      1,
		CIMode:     true,
		Infinite:   true,
		CallCount:  1,
		From:       measurementID1,
		Packets:    16,
		MStartedAt: defaultCurrentTime,
	}
	assert.Equal(t, expectedCtx, ctx)

	b, err := os.ReadFile(getMeasurementsPath())
	assert.NoError(t, err)
	expectedHistory := []byte(measurementID1 + "\n")
	assert.Equal(t, expectedHistory, b)
}

func getExpectedPingViewContext() *view.Context {
	return &view.Context{
		Cmd:        "ping",
		Target:     "jsdelivr.com",
		Limit:      1,
		CIMode:     true,
		CallCount:  1,
		From:       "world",
		MaxHistory: 1,
		MStartedAt: defaultCurrentTime,
	}
}

func getPingMeasurementCreate() *globalping.MeasurementCreate {
	return &globalping.MeasurementCreate{
		Type:    "ping",
		Target:  "jsdelivr.com",
		Limit:   1,
		Options: &globalping.MeasurementOptions{},
		Locations: []globalping.Locations{
			{Magic: "world"},
		},
	}
}
