package cmd

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	apiMocks "github.com/jsdelivr/globalping-cli/mocks/api"
	utilsMocks "github.com/jsdelivr/globalping-cli/mocks/utils"
	viewMocks "github.com/jsdelivr/globalping-cli/mocks/view"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/jsdelivr/globalping-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Execute_ComparisonCreatesBothTargetsForEveryCommand(t *testing.T) {
	for _, measurementType := range []globalping.MeasurementType{"ping", "traceroute", "mtr", "dns", "http"} {
		t.Run(string(measurementType), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			first := createDefaultMeasurementCreate(measurementType)
			first.Target = "one.example"
			first.Locations = globalping.LocationOptions{{Magic: "world"}}
			first.InProgressUpdates = false

			if measurementType == "dns" {
				first.Options.Query = &globalping.QueryOptions{}
			}

			second := *first
			second.Target = "two.example"
			second.Locations = globalping.PreviousMeasurementID(measurementID1)
			targetArgument := "one.example,two.example"

			if measurementType == "http" {
				targetArgument = "http://one.example:8080/first?x=1,https://two.example:8443/second?y=2"
				first.Target = "one.example"
				first.Options = &globalping.MeasurementOptions{
					Protocol: "HTTP",
					Port:     8080,
					Request: &globalping.RequestOptions{
						Path:    "/first",
						Query:   "x=1",
						Headers: map[string]string{},
					},
				}
				second.Target = "two.example"
				second.Options = &globalping.MeasurementOptions{
					Protocol: "HTTPS",
					Port:     8443,
					Request: &globalping.RequestOptions{
						Path:    "/second",
						Query:   "y=2",
						Headers: map[string]string{},
					},
				}
			}

			client := apiMocks.NewMockClient(ctrl)
			gomock.InOrder(
				client.EXPECT().CreateMeasurement(t.Context(), first).Return(&globalping.MeasurementCreateResponse{ID: measurementID1}, nil),
				client.EXPECT().CreateMeasurement(t.Context(), &second).Return(&globalping.MeasurementCreateResponse{ID: measurementID2}, nil),
			)
			firstMeasurement := createDefaultMeasurement(measurementType)
			secondMeasurement := createDefaultMeasurement(measurementType)
			secondMeasurement.ID = measurementID2
			client.EXPECT().GetMeasurement(t.Context(), measurementID1).Return(firstMeasurement, nil)
			client.EXPECT().GetMeasurement(t.Context(), measurementID2).Return(secondMeasurement, nil)
			viewer := viewMocks.NewMockViewer(ctrl)
			viewer.EXPECT().OutputComparisonTable(firstMeasurement, secondMeasurement).Return("", nil)
			viewer.EXPECT().OutputShare()
			utils := utilsMocks.NewMockUtils(ctrl)
			utils.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()
			ctx := createDefaultContext()
			ctx.History = view.NewHistoryBuffer(2)
			storage := createDefaultTestStorage(t, utils)
			root := NewRoot(view.NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer)), ctx, viewer, utils, client, nil, storage)
			oldArgs := os.Args
			t.Cleanup(func() { os.Args = oldArgs })
			os.Args = []string{"globalping", string(measurementType), targetArgument, "--ci"}

			err := root.Cmd.ExecuteContext(t.Context())

			require.NoError(t, err)
			assert.True(t, ctx.Comparison)
			assert.True(t, ctx.Table)
			measurements, err := storage.GetMeasurements()
			require.NoError(t, err)
			assert.Equal(t, measurementID1+"\n", string(measurements))
			history, err := storage.GetHistory(0)
			require.NoError(t, err)
			require.Len(t, history, 1)
			assert.Contains(t, history[0], "measurement="+measurementID1+","+measurementID2)
			assert.NotContains(t, history[0], "&display=table")
		})
	}
}

