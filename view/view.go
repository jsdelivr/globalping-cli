package view

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/pterm/pterm"
)

var (
	// UI styles
	highlight = lipgloss.NewStyle().
			Bold(true).Foreground(lipgloss.Color("#17D4A7"))

	arrow = lipgloss.NewStyle().SetString(">").Bold(true).Foreground(lipgloss.Color("#17D4A7")).PaddingRight(1).String()

	bold = lipgloss.NewStyle().Bold(true)
)

// Used to slice the output to fit the terminal in live view
func sliceOutput(output string, w, h int) string {
	// Split output into lines
	lines := strings.Split(output, "\n")

	// Subtract 2 lines from height to account for the header
	h = h - 2

	// If output is too long, slice it in reverse
	if len(lines) > h {
		lines = lines[len(lines)-h:]
	}

	// If any line is too long, slice it
	for i, line := range lines {
		if len(line) > w {
			lines[i] = line[:w]
		}
	}

	// Join lines back into a string
	return strings.Join(lines, "\n")
}

// Generate header that also checks if the probe has a state in it in the form %s, %s, (%s), %s, ASN:%d
func generateHeader(result model.MeasurementResponse, ctx model.Context) string {
	var output strings.Builder

	// Continent + Country + (State) + City + ASN + Network + (Region Tag)
	output.WriteString(result.Probe.Continent + ", " + result.Probe.Country + ", ")
	if result.Probe.State != "" {
		output.WriteString("(" + result.Probe.State + "), ")
	}
	output.WriteString(result.Probe.City + ", ASN:" + fmt.Sprint(result.Probe.ASN) + ", " + result.Probe.Network)

	// Check tags to see if there's a region code
	if len(result.Probe.Tags) > 0 {
		for _, tag := range result.Probe.Tags {
			// If tag ends in a number, it's likely a region code and should be displayed
			if _, err := strconv.Atoi(tag[len(tag)-1:]); err == nil {
				output.WriteString(" (" + tag + ")")
				break
			}
		}
	}

	if ctx.CI {
		return "> " + output.String()
	} else {
		return arrow + highlight.Render(output.String())
	}
}

var apiPollInterval = 500 * time.Millisecond

func LiveView(id string, data *model.GetMeasurement, ctx model.Context) {
	var err error

	// Create new writer
	areaPrinter, err := pterm.DefaultArea.Start()
	if err != nil {
		fmt.Printf("failed to start writer: %v\n", err)
		return
	}
	areaPrinter.RemoveWhenDone = true

	defer func() {
		// Stop area printer and clear area if not already done
		err := areaPrinter.Stop()
		if err != nil {
			fmt.Printf("failed to stop writer: %v\n", err)
		}
	}()

	w, h, err := pterm.GetTerminalSize()
	if err != nil {
		fmt.Printf("failed to get terminal size: %v\n", err)
		return
	}

	// String builder for output
	var output strings.Builder

	fetcher := client.NewMeasurementsFetcher(client.ApiUrl)

	// Poll API until the measurement is complete
	for data.Status == "in-progress" {
		time.Sleep(apiPollInterval)
		data, err = fetcher.GetMeasurement(id)
		if err != nil {
			fmt.Printf("failed to get data: %v\n", err)
			return
		}

		// Reset string builder
		output.Reset()

		// Output every result in case of multiple probes
		for _, result := range data.Results {
			// Output slightly different format if state is available
			output.WriteString(generateHeader(result, ctx) + "\n")

			// Output only latency values if flag is set
			output.WriteString(strings.TrimSpace(result.Result.RawOutput) + "\n\n")
		}

		areaPrinter.Update(sliceOutput(output.String(), w, h))
	}

	// Stop area printer and clear area
	err = areaPrinter.Stop()
	if err != nil {
		fmt.Printf("failed to stop writer: %v\n", err)
	}

	// Print full output
	fmt.Println(strings.TrimSpace(output.String()))
}

// If json flag is used, only output json
func OutputJson(id string, fetcher client.MeasurementsFetcher) {
	output, err := fetcher.GetRawMeasurement(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(output))
}

