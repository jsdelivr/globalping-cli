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

		v.printer.ErrPrintln(v.getProbeInfo(result))

		if v.ctx.Cmd == "http" {
			if v.ctx.Full {
				firstLineEnd := strings.Index(result.Result.RawOutput, "\n")
				if firstLineEnd > 0 {
					v.printer.ErrPrintln(result.Result.RawOutput[:firstLineEnd])
				}
				v.printer.ErrPrintln(result.Result.RawHeaders)
				v.printer.Println(strings.TrimSpace(result.Result.RawBody))
			} else if m.Options.Request.Method == "GET" {
				v.printer.Println(strings.TrimSpace(result.Result.RawBody))
			} else {
				v.printer.Println(strings.TrimSpace(result.Result.RawOutput))
			}
		} else {
			v.printer.Println(strings.TrimSpace(result.Result.RawOutput))
		}
	}

	if v.ctx.Share {
		v.printer.ErrPrintln(v.getShareMessage(id))
	}
}
