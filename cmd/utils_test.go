package cmd

import (
	"fmt"
	"testing"
	"time"

	"github.com/jsdelivr/globalping-cli/storage"
	"github.com/jsdelivr/globalping-cli/utils"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/jsdelivr/globalping-go"
)

var (
	measurementID1 = "1OOxHNyhdsBQYEjU"
	measurementID2 = "2hUicONd75250Z1b"
	measurementID3 = "3PDXL29YeGctf6iJ"
	measurementID4 = "4H3tBVPZEj5k6AcW"

	defaultCurrentTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
)

func createDefaultMeasurementCreateResponse() *globalping.MeasurementCreateResponse {
	return &globalping.MeasurementCreateResponse{
		ID:          measurementID1,
		ProbesCount: 1,
	}
}

func createDefaultMeasurementCreate(cmd globalping.MeasurementType) *globalping.MeasurementCreate {
	measurement := &globalping.MeasurementCreate{
		Type:    cmd,
		Target:  "jsdelivr.com",
		Limit:   1,
		Options: &globalping.MeasurementOptions{},
		Locations: []globalping.Locations{
			{Magic: "Berlin"},
		},
	}
	switch cmd {
	case "ping":
		measurement.Options.Protocol = "ICMP"
		measurement.Options.Port = 80
	case "traceroute":
		measurement.Options.Protocol = "ICMP"
		measurement.Options.Port = 80
	case "dns":
		measurement.Options.Protocol = "UDP"
		measurement.Options.Port = 53
	case "mtr":
		measurement.Options.Protocol = "ICMP"
		measurement.Options.Port = 80
	case "http":
		measurement.Options.Protocol = "HTTPS"
		measurement.Options.Port = 443
	}
	return measurement
}

func createDefaultMeasurement(cmd globalping.MeasurementType) *globalping.Measurement {
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

func createDefaultMeasurement_MultipleProbes(cmd globalping.MeasurementType, status globalping.MeasurementStatus) *globalping.Measurement {
	return &globalping.Measurement{
		ID:          measurementID1,
		Status:      status,
		Type:        cmd,
		ProbesCount: 3,
		Results: []globalping.ProbeMeasurement{
			{
				Result: globalping.ProbeResult{
					Status: status,
				},
			},
			{
				Result: globalping.ProbeResult{
					Status: status,
				},
			},
			{
				Result: globalping.ProbeResult{
					Status: status,
				},
			},
		},
	}
}

func createDefaultContext(_ string) *view.Context {
	ctx := &view.Context{
		History:             view.NewHistoryBuffer(1),
		From:                "world",
		Limit:               1,
		RunSessionStartedAt: defaultCurrentTime,
	}
	return ctx
}

func createDefaultTestStorage(t *testing.T, utils utils.Utils) *storage.LocalStorage {
	s := storage.NewLocalStorage(utils)
	err := s.Init("globalping-cli_" + t.Name())
	if err != nil {
		panic(err)
	}
	t.Cleanup(func() {
		s.Remove()
	})
	return s
}

func createDefaultExpectedContext(cmd string) *view.Context {
	ctx := &view.Context{
		Cmd:                 cmd,
		Target:              "jsdelivr.com",
		From:                "Berlin",
		Limit:               1,
		CIMode:              true,
		History:             view.NewHistoryBuffer(1),
		MeasurementsCreated: 1,
		RunSessionStartedAt: defaultCurrentTime,
	}
	switch cmd {
	case "ping":
		ctx.Protocol = "ICMP"
		ctx.Port = 80
	case "traceroute":
		ctx.Protocol = "ICMP"
		ctx.Port = 80
	case "dns":
		ctx.Protocol = "UDP"
		ctx.Port = 53
	case "mtr":
		ctx.Protocol = "ICMP"
		ctx.Port = 80
	case "http":
		ctx.Protocol = "HTTPS"
		ctx.Port = 443
	}
	ctx.History.Push(&view.HistoryItem{
		Id:        measurementID1,
		Status:    globalping.StatusInProgress,
		StartedAt: defaultCurrentTime,
	})
	return ctx
}

func createDefaultExpectedHistoryItem(index string, cmd string, measurements string) string {
	return fmt.Sprintf("%s | %s | %s\n> https://globalping.io?measurement=%s",
		index,
		time.Unix(defaultCurrentTime.Unix(), 0).Format("2006-01-02 15:04:05"),
		cmd,
		measurements,
	)
}
