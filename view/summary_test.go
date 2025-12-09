package view

import (
	"bytes"
	"testing"

	"github.com/jsdelivr/globalping-go"
	"github.com/stretchr/testify/assert"
)

func Test_OutputSummary(t *testing.T) {
	t.Run("No_stats", func(t *testing.T) {
		w := new(bytes.Buffer)
		ctx := createDefaultContext("ping")
		viewer := NewViewer(ctx, NewPrinter(nil, w, w), nil)
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
		viewer := NewViewer(ctx, NewPrinter(nil, w, w), nil)
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
		viewer := NewViewer(ctx, NewPrinter(nil, w, w), nil)
		viewer.OutputSummary()

		assert.Equal(t, "", w.String())
	})
}
