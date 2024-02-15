package view

import (
	"fmt"
	"math"
)

func (v *viewer) OutputSummary() {
	if len(v.ctx.InProgressStats) == 0 {
		return
	}

	if len(v.ctx.InProgressStats) == 1 {
		stats := v.ctx.InProgressStats[0]

		v.printer.Printf("\n--- %s ping statistics ---\n", v.ctx.Hostname)
		v.printer.Printf("%d packets transmitted, %d received, %.2f%% packet loss, time %.0fms\n",
			stats.Sent,
			stats.Rcv,
			stats.Loss,
			stats.Time,
		)
		min := "-"
		avg := "-"
		max := "-"
		mdev := "-"
		if stats.Min != math.MaxFloat64 {
			min = fmt.Sprintf("%.3f", stats.Min)
		}
		if stats.Avg != -1 {
			avg = fmt.Sprintf("%.3f", stats.Avg)
		}
		if stats.Max != -1 {
			max = fmt.Sprintf("%.3f", stats.Max)
		}
		if stats.Mdev != 0 {
			mdev = fmt.Sprintf("%.3f", stats.Mdev)
		}
		v.printer.Printf("rtt min/avg/max/mdev = %s/%s/%s/%s ms\n", min, avg, max, mdev)
	}

	if v.ctx.Share && v.ctx.History != nil {
		if len(v.ctx.InProgressStats) > 1 {
			v.printer.Println()
		}
		ids := v.ctx.History.ToString("+")
		if ids != "" {
			v.printer.Println(formatWithLeadingArrow(shareMessage(ids), !v.ctx.CI))
		}
		if v.ctx.CallCount > v.ctx.MaxHistory {
			v.printer.Printf("For long-running continuous mode measurements, only the last %d packets are shared.\n",
				v.ctx.Packets*v.ctx.MaxHistory)
		}
	}
}
