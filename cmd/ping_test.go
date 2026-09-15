package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	apiMocks "github.com/jsdelivr/globalping-cli/mocks/api"
	utilsMocks "github.com/jsdelivr/globalping-cli/mocks/utils"
	viewMocks "github.com/jsdelivr/globalping-cli/mocks/view"
	"github.com/jsdelivr/globalping-cli/storage"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/jsdelivr/globalping-go"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func pingOutputFor(measurement *globalping.Measurement) gomock.Matcher {
	return gomock.Cond(func(value any) bool {
		output, ok := value.(*view.InfinitePingOutput)

		return ok && output.Measurement == measurement
	})
}

func Test_Execute_Ping_Default(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("ping")
	expectedOpts.Locations = globalping.LocationOptions{{Magic: "world"}}
	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := apiMocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(t.Context(), expectedOpts).Times(1).Return(expectedResponse, nil)

	expectedMeasurement := createDefaultMeasurement("ping")
	gbMock.EXPECT().AwaitMeasurement(t.Context(), expectedResponse.ID).Times(1).Return(expectedMeasurement, nil)

	viewerMock := viewMocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().OutputDefault(measurementID1, expectedMeasurement, expectedOpts).Times(1)

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext()
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)

	os.Args = []string{"globalping", "ping", "jsdelivr.com"}
	err := root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("ping")
	expectedCtx.From = "world"
	assert.Equal(t, expectedCtx, ctx)

	b, err := _storage.GetMeasurements()
	assert.NoError(t, err)
	expectedHistory := measurementID1 + "\n"
	assert.Equal(t, expectedHistory, string(b))

	items, err := _storage.GetHistory(0)
	assert.NoError(t, err)
	expectedHistoryItems := []string{createDefaultExpectedHistoryItem(
		"1",
		"ping jsdelivr.com",
		measurementID1,
	)}
	assert.Equal(t, expectedHistoryItems, items)
}

func Test_Execute_Ping_FiniteLatencyOutputError(t *testing.T) {
	ctrl := gomock.NewController(t)
	expectedOpts := createDefaultMeasurementCreate("ping")
	expectedOpts.Locations = globalping.LocationOptions{{Magic: "world"}}
	expectedResponse := createDefaultMeasurementCreateResponse()
	expectedMeasurement := createDefaultMeasurement("ping")
	outputErr := errors.New("latency output failed")

	gbMock := apiMocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(t.Context(), expectedOpts).Return(expectedResponse, nil)
	gbMock.EXPECT().AwaitMeasurement(t.Context(), expectedResponse.ID).Return(expectedMeasurement, nil)

	viewerMock := viewMocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().OutputLatency(measurementID1, expectedMeasurement).Return(outputErr)

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	ctx := createDefaultContext()
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(view.NewPrinter(nil, stdout, stderr), ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "--latency"}

	err := root.Cmd.ExecuteContext(t.Context())

	assert.ErrorIs(t, err, outputErr)
	assert.True(t, root.Cmd.SilenceUsage)
	assert.Empty(t, stdout.String())
	assert.Equal(t, "Error: latency output failed\n", stderr.String())
}

