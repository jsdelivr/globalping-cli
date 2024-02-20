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

	expectedOpts := getDefaultMeasurementCreate("dns")
	expectedOpts.Limit = 2
	expectedOpts.Options.Protocol = "tcp"
	expectedOpts.Options.Port = 99
	expectedOpts.Options.Resolver = "1.1.1.1"
	expectedOpts.Options.Query = &globalping.QueryOptions{
		Type: "MX",
	}
	expectedOpts.Options.Trace = true

	expectedResponse := getDefaultMeasurementCreateResponse()

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(expectedOpts).Times(1).Return(expectedResponse, false, nil)

	viewerMock := mocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().Output(measurementID1, expectedOpts).Times(1).Return(nil)

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := getDefaultContext()
	root := NewRoot(printer, ctx, viewerMock, nil, gbMock, nil)

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

	expectedCtx := getDefaultExpectedContext("dns")
	expectedCtx.Limit = 2
	expectedCtx.Resolver = "1.1.1.1"
	expectedCtx.QueryType = "MX"
	expectedCtx.Protocol = "tcp"
	expectedCtx.Port = 99
	expectedCtx.Trace = true

	assert.Equal(t, expectedCtx, ctx)

	b, err := os.ReadFile(getMeasurementsPath())
	assert.NoError(t, err)
	expectedHistory := []byte(measurementID1 + "\n")
	assert.Equal(t, expectedHistory, b)
}
