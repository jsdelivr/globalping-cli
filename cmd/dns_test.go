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
	t.Cleanup(sessionCleanup)

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

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("dns")
	root := NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)

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

	b, err := os.ReadFile(getMeasurementsPath())
	assert.NoError(t, err)
	expectedHistory := measurementID1 + "\n"
	assert.Equal(t, expectedHistory, string(b))

	b, err = os.ReadFile(getHistoryPath())
	assert.NoError(t, err)
	expectedHistory = createDefaultExpectedHistoryLogItem(
		"1",
		measurementID1,
		"dns jsdelivr.com from Berlin --limit 2 --type MX --resolver 1.1.1.1 --port 99 --protocol tcp --trace",
	)
	assert.Equal(t, expectedHistory, string(b))
}

func Test_Execute_DNS_IPv4(t *testing.T) {
	t.Cleanup(sessionCleanup)

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

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("dns")
	root := NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)

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
	t.Cleanup(sessionCleanup)

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

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("dns")
	root := NewRoot(printer, ctx, viewerMock, timeMock, gbMock, nil)

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
