package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	apiMocks "github.com/jsdelivr/globalping-cli/mocks/api"
	utilsMocks "github.com/jsdelivr/globalping-cli/mocks/utils"
	viewMocks "github.com/jsdelivr/globalping-cli/mocks/view"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/jsdelivr/globalping-go"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_Execute_MeasurementTimeout(t *testing.T) {
	for _, command := range []string{"ping", "traceroute", "dns", "mtr", "http"} {
		t.Run(command+" omitted", func(t *testing.T) {
			request := executeMeasurementUntilCreate(t, command)

			assert.Zero(t, request.Timeout)
			assertMeasurementRequestTimeout(t, request, 0, false)
		})

		t.Run(command+" provided", func(t *testing.T) {
			request := executeMeasurementUntilCreate(t, command, "--timeout", "17")

			assert.Equal(t, 17, request.Timeout)
			assertMeasurementRequestTimeout(t, request, 17, true)
		})
	}

	t.Run("client-side wait limit", func(t *testing.T) {
		startedAt := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)

		for _, test := range []struct {
			name     string
			timeout  int
			limit    time.Duration
			expected time.Duration
		}{
			{
				name:     "uses fallback when omitted",
				limit:    measurementAwaitDefaultTimeout,
				expected: measurementAwaitDefaultTimeout,
			},
			{
				name:     "adds finalization buffer",
				timeout:  5,
				limit:    5*time.Second + measurementAwaitTimeoutBuffer,
				expected: 5*time.Second + measurementAwaitTimeoutBuffer,
			},
		} {
			t.Run(test.name, func(t *testing.T) {
				state := newMeasurementAwaitState(measurementID1, startedAt, test.timeout)

				assert.Equal(t, test.expected, state.maxDuration)
				assert.NoError(t, state.checkTimeout(startedAt.Add(test.limit)))

				err := state.checkTimeout(startedAt.Add(test.limit + time.Nanosecond))
				assert.EqualError(t, err, "timed out waiting for measurement "+measurementID1+" to finish: context deadline exceeded")
				assert.ErrorIs(t, err, context.DeadlineExceeded)
			})
		}
	})
}

func Test_Execute_MeasurementTimeout_Invalid(t *testing.T) {
	for _, timeout := range []string{"0", "4", "31"} {
		t.Run(timeout, func(t *testing.T) {
			w := new(bytes.Buffer)
			root := NewRoot(view.NewPrinter(nil, w, w), createDefaultContext(), nil, nil, nil, nil, nil)
			os.Args = []string{"globalping", "ping", "jsdelivr.com", "--timeout", timeout}

			err := root.Cmd.ExecuteContext(t.Context())

			assert.EqualError(t, err, "timeout must be between 5 and 30 seconds")
		})
	}
}

func Test_Execute_MeasurementTimeout_Bounds(t *testing.T) {
	for _, timeout := range []struct {
		argument string
		expected int
	}{
		{argument: "5", expected: 5},
		{argument: "30", expected: 30},
	} {
		t.Run(timeout.argument, func(t *testing.T) {
			request := executeMeasurementUntilCreate(t, "ping", "--timeout", timeout.argument)

			assert.Equal(t, timeout.expected, request.Timeout)
			assertMeasurementRequestTimeout(t, request, timeout.expected, true)
		})
	}
}

func Test_Execute_MeasurementTimeout_OutputModes(t *testing.T) {
	for _, outputMode := range []struct {
		name string
		flag string
	}{
		{name: "ci", flag: "--ci"},
		{name: "json", flag: "--json"},
		{name: "latency", flag: "--latency"},
		{name: "table", flag: "--table"},
	} {
		t.Run(outputMode.name, func(t *testing.T) {
			request := executeMeasurementUntilCreate(t, "ping", "--timeout", "17", outputMode.flag)

			assert.Equal(t, 17, request.Timeout)
		})
	}
}