func Test_Execute_Ping_Locations_And_Session(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("ping")
	expectedOpts.Locations = globalping.LocationOptions{{Magic: "Berlin"}, {Magic: "New York"}}
	expectedResponse := createDefaultMeasurementCreateResponse()

	totalCalls := 10
	gbMock := apiMocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(t.Context(), expectedOpts).Times(totalCalls).Return(expectedResponse, nil)

	expectedMeasurement := createDefaultMeasurement("ping")
	gbMock.EXPECT().AwaitMeasurement(t.Context(), measurementID1).Times(4).Return(expectedMeasurement, nil)
	gbMock.EXPECT().AwaitMeasurement(t.Context(), measurementID2).Times(3).Return(expectedMeasurement, nil)
	gbMock.EXPECT().AwaitMeasurement(t.Context(), measurementID3).Times(3).Return(expectedMeasurement, nil)

	viewerMock := viewMocks.NewMockViewer(ctrl)
	c1 := viewerMock.EXPECT().OutputDefault(measurementID1, expectedMeasurement, expectedOpts).Times(4)
	c2 := viewerMock.EXPECT().OutputDefault(measurementID2, expectedMeasurement, expectedOpts).Times(3).After(c1)
	viewerMock.EXPECT().OutputDefault(measurementID3, expectedMeasurement, expectedOpts).Times(3).After(c2)

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext()
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "Berlin,New York "}
	err := root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)

	expectedCtx := createDefaultExpectedContext("ping")
	expectedCtx.From = "Berlin,New York"
	assert.Equal(t, expectedCtx, ctx)

	ctx = createDefaultContext()
	expectedOpts.Locations = globalping.PreviousMeasurementID(measurementID1)
	root = NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "@-1"}
	err = root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)

	expectedCtx.From = "@-1"
	expectedCtx.IsLocationFromSession = true
	assert.Equal(t, expectedCtx, ctx)

	ctx = createDefaultContext()
	expectedOpts.Locations = globalping.PreviousMeasurementID(measurementID1)
	root = NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "last"}
	err = root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)

	expectedCtx.From = "last"
	assert.Equal(t, expectedCtx, ctx)

	ctx = createDefaultContext()
	expectedOpts.Locations = globalping.PreviousMeasurementID(measurementID1)
	root = NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "previous"}
	err = root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)

	expectedCtx.From = "previous"
	assert.Equal(t, expectedCtx, ctx)

	ctx = createDefaultContext()
	expectedOpts.Locations = globalping.LocationOptions{{Magic: "world"}}
	expectedResponse.ID = measurementID2
	root = NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "ping", "jsdelivr.com"}
	err = root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)

	expectedCtx.From = "world"
	expectedCtx.History.Slice[0].Id = measurementID2
	expectedCtx.IsLocationFromSession = false
	assert.Equal(t, expectedCtx, ctx)

	ctx = createDefaultContext()
	expectedOpts.Locations = globalping.PreviousMeasurementID(measurementID1)
	root = NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "@1"}
	err = root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)

	expectedCtx.From = "@1"
	expectedCtx.IsLocationFromSession = true
	assert.Equal(t, expectedCtx, ctx)

	ctx = createDefaultContext()
	expectedOpts.Locations = globalping.PreviousMeasurementID(measurementID1)
	root = NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "first"}
	err = root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)

	expectedCtx.From = "first"
	assert.Equal(t, expectedCtx, ctx)

	ctx = createDefaultContext()
	expectedOpts.Locations = globalping.LocationOptions{{Magic: "world"}}
	expectedResponse.ID = measurementID3
	root = NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "ping", "jsdelivr.com"}
	err = root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)

	expectedCtx.From = "world"
	expectedCtx.History.Slice[0].Id = measurementID3
	expectedCtx.IsLocationFromSession = false
	assert.Equal(t, expectedCtx, ctx)

	ctx = createDefaultContext()
	expectedOpts.Locations = globalping.PreviousMeasurementID(measurementID2)
	root = NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "@2"}
	err = root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)

	expectedCtx.From = "@2"
	expectedCtx.RecordToSession = false
	expectedCtx.IsLocationFromSession = true
	assert.Equal(t, expectedCtx, ctx)

	ctx = createDefaultContext()
	expectedOpts.Locations = globalping.PreviousMeasurementID(measurementID1)
	root = NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "@-3"}
	err = root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)

	expectedCtx.From = "@-3"
	assert.Equal(t, expectedCtx, ctx)

	assert.Equal(t, "", w.String())

	b, err := _storage.GetMeasurements()
	assert.NoError(t, err)
	expectedHistory := []byte(measurementID1 + "\n" + measurementID2 + "\n" + measurementID3 + "\n")
	assert.Equal(t, expectedHistory, b)

	ctx = createDefaultContext()
	root = NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "@-4"}
	err = root.Cmd.ExecuteContext(t.Context())
	assert.ErrorIs(t, err, storage.ErrIndexOutOfRange)

	expectedCtx.From = "@-4"
	expectedCtx.IsLocationFromSession = false
	expectedCtx.RecordToSession = true
	expectedCtx.MeasurementsCreated = 0
	expectedCtx.History = view.NewHistoryBuffer(1)
	assert.Equal(t, expectedCtx, ctx)
	assert.Equal(t, "Error: index out of range\n", w.String())

	assert.NoError(t, _storage.Remove())

	w.Reset()
	ctx = createDefaultContext()
	root = NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "@1"}
	err = root.Cmd.ExecuteContext(t.Context())
	assert.ErrorIs(t, err, storage.ErrNoPreviousMeasurements)

	expectedCtx.From = "@1"
	assert.Equal(t, expectedCtx, ctx)
	assert.Equal(t, "Error: no previous measurements found\n", w.String())

	w.Reset()
	ctx = createDefaultContext()
	root = NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "@0"}
	err = root.Cmd.ExecuteContext(t.Context())
	assert.ErrorIs(t, err, storage.ErrInvalidIndex)

	expectedCtx.From = "@0"
	assert.Equal(t, expectedCtx, ctx)
	assert.Equal(t, "Error: invalid index\n", w.String())

	w.Reset()
	ctx = createDefaultContext()
	root = NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "@x"}
	err = root.Cmd.ExecuteContext(t.Context())
	assert.ErrorIs(t, err, storage.ErrInvalidIndex)

	expectedCtx.From = "@x"
	assert.Equal(t, expectedCtx, ctx)
	assert.Equal(t, "Error: invalid index\n", w.String())

	w.Reset()
	ctx = createDefaultContext()
	root = NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "@"}
	err = root.Cmd.ExecuteContext(t.Context())
	assert.ErrorIs(t, err, storage.ErrInvalidIndex)

	expectedCtx.From = "@"
	assert.Equal(t, expectedCtx, ctx)
	assert.Equal(t, "Error: invalid index\n", w.String())
}

