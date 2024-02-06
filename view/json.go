package view

import (
	"fmt"
	"os"
)

// Outputs the raw JSON for a measurement
func (v *viewer) OutputJson(id string) error {
	output, err := v.gp.GetRawMeasurement(id)
	if err != nil {
		return err
	}
	fmt.Println(string(output))

	if v.ctx.Share {
		fmt.Fprintln(os.Stderr, formatWithLeadingArrow(shareMessage(id), !v.ctx.CI))
	}
	fmt.Println()

	return nil
}
