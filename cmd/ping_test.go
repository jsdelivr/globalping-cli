package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
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

	expectedOpts := createDefaultMeasurementCreate("ping")
	expectedOpts.Locations[0].Magic = "world"
	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(expectedOpts).Times(1).Return(expectedResponse, nil)

	viewerMock := mocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().Output(measurementID1, expectedOpts).Times(1).Return(nil)

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("ping")
	root := NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)

	os.Args = []string{"globalping", "ping", "jsdelivr.com"}
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("ping")
	expectedCtx.From = "world"
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
		"ping jsdelivr.com",
	)
	assert.Equal(t, expectedHistory, string(b))
}

func Test_Execute_Ping_Locations_And_Session(t *testing.T) {
	t.Cleanup(sessionCleanup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("ping")
	expectedOpts.Locations = append(expectedOpts.Locations, globalping.Locations{Magic: "New York"})
	expectedResponse := createDefaultMeasurementCreateResponse()

	totalCalls := 10
	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(expectedOpts).Times(totalCalls).Return(expectedResponse, nil)

	viewerMock := mocks.NewMockViewer(ctrl)
	c1 := viewerMock.EXPECT().Output(measurementID1, expectedOpts).Times(4).Return(nil)
	c2 := viewerMock.EXPECT().Output(measurementID2, expectedOpts).Times(3).Return(nil).After(c1)
	viewerMock.EXPECT().Output(measurementID3, expectedOpts).Times(3).Return(nil).After(c2)

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("ping")
	root := NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "Berlin,New York "}
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	expectedCtx := createDefaultExpectedContext("ping")
	expectedCtx.From = "Berlin,New York"
	assert.Equal(t, expectedCtx, ctx)

	ctx = createDefaultContext("ping")
	expectedOpts.Locations = []globalping.Locations{{Magic: measurementID1}}
	root = NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "@-1"}
	err = root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	expectedCtx.From = "@-1"
	expectedCtx.IsLocationFromSession = true
	assert.Equal(t, expectedCtx, ctx)

	ctx = createDefaultContext("ping")
	expectedOpts.Locations = []globalping.Locations{{Magic: measurementID1}}
	root = NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "last"}
	err = root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	expectedCtx.From = "last"
	assert.Equal(t, expectedCtx, ctx)

	ctx = createDefaultContext("ping")
	expectedOpts.Locations = []globalping.Locations{{Magic: measurementID1}}
	root = NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "previous"}
	err = root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	expectedCtx.From = "previous"
	assert.Equal(t, expectedCtx, ctx)

	ctx = createDefaultContext("ping")
	expectedOpts.Locations = []globalping.Locations{{Magic: "world"}}
	expectedResponse.ID = measurementID2
	root = NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)
	os.Args = []string{"globalping", "ping", "jsdelivr.com"}
	err = root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	expectedCtx.From = "world"
	expectedCtx.History.Slice[0].Id = measurementID2
	expectedCtx.IsLocationFromSession = false
	assert.Equal(t, expectedCtx, ctx)

	ctx = createDefaultContext("ping")
	expectedOpts.Locations = []globalping.Locations{{Magic: measurementID1}}
	root = NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "@1"}
	err = root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	expectedCtx.From = "@1"
	expectedCtx.IsLocationFromSession = true
	assert.Equal(t, expectedCtx, ctx)

	ctx = createDefaultContext("ping")
	expectedOpts.Locations = []globalping.Locations{{Magic: measurementID1}}
	root = NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "first"}
	err = root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	expectedCtx.From = "first"
	assert.Equal(t, expectedCtx, ctx)

	ctx = createDefaultContext("ping")
	expectedOpts.Locations = []globalping.Locations{{Magic: "world"}}
	expectedResponse.ID = measurementID3
	root = NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)
	os.Args = []string{"globalping", "ping", "jsdelivr.com"}
	err = root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	expectedCtx.From = "world"
	expectedCtx.History.Slice[0].Id = measurementID3
	expectedCtx.IsLocationFromSession = false
	assert.Equal(t, expectedCtx, ctx)

	ctx = createDefaultContext("ping")
	expectedOpts.Locations = []globalping.Locations{{Magic: measurementID2}}
	root = NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "@2"}
	err = root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	expectedCtx.From = "@2"
	expectedCtx.RecordToSession = false
	expectedCtx.IsLocationFromSession = true
	assert.Equal(t, expectedCtx, ctx)

	ctx = createDefaultContext("ping")
	expectedOpts.Locations = []globalping.Locations{{Magic: measurementID1}}
	root = NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "@-3"}
	err = root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	expectedCtx.From = "@-3"
	assert.Equal(t, expectedCtx, ctx)

	assert.Equal(t, "", w.String())

	b, err := os.ReadFile(getMeasurementsPath())
	assert.NoError(t, err)
	expectedHistory := []byte(measurementID1 + "\n" + measurementID2 + "\n" + measurementID3 + "\n")
	assert.Equal(t, expectedHistory, b)

	ctx = createDefaultContext("ping")
	root = NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "@-4"}
	err = root.Cmd.ExecuteContext(context.TODO())
	assert.Error(t, err, ErrIndexOutOfRange)

	expectedCtx.From = "@-4"
	expectedCtx.IsLocationFromSession = false
	expectedCtx.RecordToSession = true
	expectedCtx.MeasurementsCreated = 0
	expectedCtx.History = view.NewHistoryBuffer(1)
	assert.Equal(t, expectedCtx, ctx)
	assert.Equal(t, "Error: index out of range\n", w.String())

	sessionCleanup()

	w.Reset()
	ctx = createDefaultContext("ping")
	root = NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "@1"}
	err = root.Cmd.ExecuteContext(context.TODO())
	assert.Error(t, err, ErrNoPreviousMeasurements)

	expectedCtx.From = "@1"
	assert.Equal(t, expectedCtx, ctx)
	assert.Equal(t, "Error: no previous measurements found\n", w.String())

	w.Reset()
	ctx = createDefaultContext("ping")
	root = NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "@0"}
	err = root.Cmd.ExecuteContext(context.TODO())
	assert.Error(t, err, ErrInvalidIndex)

	expectedCtx.From = "@0"
	assert.Equal(t, expectedCtx, ctx)
	assert.Equal(t, "Error: invalid index\n", w.String())

	w.Reset()
	ctx = createDefaultContext("ping")
	root = NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "@x"}
	err = root.Cmd.ExecuteContext(context.TODO())
	assert.Error(t, err, ErrInvalidIndex)

	expectedCtx.From = "@x"
	assert.Equal(t, expectedCtx, ctx)
	assert.Equal(t, "Error: invalid index\n", w.String())

	w.Reset()
	ctx = createDefaultContext("ping")
	root = NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "@"}
	err = root.Cmd.ExecuteContext(context.TODO())
	assert.Error(t, err, ErrInvalidIndex)

	expectedCtx.From = "@"
	assert.Equal(t, expectedCtx, ctx)
	assert.Equal(t, "Error: invalid index\n", w.String())
}

