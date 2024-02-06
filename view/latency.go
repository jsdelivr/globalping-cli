package view

import (
	"errors"
	"fmt"
	"os"

	"github.com/jsdelivr/globalping-cli/globalping"
)

// Outputs the latency stats for a measurement
func (v *viewer) OutputLatency(id string, data *globalping.Measurement) error {
	// Output every result in case of multiple probes
	for i, result := range data.Results {
		if i > 0 {
			// new line as separator if more than 1 result
			fmt.Println()
		}

		fmt.Fprintln(os.Stderr, generateProbeInfo(&result, !v.ctx.CI))

		switch v.ctx.Cmd {
		case "ping":
			stats, err := globalping.DecodePingStats(result.Result.StatsRaw)
			if err != nil {
				return err
			}
			fmt.Println(v.latencyStatHeader("Min", v.ctx.CI) + fmt.Sprintf("%.2f ms", stats.Min))
			fmt.Println(v.latencyStatHeader("Max", v.ctx.CI) + fmt.Sprintf("%.2f ms", stats.Max))
			fmt.Println(v.latencyStatHeader("Avg", v.ctx.CI) + fmt.Sprintf("%.2f ms", stats.Avg))
		case "dns":
			timings, err := globalping.DecodeDNSTimings(result.Result.TimingsRaw)
			if err != nil {
				return err
			}
			fmt.Println(v.latencyStatHeader("Total", v.ctx.CI) + fmt.Sprintf("%v ms", timings.Total))
		case "http":
			timings, err := globalping.DecodeHTTPTimings(result.Result.TimingsRaw)
			if err != nil {
				return err
			}
			fmt.Println(v.latencyStatHeader("Total", v.ctx.CI) + fmt.Sprintf("%v ms", timings.Total))
			fmt.Println(v.latencyStatHeader("Download", v.ctx.CI) + fmt.Sprintf("%v ms", timings.Download))
			fmt.Println(v.latencyStatHeader("First byte", v.ctx.CI) + fmt.Sprintf("%v ms", timings.FirstByte))
			fmt.Println(v.latencyStatHeader("DNS", v.ctx.CI) + fmt.Sprintf("%v ms", timings.DNS))
			fmt.Println(v.latencyStatHeader("TLS", v.ctx.CI) + fmt.Sprintf("%v ms", timings.TLS))
			fmt.Println(v.latencyStatHeader("TCP", v.ctx.CI) + fmt.Sprintf("%v ms", timings.TCP))
		default:
			return errors.New("unexpected command for latency output: " + v.ctx.Cmd)
		}
	}

	if v.ctx.Share {
		fmt.Fprintln(os.Stderr, formatWithLeadingArrow(shareMessage(id), !v.ctx.CI))
	}
	fmt.Println()

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
