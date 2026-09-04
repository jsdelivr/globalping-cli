package view

import (
	"errors"
	"fmt"

	"github.com/jsdelivr/globalping-go"
)

// Outputs the latency stats for a measurement
func (v *viewer) OutputLatency(id string, measurement *globalping.Measurement) error {
	// Output every result in case of multiple probes
	for i, result := range measurement.Results {
		if i > 0 {
			// new line as separator if more than 1 result
			v.printer.Println()
		}

		v.printer.ErrPrintln(v.getProbeInfo(&result))

		if result.Result.Status != globalping.TestStatusFinished {
			v.printer.Println(result.Result.RawOutput)

			continue
		}

		switch v.ctx.Cmd {
		case "ping":
			stats, err := globalping.DecodePingStats(result.Result.StatsRaw)

			if err != nil {
				return err
			}

			formatPingValue := func(value *float64) string {
				if value == nil {
					return "-"
				}

				return fmt.Sprintf("%.2f ms", *value)
			}

			v.printer.Println(v.latencyStatHeader("Min") + formatPingValue(stats.Min))
			v.printer.Println(v.latencyStatHeader("Max") + formatPingValue(stats.Max))
			v.printer.Println(v.latencyStatHeader("Avg") + formatPingValue(stats.Avg))
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

			formatHTTPValue := func(value *int) string {
				if value == nil {
					return "-"
				}

				return fmt.Sprintf("%v ms", *value)
			}

			v.printer.Println(v.latencyStatHeader("Total") + formatHTTPValue(timings.Total))
			v.printer.Println(v.latencyStatHeader("DNS") + formatHTTPValue(timings.DNS))
			v.printer.Println(v.latencyStatHeader("TCP") + formatHTTPValue(timings.TCP))
			v.printer.Println(v.latencyStatHeader("TLS") + formatHTTPValue(timings.TLS))
			v.printer.Println(v.latencyStatHeader("First byte") + formatHTTPValue(timings.FirstByte))
			v.printer.Println(v.latencyStatHeader("Download") + formatHTTPValue(timings.Download))
		default:
			return errors.New("unexpected command for latency output: " + v.ctx.Cmd)
		}
	}

	if v.ctx.Share {
		v.printer.ErrPrintln(v.getShareMessage(id))
	}

	v.printer.Println()

	return nil
}

func (v *viewer) latencyStatHeader(title string) string {
	return v.printer.Bold(title + ": ")
}
