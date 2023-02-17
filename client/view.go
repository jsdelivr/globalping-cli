package client

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
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

func dataSetup(id string) (model.GetMeasurement, error) {
	// Get results
	data, err := GetAPI(id)
	if err != nil {
		return data, err
	}

	// Probe may not have started yet
	for len(data.Results) == 0 {
		time.Sleep(100 * time.Millisecond)
		data, err = GetAPI(id)
		if err != nil {
			return data, err
		}
	}

	return data, nil
}

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

func generateHeader(result model.MeasurementResponse) string {
	var output strings.Builder
	if result.Probe.State != "" {
		output.WriteString(arrow + highlight.Render(fmt.Sprintf("%s, %s, (%s), %s, ASN:%d", result.Probe.Continent, result.Probe.Country, result.Probe.State, result.Probe.City, result.Probe.ASN)))
	} else {
		output.WriteString(arrow + highlight.Render(fmt.Sprintf("%s, %s, %s, ASN:%d", result.Probe.Continent, result.Probe.Country, result.Probe.City, result.Probe.ASN)))
	}

	return output.String()
}

func LiveView(id string, ctx model.Context) {
	// Get results
	data, err := dataSetup(id)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create new writer
	writer, _ := pterm.DefaultArea.Start()
	w, h, _ := pterm.GetTerminalSize()

	// String builder for output
	var output strings.Builder

	// Poll API every 100 milliseconds until the measurement is complete
	for data.Status == "in-progress" {
		time.Sleep(100 * time.Millisecond)
		data, err = GetAPI(id)

		// Reset string builder
		output.Reset()

		// Output every result in case of multiple probes
		for _, result := range data.Results {
			// Output slightly different format if state is available
			output.WriteString(generateHeader(result) + "\n")

			// Output only latency values if flag is set
			output.WriteString(strings.TrimSpace(result.Result.RawOutput) + "\n\n")

		}

		if err != nil {
			writer.Stop()
			fmt.Println(err)
			return
		}

		writer.Update(sliceOutput(output.String(), w, h))
	}

	// Stop live updater and output to stdout
	writer.RemoveWhenDone = true
	writer.Stop()
	fmt.Println(output.String())
}

// If json flag is used, only output json
func OutputJson(id string) {
	// Get results
	data, err := dataSetup(id)
	if err != nil {
		fmt.Println(err)
		return
	}

	for data.Status == "in-progress" {
		time.Sleep(100 * time.Millisecond)
		data, err = GetAPI(id)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	output, err := GetApiJson(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(output)
}

// If latency flag is used, only output latency values
func OutputLatency(id string, ctx model.Context) {
	// Get results
	data, err := dataSetup(id)
	if err != nil {
		fmt.Println(err)
		return
	}

	// String builder for output
	var output strings.Builder

	// Poll API every 100 milliseconds until the measurement is complete
	for data.Status == "in-progress" {
		time.Sleep(100 * time.Millisecond)
		data, err = GetAPI(id)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// Output every result in case of multiple probes
	for _, result := range data.Results {
		// Output slightly different format if state is available
		output.WriteString(generateHeader(result) + "\n")

		// Output only latency values if flag is set
		if ctx.Cmd == "ping" || ctx.Cmd == "mtr" {
			output.WriteString(bold.Render("Min: ") + fmt.Sprintf("%v ms\n", result.Result.Stats["min"]))
			output.WriteString(bold.Render("Max: ") + fmt.Sprintf("%v ms\n", result.Result.Stats["max"]))
			output.WriteString(bold.Render("Avg: ") + fmt.Sprintf("%v ms\n", result.Result.Stats["avg"]))

			output.WriteString(bold.Render("Transmitted: ") + fmt.Sprintf("%v\n", result.Result.Stats["total"]))
			output.WriteString(bold.Render("Received: ") + fmt.Sprintf("%v\n", result.Result.Stats["rcv"]))
			output.WriteString(bold.Render("Dropped: ") + fmt.Sprintf("%v\n", result.Result.Stats["drop"]))
			output.WriteString(bold.Render("Loss: ") + fmt.Sprintf("%v%%\n\n", result.Result.Stats["loss"]))
		}
	}

	fmt.Println(strings.TrimSpace(output.String()))
}

func OutputResults(id string, ctx model.Context) {
	switch {
	case ctx.JsonOutput:
		OutputJson(id)
		return
	case ctx.Latency:
		OutputLatency(id, ctx)
		return
	default:
		LiveView(id, ctx)
		return
	}
}