func Test_CreateAndHandleMeasurements_SecondCreationFailureRendersFirstAndPreservesBothErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	createErr := &globalping.MeasurementError{
		StatusCode: http.StatusTooManyRequests,
		Type:       "too_many_requests",
		Message:    "second target rejected",
	}
	renderErr := errors.New("first target render failed")
	first := createDefaultMeasurementCreate("ping")
	first.Target = "one.example"
	second := *first
	second.Target = "two.example"
	second.Locations = globalping.PreviousMeasurementID(measurementID1)
	client := apiMocks.NewMockClient(ctrl)
	gomock.InOrder(
		client.EXPECT().CreateMeasurement(t.Context(), first).Return(&globalping.MeasurementCreateResponse{ID: measurementID1}, nil),
		client.EXPECT().CreateMeasurement(t.Context(), &second).Return(nil, createErr),
	)
	measurement := createDefaultMeasurement("ping")
	client.EXPECT().GetMeasurement(t.Context(), measurementID1).Return(measurement, nil)
	viewer := viewMocks.NewMockViewer(ctrl)
	viewer.EXPECT().OutputTable(measurement).Return("", renderErr)
	viewer.EXPECT().OutputShare()
	utils := utilsMocks.NewMockUtils(ctrl)
	utils.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()
	ctx := createDefaultContext()
	ctx.Comparison = true
	ctx.Table = true
	ctx.RecordToSession = true
	ctx.History = view.NewHistoryBuffer(2)
	storage := createDefaultTestStorage(t, utils)
	output := new(bytes.Buffer)
	root := NewRoot(view.NewPrinter(nil, output, output), ctx, viewer, utils, client, nil, storage)

	err := root.createAndHandleMeasurements(t.Context(), []*globalping.MeasurementCreate{first, &second})

	require.Error(t, err)
	assert.ErrorIs(t, err, createErr)
	assert.ErrorIs(t, err, renderErr)
	assert.True(t, root.Cmd.SilenceErrors)
	assert.Equal(t, 1, strings.Count(output.String(), "second target rejected"))
	assert.Equal(t, 1, strings.Count(output.String(), "first target render failed"))
	assert.Equal(t, measurementID1, ctx.History.ToString(","))
	measurements, storageErr := storage.GetMeasurements()
	require.NoError(t, storageErr)
	assert.Equal(t, measurementID1+"\n", string(measurements))
}