func Test_Execute_Ping_Infinite(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts1 := createDefaultMeasurementCreate("ping")
	expectedOpts1.Options.Packets = 16
	expectedOpts1.InProgressUpdates = true
	expectedOpts1.Timeout = 17
	expectedOpts2 := createDefaultMeasurementCreate("ping")
	expectedOpts2.Options.Packets = 16
	expectedOpts2.InProgressUpdates = true
	expectedOpts2.Timeout = 17
	expectedOpts2.Locations = globalping.PreviousMeasurementID(measurementID1)
	expectedOpts3 := createDefaultMeasurementCreate("ping")
	expectedOpts3.Options.Packets = 16
	expectedOpts3.InProgressUpdates = true
	expectedOpts3.Timeout = 17
	expectedOpts3.Locations = globalping.PreviousMeasurementID(measurementID2)
	expectedOpts4 := createDefaultMeasurementCreate("ping")
	expectedOpts4.Options.Packets = 16
	expectedOpts4.InProgressUpdates = true
	expectedOpts4.Timeout = 17
	expectedOpts4.Locations = globalping.PreviousMeasurementID(measurementID3)

	expectedResponse1 := createDefaultMeasurementCreateResponse()
	expectedResponse2 := createDefaultMeasurementCreateResponse()
	expectedResponse2.ID = measurementID2
	expectedResponse3 := createDefaultMeasurementCreateResponse()
	expectedResponse3.ID = measurementID3
	expectedResponse4 := createDefaultMeasurementCreateResponse()
	expectedResponse4.ID = measurementID4

	gbMock := apiMocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(gomock.Any(), expectedOpts1).Return(expectedResponse1, nil)
	gbMock.EXPECT().CreateMeasurement(gomock.Any(), expectedOpts2).Return(expectedResponse2, nil)
	gbMock.EXPECT().CreateMeasurement(gomock.Any(), expectedOpts3).Return(expectedResponse3, nil)
	gbMock.EXPECT().CreateMeasurement(gomock.Any(), expectedOpts4).Return(expectedResponse4, nil)

	expectedMeasurement1 := createDefaultMeasurement_MultipleProbes(globalping.MeasurementStatusFinished, globalping.TestStatusFinished)
	expectedMeasurement2 := createDefaultMeasurement_MultipleProbes(globalping.MeasurementStatusInProgress, globalping.TestStatusInProgress)
	expectedMeasurement2.ID = measurementID2
	expectedMeasurement2.Results[0].Result.Status = globalping.TestStatusFinished
	expectedMeasurement3 := createDefaultMeasurement_MultipleProbes(globalping.MeasurementStatusInProgress, globalping.TestStatusInProgress)
	expectedMeasurement3.ID = measurementID3
	expectedMeasurement3.Results[0].Result.Status = globalping.TestStatusFinished
	expectedMeasurement4 := createDefaultMeasurement_MultipleProbes(globalping.MeasurementStatusInProgress, globalping.TestStatusInProgress)
	expectedMeasurement4.ID = measurementID4
	expectedMeasurement4.Results[1].Result.Status = globalping.TestStatusFinished

	gbMock.EXPECT().GetMeasurement(gomock.Any(), measurementID1).Return(expectedMeasurement1, nil)
	gbMock.EXPECT().GetMeasurement(gomock.Any(), measurementID2).Return(expectedMeasurement2, nil)
	gbMock.EXPECT().GetMeasurement(gomock.Any(), measurementID3).Return(expectedMeasurement3, nil)
	gbMock.EXPECT().GetMeasurement(gomock.Any(), measurementID4).Return(expectedMeasurement4, nil)
	gbMock.EXPECT().GetMeasurement(gomock.Any(), measurementID2).Return(expectedMeasurement2, nil)
	gbMock.EXPECT().GetMeasurement(gomock.Any(), measurementID3).Return(expectedMeasurement3, nil)
	var finalRunCtx context.Context
	gbMock.EXPECT().GetMeasurement(gomock.Any(), measurementID4).DoAndReturn(func(ctx context.Context, _ string) (*globalping.Measurement, error) {
		finalRunCtx = ctx

		return expectedMeasurement4, nil
	})

	viewerMock := viewMocks.NewMockViewer(ctrl)
	waitFn := func(_ *view.InfinitePingOutput) (string, error) { time.Sleep(5 * time.Millisecond); return "", nil }
	viewerMock.EXPECT().OutputInfinite(pingOutputFor(expectedMeasurement1)).DoAndReturn(waitFn)
	viewerMock.EXPECT().OutputInfinite(pingOutputFor(expectedMeasurement2)).DoAndReturn(waitFn)
	viewerMock.EXPECT().OutputInfinite(pingOutputFor(expectedMeasurement3)).DoAndReturn(waitFn)
	viewerMock.EXPECT().OutputInfinite(pingOutputFor(expectedMeasurement4)).DoAndReturn(waitFn)
	viewerMock.EXPECT().OutputInfinite(pingOutputFor(expectedMeasurement2)).DoAndReturn(waitFn)
	viewerMock.EXPECT().OutputInfinite(pingOutputFor(expectedMeasurement3)).DoAndReturn(waitFn)
	finalOutputStarted := make(chan struct{})
	viewerMock.EXPECT().OutputInfinite(pingOutputFor(expectedMeasurement4)).DoAndReturn(func(_ *view.InfinitePingOutput) (string, error) {
		close(finalOutputStarted)
		<-finalRunCtx.Done()

		return "", nil
	})

	viewerMock.EXPECT().OutputPingSummary("", gomock.Any()).Times(1)
	viewerMock.EXPECT().OutputShare().Times(1)

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := &view.Context{
		History: view.NewHistoryBuffer(10),
		From:    "world",
		Limit:   1,
	}
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "--infinite", "--timeout", "17", "from", "Berlin"}

	go func() {
		<-finalOutputStarted
		root.cancel <- syscall.SIGINT
	}()
	err := root.Cmd.ExecuteContext(t.Context())

	assert.NoError(t, err)
	assert.Equal(t, "", w.String())

	expectedCtx := &view.Context{
		Cmd:                 "ping",
		Target:              "jsdelivr.com",
		Targets:             []string{"jsdelivr.com"},
		From:                "Berlin",
		Limit:               1,
		Packets:             16,
		Infinite:            true,
		Timeout:             17,
		CIMode:              true,
		Protocol:            "ICMP",
		Port:                80,
		MeasurementsCreated: 4,
	}
	expectedCtx.History = &view.HistoryBuffer{
		Index: 4,
		Slice: []*view.HistoryItem{
			{
				Id:        measurementID1,
				Status:    globalping.MeasurementStatusFinished,
				StartedAt: defaultCurrentTime,
			},
			{
				Id:     measurementID2,
				Status: globalping.MeasurementStatusInProgress,
				ProbeStatus: []globalping.TestStatus{
					globalping.TestStatusFinished,
					globalping.TestStatusInProgress,
					globalping.TestStatusInProgress,
				},
				StartedAt: defaultCurrentTime,
			},
			{
				Id:     measurementID3,
				Status: globalping.MeasurementStatusInProgress,
				ProbeStatus: []globalping.TestStatus{
					globalping.TestStatusFinished,
					globalping.TestStatusInProgress,
					globalping.TestStatusInProgress,
				},
				StartedAt: defaultCurrentTime,
			},
			{
				Id:     measurementID4,
				Status: globalping.MeasurementStatusInProgress,
				ProbeStatus: []globalping.TestStatus{
					globalping.TestStatusInProgress,
					globalping.TestStatusFinished,
					globalping.TestStatusInProgress,
				},
				StartedAt: defaultCurrentTime,
			},
			nil, nil, nil, nil, nil, nil,
		},
	}
	assert.Equal(t, expectedCtx, ctx)

	b, err := _storage.GetMeasurements()
	assert.NoError(t, err)
	expectedHistory := measurementID1 + "\n"
	assert.Equal(t, expectedHistory, string(b))

	items, err := _storage.GetHistory(0)
	assert.NoError(t, err)
	expectedHistoryItems := []string{createDefaultExpectedHistoryItem(
		"1",
		"ping jsdelivr.com --infinite --timeout 17 from Berlin",
		measurementID1+"."+measurementID2+"."+measurementID3+"."+measurementID4,
	)}
	assert.Equal(t, expectedHistoryItems, items)
}

