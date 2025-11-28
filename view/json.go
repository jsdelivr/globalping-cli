package view

// Outputs the raw JSON for a measurement
func (v *viewer) OutputJSON(id string, measurement []byte) {
	v.printer.Println(string(measurement))

	if v.ctx.Share {
		v.printer.ErrPrintln(v.getShareMessage(id))
	}
	v.printer.Println()
}
