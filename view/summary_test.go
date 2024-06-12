package view

import (
	"bytes"
	"fmt"
	"math"
	"testing"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/stretchr/testify/assert"
)

func Test_OutputSummary(t *testing.T) {
	t.Run("No_stats", func(t *testing.T) {
		w := new(bytes.Buffer)
		ctx := createDefaultContext("ping")
		viewer := NewViewer(ctx, NewPrinter(nil, w, w), nil, nil)
		viewer.OutputSummary()

		assert.Equal(t, "", w.String())
	})

	t.Run("With_stats_Single_location", func(t *testing.T) {
		w := new(bytes.Buffer)
		ctx := createDefaultContext("ping")
		ctx.AggregatedStats = []*MeasurementStats{
			NewMeasurementStats(),
		}
		ctx.AggregatedStats[0].Sent = 1
		ctx.AggregatedStats[0].Rcv = 0
		ctx.AggregatedStats[0].Lost = 1
		ctx.AggregatedStats[0].Loss = 100
		ctx.AggregatedStats[0].Time = 1000
		hm := &HistoryItem{
			Id:     measurementID2,
			Status: globalping.StatusInProgress,
			Stats: []*MeasurementStats{
				{Sent: 9, Rcv: 9, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 1000, Tsum: 6.93, Tsum2: 5.3361},
			},
		}
		ctx.History.Push(hm)
		viewer := NewViewer(ctx, NewPrinter(nil, w, w), nil, nil)
		viewer.OutputSummary()

		assert.Equal(t, `
---  ping statistics ---
10 packets transmitted, 9 received, 10.00% packet loss, time 2000ms
rtt min/avg/max/mdev = 0.770/0.770/0.770/0.000 ms
`,
			w.String())
	})

	t.Run("Multiple_locations", func(t *testing.T) {
		w := new(bytes.Buffer)
		ctx := createDefaultContext("ping")
		ctx.AggregatedStats = []*MeasurementStats{
			NewMeasurementStats(),
			NewMeasurementStats(),
		}
		viewer := NewViewer(ctx, NewPrinter(nil, w, w), nil, nil)
		viewer.OutputSummary()

		assert.Equal(t, "", w.String())
	})

	t.Run("Single_location_Share", func(t *testing.T) {
		w := new(bytes.Buffer)
		ctx := createDefaultContext("ping")
		ctx.AggregatedStats = []*MeasurementStats{
			{Sent: 1, Rcv: 0, Lost: 1, Loss: 100, Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1, Time: 0},
		}
		ctx.Share = true
		viewer := NewViewer(ctx, NewPrinter(nil, w, w), nil, nil)
		viewer.OutputSummary()

		expectedOutput := `
---  ping statistics ---
1 packets transmitted, 0 received, 100.00% packet loss, time 0ms
rtt min/avg/max/mdev = -/-/-/- ms
` + fmt.Sprintf("\033[1;38;2;23;212;167m> View the results online: https://www.jsdelivr.com/globalping?measurement=%s\033[0m\n", measurementID1)

		assert.Equal(t, expectedOutput, w.String())
	})

	t.Run("Multiple_locations_Share", func(t *testing.T) {
		ctx := createDefaultContext("ping")
		ctx.AggregatedStats = []*MeasurementStats{
			NewMeasurementStats(),
			NewMeasurementStats(),
		}
		ctx.History.Push(&HistoryItem{Id: measurementID2})
		ctx.Share = true
		ctx.CIMode = true
		w := new(bytes.Buffer)
		viewer := NewViewer(ctx, NewPrinter(nil, w, w), nil, nil)
		viewer.OutputSummary()

		expectedOutput := fmt.Sprintf("\n> View the results online: https://www.jsdelivr.com/globalping?measurement=%s.%s\n", measurementID1, measurementID2)
		assert.Equal(t, expectedOutput, w.String())
	})

	t.Run("Multiple_locations_Share_More_calls_than_MaxHistory", func(t *testing.T) {
		history := NewHistoryBuffer(1)
		history.Push(&HistoryItem{Id: measurementID2})
		ctx := &Context{
			AggregatedStats: []*MeasurementStats{
				NewMeasurementStats(),
				NewMeasurementStats(),
			},
			History:             history,
			Share:               true,
			CIMode:              true,
			MeasurementsCreated: 2,
			Packets:             16,
		}
		w := new(bytes.Buffer)
		viewer := NewViewer(ctx, NewPrinter(nil, w, w), nil, nil)
		viewer.OutputSummary()

		expectedOutput := fmt.Sprintf("\n> View the results online: https://www.jsdelivr.com/globalping?measurement=%s", measurementID2) +
			"\nFor long-running continuous mode measurements, only the last 16 packets are shared.\n"
		assert.Equal(t, expectedOutput, w.String())
	})
}