func Test_Execute_Ping_Infinite(t *testing.T) {
	t.Cleanup(sessionCleanup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts1 := createDefaultMeasurementCreate("ping")
	expectedOpts1.Options.Packets = 16
	expectedOpts2 := createDefaultMeasurementCreate("ping")
	expectedOpts2.Options.Packets = 16
	expectedOpts2.Locations[0].Magic = measurementID1
	expectedOpts3 := createDefaultMeasurementCreate("ping")
	expectedOpts3.Options.Packets = 16
	expectedOpts3.Locations[0].Magic = measurementID2
	expectedOpts4 := createDefaultMeasurementCreate("ping")
	expectedOpts4.Options.Packets = 16
	expectedOpts4.Locations[0].Magic = measurementID3

	expectedResponse1 := createDefaultMeasurementCreateResponse()
	expectedResponse2 := createDefaultMeasurementCreateResponse()
	expectedResponse2.ID = measurementID2
	expectedResponse3 := createDefaultMeasurementCreateResponse()
	expectedResponse3.ID = measurementID3
	expectedResponse4 := createDefaultMeasurementCreateResponse()
	expectedResponse4.ID = measurementID4

	gbMock := mocks.NewMockClient(ctrl)
	createCall1 := gbMock.EXPECT().CreateMeasurement(expectedOpts1).Return(expectedResponse1, nil)
	createCall2 := gbMock.EXPECT().CreateMeasurement(expectedOpts2).Return(expectedResponse2, nil).After(createCall1)
	createCall3 := gbMock.EXPECT().CreateMeasurement(expectedOpts3).Return(expectedResponse3, nil).After(createCall2)
	gbMock.EXPECT().CreateMeasurement(expectedOpts4).Return(expectedResponse4, nil).After(createCall3)

	expectedMeasurement1 := createDefaultMeasurement_MultipleProbes("ping", globalping.StatusFinished)
	expectedMeasurement2 := createDefaultMeasurement_MultipleProbes("ping", globalping.StatusInProgress)
	expectedMeasurement2.ID = measurementID2
	expectedMeasurement2.Results[0].Result.Status = globalping.StatusFinished
	expectedMeasurement3 := createDefaultMeasurement_MultipleProbes("ping", globalping.StatusInProgress)
	expectedMeasurement3.ID = measurementID3
	expectedMeasurement3.Results[0].Result.Status = globalping.StatusFinished
	expectedMeasurement4 := createDefaultMeasurement_MultipleProbes("ping", globalping.StatusInProgress)
	expectedMeasurement4.ID = measurementID4
	expectedMeasurement4.Results[1].Result.Status = globalping.StatusFinished

	getCall1 := gbMock.EXPECT().GetMeasurement(measurementID1).Return(expectedMeasurement1, nil)
	getCall2 := gbMock.EXPECT().GetMeasurement(measurementID2).Return(expectedMeasurement2, nil).After(getCall1)
	getCall3 := gbMock.EXPECT().GetMeasurement(measurementID3).Return(expectedMeasurement3, nil).After(getCall2)
	getCall4 := gbMock.EXPECT().GetMeasurement(measurementID4).Return(expectedMeasurement4, nil).After(getCall3)
	getCall5 := gbMock.EXPECT().GetMeasurement(measurementID2).Return(expectedMeasurement2, nil).After(getCall4)
	getCall6 := gbMock.EXPECT().GetMeasurement(measurementID3).Return(expectedMeasurement3, nil).After(getCall5)
	gbMock.EXPECT().GetMeasurement(measurementID4).Return(expectedMeasurement4, nil).After(getCall6)

	viewerMock := mocks.NewMockViewer(ctrl)
	waitFn := func(m *globalping.Measurement) error { time.Sleep(5 * time.Millisecond); return nil }
	outputCall1 := viewerMock.EXPECT().OutputInfinite(expectedMeasurement1).DoAndReturn(waitFn)
	outputCall2 := viewerMock.EXPECT().OutputInfinite(expectedMeasurement2).DoAndReturn(waitFn).After(outputCall1)
	outputCall3 := viewerMock.EXPECT().OutputInfinite(expectedMeasurement3).DoAndReturn(waitFn).After(outputCall2)
	outputCall4 := viewerMock.EXPECT().OutputInfinite(expectedMeasurement4).DoAndReturn(waitFn).After(outputCall3)
	outputCall5 := viewerMock.EXPECT().OutputInfinite(expectedMeasurement2).DoAndReturn(waitFn).After(outputCall4)
	outputCall6 := viewerMock.EXPECT().OutputInfinite(expectedMeasurement3).DoAndReturn(waitFn).After(outputCall5)
	viewerMock.EXPECT().OutputInfinite(expectedMeasurement4).DoAndReturn(func(m *globalping.Measurement) error {
		time.Sleep(500 * time.Millisecond)
		return nil
	}).After(outputCall6)

	viewerMock.EXPECT().OutputSummary().Times(1)
	viewerMock.EXPECT().OutputShare().Times(1)

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := &view.Context{
		History: view.NewHistoryBuffer(10),
		From:    "world",
		Limit:   1,
	}
	root := NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "--infinite", "from", "Berlin"}

	go func() {
		time.Sleep(150 * time.Millisecond)
		root.cancel <- syscall.SIGINT
	}()
	err := root.Cmd.ExecuteContext(context.TODO())

	assert.NoError(t, err)
	assert.Equal(t, "", w.String())

	expectedCtx := &view.Context{
		Cmd:                 "ping",
		Target:              "jsdelivr.com",
		From:                "Berlin",
		Limit:               1,
		Packets:             16,
		Infinite:            true,
		CIMode:              true,
		MeasurementsCreated: 4,
		RunSessionStartedAt: defaultCurrentTime,
	}
	expectedCtx.History = &view.HistoryBuffer{
		Index: 4,
		Slice: []*view.HistoryItem{
			{
				Id:        measurementID1,
				Status:    globalping.StatusFinished,
				StartedAt: defaultCurrentTime,
			},
			{
				Id:     measurementID2,
				Status: globalping.StatusInProgress,
				ProbeStatus: []globalping.MeasurementStatus{
					globalping.StatusFinished,
					globalping.StatusInProgress,
					globalping.StatusInProgress,
				},
				StartedAt: defaultCurrentTime,
			},
			{
				Id:     measurementID3,
				Status: globalping.StatusInProgress,
				ProbeStatus: []globalping.MeasurementStatus{
					globalping.StatusFinished,
					globalping.StatusInProgress,
					globalping.StatusInProgress,
				},
				StartedAt: defaultCurrentTime,
			},
			{
				Id:     measurementID4,
				Status: globalping.StatusInProgress,
				ProbeStatus: []globalping.MeasurementStatus{
					globalping.StatusInProgress,
					globalping.StatusFinished,
					globalping.StatusInProgress,
				},
				StartedAt: defaultCurrentTime,
			},
			nil, nil, nil, nil, nil, nil,
		},
	}
	assert.Equal(t, expectedCtx, ctx)

	b, err := os.ReadFile(getMeasurementsPath())
	assert.NoError(t, err)
	expectedHistory := measurementID1 + "\n"
	assert.Equal(t, expectedHistory, string(b))

	b, err = os.ReadFile(getHistoryPath())
	assert.NoError(t, err)
	expectedHistory = createDefaultExpectedHistoryLogItem(
		"1",
		measurementID1+"."+measurementID2+"."+measurementID3+"."+measurementID4,
		"ping jsdelivr.com --infinite from Berlin",
	)
	assert.Equal(t, expectedHistory, string(b))
}

