package view

import (
	"errors"
	"fmt"
	"os"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
)

// Outputs the latency stats for a measurement
func OutputLatency(id string, data *model.GetMeasurement, ctx model.Context) error {
	// Output every result in case of multiple probes
	for i, result := range data.Results {
		if i > 0 {
			// new line as separator if more than 1 result
			fmt.Println()
		}

		fmt.Fprintln(os.Stderr, generateProbeInfo(&result, !ctx.CI))

		switch ctx.Cmd {
		case "ping":
			stats, err := client.DecodePingStats(result.Result.StatsRaw)
			if err != nil {
				return err
			}
			fmt.Println(latencyStatHeader("Min", ctx.CI) + fmt.Sprintf("%.2f ms", stats.Min))
			fmt.Println(latencyStatHeader("Max", ctx.CI) + fmt.Sprintf("%.2f ms", stats.Max))
			fmt.Println(latencyStatHeader("Avg", ctx.CI) + fmt.Sprintf("%.2f ms", stats.Avg))
		case "dns":
			timings, err := client.DecodeDNSTimings(result.Result.TimingsRaw)
			if err != nil {
				return err
			}
			fmt.Println(latencyStatHeader("Total", ctx.CI) + fmt.Sprintf("%v ms", timings.Total))
		case "http":
			timings, err := client.DecodeHTTPTimings(result.Result.TimingsRaw)
			if err != nil {
				return err
			}
			fmt.Println(latencyStatHeader("Total", ctx.CI) + fmt.Sprintf("%v ms", timings.Total))
			fmt.Println(latencyStatHeader("Download", ctx.CI) + fmt.Sprintf("%v ms", timings.Download))
			fmt.Println(latencyStatHeader("First byte", ctx.CI) + fmt.Sprintf("%v ms", timings.FirstByte))
			fmt.Println(latencyStatHeader("DNS", ctx.CI) + fmt.Sprintf("%v ms", timings.DNS))
			fmt.Println(latencyStatHeader("TLS", ctx.CI) + fmt.Sprintf("%v ms", timings.TLS))
			fmt.Println(latencyStatHeader("TCP", ctx.CI) + fmt.Sprintf("%v ms", timings.TCP))
		default:
			return errors.New("unexpected command for latency output: " + ctx.Cmd)
		}
	}

	if ctx.Share {
		fmt.Fprintln(os.Stderr, formatWithLeadingArrow(shareMessage(id), !ctx.CI))
	}
	fmt.Println()

	return nil
}

func latencyStatHeader(title string, ci bool) string {
	text := fmt.Sprintf("%s: ", title)
	if ci {
		return text
	} else {
		return terminalLayoutBold.Render(text)
	}
}
