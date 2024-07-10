package view

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/utils"
	"github.com/mattn/go-runewidth"
)

var (
	// Table defaults
	colSeparator = " | "

	apiCreditInfo                 = "Consuming 1 API credit for every 16 packets until stopped.\n"
	apiCreditConsumptionInfo      = "Consuming ~%s/minute.\n"
	apiCreditLastConsumptionInfo  = ""
	apiCreditLastMeasurementCount = 0
)

func (v *viewer) OutputInfinite(m *globalping.Measurement) error {
	if v.ctx.ToJSON {
		if m.Status == globalping.StatusInProgress {
			return nil
		}
		return v.OutputJson(m.ID)
	}

	if isFailedMeasurement(m) {
		return v.outputFailSummary(m)
	}

	if len(m.Results) == 1 {
		if v.ctx.ToLatency {
			return v.outputTableView(m)
		}
		return v.outputStreamingPackets(m)
	}
	return v.outputTableView(m)
}

func (v *viewer) outputStreamingPackets(m *globalping.Measurement) error {
	if len(v.ctx.AggregatedStats) == 0 {
		v.ctx.AggregatedStats = []*MeasurementStats{NewMeasurementStats()}
		v.printer.Print(v.getAPICreditInfo())
	}
	probeMeasurement := &m.Results[0]
	hm := v.ctx.History.Find(m.ID)
	if probeMeasurement.Result.RawOutput != "" {
		concurrentStats := v.aggregateConcurrentStats(v.ctx.AggregatedStats[0], 0, m.ID)
		parsedOutput := v.parsePingRawOutput(hm, probeMeasurement, concurrentStats.Sent)
		if len(hm.Stats) == 0 {
			hm.Stats = make([]*MeasurementStats, 1)
		}
		hm.Stats[0] = parsedOutput.Stats
		if !v.ctx.IsHeaderPrinted {
			v.ctx.Hostname = parsedOutput.Hostname
			v.printer.Println(v.getProbeInfo(probeMeasurement))
			v.printer.Printf("PING %s (%s) %s bytes of data.\n",
				parsedOutput.Hostname,
				parsedOutput.Address,
				parsedOutput.BytesOfData,
			)
			v.ctx.IsHeaderPrinted = true
		}
		for hm.LinesPrinted < len(parsedOutput.RawPacketLines) {
			v.printer.Println(parsedOutput.RawPacketLines[hm.LinesPrinted])
			hm.LinesPrinted++
		}
		if m.Status != globalping.StatusInProgress {
			v.ctx.AggregatedStats[0] = mergeMeasurementStats(*v.ctx.AggregatedStats[0], parsedOutput.Stats)
		}
	}
	return nil
}

func (v *viewer) outputTableView(m *globalping.Measurement) error {
	if len(v.ctx.AggregatedStats) == 0 {
		// Initialize state
		v.ctx.AggregatedStats = make([]*MeasurementStats, len(m.Results))
		for i := range m.Results {
			v.ctx.AggregatedStats[i] = NewMeasurementStats()
		}
	}
	hm := v.ctx.History.Find(m.ID)
	width, _ := v.printer.GetSize()
	o, newStats, newAggregatedStats := v.generateTable(hm, m, width-2)
	hm.Stats = newStats
	output := *o + v.getAPICreditConsumptionInfo(width)
	v.printer.AreaUpdate(&output)
	if m.Status != globalping.StatusInProgress {
		v.ctx.AggregatedStats = newAggregatedStats
	}
	return nil
}

func (v *viewer) outputFailSummary(m *globalping.Measurement) error {
	for i := range m.Results {
		v.printer.Println(v.getProbeInfo(&m.Results[i]))
		v.printer.Println(m.Results[i].Result.RawOutput)
	}
	return errors.New("all probes failed")
}

