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

func Test_Execute_MTR_Default(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("mtr")
	expectedOpts.Limit = 2
	expectedOpts.Options.Protocol = "TCP"
	expectedOpts.Options.Port = 99
	expectedOpts.Options.Packets = 16

	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := apiMocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(t.Context(), expectedOpts).Times(1).Return(expectedResponse, nil)

	expectedMeasurement := createDefaultMeasurement("mtr")
	gbMock.EXPECT().AwaitMeasurement(t.Context(), expectedResponse.ID).Times(1).Return(expectedMeasurement, nil)

	viewerMock := viewMocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().OutputDefault(measurementID1, expectedMeasurement, expectedOpts).Times(1)

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("mtr")
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "mtr", "jsdelivr.com",
		"from", "Berlin",
		"--limit", "2",
		"--protocol", "tcp",
		"--port", "99",
		"--packets", "16",
	}
	err := root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("mtr")
	expectedCtx.Limit = 2
	expectedCtx.Protocol = "TCP"
	expectedCtx.Port = 99
	expectedCtx.Packets = 16

	assert.Equal(t, expectedCtx, ctx)

	b, err := _storage.GetMeasurements()
	assert.NoError(t, err)
	expectedHistory := measurementID1 + "\n"
	assert.Equal(t, expectedHistory, string(b))

	items, err := _storage.GetHistory(0)
	assert.NoError(t, err)
	expectedHistoryItems := []string{createDefaultExpectedHistoryItem(
		"1",
		"mtr jsdelivr.com from Berlin --limit 2 --protocol tcp --port 99 --packets 16",
		measurementID1,
	)}
	assert.Equal(t, expectedHistoryItems, items)
}

func Test_Execute_MTR_IPv4(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("mtr")
	expectedOpts.Options.IPVersion = globalping.IPVersion4

	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := apiMocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(t.Context(), expectedOpts).Times(1).Return(expectedResponse, nil)

	expectedMeasurement := createDefaultMeasurement("mtr")
	gbMock.EXPECT().AwaitMeasurement(t.Context(), expectedResponse.ID).Times(1).Return(expectedMeasurement, nil)

	viewerMock := viewMocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().OutputDefault(measurementID1, expectedMeasurement, expectedOpts).Times(1)

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("mtr")
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "mtr", "jsdelivr.com",
		"from", "Berlin",
		"--ipv4",
	}
	err := root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("mtr")
	expectedCtx.Ipv4 = true

	assert.Equal(t, expectedCtx, ctx)
}

func Test_Execute_MTR_IPv6(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("mtr")
	expectedOpts.Options.IPVersion = globalping.IPVersion6

	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := apiMocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(t.Context(), expectedOpts).Times(1).Return(expectedResponse, nil)

	expectedMeasurement := createDefaultMeasurement("mtr")
	gbMock.EXPECT().AwaitMeasurement(t.Context(), expectedResponse.ID).Times(1).Return(expectedMeasurement, nil)

	viewerMock := viewMocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().OutputDefault(measurementID1, expectedMeasurement, expectedOpts).Times(1)

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("mtr")
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "mtr", "jsdelivr.com",
		"from", "Berlin",
		"--ipv6",
	}
	err := root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("mtr")
	expectedCtx.Ipv6 = true

	assert.Equal(t, expectedCtx, ctx)
}

func Test_Execute_MTR_Invalid_Protocol(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("mtr")
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, nil, utilsMock, nil, nil, _storage)

	os.Args = []string{"globalping", "mtr", "jsdelivr.com", "--protocol", "invalid"}
	err := root.Cmd.ExecuteContext(t.Context())
	assert.Error(t, err, "protocol INVALID is not supported")

	items, err := _storage.GetHistory(0)
	assert.NoError(t, err)
	assert.Empty(t, items)
}