func Test_Execute_ComparisonFailureHistory(t *testing.T) {
	t.Run("partial creation stores and shares only T1", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		createErr := errors.New("second target API error")
		first := createDefaultMeasurementCreate("ping")
		first.Target = "one.example"
		first.Locations = globalping.LocationOptions{{Magic: "world"}}
		second := *first
		second.Target = "two.example"
		second.Locations = globalping.PreviousMeasurementID(measurementID1)
		client := apiMocks.NewMockClient(ctrl)
		gomock.InOrder(
			client.EXPECT().CreateMeasurement(t.Context(), first).Return(&globalping.MeasurementCreateResponse{ID: measurementID1}, nil),
			client.EXPECT().CreateMeasurement(t.Context(), &second).Return(nil, createErr),
		)
		measurement := createDefaultMeasurement("ping")
		client.EXPECT().GetMeasurement(t.Context(), measurementID1).Return(measurement, nil)
		viewer := viewMocks.NewMockViewer(ctrl)
		viewer.EXPECT().OutputTable(measurement).Return("", nil)
		viewer.EXPECT().OutputShare()
		utils := utilsMocks.NewMockUtils(ctrl)
		utils.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()
		ctx := createDefaultContext()
		ctx.History = view.NewHistoryBuffer(2)
		storage := createDefaultTestStorage(t, utils)
		output := new(bytes.Buffer)
		root := NewRoot(view.NewPrinter(nil, output, output), ctx, viewer, utils, client, nil, storage)
		oldArgs := os.Args
		t.Cleanup(func() { os.Args = oldArgs })
		os.Args = []string{"globalping", "ping", "one.example,two.example", "--ci"}

		err := root.Cmd.ExecuteContext(t.Context())

		assert.ErrorIs(t, err, createErr)
		assert.Contains(t, output.String(), "second target API error")
		measurements, storageErr := storage.GetMeasurements()
		require.NoError(t, storageErr)
		assert.Equal(t, measurementID1+"\n", string(measurements))
		history, storageErr := storage.GetHistory(0)
		require.NoError(t, storageErr)
		require.Len(t, history, 1)
		assert.Contains(t, history[0], "measurement="+measurementID1)
		assert.NotContains(t, history[0], measurementID2)
	})

	t.Run("post-creation polling failure retains both IDs", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pollErr := errors.New("comparison poll failed")
		first := createDefaultMeasurementCreate("ping")
		first.Target = "one.example"
		first.Locations = globalping.LocationOptions{{Magic: "world"}}
		second := *first
		second.Target = "two.example"
		second.Locations = globalping.PreviousMeasurementID(measurementID1)
		client := apiMocks.NewMockClient(ctrl)
		client.EXPECT().CreateMeasurement(t.Context(), first).Return(&globalping.MeasurementCreateResponse{ID: measurementID1}, nil)
		client.EXPECT().CreateMeasurement(t.Context(), &second).Return(&globalping.MeasurementCreateResponse{ID: measurementID2}, nil)
		firstProgress := createDefaultMeasurement("ping")
		firstProgress.Status = globalping.MeasurementStatusInProgress
		secondProgress := createDefaultMeasurement("ping")
		secondProgress.ID = measurementID2
		secondProgress.Status = globalping.MeasurementStatusInProgress
		client.EXPECT().GetMeasurement(t.Context(), measurementID1).Return(firstProgress, nil)
		client.EXPECT().GetMeasurement(t.Context(), measurementID2).Return(secondProgress, nil)
		client.EXPECT().GetMeasurement(t.Context(), measurementID1).Return(nil, pollErr)
		viewer := viewMocks.NewMockViewer(ctrl)
		viewer.EXPECT().OutputShare()
		utils := utilsMocks.NewMockUtils(ctrl)
		utils.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()
		ctx := createDefaultContext()
		ctx.History = view.NewHistoryBuffer(2)
		storage := createDefaultTestStorage(t, utils)
		root := NewRoot(view.NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer)), ctx, viewer, utils, client, nil, storage)
		oldArgs := os.Args
		t.Cleanup(func() { os.Args = oldArgs })
		os.Args = []string{"globalping", "ping", "one.example,two.example", "--ci"}

		err := root.Cmd.ExecuteContext(t.Context())

		assert.ErrorIs(t, err, pollErr)
		measurements, storageErr := storage.GetMeasurements()
		require.NoError(t, storageErr)
		assert.Equal(t, measurementID1+"\n", string(measurements))
		history, storageErr := storage.GetHistory(0)
		require.NoError(t, storageErr)
		require.Len(t, history, 1)
		assert.Contains(t, history[0], "measurement="+measurementID1+","+measurementID2)
		assert.NotContains(t, history[0], "&display=table")
	})
}

func Test_HandleComparisonMeasurements_PollsIndependentlyByCompleteCycle(t *testing.T) {
	ctrl := gomock.NewController(t)
	firstProgress := createDefaultMeasurement("ping")
	firstProgress.Status = globalping.MeasurementStatusInProgress
	secondProgress := createDefaultMeasurement("ping")
	secondProgress.ID = measurementID2
	secondProgress.Status = globalping.MeasurementStatusInProgress
	firstFinished := createDefaultMeasurement("ping")
	secondFinished := createDefaultMeasurement("ping")
	secondFinished.ID = measurementID2
	client := apiMocks.NewMockClient(ctrl)
	viewer := viewMocks.NewMockViewer(ctrl)
	gomock.InOrder(
		client.EXPECT().GetMeasurement(t.Context(), measurementID1).Return(firstProgress, nil),
		client.EXPECT().GetMeasurement(t.Context(), measurementID2).Return(secondProgress, nil),
		viewer.EXPECT().OutputComparisonTable(firstProgress, secondProgress).Return("", nil),
		client.EXPECT().GetMeasurement(t.Context(), measurementID1).Return(firstFinished, nil),
		client.EXPECT().GetMeasurement(t.Context(), measurementID2).Return(secondProgress, nil),
		viewer.EXPECT().OutputComparisonTable(firstFinished, secondProgress).Return("", nil),
		client.EXPECT().GetMeasurement(t.Context(), measurementID2).Return(secondFinished, nil),
		viewer.EXPECT().OutputComparisonTable(firstFinished, secondFinished).Return("", nil),
		viewer.EXPECT().OutputShare(),
	)
	utils := utilsMocks.NewMockUtils(ctrl)
	utils.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()
	ctx := createDefaultContext()
	ctx.Table = true
	root := NewRoot(view.NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer)), ctx, viewer, utils, client, nil, nil)

	err := root.handleComparisonMeasurements(t.Context(), measurementID1, measurementID2)

	require.NoError(t, err)
}