func isFailedMeasurement(m *globalping.Measurement) bool {
	for i := range m.Results {
		if m.Results[i].Result.Status != globalping.StatusFailed {
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

func (v *viewer) generateTable(hm *HistoryItem, m *globalping.Measurement, areaWidth int) (*string, []*MeasurementStats, []*MeasurementStats) {
	table := [][7]string{{"Location", "Sent", "Loss", "Last", "Min", "Avg", "Max"}}
	// Calculate max column width and max line width
	// We handle multi-line values only for the first column
	maxLineWidth := 0
	colMax := [7]int{len(table[0][0]), 4, 7, 8, 8, 8, 8}
	for i := 1; i < len(table[0]); i++ {
		maxLineWidth += len(table[i]) + len(colSeparator)
	}
	newAggregatedStats := make([]*MeasurementStats, len(m.Results))
	newStats := make([]*MeasurementStats, len(m.Results))
	for i := range m.Results {
		probeMeasurement := &m.Results[i]
		parsedOutput := v.parsePingRawOutput(hm, probeMeasurement, -1)
		newAggregatedStats[i] = mergeMeasurementStats(*v.ctx.AggregatedStats[i], parsedOutput.Stats)
		newStats[i] = parsedOutput.Stats
		row := getRowValues(v.aggregateConcurrentStats(newAggregatedStats[i], i, m.ID))
		rowWidth := 0
		for j := 1; j < len(row); j++ {
			rowWidth += len(row[j]) + len(colSeparator)
			colMax[j] = max(colMax[j], len(row[j]))
		}
		maxLineWidth = max(maxLineWidth, rowWidth)
		row[0] = getLocationText(probeMeasurement)
		colMax[0] = max(colMax[0], len(row[0]))
		table = append(table, row)
	}
	remainingWidth := max(areaWidth-maxLineWidth, 6) // Remaining width for first column
	colMax[0] = min(colMax[0], remainingWidth)       // Truncate first column if necessary
	// Generate table string
	output := ""
	for i := range table {
		table[i][0] = strings.ReplaceAll(table[i][0], "\t", "  ") // Replace tabs with spaces
		lines := strings.Split(table[i][0], "\n")                 // Split first column into lines
		color := ColorNone                                        // No color
		if i == 0 && !v.ctx.CIMode {
			color = ColorLightCyan
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
			if color != "" {
				lines[k] = v.printer.Color(lines[k], color)
			}
		}
		for j := 1; j < len(table[i]); j++ {
			lines[0] += colSeparator
			if j == 0 {
				lines[0] += v.printer.FillRightAndColor(table[i][j], colMax[j], color)
			} else {
				lines[0] += v.printer.FillLeftAndColor(table[i][j], colMax[j], color)
			}
			for k := 1; k < len(lines); k++ {
				lines[k] += colSeparator + v.printer.FillLeft("", colMax[j])
			}
		}
		for j := 0; j < len(lines); j++ {
			output += lines[j] + "\n"
		}
	}
	return &output, newStats, newAggregatedStats
}

func (v *viewer) aggregateConcurrentStats(completed *MeasurementStats, probeIndex int, excludeId string) *MeasurementStats {
	inProgressStats := v.ctx.History.FilterByStatus(globalping.StatusInProgress)
	for i := range inProgressStats {
		if inProgressStats[i].Id == excludeId {
			continue
		}
		if len(inProgressStats[i].Stats) == 0 {
			continue
		}
		completed = mergeMeasurementStats(*completed, inProgressStats[i].Stats[probeIndex])
	}
	return completed
}

func mergeMeasurementStats(stats MeasurementStats, newStats *MeasurementStats) *MeasurementStats {
	if newStats.Rcv > 0 {
		if newStats.Min < stats.Min && newStats.Min != 0 {
			stats.Min = newStats.Min
		}
		if newStats.Max > stats.Max {
			stats.Max = newStats.Max
		}
		stats.Tsum += newStats.Tsum
		stats.Tsum2 += newStats.Tsum2
		stats.Rcv += newStats.Rcv
		stats.Avg = stats.Tsum / float64(stats.Rcv)
		stats.Mdev = computeMdev(stats.Tsum, stats.Tsum2, stats.Rcv, stats.Avg)
		stats.Last = newStats.Last
	}
	stats.Sent += newStats.Sent
	stats.Lost += newStats.Lost
	stats.Time += newStats.Time
	if stats.Sent > 0 {
		stats.Loss = float64(stats.Lost) / float64(stats.Sent) * 100
	}
	return &stats
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

type ParsedPingOutput struct {
	Hostname       string
	Address        string
	BytesOfData    string
	RawPacketLines []string
	Timings        []globalping.PingTiming
	Stats          *MeasurementStats
}

// Parse ping's raw output. Adapted from iputils ping: https://github.com/iputils/iputils/tree/1c08152/ping
//
// - If startIncmpSeq is -1, RawPacketLines will be empty
func (v *viewer) parsePingRawOutput(
	hm *HistoryItem,
	m *globalping.ProbeMeasurement,
	startIncmpSeq int,
) *ParsedPingOutput {
	res := &ParsedPingOutput{
		Timings: make([]globalping.PingTiming, 0),
		Stats:   NewMeasurementStats(),
	}
	if m.Result.RawOutput == "" {
		return res
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
		if len(words) > 9 && words[1] == "packets" && words[2] == "transmitted," {
			res.Stats.Sent, _ = strconv.Atoi(words[0])
			res.Stats.Rcv, _ = strconv.Atoi(words[3])
			res.Stats.Time, _ = strconv.ParseFloat(words[9][:len(words[9])-2], 64)
		}
	} else {
		res.Stats.Time = float64(v.time.Now().Sub(hm.StartedAt).Milliseconds())
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

func (v *viewer) getAPICreditInfo() string {
	if v.ctx.CIMode {
		return apiCreditInfo
	}
	return v.printer.Color(apiCreditInfo, ColorLightYellow)
}

func (v *viewer) getAPICreditConsumptionInfo(width int) string {
	if v.ctx.MeasurementsCreated < 2 {
		return ""
	}
	if v.ctx.MeasurementsCreated == apiCreditLastMeasurementCount {
		return apiCreditLastConsumptionInfo
	}
	apiCreditLastMeasurementCount = v.ctx.MeasurementsCreated
	elapsedMinutes := v.time.Now().Sub(v.ctx.RunSessionStartedAt).Minutes()
	consumption := int64(math.Ceil(float64(apiCreditLastMeasurementCount*len(v.ctx.AggregatedStats)) / elapsedMinutes))
	info := fmt.Sprintf(apiCreditConsumptionInfo, utils.Pluralize(consumption, "API credit"))
	if len(info) > width-4 {
		info = info[:width-5] + "..."
	}
	if v.ctx.CIMode {
		apiCreditLastConsumptionInfo = info
	} else {
		apiCreditLastConsumptionInfo = v.printer.Color(info, ColorLightYellow)
	}
	return apiCreditLastConsumptionInfo
}
