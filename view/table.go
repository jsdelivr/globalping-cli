package view

import (
	"bytes"
	"encoding/json"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/jsdelivr/globalping-go"
	"github.com/mattn/go-runewidth"
)

const colSeparator = " | "

var httpBodySeparator = regexp.MustCompile(`\r?\n\r?\n`)

type httpSizeColumn int

const (
	httpSizeNone httpSizeColumn = iota
	httpSizeContentLength
	httpSizeBytes
)

type tableRenderOptions struct {
	minimumWidths        []int
	shrinkableColumns    []int
	compactStatusColumn  int
	minimumLocationWidth int
	minimumTrailingWidth int
	multilineLocation    bool
	measureLocationBytes bool
}

func (v *viewer) OutputTable(measurement *globalping.Measurement) error {
	v.ctx.TableOutputRows = 0
	if measurement.Status != globalping.MeasurementStatusInProgress && !isSomeTestFinished(measurement) {
		return v.outputFailSummary(measurement)
	}
	v.ctx.TableOutputRows = len(measurement.Results)
	return v.outputTableView(measurement)
}

func (v *viewer) outputTableView(m *globalping.Measurement) error {
	if m.Type == "ping" {
		return v.outputPingTableView(m)
	}

	width, _ := v.printer.GetSize()
	output := v.generateMeasurementTable(m, width-2)
	v.printer.AreaUpdate(&output)
	return nil
}

func (v *viewer) generateMeasurementTable(m *globalping.Measurement, areaWidth int) string {
	httpSize := httpTableSizeColumn(m)
	rows := [][]string{tableHeader(m.Type, v.ctx.Trace, httpSize)}
	for i := range m.Results {
		rows = append(rows, tableRow(m.Type, v.ctx.Trace, len(rows[0]), &m.Results[i], httpSize))
	}

	return v.renderMeasurementTable(rows, areaWidth, m.Type)
}

func tableHeader(measurementType globalping.MeasurementType, trace bool, httpSize httpSizeColumn) []string {
	switch measurementType {
	case "traceroute", "mtr":
		return []string{"Location", "Hops", "Last", "Min", "Avg", "Max"}
	case "dns":
		if trace {
			return []string{"Location", "Answers", "Time", "Resolver"}
		}
		return []string{"Location", "Status", "Answers", "Time", "Resolver"}
	case "http":
		header := []string{"Location", "Status"}
		if httpSize == httpSizeContentLength {
			header = append(header, "Content-Length")
		} else if httpSize == httpSizeBytes {
			header = append(header, "Bytes")
		}
		return append(header, "Total", "Resolved IP")
	default:
		return []string{"Location"}
	}
}

func tableRow(measurementType globalping.MeasurementType, trace bool, columns int, measurement *globalping.ProbeMeasurement, httpSize httpSizeColumn) []string {
	row := make([]string, columns)
	for i := range row {
		row[i] = "-"
	}
	row[0] = getLocationText(measurement)

	if measurement.Result.Status == globalping.TestStatusFailed {
		return []string{row[0], failureTableMessage(&measurement.Result)}
	}
	if measurement.Result.Status != globalping.TestStatusFinished {
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
		copy(row[1:], httpTableValues(&measurement.Result, httpSize))
	}

	return row
}

