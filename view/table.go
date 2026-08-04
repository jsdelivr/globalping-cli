package view

import (
	"bytes"
	"encoding/json"
	"math"
	"strconv"
	"strings"

	"github.com/jsdelivr/globalping-go"
	"github.com/mattn/go-runewidth"
)

func (v *viewer) generateMeasurementTable(m *globalping.Measurement, areaWidth int) string {
	rows := [][]string{tableHeader(m.Type, v.ctx.Trace)}
	for i := range m.Results {
		rows = append(rows, tableRow(m.Type, v.ctx.Trace, len(rows[0]), &m.Results[i]))
	}

	return v.renderMeasurementTable(rows, areaWidth)
}

func tableHeader(measurementType globalping.MeasurementType, trace bool) []string {
	switch measurementType {
	case "traceroute", "mtr":
		return []string{"Location", "Hops", "Last", "Min", "Avg", "Max"}
	case "dns":
		if trace {
			return []string{"Location", "Answers", "Time", "Resolver"}
		}
		return []string{"Location", "Status", "Answers", "Time", "Resolver"}
	case "http":
		return []string{"Location", "Status", "Content length", "Total", "Resolved IP"}
	default:
		return []string{"Location"}
	}
}

func tableRow(measurementType globalping.MeasurementType, trace bool, columns int, measurement *globalping.ProbeMeasurement) []string {
	row := make([]string, columns)
	for i := range row {
		row[i] = "-"
	}
	row[0] = getLocationText(measurement)

	if measurement.Result.Status != globalping.StatusFinished {
		return row
	}

	switch measurementType {
	case "traceroute":
		copy(row[1:], tracerouteTableValues(measurement.Result.HopsRaw))
	case "mtr":
		copy(row[1:], mtrTableValues(measurement.Result.HopsRaw))
	case "dns":
		if trace {
			copy(row[1:], dnsTraceTableValues(measurement.Result.HopsRaw))
		} else {
			copy(row[1:], dnsTableValues(&measurement.Result))
		}
	case "http":
		copy(row[1:], httpTableValues(&measurement.Result))
	}

	return row
}

type tracerouteTableTiming struct {
	RTT *float64 `json:"rtt"`
}

type tracerouteTableHop struct {
	Timings []tracerouteTableTiming `json:"timings"`
}

func tracerouteTableValues(raw json.RawMessage) []string {
	values := []string{"-", "-", "-", "-", "-"}
	var hops []tracerouteTableHop
	if len(raw) == 0 || json.Unmarshal(raw, &hops) != nil || hops == nil {
		return values
	}

	values[0] = strconv.Itoa(len(hops))
	if len(hops) == 0 {
		return values
	}
	timings := hops[len(hops)-1].Timings
	if len(timings) == 0 {
		return values
	}
	for i := len(timings) - 1; i >= 0; i-- {
		if timings[i].RTT != nil {
			values[1] = formatDuration(*timings[i].RTT)
			break
		}
	}

	minRTT := math.MaxFloat64
	maxRTT := -math.MaxFloat64
	totalRTT := 0.0
	count := 0
	for _, timing := range timings {
		if timing.RTT == nil {
			continue
		}
		minRTT = math.Min(minRTT, *timing.RTT)
		maxRTT = math.Max(maxRTT, *timing.RTT)
		totalRTT += *timing.RTT
		count++
	}
	if count > 0 {
		values[2] = formatDuration(minRTT)
		values[3] = formatDuration(totalRTT / float64(count))
		values[4] = formatDuration(maxRTT)
	}
	return values
}

type mtrTableStats struct {
	Min *float64 `json:"min"`
	Avg *float64 `json:"avg"`
	Max *float64 `json:"max"`
}

type mtrTableHop struct {
	Stats   *mtrTableStats          `json:"stats"`
	Timings []tracerouteTableTiming `json:"timings"`
}

