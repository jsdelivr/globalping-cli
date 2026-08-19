package view

import (
	"fmt"
	"math"
)

func (v *viewer) OutputSummary(infiniteTableOutput string) {
	if v.ctx.Infinite && v.ctx.Table {
		if infiniteTableOutput == "" {
			return
		}

		if v.ctx.CIMode {
			v.printer.Print(infiniteTableOutput)
		} else {
			v.printer.AreaUpdate(&infiniteTableOutput)
		}

		return
	}

	if len(v.ctx.AggregatedStats) != 1 {
		return
	}

	stats := v.aggregateConcurrentStats(v.ctx.AggregatedStats[0], 0, "")

	v.printer.Printf("\n--- %s ping statistics ---\n", v.ctx.Hostname)
	v.printer.Printf("%d packets transmitted, %d received, %.2f%% packet loss, time %.0fms\n",
		stats.Sent,
		stats.Rcv,
		stats.Loss,
		stats.Time,
	)
	minText := "-"
	avg := "-"
	maxText := "-"
	mdev := "-"

	if stats.Min != math.MaxFloat64 {
		minText = fmt.Sprintf("%.3f", stats.Min)
	}

	if stats.Avg != -1 {
		avg = fmt.Sprintf("%.3f", stats.Avg)
	}

	if stats.Max != -1 {
		maxText = fmt.Sprintf("%.3f", stats.Max)
	}

	if stats.Mdev != 0 {
		mdev = fmt.Sprintf("%.3f", stats.Mdev)
	}

	v.printer.Printf("rtt min/avg/max/mdev = %s/%s/%s/%s ms\n", minText, avg, maxText, mdev)
}
