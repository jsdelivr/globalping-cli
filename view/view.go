package view

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/mattn/go-runewidth"
	"github.com/pterm/pterm"
)

var (
	// UI styles
	terminalLayoutHighlight = lipgloss.NewStyle().
				Bold(true).Foreground(lipgloss.Color("#17D4A7"))

	terminalLayoutArrow = lipgloss.NewStyle().SetString(">").Bold(true).Foreground(lipgloss.Color("#17D4A7")).PaddingRight(1).String()

	terminalLayoutBold = lipgloss.NewStyle().Bold(true)
)

// Used to trim the output to fit the terminal in live view
func trimOutput(output string, terminalW, terminalH int) string {
	maxW := terminalW - 4 // 4 extra chars to be safe from overflow
	maxH := terminalH - 4 // 4 extra lines to be safe from overflow

	if maxW <= 0 || maxH <= 0 {
		panic("terminal width / height too limited to display results")
	}

	text := strings.ReplaceAll(output, "\t", "  ")

	// Split output into lines
	lines := strings.Split(text, "\n")

	if len(lines) > maxH {
		//  too many lines, trim first lines
		lines = lines[len(lines)-maxH:]
	}

	for i := 0; i < len(lines); i++ {
		rWidth := runewidth.StringWidth(lines[i])
		if rWidth > maxW {
			line := lines[i]
			trimmedLine := string(lines[i][:len(line)-rWidth+maxW])
			lines[i] = trimmedLine
		}
	}

	// Join lines back into a string
	txt := strings.Join(lines, "\n")

	return txt
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
		return terminalLayoutArrow + terminalLayoutHighlight.Render(output.String())
	}
}

var apiPollInterval = 500 * time.Millisecond

func LiveView(id string, data *model.GetMeasurement, ctx model.Context, m model.PostMeasurement) {
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

			if isBodyOnlyHttpGet(ctx, m) {
				output.WriteString(strings.TrimSpace(result.Result.RawBody) + "\n\n")
			} else {
				output.WriteString(strings.TrimSpace(result.Result.RawOutput) + "\n\n")
			}
		}

		areaPrinter.Update(trimOutput(output.String(), w, h))
	}

	// Stop area printer and clear area
	err = areaPrinter.Stop()
	if err != nil {
		fmt.Printf("failed to stop writer: %v\n", err)
	}

	if os.Getenv("LIVE_DEBUG") != "1" {
		PrintStandardResults(data, ctx, m)
	}
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
				output.WriteString(terminalLayoutBold.Render("Min: ") + fmt.Sprintf("%v ms\n", result.Result.Stats["min"]))
				output.WriteString(terminalLayoutBold.Render("Max: ") + fmt.Sprintf("%v ms\n", result.Result.Stats["max"]))
				output.WriteString(terminalLayoutBold.Render("Avg: ") + fmt.Sprintf("%v ms\n\n", result.Result.Stats["avg"]))
			}

			if ctx.Cmd == "dns" {
				timings, err := client.DecodeTimings(ctx.Cmd, result.Result.TimingsRaw)
				if err != nil {
					fmt.Println(err)
					return
				}
				output.WriteString(terminalLayoutBold.Render("Total: ") + fmt.Sprintf("%v ms\n", timings.Interface["total"]))
			}

			if ctx.Cmd == "http" {
				timings, err := client.DecodeTimings(ctx.Cmd, result.Result.TimingsRaw)
				if err != nil {
					fmt.Println(err)
					return
				}
				output.WriteString(terminalLayoutBold.Render("Total: ") + fmt.Sprintf("%v ms\n", timings.Interface["total"]))
				output.WriteString(terminalLayoutBold.Render("Download: ") + fmt.Sprintf("%v ms\n", timings.Interface["download"]))
				output.WriteString(terminalLayoutBold.Render("First byte: ") + fmt.Sprintf("%v ms\n", timings.Interface["firstByte"]))
				output.WriteString(terminalLayoutBold.Render("DNS: ") + fmt.Sprintf("%v ms\n", timings.Interface["dns"]))
				output.WriteString(terminalLayoutBold.Render("TLS: ") + fmt.Sprintf("%v ms\n", timings.Interface["tls"]))
				output.WriteString(terminalLayoutBold.Render("TCP: ") + fmt.Sprintf("%v ms\n", timings.Interface["tcp"]))
			}
		}

	}

	fmt.Println(strings.TrimSpace(output.String()))
}

// Prints non-json non-latency results to the screen
func PrintStandardResults(data *model.GetMeasurement, ctx model.Context, m model.PostMeasurement) {
	for i, result := range data.Results {
		if i > 0 {
			// new line as separator if more than 1 result
			fmt.Println()
		}

		// Output slightly different format if state is available
		fmt.Fprintln(os.Stderr, generateHeader(result, ctx))

		if isBodyOnlyHttpGet(ctx, m) {
			fmt.Println(strings.TrimSpace(result.Result.RawBody))
		} else {
			fmt.Println(strings.TrimSpace(result.Result.RawOutput))
		}
	}
}

func isBodyOnlyHttpGet(ctx model.Context, m model.PostMeasurement) bool {
	return ctx.Cmd == "http" && m.Options != nil && m.Options.Request != nil && m.Options.Request.Method == "GET" && !ctx.Full
}

func OutputResults(id string, ctx model.Context, m model.PostMeasurement) {
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

		if ctx.Latency {
			OutputLatency(id, data, ctx)
			return
		}

		if ctx.JsonOutput {
			OutputJson(id, fetcher)
			return
		}

		if ctx.CI {
			PrintStandardResults(data, ctx, m)
			return
		}

		panic(fmt.Sprintf("case not handled. %+v", ctx))
	}

	LiveView(id, data, ctx, m)
}