func Test_Execute_Ping_Infinite_TableInCI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("ping")
	expectedOpts.Limit = 2
	expectedOpts.Options.Packets = 16
	expectedOpts.InProgressUpdates = true

	expectedResponse := createDefaultMeasurementCreateResponse()
	expectedMeasurement := createDefaultMeasurement("ping")
	expectedMeasurement.Results[0].Result.StatsRaw = []byte(`{"total":0,"rcv":0,"drop":0,"loss":0}`)

	gbMock := apiMocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(gomock.Any(), expectedOpts).Return(expectedResponse, nil)
	var runCtx context.Context
	gbMock.EXPECT().GetMeasurement(gomock.Any(), measurementID1).DoAndReturn(func(ctx context.Context, _ string) (*globalping.Measurement, error) {
		runCtx = ctx

		return expectedMeasurement, nil
	})

	outputStarted := make(chan struct{})
	viewerMock := viewMocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().OutputInfinite(pingOutputFor(expectedMeasurement)).DoAndReturn(func(*view.InfinitePingOutput) (string, error) {
		close(outputStarted)
		<-runCtx.Done()

		return "final table", nil
	})
	viewerMock.EXPECT().OutputPingSummary("final table", gomock.Any()).Times(1)
	viewerMock.EXPECT().OutputShare().Times(1)

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	ctx := createDefaultContext()
	storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(view.NewPrinter(nil, w, w), ctx, viewerMock, utilsMock, gbMock, nil, storage)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "--infinite", "--limit", "2", "--ci", "from", "Berlin"}

	go func() {
		<-outputStarted
		root.cancel <- syscall.SIGINT
	}()
	err := root.Cmd.ExecuteContext(t.Context())

	assert.NoError(t, err)
	assert.True(t, ctx.Infinite)
	assert.True(t, ctx.Table)
	assert.False(t, ctx.ToLatency)
	assert.True(t, ctx.CIMode)
	assert.Empty(t, w.String())
}

