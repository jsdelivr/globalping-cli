package view

import (
	"fmt"
	"os"
	"strings"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
)

// Output latency values
func OutputLatency(id string, data *model.GetMeasurement, ctx model.Context) {
	// String builder for output
	var output strings.Builder

	// Output every result in case of multiple probes
	for _, result := range data.Results {
		// Output slightly different format if state is available
		output.WriteString(generateHeader(result, ctx) + "\n")

		if ctx.Cmd == "ping" {
			output.WriteString(latencyStatHeader("Min", ctx.CI) + fmt.Sprintf("%v ms\n", result.Result.Stats["min"]))
			output.WriteString(latencyStatHeader("Max", ctx.CI) + fmt.Sprintf("%v ms\n", result.Result.Stats["max"]))
			output.WriteString(latencyStatHeader("Avg", ctx.CI) + fmt.Sprintf("%v ms\n\n", result.Result.Stats["avg"]))
		}

		if ctx.Cmd == "dns" {
			timings, err := client.DecodeTimings(ctx.Cmd, result.Result.TimingsRaw)
			if err != nil {
				fmt.Println(err)
				return
			}
			output.WriteString(latencyStatHeader("Total", ctx.CI) + fmt.Sprintf("%v ms\n\n", timings.Interface["total"]))
		}

		if ctx.Cmd == "http" {
			timings, err := client.DecodeTimings(ctx.Cmd, result.Result.TimingsRaw)
			if err != nil {
				fmt.Println(err)
				return
			}
			output.WriteString(latencyStatHeader("Total", ctx.CI) + fmt.Sprintf("%v ms\n", timings.Interface["total"]))
			output.WriteString(latencyStatHeader("Download", ctx.CI) + fmt.Sprintf("%v ms\n", timings.Interface["download"]))
			output.WriteString(latencyStatHeader("First byte", ctx.CI) + fmt.Sprintf("%v ms\n", timings.Interface["firstByte"]))
			output.WriteString(latencyStatHeader("DNS", ctx.CI) + fmt.Sprintf("%v ms\n", timings.Interface["dns"]))
			output.WriteString(latencyStatHeader("TLS", ctx.CI) + fmt.Sprintf("%v ms\n", timings.Interface["tls"]))
			output.WriteString(latencyStatHeader("TCP", ctx.CI) + fmt.Sprintf("%v ms\n\n", timings.Interface["tcp"]))
		}
	}

	fmt.Println(strings.TrimSpace(output.String()))

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
