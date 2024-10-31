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

func Test_Execute_DNS_Default(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("dns")
	expectedOpts.Limit = 2
	expectedOpts.Options.Protocol = "tcp"
	expectedOpts.Options.Port = 99
	expectedOpts.Options.Resolver = "1.1.1.1"
	expectedOpts.Options.Query = &globalping.QueryOptions{
		Type: "MX",
	}
	expectedOpts.Options.Trace = true

	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(expectedOpts).Times(1).Return(expectedResponse, nil)

	viewerMock := mocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().Output(measurementID1, expectedOpts).Times(1).Return(nil)

	utilsMock := mocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("dns")
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)

	os.Args = []string{"globalping", "dns", "jsdelivr.com",
		"from", "Berlin",
		"--limit", "2",
		"--type", "MX",
		"--resolver", "1.1.1.1",
		"--port", "99",
		"--protocol", "tcp",
		"--trace"}
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("dns")
	expectedCtx.Limit = 2
	expectedCtx.Resolver = "1.1.1.1"
	expectedCtx.QueryType = "MX"
	expectedCtx.Protocol = "tcp"
	expectedCtx.Port = 99
	expectedCtx.Trace = true

	assert.Equal(t, expectedCtx, ctx)

	b, err := _storage.GetMeasurements()
	assert.NoError(t, err)
	expectedHistory := measurementID1 + "\n"
	assert.Equal(t, expectedHistory, string(b))

	items, err := _storage.GetHistory(0)
	assert.NoError(t, err)
	expectedHistoryItems := []string{createDefaultExpectedHistoryItem(
		"1",
		"dns jsdelivr.com from Berlin --limit 2 --type MX --resolver 1.1.1.1 --port 99 --protocol tcp --trace",
		measurementID1,
	)}
	assert.Equal(t, expectedHistoryItems, items)
}

func Test_Execute_DNS_IPv4(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("dns")
	expectedOpts.Options.IPVersion = globalping.IPVersion4
	expectedOpts.Options.Query = &globalping.QueryOptions{}

	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(expectedOpts).Times(1).Return(expectedResponse, nil)

	viewerMock := mocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().Output(measurementID1, expectedOpts).Times(1).Return(nil)

	utilsMock := mocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("dns")
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)

	os.Args = []string{"globalping", "dns", "jsdelivr.com",
		"from", "Berlin",
		"--ipv4"}
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("dns")
	expectedCtx.Ipv4 = true

	assert.Equal(t, expectedCtx, ctx)
}

func Test_Execute_DNS_IPv6(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("dns")
	expectedOpts.Options.IPVersion = globalping.IPVersion6
	expectedOpts.Options.Query = &globalping.QueryOptions{}

	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(expectedOpts).Times(1).Return(expectedResponse, nil)

	viewerMock := mocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().Output(measurementID1, expectedOpts).Times(1).Return(nil)

	utilsMock := mocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("dns")
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)

	os.Args = []string{"globalping", "dns", "jsdelivr.com",
		"from", "Berlin",
		"--ipv6"}
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("dns")
	expectedCtx.Ipv6 = true

	assert.Equal(t, expectedCtx, ctx)
}