func Test_Execute_Ping_Infinite_Output_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts1 := createDefaultMeasurementCreate("ping")
	expectedOpts1.Options.Packets = 16
	expectedOpts1.InProgressUpdates = true

	expectedResponse1 := createDefaultMeasurementCreateResponse()

	gbMock := apiMocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(gomock.Any(), expectedOpts1).Return(expectedResponse1, nil)

	expectedMeasurement := createDefaultMeasurement("ping")
	gbMock.EXPECT().GetMeasurement(gomock.Any(), measurementID1).Return(expectedMeasurement, nil)

	viewerMock := viewMocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().OutputInfinite(pingOutputFor(expectedMeasurement)).Return("", errors.New("error message"))
	viewerMock.EXPECT().OutputPingSummary(gomock.Any(), gomock.Any()).Times(0)
	viewerMock.EXPECT().OutputShare().Times(1)

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext()
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "--infinite", "from", "Berlin"}
	err := root.Cmd.ExecuteContext(t.Context())
	assert.Equal(t, "error message", err.Error())

	assert.Equal(t, "Error: error message\n", w.String())

	expectedCtx := createDefaultExpectedContext("ping")
	expectedCtx.History.Find(measurementID1).Status = globalping.MeasurementStatusFinished
	expectedCtx.Packets = 16
	expectedCtx.Infinite = true
	assert.Equal(t, expectedCtx, ctx)

	b, err := _storage.GetMeasurements()
	assert.NoError(t, err)
	expectedHistory := measurementID1 + "\n"
	assert.Equal(t, expectedHistory, string(b))

	items, err := _storage.GetHistory(0)
	assert.NoError(t, err)
	expectedHistoryItems := []string{createDefaultExpectedHistoryItem(
		"1",
		"ping jsdelivr.com --infinite from Berlin",
		measurementID1,
	)}
	assert.Equal(t, expectedHistoryItems, items)
}

