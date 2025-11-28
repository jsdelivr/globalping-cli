package cmd

import (
	"bytes"
	"os"
	"testing"

	utilsMocks "github.com/jsdelivr/globalping-cli/mocks/utils"
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

	ctx := createDefaultContext("ping")
	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, nil, utilsMock, nil, nil, _storage)
	os.Args = []string{"globalping", "ping", "jsdelivr.com"}

	ctx.History.Push(&view.HistoryItem{
		Id:        measurementID1,
		Status:    globalping.StatusInProgress,
		StartedAt: defaultCurrentTime,
	})
	root.UpdateHistory()

	os.Args = []string{"globalping", "ping", "jsdelivr.com", "from", "last"}
	ctx.IsLocationFromSession = true
	ctx.History.Push(&view.HistoryItem{
		Id:        measurementID2,
		Status:    globalping.StatusInProgress,
		StartedAt: defaultCurrentTime,
	})
	root.UpdateHistory()

	os.Args = []string{"globalping", "ping", "jsdelivr.com"}
	ctx.IsLocationFromSession = false
	root.UpdateHistory()
	root.UpdateHistory()
	root.UpdateHistory()
	ctx.History.Push(&view.HistoryItem{
		Id:        measurementID3,
		Status:    globalping.StatusInProgress,
		StartedAt: defaultCurrentTime,
	})
	root.UpdateHistory()

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
