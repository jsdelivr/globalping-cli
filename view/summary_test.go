package view

import (
	"bytes"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_OutputSummary(t *testing.T) {
	t.Run("No_stats", func(t *testing.T) {
		w := new(bytes.Buffer)
		viewer := NewViewer(&Context{}, NewPrinter(nil, w, w), nil, nil)
		viewer.OutputSummary()

		assert.Equal(t, "", w.String())
	})

	t.Run("With_stats_Single_location", func(t *testing.T) {
		w := new(bytes.Buffer)
		ctx := &Context{
			InProgressStats: []MeasurementStats{
				{Sent: 10, Rcv: 9, Lost: 1, Loss: 10, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 1000, Mdev: 0.001},
			},
		}
		viewer := NewViewer(ctx, NewPrinter(nil, w, w), nil, nil)
		viewer.OutputSummary()

		assert.Equal(t, `
---  ping statistics ---
10 packets transmitted, 9 received, 10.00% packet loss, time 1000ms
rtt min/avg/max/mdev = 0.770/0.770/0.770/0.001 ms
`,
			w.String())
	})

	t.Run("With_stats_In_progress", func(t *testing.T) {
		w := new(bytes.Buffer)
		ctx := &Context{
			InProgressStats: []MeasurementStats{
				{Sent: 1, Rcv: 0, Lost: 1, Loss: 100, Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1, Time: 0},
			},
		}
		viewer := NewViewer(ctx, NewPrinter(nil, w, w), nil, nil)
		viewer.OutputSummary()

		assert.Equal(t, `
---  ping statistics ---
1 packets transmitted, 0 received, 100.00% packet loss, time 0ms
rtt min/avg/max/mdev = -/-/-/- ms
`,
			w.String())
	})

	t.Run("Multiple_locations", func(t *testing.T) {
		w := new(bytes.Buffer)
		ctx := &Context{
			InProgressStats: []MeasurementStats{
				NewMeasurementStats(),
				NewMeasurementStats(),
			},
		}
		viewer := NewViewer(ctx, NewPrinter(nil, w, w), nil, nil)
		viewer.OutputSummary()

		assert.Equal(t, "", w.String())
	})

	t.Run("Single_location_Share", func(t *testing.T) {
		history := NewHistoryBuffer(1)
		history.Push(&HistoryItem{Id: measurementID1})

		w := new(bytes.Buffer)
		ctx := &Context{
			History: history,
			InProgressStats: []MeasurementStats{
				{Sent: 1, Rcv: 0, Lost: 1, Loss: 100, Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1, Time: 0},
			},
			Share: true,
		}
		viewer := NewViewer(ctx, NewPrinter(nil, w, w), nil, nil)
		viewer.OutputSummary()

		expectedOutput := `
---  ping statistics ---
1 packets transmitted, 0 received, 100.00% packet loss, time 0ms
rtt min/avg/max/mdev = -/-/-/- ms
` + formatWithLeadingArrow(shareMessage(measurementID1), true) + "\n"

		assert.Equal(t, expectedOutput, w.String())
	})

	t.Run("Multiple_locations_Share", func(t *testing.T) {
		history := NewHistoryBuffer(2)
		history.Push(&HistoryItem{Id: measurementID1})
		history.Push(&HistoryItem{Id: measurementID2})
		ctx := &Context{
			History: history,
			InProgressStats: []MeasurementStats{
				NewMeasurementStats(),
				NewMeasurementStats(),
			},
			Share: true,
		}
		w := new(bytes.Buffer)
		viewer := NewViewer(ctx, NewPrinter(nil, w, w), nil, nil)
		viewer.OutputSummary()

		expectedOutput := "\n" + formatWithLeadingArrow(shareMessage(measurementID1+"+"+measurementID2), true) + "\n"
		assert.Equal(t, expectedOutput, w.String())
	})

	t.Run("Multiple_locations_Share_More_calls_than_MaxHistory", func(t *testing.T) {
		history := NewHistoryBuffer(1)
		history.Push(&HistoryItem{Id: measurementID2})
		ctx := &Context{
			History: history,
			InProgressStats: []MeasurementStats{
				NewMeasurementStats(),
				NewMeasurementStats(),
			},
			Share:               true,
			MeasurementsCreated: 2,
			Packets:             16,
		}
		w := new(bytes.Buffer)
		viewer := NewViewer(ctx, NewPrinter(nil, w, w), nil, nil)
		viewer.OutputSummary()

		expectedOutput := "\n" + formatWithLeadingArrow(shareMessage(measurementID2), true) +
			"\nFor long-running continuous mode measurements, only the last 16 packets are shared.\n"
		assert.Equal(t, expectedOutput, w.String())
	})
}