func Test_Execute_Ping_Infinite_AllProbesFailedExitsSuccessfully(t *testing.T) {
	ctrl := gomock.NewController(t)
	expectedOpts := createDefaultMeasurementCreate("ping")
	expectedOpts.Options.Packets = 16
	expectedOpts.InProgressUpdates = true
	client := apiMocks.NewMockClient(ctrl)
	client.EXPECT().CreateMeasurement(gomock.Any(), expectedOpts).Return(createDefaultMeasurementCreateResponse(), nil)
	measurement := createDefaultMeasurement("ping")
	client.EXPECT().GetMeasurement(gomock.Any(), measurementID1).Return(measurement, nil)
	viewer := viewMocks.NewMockViewer(ctrl)
	viewer.EXPECT().OutputInfinite(pingOutputFor(measurement)).Return("failed output", view.ErrAllProbesFailed)
	viewer.EXPECT().OutputPingSummary(gomock.Any(), gomock.Any()).Times(0)
	viewer.EXPECT().OutputShare()
	utils := utilsMocks.NewMockUtils(ctrl)
	utils.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()
	ctx := createDefaultContext()
	storage := createDefaultTestStorage(t, utils)
	w := new(bytes.Buffer)
	root := NewRoot(view.NewPrinter(nil, w, w), ctx, viewer, utils, client, nil, storage)
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "--infinite", "from", "Berlin"}

	err := root.Cmd.ExecuteContext(t.Context())

	assert.NoError(t, err)
	assert.Empty(t, w.String())
}