func Test_HandleComparisonMeasurements_CIRendersOnlyAfterBothFinish(t *testing.T) {
	ctrl := gomock.NewController(t)
	firstProgress := createDefaultMeasurement("ping")
	firstProgress.Status = globalping.MeasurementStatusInProgress
	secondProgress := createDefaultMeasurement("ping")
	secondProgress.ID = measurementID2
	secondProgress.Status = globalping.MeasurementStatusInProgress
	firstFinished := createDefaultMeasurement("ping")
	secondFinished := createDefaultMeasurement("ping")
	secondFinished.ID = measurementID2
	client := apiMocks.NewMockClient(ctrl)
	viewer := viewMocks.NewMockViewer(ctrl)
	gomock.InOrder(
		client.EXPECT().GetMeasurement(t.Context(), measurementID1).Return(firstProgress, nil),
		client.EXPECT().GetMeasurement(t.Context(), measurementID2).Return(secondProgress, nil),
		client.EXPECT().GetMeasurement(t.Context(), measurementID1).Return(firstFinished, nil),
		client.EXPECT().GetMeasurement(t.Context(), measurementID2).Return(secondFinished, nil),
		viewer.EXPECT().OutputComparisonTable(firstFinished, secondFinished).Return("", nil),
		viewer.EXPECT().OutputShare(),
	)
	utils := utilsMocks.NewMockUtils(ctrl)
	utils.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()
	ctx := createDefaultContext()
	ctx.CIMode = true
	ctx.Table = true
	root := NewRoot(view.NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer)), ctx, viewer, utils, client, nil, nil)

	err := root.handleComparisonMeasurements(t.Context(), measurementID1, measurementID2)

	require.NoError(t, err)
}

func Test_HandleComparisonMeasurements_StopsOnPollingErrorAndKeepsShareIDs(t *testing.T) {
	ctrl := gomock.NewController(t)
	pollErr := errors.New("poll failed")
	firstProgress := createDefaultMeasurement("ping")
	firstProgress.Status = globalping.MeasurementStatusInProgress
	secondProgress := createDefaultMeasurement("ping")
	secondProgress.ID = measurementID2
	secondProgress.Status = globalping.MeasurementStatusInProgress
	client := apiMocks.NewMockClient(ctrl)
	viewer := viewMocks.NewMockViewer(ctrl)
	gomock.InOrder(
		client.EXPECT().GetMeasurement(t.Context(), measurementID1).Return(firstProgress, nil),
		client.EXPECT().GetMeasurement(t.Context(), measurementID2).Return(secondProgress, nil),
		viewer.EXPECT().OutputComparisonTable(firstProgress, secondProgress).Return("", nil),
		client.EXPECT().GetMeasurement(t.Context(), measurementID1).Return(nil, pollErr),
		viewer.EXPECT().OutputShare(),
	)
	utils := utilsMocks.NewMockUtils(ctrl)
	utils.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()
	ctx := createDefaultContext()
	ctx.Table = true
	ctx.Comparison = true
	ctx.History = view.NewHistoryBuffer(2)
	ctx.History.Push(&view.HistoryItem{Id: measurementID1})
	ctx.History.Push(&view.HistoryItem{Id: measurementID2})
	root := NewRoot(view.NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer)), ctx, viewer, utils, client, nil, nil)

	err := root.handleComparisonMeasurements(t.Context(), measurementID1, measurementID2)

	assert.ErrorIs(t, err, pollErr)
	assert.Equal(t, measurementID1+","+measurementID2, ctx.History.ToString(","))
	assert.True(t, root.Cmd.SilenceUsage)
}