func Test_Execute_Ping_Infinite_Output_Error(t *testing.T) {
	t.Cleanup(sessionCleanup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts1 := createDefaultMeasurementCreate("ping")
	expectedOpts1.Options.Packets = 16

	expectedResponse1 := createDefaultMeasurementCreateResponse()

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(expectedOpts1).Return(expectedResponse1, nil)

	expectedMeasurement := createDefaultMeasurement("ping")
	gbMock.EXPECT().GetMeasurement(measurementID1).Return(expectedMeasurement, nil)

	viewerMock := mocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().OutputInfinite(expectedMeasurement).Return(errors.New("error message"))
	viewerMock.EXPECT().OutputSummary().Times(1)
	viewerMock.EXPECT().OutputShare().Times(1)

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("ping")
	root := NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "--infinite", "from", "Berlin"}
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.Equal(t, "error message", err.Error())

	assert.Equal(t, "Error: error message\n", w.String())

	expectedCtx := createDefaultExpectedContext("ping")
	expectedCtx.History.Find(measurementID1).Status = globalping.StatusFinished
	expectedCtx.Packets = 16
	expectedCtx.Infinite = true
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
		"ping jsdelivr.com --infinite from Berlin",
	)
	assert.Equal(t, expectedHistory, string(b))
}

