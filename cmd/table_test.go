package cmd

import (
	"bytes"
	"os"
	"testing"

	apiMocks "github.com/jsdelivr/globalping-cli/mocks/api"
	utilsMocks "github.com/jsdelivr/globalping-cli/mocks/utils"
	viewMocks "github.com/jsdelivr/globalping-cli/mocks/view"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/jsdelivr/globalping-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_TableFlag_IsAvailableOnlyForMeasurementCommands(t *testing.T) {
	ctx := createDefaultContext("")
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
			expectedOpts.Locations.(globalping.LocationOptions)[0].Magic = "world"
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
			gbMock.EXPECT().AwaitMeasurement(t.Context(), expectedResponse.ID).Return(expectedMeasurement, nil)

			viewerMock := viewMocks.NewMockViewer(ctrl)
			viewerMock.EXPECT().OutputTable(expectedMeasurement).Return(nil)
			viewerMock.EXPECT().OutputShare()

			utilsMock := utilsMocks.NewMockUtils(ctrl)
			utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

			w := new(bytes.Buffer)
			ctx := createDefaultContext(string(measurementType))
			storage := createDefaultTestStorage(t, utilsMock)
			root := NewRoot(view.NewPrinter(nil, w, w), ctx, viewerMock, utilsMock, gbMock, nil, storage)
			oldArgs := os.Args
			t.Cleanup(func() { os.Args = oldArgs })
			os.Args = []string{"globalping", string(measurementType), "jsdelivr.com", "--table", "--latency", "--ci"}

			err := root.Cmd.ExecuteContext(t.Context())

			require.NoError(t, err)
			assert.True(t, ctx.Table)
			assert.Empty(t, w.String())
		})
	}
}

func Test_HandleMeasurement_TableTakesOutputPrecedence(t *testing.T) {
	t.Run("table before latency", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		measurement := createDefaultMeasurement("ping")
		client := apiMocks.NewMockClient(ctrl)
		client.EXPECT().AwaitMeasurement(t.Context(), measurement.ID).Return(measurement, nil)
		viewer := viewMocks.NewMockViewer(ctrl)
		viewer.EXPECT().OutputTable(measurement).Return(nil)
		viewer.EXPECT().OutputShare()
		ctx := createDefaultContext("ping")
		ctx.Table = true
		ctx.ToLatency = true
		root := NewRoot(view.NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer)), ctx, viewer, nil, client, nil, nil)

		require.NoError(t, root.handleMeasurement(t.Context(), measurement.ID, nil))
	})

	t.Run("table before json", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		measurement := createDefaultMeasurement("ping")
		client := apiMocks.NewMockClient(ctrl)
		client.EXPECT().AwaitMeasurement(t.Context(), measurement.ID).Return(measurement, nil)
		viewer := viewMocks.NewMockViewer(ctrl)
		viewer.EXPECT().OutputTable(measurement).Return(nil)
		viewer.EXPECT().OutputShare()
		ctx := createDefaultContext("ping")
		ctx.Table = true
		ctx.ToJSON = true
		root := NewRoot(view.NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer)), ctx, viewer, nil, client, nil, nil)

		require.NoError(t, root.handleMeasurement(t.Context(), measurement.ID, nil))
	})

	t.Run("table output error suppresses usage", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		measurement := createDefaultMeasurement("ping")
		client := apiMocks.NewMockClient(ctrl)
		client.EXPECT().AwaitMeasurement(t.Context(), measurement.ID).Return(measurement, nil)
		viewer := viewMocks.NewMockViewer(ctrl)
		viewer.EXPECT().OutputTable(measurement).Return(assert.AnError)
		viewer.EXPECT().OutputShare()
		ctx := createDefaultContext("ping")
		ctx.Table = true
		root := NewRoot(view.NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer)), ctx, viewer, nil, client, nil, nil)

		require.ErrorIs(t, root.handleMeasurement(t.Context(), measurement.ID, nil), assert.AnError)
		assert.True(t, root.Cmd.SilenceUsage)
	})
}