func Test_HandleComparisonMeasurements_TracksWaitDeadlinesPerTarget(t *testing.T) {
	ctrl := gomock.NewController(t)
	firstFinished := createDefaultMeasurement("ping")
	secondProgress := createDefaultMeasurement("ping")
	secondProgress.ID = measurementID2
	secondProgress.Status = globalping.MeasurementStatusInProgress
	secondProgress.Timeout = 5
	client := apiMocks.NewMockClient(ctrl)
	viewer := viewMocks.NewMockViewer(ctrl)
	client.EXPECT().GetMeasurement(t.Context(), measurementID1).Return(firstFinished, nil)
	client.EXPECT().GetMeasurement(t.Context(), measurementID2).Return(secondProgress, nil)
	viewer.EXPECT().OutputComparisonTable(firstFinished, secondProgress).Return("", nil)
	viewer.EXPECT().OutputShare()
	utils := utilsMocks.NewMockUtils(ctrl)
	gomock.InOrder(
		utils.EXPECT().Now().Return(defaultCurrentTime),
		utils.EXPECT().Now().Return(defaultCurrentTime),
		utils.EXPECT().Now().Return(defaultCurrentTime.Add(16*time.Second)),
	)
	ctx := createDefaultContext()
	ctx.Table = true
	root := NewRoot(view.NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer)), ctx, viewer, utils, client, nil, nil)

	err := root.handleComparisonMeasurements(t.Context(), measurementID1, measurementID2)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.ErrorContains(t, err, measurementID2)
}

func Test_TableFlag_IsAvailableOnlyForMeasurementCommands(t *testing.T) {
	ctx := createDefaultContext()
	root := NewRoot(view.NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer)), ctx, nil, nil, nil, nil, nil)

	for _, command := range []string{"ping", "traceroute", "mtr", "dns", "http"} {
		cmd, _, err := root.Cmd.Find([]string{command})
		require.NoError(t, err)
		assert.NotNil(t, cmd.Flags().Lookup("table"), "%s should expose --table", command)
	}

	assert.Nil(t, root.Cmd.Flags().Lookup("table"), "the root command must not expose --table")

	for _, command := range []string{"auth", "history", "install-probe", "limits", "version"} {
		cmd, _, err := root.Cmd.Find([]string{command})
		require.NoError(t, err)
		assert.Nil(t, cmd.Flags().Lookup("table"), "%s must not expose --table", command)
	}
}

func Test_Execute_TableMeasurement(t *testing.T) {
	for _, measurementType := range []globalping.MeasurementType{"ping", "traceroute", "mtr", "dns", "http"} {
		t.Run(string(measurementType), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			expectedOpts := createDefaultMeasurementCreate(measurementType)
			locations, ok := expectedOpts.Locations.(globalping.LocationOptions)
			require.True(t, ok)
			locations[0].Magic = "world"

			switch measurementType {
			case "dns":
				expectedOpts.Options.Query = &globalping.QueryOptions{}
			case "http":
				expectedOpts.Options.Request = &globalping.RequestOptions{Headers: map[string]string{}}
			}

			expectedResponse := createDefaultMeasurementCreateResponse()
			expectedMeasurement := createDefaultMeasurement(measurementType)
			gbMock := apiMocks.NewMockClient(ctrl)
			gbMock.EXPECT().CreateMeasurement(t.Context(), expectedOpts).Return(expectedResponse, nil)
			gbMock.EXPECT().GetMeasurement(t.Context(), expectedResponse.ID).Return(expectedMeasurement, nil)

			viewerMock := viewMocks.NewMockViewer(ctrl)
			viewerMock.EXPECT().OutputTable(expectedMeasurement).Return("", nil)
			viewerMock.EXPECT().OutputShare()

			utilsMock := utilsMocks.NewMockUtils(ctrl)
			utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

			w := new(bytes.Buffer)
			ctx := createDefaultContext()
			storage := createDefaultTestStorage(t, utilsMock)
			root := NewRoot(view.NewPrinter(nil, w, w), ctx, viewerMock, utilsMock, gbMock, nil, storage)
			oldArgs := os.Args
			t.Cleanup(func() { os.Args = oldArgs })
			os.Args = []string{"globalping", string(measurementType), "jsdelivr.com", "--table", "--latency", "--json", "--ci"}

			err := root.Cmd.ExecuteContext(t.Context())

			require.NoError(t, err)
			assert.True(t, ctx.Table)
			assert.False(t, ctx.ToLatency)
			assert.False(t, ctx.ToJSON)
			assert.Empty(t, w.String())
		})
	}
}

