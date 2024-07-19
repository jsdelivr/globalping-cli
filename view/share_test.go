package view

import (
	"bytes"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_OutputShare(t *testing.T) {
	t.Run("Single_location", func(t *testing.T) {
		w := new(bytes.Buffer)
		ctx := createDefaultContext("ping")
		ctx.AggregatedStats = []*MeasurementStats{
			{Sent: 1, Rcv: 0, Lost: 1, Loss: 100, Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1, Time: 0},
		}
		ctx.Share = true
		viewer := NewViewer(ctx, NewPrinter(nil, w, w), nil, nil)
		viewer.OutputShare()

		expectedOutput := fmt.Sprintf("\033[1;38;2;23;212;167m> View the results online: https://www.jsdelivr.com/globalping?measurement=%s\033[0m\n", measurementID1)

		assert.Equal(t, expectedOutput, w.String())
	})

	t.Run("Multiple_locations", func(t *testing.T) {
		ctx := createDefaultContext("ping")
		ctx.AggregatedStats = []*MeasurementStats{
			NewMeasurementStats(),
			NewMeasurementStats(),
		}
		ctx.History.Push(&HistoryItem{Id: measurementID2})
		ctx.Share = true
		w := new(bytes.Buffer)
		printer := NewPrinter(nil, w, w)
		printer.DisableStyling()
		viewer := NewViewer(ctx, printer, nil, nil)
		viewer.OutputShare()

		expectedOutput := fmt.Sprintf("\n> View the results online: https://www.jsdelivr.com/globalping?measurement=%s.%s\n", measurementID1, measurementID2)
		assert.Equal(t, expectedOutput, w.String())
	})

	t.Run("Multiple_locations_More_calls_than_MaxHistory", func(t *testing.T) {
		history := NewHistoryBuffer(1)
		history.Push(&HistoryItem{Id: measurementID2})
		ctx := &Context{
			AggregatedStats: []*MeasurementStats{
				NewMeasurementStats(),
				NewMeasurementStats(),
			},
			History:             history,
			Share:               true,
			MeasurementsCreated: 2,
			Packets:             16,
		}
		w := new(bytes.Buffer)
		printer := NewPrinter(nil, w, w)
		printer.DisableStyling()
		viewer := NewViewer(ctx, printer, nil, nil)
		viewer.OutputShare()

		expectedOutput := fmt.Sprintf("\n> View the results online: https://www.jsdelivr.com/globalping?measurement=%s", measurementID2) +
			"\nFor long-running continuous mode measurements, only the last 16 packets are shared.\n"
		assert.Equal(t, expectedOutput, w.String())
	})
}
