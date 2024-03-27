package cmd

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/mocks"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_Execute_History_Default(t *testing.T) {
	t.Cleanup(sessionCleanup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	ctx := createDefaultContext("ping")
	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	root := NewRoot(printer, ctx, nil, timeMock, nil, nil)
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
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	timeStr := time.Unix(defaultCurrentTime.Unix(), 0).Format("2006-01-02 15:04:05")
	assert.Equal(t,
		createDefaultExpectedHistoryItem("1", timeStr, "ping jsdelivr.com", measurementID1)+
			createDefaultExpectedHistoryItem("-", timeStr, "ping jsdelivr.com from last", measurementID2)+
			createDefaultExpectedHistoryItem("2", timeStr, "ping jsdelivr.com", measurementID2)+
			createDefaultExpectedHistoryItem("3", timeStr, "ping jsdelivr.com", measurementID2)+
			createDefaultExpectedHistoryItem("4", timeStr, "ping jsdelivr.com", measurementID2)+
			createDefaultExpectedHistoryItem("5", timeStr, "ping jsdelivr.com", measurementID3),
		w.String())

	w.Reset()
	os.Args = []string{"globalping", "history", "--tail", "2"}
	err = root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t,
		createDefaultExpectedHistoryItem("5", timeStr, "ping jsdelivr.com", measurementID3)+
			createDefaultExpectedHistoryItem("4", timeStr, "ping jsdelivr.com", measurementID2),
		w.String())

	w.Reset()
	os.Args = []string{"globalping", "history", "--head", "2"}
	err = root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t,
		createDefaultExpectedHistoryItem("1", timeStr, "ping jsdelivr.com", measurementID1)+
			createDefaultExpectedHistoryItem("-", timeStr, "ping jsdelivr.com from last", measurementID2),
		w.String())
}
