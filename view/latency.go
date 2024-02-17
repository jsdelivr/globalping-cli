package view

import (
	"errors"
	"fmt"

	"github.com/jsdelivr/globalping-cli/globalping"
)

// Outputs the latency stats for a measurement
func (v *viewer) OutputLatency(id string, data *globalping.Measurement) error {
	// Output every result in case of multiple probes
	for i, result := range data.Results {
		if i > 0 {
			// new line as separator if more than 1 result
			v.printer.Println()
		}

		v.printer.Println(generateProbeInfo(&result, !v.ctx.CIMode))

		switch v.ctx.Cmd {
		case "ping":
			stats, err := globalping.DecodePingStats(result.Result.StatsRaw)
			if err != nil {
				return err
			}
			v.printer.Println(v.latencyStatHeader("Min", v.ctx.CIMode) + fmt.Sprintf("%.2f ms", stats.Min))
			v.printer.Println(v.latencyStatHeader("Max", v.ctx.CIMode) + fmt.Sprintf("%.2f ms", stats.Max))
			v.printer.Println(v.latencyStatHeader("Avg", v.ctx.CIMode) + fmt.Sprintf("%.2f ms", stats.Avg))
		case "dns":
			timings, err := globalping.DecodeDNSTimings(result.Result.TimingsRaw)
			if err != nil {
				return err
			}
			v.printer.Println(v.latencyStatHeader("Total", v.ctx.CIMode) + fmt.Sprintf("%v ms", timings.Total))
		case "http":
			timings, err := globalping.DecodeHTTPTimings(result.Result.TimingsRaw)
			if err != nil {
				return err
			}
			v.printer.Println(v.latencyStatHeader("Total", v.ctx.CIMode) + fmt.Sprintf("%v ms", timings.Total))
			v.printer.Println(v.latencyStatHeader("Download", v.ctx.CIMode) + fmt.Sprintf("%v ms", timings.Download))
			v.printer.Println(v.latencyStatHeader("First byte", v.ctx.CIMode) + fmt.Sprintf("%v ms", timings.FirstByte))
			v.printer.Println(v.latencyStatHeader("DNS", v.ctx.CIMode) + fmt.Sprintf("%v ms", timings.DNS))
			v.printer.Println(v.latencyStatHeader("TLS", v.ctx.CIMode) + fmt.Sprintf("%v ms", timings.TLS))
			v.printer.Println(v.latencyStatHeader("TCP", v.ctx.CIMode) + fmt.Sprintf("%v ms", timings.TCP))
		default:
			return errors.New("unexpected command for latency output: " + v.ctx.Cmd)
		}
	}

	if v.ctx.Share {
		v.printer.Println(formatWithLeadingArrow(shareMessage(id), !v.ctx.CIMode))
	}
	v.printer.Println()

	return nil
}

func (v *viewer) latencyStatHeader(title string, ci bool) string {
	text := fmt.Sprintf("%s: ", title)
	if ci {
		return text
	} else {
		return terminalLayoutBold.Render(text)
	}
}