func Test_Execute_Ping_Infinite_Output_TooManyRequests_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts1 := createDefaultMeasurementCreate("ping")
	expectedOpts1.Options.Packets = 16
	expectedOpts1.InProgressUpdates = true
	expectedOpts2 := createDefaultMeasurementCreate("ping")
	expectedOpts2.Options.Packets = 16
	expectedOpts2.InProgressUpdates = true
	expectedOpts2.Locations = globalping.PreviousMeasurementID(measurementID1)

	expectedResponse1 := createDefaultMeasurementCreateResponse()

	gbMock := apiMocks.NewMockClient(ctrl)
	createCall1 := gbMock.EXPECT().CreateMeasurement(gomock.Any(), expectedOpts1).Return(expectedResponse1, nil)
	gbMock.EXPECT().CreateMeasurement(gomock.Any(), expectedOpts2).Return(nil, &globalping.MeasurementError{
		StatusCode: 429,
		Type:       "too_many_requests",
		Message:    "too many requests",
	}).After(createCall1)

	expectedMeasurement := createDefaultMeasurement("ping")
	gbMock.EXPECT().GetMeasurement(gomock.Any(), measurementID1).Return(expectedMeasurement, nil)

	viewerMock := viewMocks.NewMockViewer(ctrl)
	waitFn := func(_ *view.InfinitePingOutput) (string, error) { time.Sleep(5 * time.Millisecond); return "", nil }
	viewerMock.EXPECT().OutputInfinite(pingOutputFor(expectedMeasurement)).DoAndReturn(waitFn)

	viewerMock.EXPECT().OutputPingSummary(gomock.Any(), gomock.Any()).Times(0)
	viewerMock.EXPECT().OutputShare().Times(1)

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	errW := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, errW)
	ctx := createDefaultContext()
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "Berlin", "--infinite", "--share"}
	err := root.Cmd.ExecuteContext(t.Context())
	assert.Equal(t, "too_many_requests: too many requests", err.Error())

	assert.Equal(t, "> too many requests\n", errW.String())
	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("ping")
	expectedCtx.History.Find(measurementID1).Status = globalping.MeasurementStatusFinished
	expectedCtx.Packets = 16
	expectedCtx.Infinite = true
	expectedCtx.Share = true
	assert.Equal(t, expectedCtx, ctx)

	b, err := _storage.GetMeasurements()
	assert.NoError(t, err)
	expectedHistory := measurementID1 + "\n"
	assert.Equal(t, expectedHistory, string(b))

	items, err := _storage.GetHistory(0)
	assert.NoError(t, err)
	expectedHistoryItems := []string{createDefaultExpectedHistoryItem(
		"1",
		"ping jsdelivr.com from Berlin --infinite --share",
		measurementID1,
	)}
	assert.Equal(t, expectedHistoryItems, items)
}