func mtrTableValues(raw json.RawMessage) []string {
	values := []string{"-", "-", "-", "-", "-"}
	var hops []mtrTableHop
	if len(raw) == 0 || json.Unmarshal(raw, &hops) != nil || hops == nil {
		return values
	}

	values[0] = strconv.Itoa(len(hops))
	if len(hops) == 0 {
		return values
	}
	lastHop := hops[len(hops)-1]
	for i := len(lastHop.Timings) - 1; i >= 0; i-- {
		if lastHop.Timings[i].RTT != nil {
			values[1] = formatDuration(*lastHop.Timings[i].RTT)
			break
		}
	}
	if lastHop.Stats != nil {
		if lastHop.Stats.Min != nil {
			values[2] = formatDuration(*lastHop.Stats.Min)
		}
		if lastHop.Stats.Avg != nil {
			values[3] = formatDuration(*lastHop.Stats.Avg)
		}
		if lastHop.Stats.Max != nil {
			values[4] = formatDuration(*lastHop.Stats.Max)
		}
	}
	return values
}

type totalTiming struct {
	Total *float64 `json:"total"`
}

func dnsTableValues(result *globalping.ProbeResult) []string {
	values := []string{"-", "-", "-", "-"}
	if result.StatusCodeName != "" {
		values[0] = result.StatusCodeName
	} else {
		values[0] = strconv.Itoa(result.StatusCode)
	}

	var answers []json.RawMessage
	if len(result.AnswersRaw) > 0 && json.Unmarshal(result.AnswersRaw, &answers) == nil && answers != nil {
		values[1] = strconv.Itoa(len(answers))
	}
	if total, ok := decodeTotalTiming(result.TimingsRaw); ok {
		values[2] = formatDuration(total)
	}
	if result.Resolver != "" {
		values[3] = result.Resolver
	}
	return values
}

type dnsTraceTableHop struct {
	Answers  *[]json.RawMessage `json:"answers"`
	Timings  *totalTiming       `json:"timings"`
	Resolver *string            `json:"resolver"`
}

func dnsTraceTableValues(raw json.RawMessage) []string {
	values := []string{"-", "-", "-"}
	var hops []dnsTraceTableHop
	if len(raw) == 0 || json.Unmarshal(raw, &hops) != nil || len(hops) == 0 {
		return values
	}

	lastHop := hops[len(hops)-1]
	if lastHop.Answers != nil {
		values[0] = strconv.Itoa(len(*lastHop.Answers))
	}
	if lastHop.Timings != nil && lastHop.Timings.Total != nil {
		values[1] = formatDuration(*lastHop.Timings.Total)
	}
	if lastHop.Resolver != nil && *lastHop.Resolver != "" {
		values[2] = *lastHop.Resolver
	}
	return values
}

func httpTableValues(result *globalping.ProbeResult) []string {
	values := []string{"-", "-", "-", "-"}
	if result.StatusCode > 0 {
		values[0] = strconv.Itoa(result.StatusCode)
		if result.StatusCodeName != "" {
			values[0] += " " + result.StatusCodeName
		}
	} else if result.StatusCodeName != "" {
		values[0] = result.StatusCodeName
	}
	if length, ok := contentLength(result.HeadersRaw); ok {
		values[1] = strconv.FormatUint(length, 10) + " B"
	}
	if total, ok := decodeTotalTiming(result.TimingsRaw); ok {
		values[2] = formatDuration(total)
	}
	if result.ResolvedAddress != "" {
		values[3] = result.ResolvedAddress
	}
	return values
}

func decodeTotalTiming(raw json.RawMessage) (float64, bool) {
	var timings totalTiming
	if len(raw) == 0 || json.Unmarshal(raw, &timings) != nil || timings.Total == nil {
		return 0, false
	}
	return *timings.Total, true
}

