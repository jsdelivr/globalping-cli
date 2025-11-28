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
	"go.uber.org/mock/gomock"
)

func Test_Execute_Traceroute_Default(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("traceroute")
	expectedOpts.Limit = 2
	expectedOpts.Options.Protocol = "TCP"
	expectedOpts.Options.Port = 99

	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := apiMocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(t.Context(), expectedOpts).Times(1).Return(expectedResponse, nil)

	expectedMeasurement := createDefaultMeasurement("traceroute")
	gbMock.EXPECT().AwaitMeasurement(t.Context(), expectedResponse.ID).Times(1).Return(expectedMeasurement, nil)

	viewerMock := viewMocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().OutputDefault(measurementID1, expectedMeasurement, expectedOpts).Times(1)

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("traceroute")
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "traceroute", "jsdelivr.com",
		"from", "Berlin",
		"--limit", "2",
		"--protocol", "tcp",
		"--port", "99",
	}
	err := root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("traceroute")
	expectedCtx.Limit = 2
	expectedCtx.Protocol = "TCP"
	expectedCtx.Port = 99
	assert.Equal(t, expectedCtx, ctx)

	b, err := _storage.GetMeasurements()
	assert.NoError(t, err)
	expectedHistory := measurementID1 + "\n"
	assert.Equal(t, expectedHistory, string(b))

	items, err := _storage.GetHistory(0)
	assert.NoError(t, err)
	expectedHistoryItems := []string{createDefaultExpectedHistoryItem(
		"1",
		"traceroute jsdelivr.com from Berlin --limit 2 --protocol tcp --port 99",
		measurementID1,
	)}
	assert.Equal(t, expectedHistoryItems, items)
}

func Test_Execute_Traceroute_IPv4(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("traceroute")
	expectedOpts.Options.IPVersion = globalping.IPVersion4

	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := apiMocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(t.Context(), expectedOpts).Times(1).Return(expectedResponse, nil)

	expectedMeasurement := createDefaultMeasurement("traceroute")
	gbMock.EXPECT().AwaitMeasurement(t.Context(), expectedResponse.ID).Times(1).Return(expectedMeasurement, nil)

	viewerMock := viewMocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().OutputDefault(measurementID1, expectedMeasurement, expectedOpts).Times(1)

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("traceroute")
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "traceroute", "jsdelivr.com",
		"from", "Berlin",
		"--ipv4",
	}
	err := root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("traceroute")
	expectedCtx.Ipv4 = true
	assert.Equal(t, expectedCtx, ctx)
}

func Test_Execute_Traceroute_IPv6(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("traceroute")
	expectedOpts.Options.IPVersion = globalping.IPVersion6

	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := apiMocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(t.Context(), expectedOpts).Times(1).Return(expectedResponse, nil)

	expectedMeasurement := createDefaultMeasurement("traceroute")
	gbMock.EXPECT().AwaitMeasurement(t.Context(), expectedResponse.ID).Times(1).Return(expectedMeasurement, nil)

	viewerMock := viewMocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().OutputDefault(measurementID1, expectedMeasurement, expectedOpts).Times(1)

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("traceroute")
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "traceroute", "jsdelivr.com",
		"from", "Berlin",
		"--ipv6",
	}
	err := root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("traceroute")
	expectedCtx.Ipv6 = true
	assert.Equal(t, expectedCtx, ctx)
}

func Test_Execute_Traceroute_Invalid_Protocol(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("traceroute")
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, nil, utilsMock, nil, nil, _storage)

	os.Args = []string{"globalping", "traceroute", "jsdelivr.com", "--protocol", "invalid"}
	err := root.Cmd.ExecuteContext(t.Context())
	assert.Error(t, err, "protocol INVALID is not supported")

	items, err := _storage.GetHistory(0)
	assert.NoError(t, err)
	assert.Empty(t, items)
}
