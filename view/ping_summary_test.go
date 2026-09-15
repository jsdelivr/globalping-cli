package view

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_OutputPingSummary(t *testing.T) {
	t.Run("No_stats", func(t *testing.T) {
		w := new(bytes.Buffer)
		ctx := createDefaultContext("ping")
		run := newPingTestRun(ctx, nil)
		viewer := NewViewer(ctx, NewPrinter(nil, w, w))
		viewer.OutputPingSummary("", run.Summary())

		assert.Equal(t, "", w.String())
	})

	t.Run("With_stats_Single_location", func(t *testing.T) {
		w := new(bytes.Buffer)
		ctx := createDefaultContext("ping")
		run := newPingTestRun(ctx, nil)
		run.completed = []*MeasurementStats{
			NewMeasurementStats(),
		}
		run.completed[0].Sent = 1
		run.completed[0].Rcv = 0
		run.completed[0].Lost = 1
		run.completed[0].Loss = 100
		run.completed[0].Time = 1000
		round := &pingRound{
			id: measurementID2,
			stats: []*MeasurementStats{
				{Sent: 9, Rcv: 9, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 1000, Tsum: 6.93, Tsum2: 5.3361},
			},
		}
		run.rounds = append(run.rounds, round)
		viewer := NewViewer(ctx, NewPrinter(nil, w, w))
		viewer.OutputPingSummary("", run.Summary())

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
		run := newPingTestRun(ctx, nil)
		run.completed = []*MeasurementStats{
			NewMeasurementStats(),
			NewMeasurementStats(),
		}
		viewer := NewViewer(ctx, NewPrinter(nil, w, w))
		viewer.OutputPingSummary("", run.Summary())

		assert.Equal(t, "", w.String())
	})
}