// If latency flag is used, only output latency values
func OutputLatency(id string, data *model.GetMeasurement, ctx model.Context) {
	// String builder for output
	var output strings.Builder

	// Output every result in case of multiple probes
	for _, result := range data.Results {
		// Output slightly different format if state is available
		output.WriteString(generateHeader(result, ctx) + "\n")

		if ctx.CI {
			if ctx.Cmd == "ping" {
				output.WriteString(fmt.Sprintf("Min: %v ms\n", result.Result.Stats["min"]))
				output.WriteString(fmt.Sprintf("Max: %v ms\n", result.Result.Stats["max"]))
				output.WriteString(fmt.Sprintf("Avg: %v ms\n\n", result.Result.Stats["avg"]))
			}

			if ctx.Cmd == "dns" {
				timings, err := client.DecodeTimings(ctx.Cmd, result.Result.TimingsRaw)
				if err != nil {
					fmt.Println(err)
					return
				}
				output.WriteString(fmt.Sprintf("Total: %v ms\n", timings.Interface["total"]))
			}

			if ctx.Cmd == "http" {
				timings, err := client.DecodeTimings(ctx.Cmd, result.Result.TimingsRaw)
				if err != nil {
					fmt.Println(err)
					return
				}
				output.WriteString(fmt.Sprintf("Total: %v ms\n", timings.Interface["total"]))
				output.WriteString(fmt.Sprintf("Download: %v ms\n", timings.Interface["download"]))
				output.WriteString(fmt.Sprintf("First byte: %v ms\n", timings.Interface["firstByte"]))
				output.WriteString(fmt.Sprintf("DNS: %v ms\n", timings.Interface["dns"]))
				output.WriteString(fmt.Sprintf("TLS: %v ms\n", timings.Interface["tls"]))
				output.WriteString(fmt.Sprintf("TCP: %v ms\n", timings.Interface["tcp"]))
			}
		} else {
			if ctx.Cmd == "ping" {
				output.WriteString(bold.Render("Min: ") + fmt.Sprintf("%v ms\n", result.Result.Stats["min"]))
				output.WriteString(bold.Render("Max: ") + fmt.Sprintf("%v ms\n", result.Result.Stats["max"]))
				output.WriteString(bold.Render("Avg: ") + fmt.Sprintf("%v ms\n\n", result.Result.Stats["avg"]))
			}

			if ctx.Cmd == "dns" {
				timings, err := client.DecodeTimings(ctx.Cmd, result.Result.TimingsRaw)
				if err != nil {
					fmt.Println(err)
					return
				}
				output.WriteString(bold.Render("Total: ") + fmt.Sprintf("%v ms\n", timings.Interface["total"]))
			}

			if ctx.Cmd == "http" {
				timings, err := client.DecodeTimings(ctx.Cmd, result.Result.TimingsRaw)
				if err != nil {
					fmt.Println(err)
					return
				}
				output.WriteString(bold.Render("Total: ") + fmt.Sprintf("%v ms\n", timings.Interface["total"]))
				output.WriteString(bold.Render("Download: ") + fmt.Sprintf("%v ms\n", timings.Interface["download"]))
				output.WriteString(bold.Render("First byte: ") + fmt.Sprintf("%v ms\n", timings.Interface["firstByte"]))
				output.WriteString(bold.Render("DNS: ") + fmt.Sprintf("%v ms\n", timings.Interface["dns"]))
				output.WriteString(bold.Render("TLS: ") + fmt.Sprintf("%v ms\n", timings.Interface["tls"]))
				output.WriteString(bold.Render("TCP: ") + fmt.Sprintf("%v ms\n", timings.Interface["tcp"]))
			}
		}

	}

	fmt.Println(strings.TrimSpace(output.String()))
}

func OutputCI(id string, data *model.GetMeasurement, ctx model.Context) {
	// String builder for output
	var output strings.Builder

	// Output every result in case of multiple probes
	for _, result := range data.Results {
		// Output slightly different format if state is available
		output.WriteString(generateHeader(result, ctx) + "\n")

		// Output only latency values if flag is set
		output.WriteString(strings.TrimSpace(result.Result.RawOutput) + "\n\n")
	}

	fmt.Println(strings.TrimSpace(output.String()))
}

func OutputResults(id string, ctx model.Context) {
	fetcher := client.NewMeasurementsFetcher(client.ApiUrl)

	// Wait for first result to arrive from a probe before starting display (can be in-progress)
	data, err := fetcher.GetMeasurement(id)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Probe may not have started yet
	for len(data.Results) == 0 {
		time.Sleep(apiPollInterval)
		data, err = fetcher.GetMeasurement(id)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	if ctx.CI || ctx.JsonOutput || ctx.Latency {
		// Poll API until the measurement is complete
		for data.Status == "in-progress" {
			time.Sleep(apiPollInterval)
			data, err = fetcher.GetMeasurement(id)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		switch {
		case ctx.CI:
			OutputCI(id, data, ctx)
			return
		case ctx.JsonOutput:
			OutputJson(id, fetcher)
			return
		case ctx.Latency:
			OutputLatency(id, data, ctx)
			return
		default:
			panic(fmt.Sprintf("case not handled. %+v", ctx))
		}
	}

	LiveView(id, data, ctx)
}
