package view

// Outputs the raw JSON for a measurement
func (v *viewer) OutputJson(id string) error {
	output, err := v.globalping.GetMeasurementRaw(id)
	if err != nil {
		return err
	}
	v.printer.Println(string(output))

	if v.ctx.Share {
		v.printer.Println(formatWithLeadingArrow(shareMessage(id), !v.ctx.CIMode))
	}
	v.printer.Println()

	return nil
}