func contentLength(raw json.RawMessage) (uint64, bool) {
	var headers map[string]json.RawMessage
	if len(raw) == 0 || json.Unmarshal(raw, &headers) != nil {
		return 0, false
	}
	for name, value := range headers {
		if !strings.EqualFold(name, "content-length") {
			continue
		}
		var stringValue string
		if len(value) > 0 && value[0] == '"' {
			if json.Unmarshal(value, &stringValue) != nil {
				return 0, false
			}
			stringValue = strings.TrimSpace(stringValue)
		} else {
			stringValue = strings.TrimSpace(string(value))
		}
		if stringValue == "" || strings.IndexFunc(stringValue, func(r rune) bool { return r < '0' || r > '9' }) != -1 {
			return 0, false
		}
		length, err := strconv.ParseUint(stringValue, 10, 64)
		return length, err == nil
	}
	return 0, false
}

func (v *viewer) renderMeasurementTable(rows [][]string, areaWidth int) string {
	if len(rows) == 0 || len(rows[0]) == 0 {
		return ""
	}

	columnWidths := make([]int, len(rows[0]))
	for _, row := range rows {
		for column, value := range row {
			if column < len(columnWidths) {
				columnWidths[column] = max(columnWidths[column], runewidth.StringWidth(value))
			}
		}
	}
	fitTableColumnWidths(columnWidths, rows[0], areaWidth)

	var output bytes.Buffer
	for rowIndex, row := range rows {
		color := ColorNone
		if rowIndex == 0 {
			color = FGBrightCyan
		}
		for column, value := range row {
			if column > 0 {
				output.WriteString(colSeparator)
			}
			value = strings.ReplaceAll(value, "\t", "  ")
			value = truncateTableCell(value, columnWidths[column])
			value = padTableCell(value, columnWidths[column], column != 0)
			if color != ColorNone {
				value = v.printer.Color(value, color)
			}
			output.WriteString(value)
		}
		output.WriteByte('\n')
	}
	return output.String()
}

func fitTableColumnWidths(widths []int, header []string, areaWidth int) {
	separatorWidth := runewidth.StringWidth(colSeparator) * (len(widths) - 1)
	contentWidth := areaWidth - separatorWidth
	usedWidth := 0
	for _, width := range widths {
		usedWidth += width
	}
	overflow := usedWidth - contentWidth
	if overflow <= 0 {
		return
	}

	minLocationWidth := runewidth.StringWidth(header[0])
	locationReduction := min(max(widths[0]-minLocationWidth, 0), overflow)
	widths[0] -= locationReduction
	overflow -= locationReduction

	if overflow > 0 && isHTTPTableHeader(header) {
		for i := 1; i < len(widths); i++ {
			if header[i] != "Status" {
				continue
			}
			minWidth := runewidth.StringWidth(header[i])
			reduction := min(max(widths[i]-minWidth, 0), overflow)
			widths[i] -= reduction
			overflow -= reduction
			break
		}
	}

	for i := 1; i < len(widths) && overflow > 0; i++ {
		if header[i] != "Resolved IP" {
			continue
		}
		minWidth := runewidth.StringWidth(header[i])
		reduction := min(max(widths[i]-minWidth, 0), overflow)
		widths[i] -= reduction
		overflow -= reduction
	}
}

func isHTTPTableHeader(header []string) bool {
	hasContentLength := false
	hasResolvedIP := false
	for _, value := range header {
		hasContentLength = hasContentLength || value == "Content length"
		hasResolvedIP = hasResolvedIP || value == "Resolved IP"
	}
	return hasContentLength && hasResolvedIP
}

func truncateTableCell(value string, width int) string {
	if runewidth.StringWidth(value) <= width {
		return value
	}
	tail := ""
	if width >= 3 {
		tail = "..."
	}
	return runewidth.Truncate(value, width, tail)
}

func padTableCell(value string, width int, left bool) string {
	padding := strings.Repeat(" ", max(width-runewidth.StringWidth(value), 0))
	if left {
		return padding + value
	}
	return value + padding
}
