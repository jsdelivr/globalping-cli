package view

import (
	"strings"

	"github.com/jsdelivr/globalping-cli/globalping"
)

// Outputs non-json non-latency results for a measurement
func (v *viewer) outputDefault(id string, data *globalping.Measurement, m *globalping.MeasurementCreate) {
	for i := range data.Results {
		result := &data.Results[i]
		if i > 0 {
			// new line as separator if more than 1 result
			v.printer.Println()
		}

		// Output slightly different format if state is available
		v.printer.Println(generateProbeInfo(result, !v.ctx.CIMode))

		if v.isBodyOnlyHttpGet(m) {
			v.printer.Println(strings.TrimSpace(result.Result.RawBody))
		} else {
			v.printer.Println(strings.TrimSpace(result.Result.RawOutput))
		}
	}

	if v.ctx.Share {
		v.printer.Println(formatWithLeadingArrow(shareMessage(id), !v.ctx.CIMode))
	}
}
