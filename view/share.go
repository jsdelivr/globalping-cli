package view

func (v *viewer) OutputShare() {
	if !v.ctx.Share {
		return
	}
	if v.ctx.History == nil {
		return
	}

	if len(v.ctx.AggregatedStats) > 1 {
		v.printer.Println() // Add a newline in table view
	}
	ids := v.ctx.History.ToString(".")
	if ids != "" {
		v.printer.Println(v.getShareMessage(ids))
	}
	if v.ctx.MeasurementsCreated > v.ctx.History.Capacity() {
		v.printer.Printf("For long-running continuous mode measurements, only the last %d packets are shared.\n",
			v.ctx.Packets*v.ctx.History.Capacity())
	}
}
