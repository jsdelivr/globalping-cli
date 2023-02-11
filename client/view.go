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

// Generate latency strings
/* func latencyString(cmd string, data model.ResultStruct) string {
	var output strings.Builder
	// Only return stats field data
	if cmd == "ping" || cmd == "mtr" {
		output.WriteString(highlight.Render("Min: ") + fmt.Sprintf("%v ms", data.Stats.Min))
		output.WriteString(highlight.Render("Max: ") + fmt.Sprintf("%v ms", data.Stats.Max))
		output.WriteString(highlight.Render("Avg: ") + fmt.Sprintf("%v ms", data.Stats.Avg))
		output.WriteString(highlight.Render("Loss: ") + fmt.Sprintf("%v%%", data.Stats.Loss))
		return output.String()
	}

	// Only return timings field data
	return ""
} */

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
			if result.Probe.State != "" {
				output.WriteString(arrow + highlight.Render(fmt.Sprintf("%s, %s, (%s), %s, ASN:%d", result.Probe.Continent, result.Probe.Country, result.Probe.State, result.Probe.City, result.Probe.ASN)))
			} else {
				output.WriteString(arrow + highlight.Render(fmt.Sprintf("%s, %s, %s, ASN:%d", result.Probe.Continent, result.Probe.Country, result.Probe.City, result.Probe.ASN)))
			}

			// Output only latency values if flag is set
			if ctx.Latency {

			} else {
				output.WriteString("\n" + strings.TrimSpace(result.Result["rawOutput"].(string)) + "\n\n")
			}

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

func OutputResults(id string, ctx model.Context) {
	switch {
	case ctx.JsonOutput:
		OutputJson(id)
		return
	default:
		LiveView(id, ctx)
		return
	}
}