func Test_Execute_Ping_Infinite_Output_TooManyRequests_Error(t *testing.T) {
	t.Cleanup(sessionCleanup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts1 := createDefaultMeasurementCreate("ping")
	expectedOpts1.Options.Packets = 16
	expectedOpts2 := createDefaultMeasurementCreate("ping")
	expectedOpts2.Options.Packets = 16
	expectedOpts2.Locations[0].Magic = measurementID1

	expectedResponse1 := createDefaultMeasurementCreateResponse()

	gbMock := mocks.NewMockClient(ctrl)
	createCall1 := gbMock.EXPECT().CreateMeasurement(expectedOpts1).Return(expectedResponse1, nil)
	gbMock.EXPECT().CreateMeasurement(expectedOpts2).Return(nil, &globalping.MeasurementError{
		Code:    429,
		Message: "too many requests",
	}).After(createCall1)

	expectedMeasurement := createDefaultMeasurement("ping")
	gbMock.EXPECT().GetMeasurement(measurementID1).Return(expectedMeasurement, nil)

	viewerMock := mocks.NewMockViewer(ctrl)
	waitFn := func(m *globalping.Measurement) error { time.Sleep(5 * time.Millisecond); return nil }
	viewerMock.EXPECT().OutputInfinite(expectedMeasurement).DoAndReturn(waitFn)

	viewerMock.EXPECT().OutputSummary().Times(1)
	viewerMock.EXPECT().OutputShare().Times(1)

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	errW := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, errW)
	ctx := createDefaultContext("ping")
	root := NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "Berlin", "--infinite", "--share"}
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.Equal(t, "too many requests", err.Error())

	assert.Equal(t, "> too many requests\n", errW.String())
	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("ping")
	expectedCtx.History.Find(measurementID1).Status = globalping.StatusFinished
	expectedCtx.Packets = 16
	expectedCtx.Infinite = true
	expectedCtx.Share = true
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
		"ping jsdelivr.com from Berlin --infinite --share",
	)
	assert.Equal(t, expectedHistory, string(b))
}

