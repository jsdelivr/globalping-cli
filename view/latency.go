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

		v.printer.Println(v.getProbeInfo(&result))

		switch v.ctx.Cmd {
		case "ping":
			stats, err := globalping.DecodePingStats(result.Result.StatsRaw)
			if err != nil {
				return err
			}
			v.printer.Println(v.latencyStatHeader("Min") + fmt.Sprintf("%.2f ms", stats.Min))
			v.printer.Println(v.latencyStatHeader("Max") + fmt.Sprintf("%.2f ms", stats.Max))
			v.printer.Println(v.latencyStatHeader("Avg") + fmt.Sprintf("%.2f ms", stats.Avg))
		case "dns":
			timings, err := globalping.DecodeDNSTimings(result.Result.TimingsRaw)
			if err != nil {
				return err
			}
			v.printer.Println(v.latencyStatHeader("Total") + fmt.Sprintf("%v ms", timings.Total))
		case "http":
			timings, err := globalping.DecodeHTTPTimings(result.Result.TimingsRaw)
			if err != nil {
				return err
			}
			v.printer.Println(v.latencyStatHeader("Total") + fmt.Sprintf("%v ms", timings.Total))
			v.printer.Println(v.latencyStatHeader("Download") + fmt.Sprintf("%v ms", timings.Download))
			v.printer.Println(v.latencyStatHeader("First byte") + fmt.Sprintf("%v ms", timings.FirstByte))
			v.printer.Println(v.latencyStatHeader("DNS") + fmt.Sprintf("%v ms", timings.DNS))
			v.printer.Println(v.latencyStatHeader("TLS") + fmt.Sprintf("%v ms", timings.TLS))
			v.printer.Println(v.latencyStatHeader("TCP") + fmt.Sprintf("%v ms", timings.TCP))
		default:
			return errors.New("unexpected command for latency output: " + v.ctx.Cmd)
		}
	}

	if v.ctx.Share {
		v.printer.Println(v.getShareMessage(id))
	}
	v.printer.Println()

	return nil
}

func (v *viewer) latencyStatHeader(title string) string {
	return v.printer.Bold(fmt.Sprintf("%s: ", title))
}
