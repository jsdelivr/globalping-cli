package view

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jsdelivr/globalping-cli/globalping"
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

func (v *viewer) Output(id string, m *globalping.MeasurementCreate) error {
	// Wait for first result to arrive from a probe before starting display (can be in-progress)
	data, err := v.globalping.GetMeasurement(id)
	if err != nil {
		return err
	}
	// Probe may not have started yet
	for len(data.Results) == 0 {
		time.Sleep(v.ctx.APIMinInterval)
		data, err = v.globalping.GetMeasurement(id)
		if err != nil {
			return err
		}
	}

	if v.ctx.CIMode || v.ctx.ToJSON || v.ctx.ToLatency {
		// Poll API until the measurement is complete
		for data.Status == globalping.StatusInProgress {
			time.Sleep(v.ctx.APIMinInterval)
			data, err = v.globalping.GetMeasurement(id)
			if err != nil {
				return err
			}
		}

		if v.ctx.ToLatency {
			return v.OutputLatency(id, data)
		}

		if v.ctx.ToJSON {
			return v.OutputJson(id)
		}

		if v.ctx.CIMode {
			v.outputDefault(id, data, m)
			return nil
		}
	}

	return v.liveView(id, data, m)
}

func (v *viewer) liveView(id string, data *globalping.Measurement, m *globalping.MeasurementCreate) error {
	var err error

	// Create new writer
	areaPrinter, err := pterm.DefaultArea.Start()
	if err != nil {
		return fmt.Errorf("failed to start writer: %v", err)
	}
	areaPrinter.RemoveWhenDone = true

	defer func() {
		// Stop area printer and clear area if not already done
		err := areaPrinter.Stop()
		if err != nil {
			v.printer.Printf("failed to stop writer: %v", err)
		}
	}()

	w, h, err := pterm.GetTerminalSize()
	if err != nil {
		return fmt.Errorf("failed to get terminal size: %v", err)
	}

	// String builder for output
	var output strings.Builder

	// Poll API until the measurement is complete
	for data.Status == globalping.StatusInProgress {
		time.Sleep(v.ctx.APIMinInterval)
		data, err = v.globalping.GetMeasurement(id)
		if err != nil {
			return fmt.Errorf("failed to get data: %v", err)
		}

		// Reset string builder
		output.Reset()

		// Output every result in case of multiple probes
		for i := range data.Results {
			result := &data.Results[i]
			// Output slightly different format if state is available
			output.WriteString(generateProbeInfo(result, !v.ctx.CIMode) + "\n")

			if v.isBodyOnlyHttpGet(m) {
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
		return fmt.Errorf("failed to stop writer: %v", err)
	}

	v.outputDefault(id, data, m)
	return nil
}

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

// Also checks if the probe has a state in it in the form %s, %s, (%s), %s, ASN:%d
func generateProbeInfo(result *globalping.ProbeMeasurement, useStyling bool) string {
	var output strings.Builder

	// Continent + Country + (State) + City + ASN + Network + (Region Tag)
	output.WriteString(getLocationText(result))

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

	headerWithFormat := formatWithLeadingArrow(output.String(), useStyling)
	return headerWithFormat
}

func formatWithLeadingArrow(text string, useStyling bool) string {
	if useStyling {
		return terminalLayoutArrow + terminalLayoutHighlight.Render(text)
	}
	return "> " + text
}

func (v *viewer) isBodyOnlyHttpGet(m *globalping.MeasurementCreate) bool {
	return v.ctx.Cmd == "http" && m.Options != nil && m.Options.Request != nil && m.Options.Request.Method == "GET" && !v.ctx.Full
}

func shareMessage(id string) string {
	return fmt.Sprintf("View the results online: https://www.jsdelivr.com/globalping?measurement=%s", id)
}

func getLocationText(m *globalping.ProbeMeasurement) string {
	state := ""
	if m.Probe.State != "" {
		state = "(" + m.Probe.State + "), "
	}
	return m.Probe.Continent +
		", " + m.Probe.Country +
		", " + state + m.Probe.City +
		", ASN:" + fmt.Sprint(m.Probe.ASN) +
		", " + m.Probe.Network
}