func Test_Ping_InfiniteCreationFailureFinishesPendingMeasurement(t *testing.T) {
	for _, finalStatus := range []globalping.TestStatus{globalping.TestStatusFinished, globalping.TestStatusFailed} {
		t.Run(string(finalStatus), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			firstResponse := createDefaultMeasurementCreateResponse()
			measurement := createDefaultMeasurement_MultipleProbes(globalping.MeasurementStatusInProgress, globalping.TestStatusInProgress)
			measurement.Results[0].Result.Status = globalping.TestStatusFinished

			client := apiMocks.NewMockClient(ctrl)
			firstCreate := client.EXPECT().CreateMeasurement(gomock.Any(), gomock.Any()).Return(firstResponse, nil)
			firstPoll := client.EXPECT().GetMeasurement(gomock.Any(), firstResponse.ID).Return(measurement, nil).After(firstCreate)
			viewer := viewMocks.NewMockViewer(ctrl)
			firstOutput := viewer.EXPECT().OutputInfinite(pingOutputFor(measurement)).Return("current output", nil).After(firstPoll)
			failedCreate := client.EXPECT().CreateMeasurement(gomock.Any(), gomock.Any()).Return(nil, assert.AnError).After(firstOutput)
			completed := createDefaultMeasurement_MultipleProbes(globalping.MeasurementStatusFinished, finalStatus)
			var outputErr error

			if finalStatus == globalping.TestStatusFailed {
				outputErr = view.ErrAllProbesFailed
			}

			finalPoll := client.EXPECT().GetMeasurement(gomock.Any(), firstResponse.ID).Return(completed, nil).After(failedCreate)
			viewer.EXPECT().OutputInfinite(pingOutputFor(completed)).Return("completed output", outputErr).After(finalPoll)
			utilsMock := utilsMocks.NewMockUtils(ctrl)
			utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()
			ctx := createDefaultContext()
			ctx.Infinite = true
			ctx.Packets = 16
			root := NewRoot(view.NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer)), ctx, viewer, utilsMock, client, nil, nil)
			opts := createDefaultMeasurementCreate("ping")

			output, err := root.runInfinitePing(t.Context(), opts, view.NewInfinitePingRun(ctx.Protocol, ctx.Packets, utilsMock.Now(), utilsMock.Now))

			assert.ErrorIs(t, err, assert.AnError)
			assert.Equal(t, "completed output", output)
			assert.Equal(t, globalping.MeasurementStatusFinished, ctx.History.Find(firstResponse.ID).Status)
		})
	}
}

func Test_Execute_Ping_IPv4(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("ping")
	expectedOpts.Locations = globalping.LocationOptions{{Magic: "world"}}
	expectedOpts.Options.IPVersion = globalping.IPVersion4
	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := apiMocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(t.Context(), expectedOpts).Times(1).Return(expectedResponse, nil)

	expectedMeasurement := createDefaultMeasurement("ping")
	gbMock.EXPECT().AwaitMeasurement(t.Context(), expectedResponse.ID).Times(1).Return(expectedMeasurement, nil)

	viewerMock := viewMocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().OutputDefault(measurementID1, expectedMeasurement, expectedOpts).Times(1)

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext()
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)

	os.Args = []string{"globalping", "ping", "jsdelivr.com", "--ipv4"}
	err := root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("ping")
	expectedCtx.From = "world"
	expectedCtx.Ipv4 = true
	assert.Equal(t, expectedCtx, ctx)
}

func Test_Execute_Ping_IPv6(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("ping")
	expectedOpts.Locations = globalping.LocationOptions{{Magic: "world"}}
	expectedOpts.Options.IPVersion = globalping.IPVersion6
	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := apiMocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(t.Context(), expectedOpts).Times(1).Return(expectedResponse, nil)

	expectedMeasurement := createDefaultMeasurement("ping")
	gbMock.EXPECT().AwaitMeasurement(t.Context(), expectedResponse.ID).Times(1).Return(expectedMeasurement, nil)

	viewerMock := viewMocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().OutputDefault(measurementID1, expectedMeasurement, expectedOpts).Times(1)

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext()
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)

	os.Args = []string{"globalping", "ping", "jsdelivr.com", "--ipv6"}
	err := root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("ping")
	expectedCtx.From = "world"
	expectedCtx.Ipv6 = true
	assert.Equal(t, expectedCtx, ctx)
}

func Test_Execute_Ping_Invalid_Protocol(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext()
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, nil, utilsMock, nil, nil, _storage)

	os.Args = []string{"globalping", "ping", "jsdelivr.com", "--protocol", "invalid"}
	err := root.Cmd.ExecuteContext(t.Context())
	assert.Error(t, err, "protocol INVALID is not supported")

	items, err := _storage.GetHistory(0)
	assert.NoError(t, err)
	assert.Empty(t, items)
}