func Test_Execute_Ping_IPv4(t *testing.T) {
	t.Cleanup(sessionCleanup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("ping")
	expectedOpts.Locations[0].Magic = "world"
	expectedOpts.Options.IPVersion = globalping.IPVersion4
	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(expectedOpts).Times(1).Return(expectedResponse, nil)

	viewerMock := mocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().Output(measurementID1, expectedOpts).Times(1).Return(nil)

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("ping")
	root := NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)

	os.Args = []string{"globalping", "ping", "jsdelivr.com", "--ipv4"}
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("ping")
	expectedCtx.From = "world"
	expectedCtx.Ipv4 = true
	assert.Equal(t, expectedCtx, ctx)
}

func Test_Execute_Ping_IPv6(t *testing.T) {
	t.Cleanup(sessionCleanup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("ping")
	expectedOpts.Locations[0].Magic = "world"
	expectedOpts.Options.IPVersion = globalping.IPVersion6
	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(expectedOpts).Times(1).Return(expectedResponse, nil)

	viewerMock := mocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().Output(measurementID1, expectedOpts).Times(1).Return(nil)

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("ping")
	root := NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)

	os.Args = []string{"globalping", "ping", "jsdelivr.com", "--ipv6"}
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("ping")
	expectedCtx.From = "world"
	expectedCtx.Ipv6 = true
	assert.Equal(t, expectedCtx, ctx)
}
