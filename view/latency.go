package view

import (
	"fmt"
	"os"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
)

// Output latency values
func OutputLatency(id string, data *model.GetMeasurement, ctx model.Context) {
	// Output every result in case of multiple probes
	for i, result := range data.Results {
		if i > 0 {
			// new line as separator if more than 1 result
			fmt.Println()
		}

		fmt.Fprintln(os.Stderr, generateHeader(result, ctx))

		switch ctx.Cmd {
		case "ping":
			fmt.Println(latencyStatHeader("Min", ctx.CI) + fmt.Sprintf("%v ms", result.Result.Stats["min"]))
			fmt.Println(latencyStatHeader("Max", ctx.CI) + fmt.Sprintf("%v ms", result.Result.Stats["max"]))
			fmt.Println(latencyStatHeader("Avg", ctx.CI) + fmt.Sprintf("%v ms", result.Result.Stats["avg"]))
		case "dns":
			timings, err := client.DecodeTimings(ctx.Cmd, result.Result.TimingsRaw)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(latencyStatHeader("Total", ctx.CI) + fmt.Sprintf("%v ms", timings.Interface["total"]))
		case "http":
			timings, err := client.DecodeTimings(ctx.Cmd, result.Result.TimingsRaw)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(latencyStatHeader("Total", ctx.CI) + fmt.Sprintf("%v ms", timings.Interface["total"]))
			fmt.Println(latencyStatHeader("Download", ctx.CI) + fmt.Sprintf("%v ms", timings.Interface["download"]))
			fmt.Println(latencyStatHeader("First byte", ctx.CI) + fmt.Sprintf("%v ms", timings.Interface["firstByte"]))
			fmt.Println(latencyStatHeader("DNS", ctx.CI) + fmt.Sprintf("%v ms", timings.Interface["dns"]))
			fmt.Println(latencyStatHeader("TLS", ctx.CI) + fmt.Sprintf("%v ms", timings.Interface["tls"]))
			fmt.Println(latencyStatHeader("TCP", ctx.CI) + fmt.Sprintf("%v ms", timings.Interface["tcp"]))
		default:
			panic("unexpected command for latency output: " + ctx.Cmd)
		}

	}

	if ctx.Share {
		fmt.Fprintln(os.Stderr, formatWithLeadingArrow(ctx, shareMessage(id)))
	}
}

func latencyStatHeader(title string, ci bool) string {
	text := fmt.Sprintf("%s: ", title)
	if ci {
		return text
	} else {
		return terminalLayoutBold.Render(text)
	}
}
