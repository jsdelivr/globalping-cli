package cmd

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"testing"

	utilsMocks "github.com/jsdelivr/globalping-cli/mocks/utils"
	"github.com/jsdelivr/globalping-cli/storage"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/jsdelivr/globalping-go"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_Execute_History_Default(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	ctx := createDefaultContext()
	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, nil, utilsMock, nil, nil, _storage)
	os.Args = []string{"globalping", "ping", "jsdelivr.com"}

	ctx.History.Push(&view.HistoryItem{
		Id:        measurementID1,
		Status:    globalping.MeasurementStatusInProgress,
		StartedAt: defaultCurrentTime,
	})
	assert.NoError(t, root.UpdateHistory())

	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "last"}
	ctx.IsLocationFromSession = true
	ctx.History.Push(&view.HistoryItem{
		Id:        measurementID2,
		Status:    globalping.MeasurementStatusInProgress,
		StartedAt: defaultCurrentTime,
	})
	assert.NoError(t, root.UpdateHistory())

	os.Args = []string{"globalping", "ping", "jsdelivr.com"}
	ctx.IsLocationFromSession = false
	assert.NoError(t, root.UpdateHistory())
	assert.NoError(t, root.UpdateHistory())
	assert.NoError(t, root.UpdateHistory())
	ctx.History.Push(&view.HistoryItem{
		Id:        measurementID3,
		Status:    globalping.MeasurementStatusInProgress,
		StartedAt: defaultCurrentTime,
	})
	assert.NoError(t, root.UpdateHistory())

	os.Args = []string{"globalping", "history"}
	err := root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)

	assert.Equal(t,
		createDefaultExpectedHistoryItem("1", "ping jsdelivr.com", measurementID1)+"\n"+
			createDefaultExpectedHistoryItem("-", "ping jsdelivr.com from last", measurementID2)+"\n"+
			createDefaultExpectedHistoryItem("2", "ping jsdelivr.com", measurementID2)+"\n"+
			createDefaultExpectedHistoryItem("3", "ping jsdelivr.com", measurementID2)+"\n"+
			createDefaultExpectedHistoryItem("4", "ping jsdelivr.com", measurementID2)+"\n"+
			createDefaultExpectedHistoryItem("5", "ping jsdelivr.com", measurementID3)+"\n",
		w.String())

	w.Reset()
	os.Args = []string{"globalping", "history", "--tail", "2"}
	err = root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)
	assert.Equal(t,
		createDefaultExpectedHistoryItem("5", "ping jsdelivr.com", measurementID3)+"\n"+
			createDefaultExpectedHistoryItem("4", "ping jsdelivr.com", measurementID2)+"\n",
		w.String())

	w.Reset()
	os.Args = []string{"globalping", "history", "--head", "2"}
	err = root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)
	assert.Equal(t,
		createDefaultExpectedHistoryItem("1", "ping jsdelivr.com", measurementID1)+"\n"+
			createDefaultExpectedHistoryItem("-", "ping jsdelivr.com from last", measurementID2)+"\n",
		w.String())
}

func Test_Execute_History_Empty(t *testing.T) {
	utilsMock := utilsMocks.NewMockUtils(gomock.NewController(t))
	_storage := createDefaultTestStorage(t, utilsMock)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	root := NewRoot(view.NewPrinter(nil, stdout, stderr), createDefaultContext(), nil, utilsMock, nil, nil, _storage)
	os.Args = []string{"globalping", "history"}

	err := root.Cmd.ExecuteContext(t.Context())

	assert.NoError(t, err)
	assert.Equal(t, "No history items found\n", stdout.String())
	assert.Empty(t, stderr.String())
}

func Test_Execute_History_ForwardScannerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	utilsMock := utilsMocks.NewMockUtils(ctrl)
	_storage := createDefaultTestStorage(t, utilsMock)
	assert.NoError(t, _storage.SaveCommandToHistory("1", defaultCurrentTime.Unix(), measurementID1, "command1"))
	assert.NoError(t, _storage.SaveCommandToHistory("2", defaultCurrentTime.Unix(), measurementID2, strings.Repeat("x", bufio.MaxScanTokenSize)))
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	root := NewRoot(view.NewPrinter(nil, stdout, stderr), createDefaultContext(), nil, utilsMock, nil, nil, _storage)
	os.Args = []string{"globalping", "history"}

	err := root.Cmd.ExecuteContext(t.Context())

	assert.ErrorIs(t, err, storage.ErrReadHistory)
	assert.ErrorContains(t, err, "token too long")
	assert.True(t, root.Cmd.SilenceUsage)
	assert.Empty(t, stdout.String())
	assert.Equal(t, 1, strings.Count(stderr.String(), "Error:"))
	assert.Contains(t, stderr.String(), "failed to get history")
	assert.Contains(t, stderr.String(), "token too long")
}
