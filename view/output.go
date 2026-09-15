package view

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/jsdelivr/globalping-cli/utils"
	"github.com/jsdelivr/globalping-go"
	"github.com/mattn/go-runewidth"
)

var ErrAllProbesFailed = errors.New("all probes failed")

func (v *viewer) outputFailSummary(m *globalping.Measurement) error {
	for i := range m.Results {
		v.printer.ErrPrintln(v.getProbeInfoWithStatus(&m.Results[i], true))
		v.printer.Println(m.Results[i].Result.RawOutput)
	}

	return ErrAllProbesFailed
}

func isSomeTestFinished(m *globalping.Measurement) bool {
	for i := range m.Results {
		if m.Results[i].Result.Status == globalping.TestStatusFinished {
			return true
		}
	}

	return false
}

func (v *viewer) OutputLive(measurement *globalping.Measurement, opts *globalping.MeasurementCreate, w, h int) {
	output := &strings.Builder{}
	headerLines := make(map[int]bool)
	lineCount := 0

	// Output every result in case of multiple probes
	for i := range measurement.Results {
		result := &measurement.Results[i]
		// Keep headers plain until both width and height have been trimmed.
		header := normalizeTableLocation(getProbeInfoText(result))
		status := ""

		if result.Result.Status == globalping.TestStatusFailed {
			status = " — " + resultStatusLabel(&result.Result)
		}

		maxW := w - 4
		status = runewidth.Truncate(status, max(maxW, 0), "")
		header = truncateTableCell(strings.ReplaceAll(header, "\t", "  "), max(maxW-runewidth.StringWidth(status), 0))
		headerLines[lineCount] = true
		output.WriteString(header + status + "\n")
		bodyStart := output.Len()

		switch {
		case result.Result.Status == globalping.TestStatusFailed:
			output.WriteString(result.Result.RawOutput + "\n\n")
		case v.isBodyOnlyHttpGet(opts):
			if result.Result.RawBody != nil {
				output.WriteString(strings.TrimSpace(*result.Result.RawBody))
			}

			output.WriteString("\n\n")
		default:
			output.WriteString(strings.TrimSpace(result.Result.RawOutput) + "\n\n")
		}

		lineCount += 1 + strings.Count(output.String()[bodyStart:], "\n")
	}

	trimmed := trimOutput(output, w, h)
	lines := strings.Split(*trimmed, "\n")
	firstLine := lineCount + 1 - len(lines)

	for i := range lines {
		if headerLines[firstLine+i] {
			lines[i] = v.printer.BoldForeground(lines[i], BGYellow)
		}
	}

	text := strings.Join(lines, "\n")
	v.printer.AreaUpdate(&text)
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

	for i := range lines {
		lines[i] = runewidth.Truncate(lines[i], maxW, "")
	}

	txt := strings.Join(lines, "\n")

	return &txt
}

func (v *viewer) getProbeInfo(result *globalping.ProbeMeasurement) string {
	return v.getProbeInfoWithStatus(result, false)
}

func (v *viewer) getProbeInfoWithStatus(result *globalping.ProbeMeasurement, includeOffline bool) string {
	text := getProbeInfoText(result)

	if result.Result.Status == globalping.TestStatusFailed || includeOffline && result.Result.Status == globalping.TestStatusOffline {
		text += " — " + resultStatusLabel(&result.Result)
	}

	return v.printer.BoldForeground(text, BGYellow)
}

func getProbeInfoText(result *globalping.ProbeMeasurement) string {
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

	return output.String()
}

func resultStatusLabel(result *globalping.ProbeResult) string {
	if result.Status == globalping.TestStatusOffline {
		return "Probe offline"
	}

	switch result.FailureSource {
	case globalping.FailureSourceTarget:
		return "Target error"
	case globalping.FailureSourceResolver:
		return "Resolver error"
	case globalping.FailureSourceInternal:
		return "Internal error"
	default:
		return "Error"
	}
}

func (v *viewer) getShareMessage(id string) string {
	shareURL := utils.ShareURL + id

	if v.ctx.Table && !v.ctx.Comparison {
		shareURL += "&display=table"
	}

	return v.printer.BoldForeground("> View the results online: "+shareURL, BGYellow)
}

func (v *viewer) isBodyOnlyHttpGet(m *globalping.MeasurementCreate) bool {
	return v.ctx.Cmd == "http" && m.Options != nil && m.Options.Request != nil && m.Options.Request.Method == http.MethodGet && !v.ctx.Full
}

func getLocationText(m *globalping.ProbeMeasurement) string {
	state := ""

	if m.Probe.State != nil && *m.Probe.State != "" {
		state = " (" + *m.Probe.State + ")"
	}

	return m.Probe.City + state + ", " +
		m.Probe.Country + ", " +
		m.Probe.Continent + ", " +
		m.Probe.Network + " " +
		"(AS" + strconv.Itoa(m.Probe.ASN) + ")"
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