func Test_HandleMeasurement_TimeoutAndCancellation(t *testing.T) {
	for _, table := range []bool{false, true} {
		mode := "interactive"

		if table {
			mode = "table"
		}

		t.Run(mode+" timeout", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			measurement := createDefaultMeasurement("ping")
			measurement.Status = globalping.MeasurementStatusInProgress
			measurement.Timeout = 5

			client := apiMocks.NewMockClient(ctrl)
			client.EXPECT().GetMeasurement(t.Context(), measurement.ID).Return(measurement, nil)

			viewer := viewMocks.NewMockViewer(ctrl)

			if table {
				viewer.EXPECT().OutputTable(measurement).Return("", nil)
				viewer.EXPECT().OutputShare()
			}

			utils := utilsMocks.NewMockUtils(ctrl)
			gomock.InOrder(
				utils.EXPECT().Now().Return(defaultCurrentTime),
				utils.EXPECT().Now().Return(defaultCurrentTime.Add(15*time.Second+time.Nanosecond)),
			)

			ctx := createDefaultContext()
			ctx.Table = table
			root := NewRoot(view.NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer)), ctx, viewer, utils, client, nil, nil)

			err := root.handleMeasurement(t.Context(), measurement.ID, nil)

			assert.EqualError(t, err, "timed out waiting for measurement "+measurement.ID+" to finish: context deadline exceeded")
			assert.ErrorIs(t, err, context.DeadlineExceeded)
			assert.True(t, root.Cmd.SilenceUsage)
		})

		t.Run(mode+" cancellation", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			measurement := createDefaultMeasurement("ping")
			measurement.Status = globalping.MeasurementStatusInProgress
			measurement.Timeout = 5
			ctx, cancel := context.WithCancel(t.Context())

			client := apiMocks.NewMockClient(ctrl)
			client.EXPECT().GetMeasurement(ctx, measurement.ID).DoAndReturn(
				func(context.Context, string) (*globalping.Measurement, error) {
					cancel()

					return measurement, nil
				},
			)

			viewer := viewMocks.NewMockViewer(ctrl)

			if table {
				viewer.EXPECT().OutputTable(measurement).Return("", nil)
				viewer.EXPECT().OutputShare()
			}

			utils := utilsMocks.NewMockUtils(ctrl)
			utils.EXPECT().Now().Return(defaultCurrentTime).Times(2)

			viewCtx := createDefaultContext()
			viewCtx.Table = table
			viewCtx.APIMinInterval = time.Hour
			root := NewRoot(view.NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer)), viewCtx, viewer, utils, client, nil, nil)

			err := root.handleMeasurement(ctx, measurement.ID, nil)

			assert.ErrorIs(t, err, context.Canceled)
			assert.True(t, root.Cmd.SilenceUsage)
		})
	}
}

func Test_PingInfinite_MeasurementTimeoutAndCancellation(t *testing.T) {
	for _, test := range []struct {
		name      string
		finalNow  func(context.CancelFunc) time.Time
		expected  error
		exactText string
		timedOut  bool
	}{
		{
			name: "timeout prevents successor creation",
			finalNow: func(context.CancelFunc) time.Time {
				return defaultCurrentTime.Add(15*time.Second + time.Nanosecond)
			},
			expected:  context.DeadlineExceeded,
			exactText: "timed out waiting for measurement " + measurementID1 + " to finish: context deadline exceeded",
			timedOut:  true,
		},
		{
			name: "cancellation",
			finalNow: func(cancel context.CancelFunc) time.Time {
				cancel()

				return defaultCurrentTime
			},
			expected: context.Canceled,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			opts := createDefaultMeasurementCreate("ping")
			createResponse := createDefaultMeasurementCreateResponse()
			measurement := createDefaultMeasurement("ping")
			measurement.Status = globalping.MeasurementStatusInProgress
			measurement.Timeout = 5
			measurement.Results[0].Result.Status = globalping.TestStatusInProgress

			if test.timedOut {
				measurement = createDefaultMeasurement_MultipleProbes(globalping.MeasurementStatusInProgress, globalping.TestStatusInProgress)
				measurement.ProbesCount = 2
				measurement.Results = measurement.Results[:2]
				measurement.Results[0].Result.Status = globalping.TestStatusFinished
				measurement.Timeout = 5
			}

			client := apiMocks.NewMockClient(ctrl)
			createCalls := 0
			client.EXPECT().CreateMeasurement(ctx, opts).DoAndReturn(
				func(context.Context, *globalping.MeasurementCreate) (*globalping.MeasurementCreateResponse, error) {
					createCalls++

					return createResponse, nil
				},
			)
			client.EXPECT().GetMeasurement(ctx, measurement.ID).Return(measurement, nil)

			viewer := viewMocks.NewMockViewer(ctrl)

			if test.timedOut {
				viewer.EXPECT().OutputInfinite(measurement).Times(0)
			} else {
				viewer.EXPECT().OutputInfinite(measurement).Return("", nil)
			}

			utils := utilsMocks.NewMockUtils(ctrl)
			calls := []any{
				utils.EXPECT().Now().Return(defaultCurrentTime),
				utils.EXPECT().Now().Return(defaultCurrentTime),
				utils.EXPECT().Now().Return(defaultCurrentTime),
			}

			if test.timedOut {
				calls = append(calls, utils.EXPECT().Now().DoAndReturn(func() time.Time { return test.finalNow(cancel) }))
			} else {
				calls = append(
					calls,
					utils.EXPECT().Now().Return(defaultCurrentTime),
					utils.EXPECT().Now().DoAndReturn(func() time.Time { return test.finalNow(cancel) }),
				)
			}

			gomock.InOrder(calls...)

			viewCtx := createDefaultContext()
			viewCtx.APIMinInterval = time.Hour
			root := NewRoot(view.NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer)), viewCtx, viewer, utils, client, nil, nil)

			_, err := root.ping(ctx, opts)

			assert.ErrorIs(t, err, test.expected)
			assert.Equal(t, 1, createCalls)

			if test.exactText != "" {
				assert.EqualError(t, err, test.exactText)
				assert.True(t, root.Cmd.SilenceUsage)
			}
		})
	}
}

