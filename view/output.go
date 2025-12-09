package view

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jsdelivr/globalping-cli/utils"
	"github.com/jsdelivr/globalping-go"
	"github.com/mattn/go-runewidth"
)

func (v *viewer) OutputLive(measurement *globalping.Measurement, opts *globalping.MeasurementCreate, w, h int) {
	output := &strings.Builder{}

	// Output every result in case of multiple probes
	for i := range measurement.Results {
		result := &measurement.Results[i]
		// Output slightly different format if state is available
		output.WriteString(v.getProbeInfo(result) + "\n")

		if v.isBodyOnlyHttpGet(opts) {
			output.WriteString(strings.TrimSpace(result.Result.RawBody) + "\n\n")
		} else {
			output.WriteString(strings.TrimSpace(result.Result.RawOutput) + "\n\n")
		}
	}

	v.printer.AreaUpdate(trimOutput(output, w, h))
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
	if len(result.Probe.Tags) > 0 {
		regionCode := ""
		userInfoTags := []string{}
		for _, tag := range result.Probe.Tags {
			if strings.HasPrefix(tag, "u-") && !strings.Contains(tag, ":") {
				userInfoTags = append(userInfoTags, tag)
			} else if regionCode == "" {
				// If tag ends in a number, it's likely a region code and should be displayed
				if _, err := strconv.Atoi(tag[len(tag)-1:]); err == nil {
					regionCode = tag
				}
			}
		}
		userInfo := largestCommonPrefix(userInfoTags)
		if userInfo != "" {
			output.WriteString(", " + userInfo)
		}
		if regionCode != "" {
			output.WriteString(" (" + regionCode + ")")
		}
	}
	return v.printer.BoldForeground(output.String(), BGYellow)
}

func (v *viewer) getShareMessage(id string) string {
	return v.printer.BoldForeground(fmt.Sprintf("> View the results online: %s%s", utils.ShareURL, id), BGYellow)
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

func largestCommonPrefix(items []string) string {
	if len(items) == 0 {
		return ""
	}
	if len(items) == 1 {
		return items[0]
	}
	prefix := items[0]
	for i := 1; i < len(items); i++ {
		for j := 0; j < len(prefix); j++ {
			if j >= len(items[i]) || prefix[j] != items[i][j] {
				prefix = prefix[:j]
				break
			}
		}
	}
	return prefix
}
