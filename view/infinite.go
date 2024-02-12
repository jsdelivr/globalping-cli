package view

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/mattn/go-runewidth"
	"github.com/pterm/pterm"
)

// Table defaults
var (
	colSeparator = " | "
)

func (v *viewer) OutputInfinite(id string) error {
	if v.ctx.History == nil {
		v.ctx.History = NewRbuffer(v.ctx.MaxHistory)
	}
	v.ctx.History.Push(id)

	res, err := v.gp.GetMeasurement(id)
	if err != nil {
		return err
	}
	// Probe may not have started yet
	for len(res.Results) == 0 {
		time.Sleep(v.ctx.APIMinInterval)
		res, err = v.gp.GetMeasurement(id)
		if err != nil {
			return err
		}
	}

	if v.ctx.ToJSON {
		for res.Status == globalping.StatusInProgress {
			time.Sleep(v.ctx.APIMinInterval)
			res, err = v.gp.GetMeasurement(res.ID)
			if err != nil {
				return err
			}
		}
		return v.OutputJson(id)
	}

	if len(res.Results) == 1 {
		if v.ctx.ToLatency {
			return v.outputTableView(res)
		}
		return v.outputStreamingPackets(res)
	}
	return v.outputTableView(res)
}

func (v *viewer) outputStreamingPackets(res *globalping.Measurement) error {
	if len(v.ctx.CompletedStats) == 0 {
		v.ctx.CompletedStats = []MeasurementStats{NewMeasurementStats()}
		v.ctx.InProgressStats = []MeasurementStats{NewMeasurementStats()}
	}
	printHeader := true
	linesPrinted := 0
	var err error
	for {
		measurement := &res.Results[0]
		if isFailedMeasurement(res) {
			return v.outputFailSummary(res)
		}
		if measurement.Result.RawOutput != "" {
			parsedOutput := parsePingRawOutput(measurement, v.ctx.CompletedStats[0].Sent)
			if printHeader && v.ctx.CompletedStats[0].Sent == 0 {
				v.ctx.Hostname = parsedOutput.Hostname
				v.printer.Println(generateProbeInfo(measurement, !v.ctx.CI))
				v.printer.Printf("PING %s (%s) %s bytes of data.\n",
					parsedOutput.Hostname,
					parsedOutput.Address,
					parsedOutput.BytesOfData,
				)
				printHeader = false
			}
			for linesPrinted < len(parsedOutput.RawPacketLines) {
				v.printer.Println(parsedOutput.RawPacketLines[linesPrinted])
				linesPrinted++
			}
			v.ctx.InProgressStats[0] = mergeMeasurementStats(v.ctx.CompletedStats[0], parsedOutput)
			if res.Status != globalping.StatusInProgress {
				v.ctx.CompletedStats[0] = v.ctx.InProgressStats[0]
			}
		}
		if res.Status != globalping.StatusInProgress {
			break
		}
		time.Sleep(v.ctx.APIMinInterval)
		res, err = v.gp.GetMeasurement(res.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *viewer) outputTableView(res *globalping.Measurement) error {
	var err error
	if len(v.ctx.CompletedStats) == 0 {
		// Initialize state
		v.ctx.CompletedStats = make([]MeasurementStats, len(res.Results))
		for i := range v.ctx.CompletedStats {
			v.ctx.CompletedStats[i].Last = -1
			v.ctx.CompletedStats[i].Min = math.MaxFloat64
			v.ctx.CompletedStats[i].Avg = -1
			v.ctx.CompletedStats[i].Max = -1
		}
		// Create new writer
		v.ctx.Area, err = pterm.DefaultArea.Start()
		if err != nil {
			return errors.New("failed to start writer: " + err.Error())
		}
	}
	for {
		if isFailedMeasurement(res) {
			return v.outputFailSummary(res)
		}
		o, stats := v.generateTable(res, pterm.GetTerminalWidth()-2)
		if o != nil {
			v.ctx.Area.Update(*o)
		}
		if stats != nil {
			v.ctx.InProgressStats = stats
		}
		if res.Status != globalping.StatusInProgress {
			if stats != nil {
				v.ctx.CompletedStats = stats
			}
			break
		}
		time.Sleep(v.ctx.APIMinInterval)
		res, err = v.gp.GetMeasurement(res.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *viewer) outputFailSummary(res *globalping.Measurement) error {
	for i := range res.Results {
		v.printer.Println(generateProbeInfo(&res.Results[i], !v.ctx.CI))
		v.printer.Println(res.Results[i].Result.RawOutput)
	}
	return errors.New("all probes failed")
}

func isFailedMeasurement(res *globalping.Measurement) bool {
	for i := range res.Results {
		if res.Results[i].Result.Status != globalping.StatusFailed {
			return false
		}
	}
	return true
}

func formatDuration(ms float64) string {
	if ms < 10 {
		return fmt.Sprintf("%.2f ms", ms)
	}
	if ms < 100 {
		return fmt.Sprintf("%.1f ms", ms)
	}
	return fmt.Sprintf("%.0f ms", ms)
}

func (v *viewer) generateTable(res *globalping.Measurement, areaWidth int) (*string, []MeasurementStats) {
	table := [][7]string{{"Location", "Sent", "Loss", "Last", "Min", "Avg", "Max"}}
	// Calculate max column width and max line width
	// We handle multi-line values only for the first column
	maxLineWidth := 0
	colMax := [7]int{len(table[0][0]), 4, 7, 8, 8, 8, 8}
	for i := 1; i < len(table[0]); i++ {
		maxLineWidth += len(table[i]) + len(colSeparator)
	}
	skip := false
	newStats := make([]MeasurementStats, len(res.Results))
	for i := range res.Results {
		measurement := &res.Results[i]
		if measurement.Result.RawOutput == "" {
			skip = true
			break
		}
		parsedOutput := parsePingRawOutput(measurement, v.ctx.CompletedStats[i].Sent)
		newStats[i] = mergeMeasurementStats(v.ctx.CompletedStats[i], parsedOutput)
		row := getRowValues(&newStats[i])
		rowWidth := 0
		for j := 1; j < len(row); j++ {
			rowWidth += len(row[j]) + len(colSeparator)
			colMax[j] = max(colMax[j], len(row[j]))
		}
		maxLineWidth = max(maxLineWidth, rowWidth)
		row[0] = getLocationText(measurement)
		colMax[0] = max(colMax[0], len(row[0]))
		table = append(table, row)
	}
	if skip {
		return nil, nil
	}
	remainingWidth := max(areaWidth-maxLineWidth, 6) // Remaining width for first column
	colMax[0] = min(colMax[0], remainingWidth)       // Truncate first column if necessary
	// Generate table string
	output := ""
	for i := range table {
		table[i][0] = strings.ReplaceAll(table[i][0], "\t", "  ") // Replace tabs with spaces
		lines := strings.Split(table[i][0], "\n")                 // Split first column into lines
		color := pterm.Reset                                      // No color
		if i == 0 {
			color = pterm.FgLightCyan
		}
		for k := range lines {
			width := runewidth.StringWidth(lines[k])
			if colMax[0] < width {
				lines[k] = runewidth.FillRight(
					runewidth.Truncate(lines[k], colMax[0], "..."),
					colMax[0],
				)
			} else if colMax[0] > width {
				lines[k] = runewidth.FillRight(lines[k], colMax[0])
			}
			if color != 0 {
				lines[k] = pterm.NewStyle(color).Sprint(lines[k])
			}
		}
		for j := 1; j < len(table[i]); j++ {
			lines[0] += colSeparator + formatValue(table[i][j], color, colMax[j], j != 0)
			for k := 1; k < len(lines); k++ {
				lines[k] += colSeparator + formatValue("", 0, colMax[j], false)
			}
		}
		for j := 0; j < len(lines); j++ {
			output += lines[j] + "\n"
		}
	}
	return &output, newStats
}

func mergeMeasurementStats(stats MeasurementStats, o *ParsedPingOutput) MeasurementStats {
	if o.Stats.Rcv > 0 {
		if o.Stats.Min < stats.Min && o.Stats.Min != 0 {
			stats.Min = o.Stats.Min
		}
		if o.Stats.Max > stats.Max {
			stats.Max = o.Stats.Max
		}
		stats.Tsum += o.Stats.Tsum
		stats.Tsum2 += o.Stats.Tsum2
		stats.Rcv += o.Stats.Rcv
		stats.Avg = stats.Tsum / float64(stats.Rcv)
		stats.Mdev = computeMdev(stats.Tsum, stats.Tsum2, stats.Rcv, stats.Avg)
		stats.Last = o.Timings[len(o.Timings)-1].RTT
	}
	stats.Sent += o.Stats.Sent
	stats.Lost += o.Stats.Lost
	stats.Time += o.Time
	if stats.Sent > 0 {
		stats.Loss = float64(stats.Lost) / float64(stats.Sent) * 100
	}
	return stats
}

func getRowValues(stats *MeasurementStats) [7]string {
	last := "-"
	min := "-"
	avg := "-"
	max := "-"
	if stats.Last != -1 {
		last = formatDuration(stats.Last)
	}
	if stats.Min != math.MaxFloat64 {
		min = formatDuration(stats.Min)
	}
	if stats.Avg != -1 {
		avg = formatDuration(stats.Avg)
	}
	if stats.Max != -1 {
		max = formatDuration(stats.Max)
	}
	return [7]string{
		"",
		fmt.Sprintf("%d", stats.Sent),
		fmt.Sprintf("%.2f", stats.Loss) + "%",
		last,
		min,
		avg,
		max,
	}
}

func formatValue(v string, color pterm.Color, width int, toRight bool) string {
	for len(v) < width {
		if toRight {
			v = " " + v
		} else {
			v = v + " "
		}
	}
	if color != 0 {
		v = pterm.NewStyle(color).Sprint(v)
	}
	return v
}

type ParsedPingOutput struct {
	Hostname       string
	Address        string
	BytesOfData    string
	RawPacketLines []string
	Timings        []globalping.PingTiming
	Stats          *MeasurementStats
	Time           float64
}

// Parse ping's raw output. Adapted from iputils ping: https://github.com/iputils/iputils/tree/1c08152/ping
//
// - If startIncmpSeq is -1, RawPacketLines will be empty
//
// - Stats.Time will be 0 if no summary is found
func parsePingRawOutput(m *globalping.ProbeMeasurement, startIncmpSeq int) *ParsedPingOutput {
	res := &ParsedPingOutput{
		Timings: make([]globalping.PingTiming, 0),
		Stats: &MeasurementStats{
			Min: math.MaxFloat64,
			Max: -1,
			Avg: -1,
		},
	}
	scanner := bufio.NewScanner(strings.NewReader(m.Result.RawOutput))
	scanner.Scan()
	header := scanner.Text()
	words := strings.Split(header, " ")
	if len(words) > 2 {
		res.Hostname = words[1]
		if len(words[2]) > 1 && words[2][0] == '(' {
			res.Address = words[2][1 : len(words[2])-1]
		} else {
			res.Address = words[2]
		}
		res.BytesOfData = words[3]
	}

	sentMap := make([]bool, 0)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			break
		}
		// Find icmp_seq
		icmp_seq := -1
		icmp_seq_index := 0
		words := strings.Split(line, " ")
		for icmp_seq_index < len(words) {
			if strings.HasPrefix(words[icmp_seq_index], "icmp_seq=") {
				n, err := strconv.Atoi(words[icmp_seq_index][9:])
				if err != nil {
				} else {
					icmp_seq = n - 1 // icmp_seq starts at 1
				}
				break
			}
			icmp_seq_index++
		}
		if icmp_seq >= len(sentMap) {
			sentMap = append(sentMap, false)
		}
		// Get timing
		if icmp_seq != -1 {
			if words[1] == "bytes" && words[2] == "from" {
				if !sentMap[icmp_seq] {
					res.Stats.Sent++
				}
				res.Stats.Rcv++
				ttl, _ := strconv.Atoi(words[icmp_seq_index+1][4:])
				rtt, _ := strconv.ParseFloat(words[icmp_seq_index+2][5:], 64)
				res.Stats.Min = math.Min(res.Stats.Min, rtt)
				res.Stats.Max = math.Max(res.Stats.Max, rtt)
				res.Stats.Tsum += rtt
				res.Stats.Tsum2 += rtt * rtt
				res.Timings = append(res.Timings, globalping.PingTiming{
					TTL: ttl,
					RTT: rtt,
				})
			} else {
				if !sentMap[icmp_seq] {
					res.Stats.Sent++
				}
				sentMap[icmp_seq] = true
			}
			if startIncmpSeq != -1 {
				words[icmp_seq_index] = "icmp_seq=" + strconv.Itoa(startIncmpSeq+icmp_seq+1)
				line = strings.Join(words, " ")
			}
		}
		if startIncmpSeq != -1 {
			res.RawPacketLines = append(res.RawPacketLines, line)
		}
	}
	hasSummary := scanner.Scan()
	if hasSummary {
		// Parse summary
		scanner.Scan() // skip ---  ping statistics ---
		line := scanner.Text()
		words = strings.Split(line, " ")
		if len(words) < 3 {
			return res
		}
		if words[1] == "packets" && words[2] == "transmitted," {
			res.Stats.Sent, _ = strconv.Atoi(words[0])
			res.Stats.Rcv, _ = strconv.Atoi(words[3])
			res.Time, _ = strconv.ParseFloat(words[9][:len(words[9])-2], 64)
		}
	}
	if res.Stats.Sent > 0 {
		res.Stats.Lost = res.Stats.Sent - res.Stats.Rcv
		res.Stats.Loss = float64(res.Stats.Lost) / float64(res.Stats.Sent) * 100
		if res.Stats.Rcv > 0 {
			res.Stats.Avg = res.Stats.Tsum / float64(res.Stats.Rcv)
			res.Stats.Mdev = computeMdev(res.Stats.Tsum, res.Stats.Tsum2, res.Stats.Rcv, res.Stats.Avg)
			res.Stats.Last = res.Timings[len(res.Timings)-1].RTT
		}
	}
	return res
}

// https://github.com/iputils/iputils/tree/1c08152/ping/ping_common.c#L917
func computeMdev(tsum float64, tsum2 float64, rcv int, avg float64) float64 {
	if tsum < math.MaxInt32 {
		return math.Sqrt((tsum2 - ((tsum * tsum) / float64(rcv))) / float64(rcv))
	}
	return math.Sqrt(tsum2/float64(rcv) - avg*avg)
}