func Test_HandleMeasurement_TableTakesOutputPrecedence(t *testing.T) {
	t.Run("table before latency", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		inProgressMeasurement := createDefaultMeasurement("ping")
		inProgressMeasurement.Status = globalping.MeasurementStatusInProgress
		measurement := createDefaultMeasurement("ping")
		client := apiMocks.NewMockClient(ctrl)
		client.EXPECT().GetMeasurement(t.Context(), measurement.ID).Return(inProgressMeasurement, nil)
		client.EXPECT().GetMeasurement(t.Context(), measurement.ID).Return(measurement, nil)
		viewer := viewMocks.NewMockViewer(ctrl)
		viewer.EXPECT().OutputTable(inProgressMeasurement).Return("", nil)
		viewer.EXPECT().OutputTable(measurement).Return("", nil)
		viewer.EXPECT().OutputShare()
		ctx := createDefaultContext()
		ctx.Table = true
		ctx.ToLatency = true
		utils := utilsMocks.NewMockUtils(ctrl)
		utils.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()
		root := NewRoot(view.NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer)), ctx, viewer, utils, client, nil, nil)

		require.NoError(t, root.handleMeasurement(t.Context(), measurement.ID, nil))
	})

	t.Run("table before json", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		measurement := createDefaultMeasurement("ping")
		client := apiMocks.NewMockClient(ctrl)
		client.EXPECT().GetMeasurement(t.Context(), measurement.ID).Return(measurement, nil)
		viewer := viewMocks.NewMockViewer(ctrl)
		viewer.EXPECT().OutputTable(measurement).Return("", nil)
		viewer.EXPECT().OutputShare()
		ctx := createDefaultContext()
		ctx.Table = true
		ctx.ToJSON = true
		utils := utilsMocks.NewMockUtils(ctrl)
		utils.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()
		root := NewRoot(view.NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer)), ctx, viewer, utils, client, nil, nil)

		require.NoError(t, root.handleMeasurement(t.Context(), measurement.ID, nil))
	})

	t.Run("table output error suppresses usage", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		measurement := createDefaultMeasurement("ping")
		client := apiMocks.NewMockClient(ctrl)
		client.EXPECT().GetMeasurement(t.Context(), measurement.ID).Return(measurement, nil)
		viewer := viewMocks.NewMockViewer(ctrl)
		viewer.EXPECT().OutputTable(measurement).Return("", assert.AnError)
		viewer.EXPECT().OutputShare()
		ctx := createDefaultContext()
		ctx.Table = true
		utils := utilsMocks.NewMockUtils(ctrl)
		utils.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()
		root := NewRoot(view.NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer)), ctx, viewer, utils, client, nil, nil)

		require.ErrorIs(t, root.handleMeasurement(t.Context(), measurement.ID, nil), assert.AnError)
		assert.True(t, root.Cmd.SilenceUsage)
		assert.False(t, root.Cmd.SilenceErrors)
	})

	t.Run("table output errors do not receive outcome-specific handling", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		measurement := createDefaultMeasurement("ping")
		client := apiMocks.NewMockClient(ctrl)
		client.EXPECT().GetMeasurement(t.Context(), measurement.ID).Return(measurement, nil)
		viewer := viewMocks.NewMockViewer(ctrl)
		viewer.EXPECT().OutputTable(measurement).Return("", view.ErrAllProbesFailed)
		viewer.EXPECT().OutputShare()
		ctx := createDefaultContext()
		ctx.Table = true
		utils := utilsMocks.NewMockUtils(ctrl)
		utils.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()
		root := NewRoot(view.NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer)), ctx, viewer, utils, client, nil, nil)

		require.ErrorIs(t, root.handleMeasurement(t.Context(), measurement.ID, nil), view.ErrAllProbesFailed)
		assert.True(t, root.Cmd.SilenceUsage)
		assert.False(t, root.Cmd.SilenceErrors)
	})

	t.Run("polling error finalizes table output", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		measurement := createDefaultMeasurement("ping")
		client := apiMocks.NewMockClient(ctrl)
		client.EXPECT().GetMeasurement(t.Context(), measurement.ID).Return(nil, assert.AnError)
		viewer := viewMocks.NewMockViewer(ctrl)
		viewer.EXPECT().OutputShare()
		ctx := createDefaultContext()
		ctx.Table = true
		utils := utilsMocks.NewMockUtils(ctrl)
		utils.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()
		root := NewRoot(view.NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer)), ctx, viewer, utils, client, nil, nil)

		require.ErrorIs(t, root.handleMeasurement(t.Context(), measurement.ID, nil), assert.AnError)
		assert.True(t, root.Cmd.SilenceUsage)
	})
}

