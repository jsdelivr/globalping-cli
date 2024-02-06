package view

import (
	"fmt"
	"os"
	"strings"

	"github.com/jsdelivr/globalping-cli/globalping"
)

// Outputs non-json non-latency results for a measurement
func (v *viewer) outputDefault(id string, data *globalping.Measurement, m *globalping.MeasurementCreate) {
	for i := range data.Results {
		result := &data.Results[i]
		if i > 0 {
			// new line as separator if more than 1 result
			fmt.Println()
		}

		// Output slightly different format if state is available
		fmt.Fprintln(os.Stderr, generateProbeInfo(result, !v.ctx.CI))

		if v.isBodyOnlyHttpGet(m) {
			fmt.Println(strings.TrimSpace(result.Result.RawBody))
		} else {
			fmt.Println(strings.TrimSpace(result.Result.RawOutput))
		}
	}

	if v.ctx.Share {
		fmt.Fprintln(os.Stderr, formatWithLeadingArrow(shareMessage(id), !v.ctx.CI))
	}
}