func failureTableMessage(result *globalping.ProbeResult) string {
	label := "Error"
	switch result.FailureSource {
	case globalping.FailureSourceTarget:
		label = "Target error"
	case globalping.FailureSourceResolver:
		label = "Resolver error"
	case globalping.FailureSourceInternal:
		label = "Internal error"
	}

	for _, line := range strings.Split(result.RawOutput, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return label + ": " + line
		}
	}
	return label
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
	if lastRTT := lastTimingRTT(timings); lastRTT != nil {
		values[1] = formatDuration(*lastRTT)
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

func lastTimingRTT(timings []tracerouteTableTiming) *float64 {
	for i := len(timings) - 1; i >= 0; i-- {
		if timings[i].RTT != nil {
			return timings[i].RTT
		}
	}
	return nil
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
	if lastRTT := lastTimingRTT(lastHop.Timings); lastRTT != nil {
		values[1] = formatDuration(*lastRTT)
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
		values[2] = formatTotalDuration(total)
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
		values[1] = formatTotalDuration(*lastHop.Timings.Total)
	}
	if lastHop.Resolver != nil && *lastHop.Resolver != "" {
		values[2] = *lastHop.Resolver
	}
	return values
}

func httpTableSizeColumn(measurement *globalping.Measurement) httpSizeColumn {
	if measurement.Type != "http" {
		return httpSizeNone
	}
	if hasHTTPContentLength(measurement.Results) {
		return httpSizeContentLength
	}
	if hasHTTPBody(measurement.Results) {
		return httpSizeBytes
	}
	return httpSizeNone
}

func hasHTTPContentLength(results []globalping.ProbeMeasurement) bool {
	for i := range results {
		if results[i].Result.Status == globalping.TestStatusFinished {
			if _, ok := contentLength(results[i].Result.HeadersRaw); ok {
				return true
			}
		}
	}
	return false
}

func hasHTTPBody(results []globalping.ProbeMeasurement) bool {
	for i := range results {
		if results[i].Result.Status == globalping.TestStatusFinished {
			if _, ok := httpBodyLength(results[i].Result.RawOutput); ok {
				return true
			}
		}
	}
	return false
}

func httpTableValues(result *globalping.ProbeResult, httpSize httpSizeColumn) []string {
	values := []string{"-", "-", "-", "-"}
	if result.StatusCode > 0 {
		values[0] = strconv.Itoa(result.StatusCode)
		if result.StatusCodeName != "" {
			values[0] += " " + result.StatusCodeName
		}
	} else if result.StatusCodeName != "" {
		values[0] = result.StatusCodeName
	}
	if httpSize == httpSizeContentLength {
		if length, ok := contentLength(result.HeadersRaw); ok {
			values[1] = strconv.FormatUint(length, 10) + " B"
		}
	} else if httpSize == httpSizeBytes {
		if length, ok := httpBodyLength(result.RawOutput); ok {
			suffix := " B"
			if result.Truncated {
				suffix = "+ B"
			}
			values[1] = strconv.Itoa(length) + suffix
		}
	}
	if total, ok := decodeTotalTiming(result.TimingsRaw); ok {
		values[2] = formatTotalDuration(total)
	}
	if result.ResolvedAddress != "" {
		values[3] = result.ResolvedAddress
	}
	if httpSize == httpSizeNone {
		return append(values[:1], values[2:]...)
	}
	return values
}

func httpBodyLength(rawOutput string) (int, bool) {
	separator := httpBodySeparator.FindStringIndex(rawOutput)
	if separator == nil {
		return 0, false
	}
	return len(rawOutput[separator[1]:]), true
}

func decodeTotalTiming(raw json.RawMessage) (float64, bool) {
	var timings totalTiming
	if len(raw) == 0 || json.Unmarshal(raw, &timings) != nil || timings.Total == nil {
		return 0, false
	}
	return *timings.Total, true
}

func formatTotalDuration(ms float64) string {
	if math.Trunc(ms) == ms {
		return strconv.FormatFloat(ms, 'f', 0, 64) + " ms"
	}
	return formatDuration(ms)
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

func (v *viewer) renderMeasurementTable(rows [][]string, areaWidth int, measurementType globalping.MeasurementType) string {
	options := tableRenderOptions{}
	if measurementType == "http" {
		options.shrinkableColumns = []int{len(rows[0]) - 1}
		options.compactStatusColumn = 1
	}
	return v.renderTable(rows, areaWidth, options)
}

func (v *viewer) renderPingTable(rows [][]string, areaWidth int) string {
	return v.renderTable(rows, areaWidth, tableRenderOptions{
		minimumWidths:        []int{0, 4, 7, 8, 8, 8, 8},
		minimumLocationWidth: 6,
		minimumTrailingWidth: 60,
		multilineLocation:    true,
		measureLocationBytes: true,
	})
}

func (v *viewer) renderTable(rows [][]string, areaWidth int, options tableRenderOptions) string {
	if len(rows) == 0 || len(rows[0]) == 0 {
		return ""
	}

	columnWidths := make([]int, len(rows[0]))
	for rowIndex, row := range rows {
		for column, value := range row {
			if column < len(columnWidths) {
				if rowIndex > 0 && isSpanningRow(row, len(columnWidths)) && column == 1 {
					continue
				}
				width := runewidth.StringWidth(value)
				if column == 0 && options.measureLocationBytes {
					width = len(value)
				}
				columnWidths[column] = max(columnWidths[column], width)
			}
		}
	}
	for column, width := range options.minimumWidths {
		if column < len(columnWidths) {
			columnWidths[column] = max(columnWidths[column], width)
		}
	}
	if options.compactStatusColumn > 0 && options.compactStatusColumn < len(rows[0]) {
		tableWidth := runewidth.StringWidth(colSeparator) * (len(columnWidths) - 1)
		for _, width := range columnWidths {
			tableWidth += width
		}
		if tableWidth > areaWidth {
			statusColumn := options.compactStatusColumn
			adjustedRows := make([][]string, len(rows))
			copy(adjustedRows, rows)
			statusWidth := runewidth.StringWidth(rows[0][statusColumn])
			for rowIndex := 1; rowIndex < len(rows); rowIndex++ {
				if len(rows[rowIndex]) <= statusColumn {
					continue
				}
				if isSpanningRow(rows[rowIndex], len(columnWidths)) {
					continue
				}
				status := rows[rowIndex][statusColumn]
				if code, ok := httpStatusCode(status); ok {
					adjustedRows[rowIndex] = append([]string(nil), rows[rowIndex]...)
					adjustedRows[rowIndex][statusColumn] = code
					status = code
				}
				statusWidth = max(statusWidth, runewidth.StringWidth(status))
			}
			rows = adjustedRows
			columnWidths[statusColumn] = statusWidth
		}
	}
	if options.minimumTrailingWidth > 0 {
		trailingWidth := options.minimumTrailingWidth
		for _, row := range rows[1:] {
			if isSpanningRow(row, len(columnWidths)) {
				continue
			}
			rowWidth := 0
			for column := 1; column < len(row); column++ {
				rowWidth += len(row[column]) + len(colSeparator)
			}
			trailingWidth = max(trailingWidth, rowWidth)
		}
		locationWidth := max(areaWidth-trailingWidth, options.minimumLocationWidth)
		columnWidths[0] = min(columnWidths[0], locationWidth)
	} else {
		fitTableColumnWidths(columnWidths, rows[0], areaWidth, options.shrinkableColumns)
	}

	var output bytes.Buffer
	for rowIndex, row := range rows {
		if len(row) == 0 {
			output.WriteByte('\n')
			continue
		}
		color := ColorNone
		if rowIndex == 0 {
			color = FGBrightCyan
		}
		location := strings.ReplaceAll(row[0], "\t", "  ")
		locationLines := []string{location}
		if options.multilineLocation {
			locationLines = strings.Split(location, "\n")
		}
		for lineIndex, line := range locationLines {
			line = truncateTableCell(line, columnWidths[0])
			line = padTableCell(line, columnWidths[0], false)
			if color != ColorNone {
				line = v.printer.Color(line, color)
			}
			output.WriteString(line)
			if isSpanningRow(row, len(columnWidths)) {
				output.WriteString(colSeparator)
				value := ""
				if lineIndex == 0 {
					value = strings.ReplaceAll(row[1], "\t", "  ")
				}
				availableWidth := max(areaWidth-columnWidths[0]-runewidth.StringWidth(colSeparator), 0)
				spanWidth := runewidth.StringWidth(colSeparator) * (len(columnWidths) - 2)
				for _, width := range columnWidths[1:] {
					spanWidth += width
				}
				spanWidth = min(max(spanWidth, runewidth.StringWidth(value)), availableWidth)
				value = truncateTableCell(value, spanWidth)
				output.WriteString(padTableCell(value, spanWidth, false))
				output.WriteByte('\n')
				continue
			}

			for column := 1; column < len(row); column++ {
				output.WriteString(colSeparator)
				value := ""
				if lineIndex == 0 {
					value = strings.ReplaceAll(row[column], "\t", "  ")
				}
				value = truncateTableCell(value, columnWidths[column])
				value = padTableCell(value, columnWidths[column], true)
				if color != ColorNone {
					value = v.printer.Color(value, color)
				}
				output.WriteString(value)
			}
			output.WriteByte('\n')
		}
	}
	return output.String()
}

func isSpanningRow(row []string, columns int) bool {
	return columns > 2 && len(row) == 2
}

func fitTableColumnWidths(widths []int, header []string, areaWidth int, shrinkableColumns []int) {
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

	for _, column := range shrinkableColumns {
		if overflow <= 0 {
			break
		}
		if column <= 0 || column >= len(widths) || column >= len(header) {
			continue
		}
		minWidth := runewidth.StringWidth(header[column])
		reduction := min(max(widths[column]-minWidth, 0), overflow)
		widths[column] -= reduction
		overflow -= reduction
	}
}

func httpStatusCode(status string) (string, bool) {
	code, _, found := strings.Cut(status, " ")
	if !found {
		return "", false
	}
	if _, err := strconv.Atoi(code); err != nil {
		return "", false
	}
	return code, true
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