func Test_HandleMeasurement_AllFailedOutcomesSucceedInFiniteModes(t *testing.T) {
	for _, mode := range []string{"default", "CI", "JSON", "latency", "table"} {
		t.Run(mode, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			measurement := createDefaultMeasurement("ping")
			measurement.Results[0].Result.Status = globalping.TestStatusFailed
			measurement.Results[0].Result.RawOutput = "target failed"
			client := apiMocks.NewMockClient(ctrl)
			viewer := viewMocks.NewMockViewer(ctrl)
			ctx := createDefaultContext()
			utils := utilsMocks.NewMockUtils(ctrl)
			utils.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

			switch mode {
			case "default":
				client.EXPECT().GetMeasurement(t.Context(), measurementID1).Return(measurement, nil)
				viewer.EXPECT().OutputDefault(measurementID1, measurement, gomock.Any())
			case "CI":
				ctx.CIMode = true
				client.EXPECT().AwaitMeasurement(t.Context(), measurementID1).Return(measurement, nil)
				viewer.EXPECT().OutputDefault(measurementID1, measurement, gomock.Any())
			case "JSON":
				ctx.ToJSON = true
				client.EXPECT().AwaitMeasurement(t.Context(), measurementID1).Return(measurement, nil)
				client.EXPECT().GetMeasurementRaw(t.Context(), measurementID1).Return([]byte(`{"status":"finished"}`), nil)
				viewer.EXPECT().OutputJSON(measurementID1, []byte(`{"status":"finished"}`))
			case "latency":
				ctx.ToLatency = true
				client.EXPECT().AwaitMeasurement(t.Context(), measurementID1).Return(measurement, nil)
				viewer.EXPECT().OutputLatency(measurementID1, measurement).Return(nil)
			case "table":
				ctx.Table = true
				client.EXPECT().GetMeasurement(t.Context(), measurementID1).Return(measurement, nil)
				viewer.EXPECT().OutputTable(measurement).Return("", nil)
				viewer.EXPECT().OutputShare()
			}

			root := NewRoot(view.NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer)), ctx, viewer, utils, client, nil, nil)

			err := root.handleMeasurement(t.Context(), measurementID1, &globalping.MeasurementCreate{})

			require.NoError(t, err)
		})
	}
}