func Test_MeasurementTimeout_SharedHelpFlag(t *testing.T) {
	w := new(bytes.Buffer)
	root := NewRoot(view.NewPrinter(nil, w, w), createDefaultContext(), nil, nil, nil, nil, nil)

	assert.NoError(t, root.Cmd.Help())
	assert.Contains(t, w.String(), "--timeout int")

	var sharedFlag any

	for _, commandName := range []string{"ping", "traceroute", "dns", "mtr", "http"} {
		command, _, err := root.Cmd.Find([]string{commandName})
		assert.NoError(t, err)
		assert.NotNil(t, command)

		flag := command.Flags().Lookup("timeout")
		assert.NotNil(t, flag, "%s should expose --timeout", commandName)

		if sharedFlag == nil {
			sharedFlag = flag
		} else {
			assert.Same(t, sharedFlag, flag, "%s should use the shared measurement flag", commandName)
		}

		w.Reset()
		assert.NoError(t, command.Help())
		assert.Contains(t, w.String(), "--timeout int", "%s help should list --timeout", commandName)
	}
}

func executeMeasurementUntilCreate(t *testing.T, command string, extraArgs ...string) *globalping.MeasurementCreate {
	t.Helper()

	ctrl := gomock.NewController(t)
	expectedErr := errors.New("stop after capturing measurement request")
	var request *globalping.MeasurementCreate

	client := apiMocks.NewMockClient(ctrl)
	client.EXPECT().CreateMeasurement(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, measurement *globalping.MeasurementCreate) (*globalping.MeasurementCreateResponse, error) {
			request = measurement

			return nil, expectedErr
		},
	)

	w := new(bytes.Buffer)
	root := NewRoot(view.NewPrinter(nil, w, w), createDefaultContext(), nil, nil, client, nil, nil)
	os.Args = append([]string{"globalping", command, "jsdelivr.com"}, extraArgs...)

	err := root.Cmd.ExecuteContext(t.Context())

	assert.ErrorIs(t, err, expectedErr)
	assert.NotNil(t, request)

	return request
}

func assertMeasurementRequestTimeout(t *testing.T, request *globalping.MeasurementCreate, expected int, present bool) {
	t.Helper()

	body, err := json.Marshal(request)
	assert.NoError(t, err)

	var properties map[string]json.RawMessage
	assert.NoError(t, json.Unmarshal(body, &properties))
	timeout, ok := properties["timeout"]
	assert.Equal(t, present, ok)

	if present {
		var value int
		assert.NoError(t, json.Unmarshal(timeout, &value))
		assert.Equal(t, expected, value)
	}
}
