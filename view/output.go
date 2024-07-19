package view

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/mattn/go-runewidth"
)

var ShareURL = "https://www.jsdelivr.com/globalping?measurement="

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

	w, h := v.printer.GetSize()

	output := &strings.Builder{}

	// Poll API until the measurement is complete
	for data.Status == globalping.StatusInProgress {
		time.Sleep(v.ctx.APIMinInterval)
		data, err = v.globalping.GetMeasurement(id)
		if err != nil {
			return fmt.Errorf("failed to get data: %v", err)
		}

		output.Reset()

		// Output every result in case of multiple probes
		for i := range data.Results {
			result := &data.Results[i]
			// Output slightly different format if state is available
			output.WriteString(v.getProbeInfo(result) + "\n")

			if v.isBodyOnlyHttpGet(m) {
				output.WriteString(strings.TrimSpace(result.Result.RawBody) + "\n\n")
			} else {
				output.WriteString(strings.TrimSpace(result.Result.RawOutput) + "\n\n")
			}
		}

		v.printer.AreaUpdate(trimOutput(output, w, h))
	}
	v.printer.AreaClear()

	v.outputDefault(id, data, m)
	return nil
}

// Used to trim the output to fit the terminal in live view
func trimOutput(output *strings.Builder, terminalW, terminalH int) *string {
	maxW := terminalW - 4 // 4 extra chars to be safe from overflow
	maxH := terminalH - 4 // 4 extra lines to be safe from overflow
	if maxW <= 0 || maxH <= 0 {
		panic("terminal width / height too limited to display results")
	}

	text := strings.ReplaceAll(output.String(), "\t", "  ")
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

	txt := strings.Join(lines, "\n")
	return &txt
}

func (v *viewer) getProbeInfo(result *globalping.ProbeMeasurement) string {
	var output strings.Builder
	output.WriteString("> ")
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
	return v.printer.BoldWithColor(output.String(), ColorHighlight)
}

func (v *viewer) getShareMessage(id string) string {
	return v.printer.BoldWithColor(fmt.Sprintf("> View the results online: %s%s", ShareURL, id), ColorHighlight)
}

func (v *viewer) isBodyOnlyHttpGet(m *globalping.MeasurementCreate) bool {
	return v.ctx.Cmd == "http" && m.Options != nil && m.Options.Request != nil && m.Options.Request.Method == "GET" && !v.ctx.Full
}

func getLocationText(m *globalping.ProbeMeasurement) string {
	state := ""
	if m.Probe.State != "" {
		state = " (" + m.Probe.State + ")"
	}
	return m.Probe.City + state + ", " +
		m.Probe.Country + ", " +
		m.Probe.Continent + ", " +
		m.Probe.Network + " " +
		"(AS" + fmt.Sprint(m.Probe.ASN) + ")"
}
