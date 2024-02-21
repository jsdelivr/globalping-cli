package cmd

import (
	"errors"
	"io/fs"
	"os"
	"time"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/view"
)

var (
	measurementID1 = "1OOxHNyhdsBQYEjU"
	measurementID2 = "2hUicONd75250Z1b"
	measurementID3 = "3PDXL29YeGctf6iJ"
	// measurementID4 = "4H3tBVPZEj5k6AcW"

	defaultCurrentTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
)

func sessionCleanup() {
	sessionPath := getSessionPath()
	err := os.RemoveAll(sessionPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		panic("Failed to remove session path: " + err.Error())
	}
}

func getDefaultMeasurementCreateResponse() *globalping.MeasurementCreateResponse {
	return &globalping.MeasurementCreateResponse{
		ID:          measurementID1,
		ProbesCount: 1,
	}
}

func getDefaultMeasurementCreate(cmd string) *globalping.MeasurementCreate {
	return &globalping.MeasurementCreate{
		Type:    cmd,
		Target:  "jsdelivr.com",
		Limit:   1,
		Options: &globalping.MeasurementOptions{},
		Locations: []globalping.Locations{
			{Magic: "Berlin"},
		},
	}
}

func getDefaultMeasurement(cmd string) *globalping.Measurement {
	return &globalping.Measurement{
		ID:          measurementID1,
		Status:      globalping.StatusFinished,
		Type:        cmd,
		ProbesCount: 1,
		Results: []globalping.ProbeMeasurement{
			{
				Result: globalping.ProbeResult{
					Status: globalping.StatusFinished,
				},
			},
		},
	}
}

func getDefaultContext() *view.Context {
	return &view.Context{
		History: view.NewHistoryBuffer(1),
	}
}

func getDefaultExpectedContext(cmd string) *view.Context {
	ctx := &view.Context{
		Cmd:                 cmd,
		Target:              "jsdelivr.com",
		From:                "Berlin",
		Limit:               1,
		CIMode:              true,
		History:             view.NewHistoryBuffer(1),
		MeasurementsCreated: 1,
	}
	if cmd == "ping" {
		ctx.History.Push(&view.HistoryItem{
			Id:        measurementID1,
			Status:    globalping.StatusInProgress,
			StartedAt: defaultCurrentTime,
		})
	}
	return ctx
}
