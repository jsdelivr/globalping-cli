package view

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/jsdelivr/globalping-cli/utils"
	"github.com/jsdelivr/globalping-go"
)

var (
	apiCreditInfo                 = "Consuming 1 API credit for every 16 packets until stopped.\n"
	apiCreditConsumptionInfo      = "Consuming ~%s/minute.\n"
	apiCreditLastConsumptionInfo  = ""
	apiCreditLastMeasurementCount = 0
)

func (v *viewer) OutputInfinite(measurement *globalping.Measurement) error {
	if len(measurement.Results) == 1 && !v.ctx.ToLatency && !v.ctx.Table {
		if measurement.Status != globalping.MeasurementStatusInProgress && !isSomeTestFinished(measurement) {
			return v.outputFailSummary(measurement)
		}
		return v.outputStreamingPackets(measurement)
	}
	return v.OutputTable(measurement)
}

func (v *viewer) outputStreamingPackets(m *globalping.Measurement) error {
	if len(v.ctx.AggregatedStats) == 0 {
		v.ctx.AggregatedStats = []*MeasurementStats{NewMeasurementStats()}
		v.printer.ErrPrint(v.getAPICreditInfo())
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
			v.printer.ErrPrintln(v.getProbeInfo(probeMeasurement))
			if v.ctx.Protocol == "ICMP" {
				v.printer.Printf("PING %s (%s) %s bytes of data.\n",
					parsedOutput.Hostname,
					parsedOutput.Address,
					parsedOutput.BytesOfData,
				)
			} else {
				v.printer.Println(parsedOutput.Header)
			}
			v.ctx.IsHeaderPrinted = true
		}
		for hm.LinesPrinted < len(parsedOutput.RawPacketLines) {
			v.printer.Println(parsedOutput.RawPacketLines[hm.LinesPrinted])
			hm.LinesPrinted++
		}
		if m.Status != globalping.MeasurementStatusInProgress {
			v.ctx.AggregatedStats[0] = mergeMeasurementStats(*v.ctx.AggregatedStats[0], parsedOutput.Stats)
		}
	}
	return nil
}

func (v *viewer) outputPingTableView(m *globalping.Measurement) error {
	if len(v.ctx.AggregatedStats) == 0 {
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
	if m.Status != globalping.MeasurementStatusInProgress {
		v.ctx.AggregatedStats = newAggregatedStats
	}
	return nil
}

func (v *viewer) outputFailSummary(m *globalping.Measurement) error {
	for i := range m.Results {
		v.printer.ErrPrintln(v.getProbeInfo(&m.Results[i]))
		v.printer.Println(m.Results[i].Result.RawOutput)
	}
	return errors.New("all probes failed")
}

func isSomeTestFinished(m *globalping.Measurement) bool {
	for i := range m.Results {
		if m.Results[i].Result.Status == globalping.TestStatusFinished {
			return true
		}
	}
	return false
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
	rows := [][]string{{"Location", "Sent", "Loss", "Last", "Min", "Avg", "Max"}}
	newAggregatedStats := make([]*MeasurementStats, len(m.Results))
	newStats := make([]*MeasurementStats, len(m.Results))
	for i := range m.Results {
		probeMeasurement := &m.Results[i]
		var row []string
		if probeMeasurement.Result.Status == globalping.TestStatusFailed || probeMeasurement.Result.Status == globalping.TestStatusOffline {
			preservedStats := *v.ctx.AggregatedStats[i]
			newAggregatedStats[i] = &preservedStats
			newStats[i] = NewMeasurementStats()
			if probeMeasurement.Result.Status == globalping.TestStatusFailed && !v.ctx.Infinite {
				row = []string{"", failureTableMessage(&probeMeasurement.Result)}
			} else {
				row = []string{"", "-", "-", "-", "-", "-", "-"}
			}
		} else {
			parsedOutput := v.parsePingRawOutput(hm, probeMeasurement, -1)
			newAggregatedStats[i] = mergeMeasurementStats(*v.ctx.AggregatedStats[i], parsedOutput.Stats)
			newStats[i] = parsedOutput.Stats
			values := getRowValues(v.aggregateConcurrentStats(newAggregatedStats[i], i, m.ID))
			row = values[:]
		}
		row[0] = getLocationText(probeMeasurement)
		rows = append(rows, row)
	}
	output := v.renderPingTable(rows, areaWidth)
	return &output, newStats, newAggregatedStats
}

func (v *viewer) aggregateConcurrentStats(completed *MeasurementStats, probeIndex int, excludeId string) *MeasurementStats {
	inProgressStats := v.ctx.History.FilterByStatus(globalping.MeasurementStatusInProgress)
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
	Header         string
	Hostname       string
	Address        string
	BytesOfData    string
	RawPacketLines []string
	Timings        []globalping.PingTiming
	Stats          *MeasurementStats
}

// Parse ping's raw output. Adapted from iputils ping: https://github.com/iputils/iputils/tree/1c08152/ping
//
// - If startSequence is -1, RawPacketLines will be empty
func (v *viewer) parsePingRawOutput(
	hm *HistoryItem,
	m *globalping.ProbeMeasurement,
	startSequence int,
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
	res.Header = scanner.Text()
	words := strings.Split(res.Header, " ")
	if len(words) > 2 {
		res.Hostname = words[1]
		if len(words[2]) > 1 && words[2][0] == '(' {
			res.Address = words[2][1 : len(words[2])-1]
		} else {
			res.Address = words[2]
		}
		if v.ctx.Protocol == "ICMP" {
			res.BytesOfData = words[3]
		}
	}

	sentMap := make([]bool, 0)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			break
		}
		if v.ctx.Protocol == "TCP" {
			line, sentMap = parseTCPLine(line, sentMap, startSequence, res)
		} else {
			line, sentMap = parseICMPLine(line, sentMap, startSequence, res)
		}
		if startSequence != -1 {
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
			if v.ctx.Protocol == "TCP" {
				res.Stats.Time, _ = strconv.ParseFloat(words[9], 64)
			} else {
				res.Stats.Time, _ = strconv.ParseFloat(words[9][:len(words[9])-2], 64)
			}
		}
	} else {
		res.Stats.Time = float64(v.utils.Now().Sub(hm.StartedAt).Milliseconds())
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

func parseICMPLine(line string, sentMap []bool, startSequence int, res *ParsedPingOutput) (string, []bool) {
	seq := -1
	seqIndex := 0
	words := strings.Split(line, " ")
	for seqIndex < len(words) {
		if strings.HasPrefix(words[seqIndex], "icmp_seq=") {
			n, err := strconv.Atoi(words[seqIndex][9:])
			if err == nil {
				seq = n - 1 // seq starts at 1
			}
			break
		}
		seqIndex++
	}
	if seq >= len(sentMap) {
		sentMap = append(sentMap, false)
	}
	// Get timing
	if seq != -1 {
		if words[1] == "bytes" && words[2] == "from" {
			if !sentMap[seq] {
				res.Stats.Sent++
			}
			res.Stats.Rcv++
			ttl, _ := strconv.Atoi(words[seqIndex+1][4:])
			rtt, _ := strconv.ParseFloat(words[seqIndex+2][5:], 64)
			res.Stats.Min = math.Min(res.Stats.Min, rtt)
			res.Stats.Max = math.Max(res.Stats.Max, rtt)
			res.Stats.Tsum += rtt
			res.Stats.Tsum2 += rtt * rtt
			res.Timings = append(res.Timings, globalping.PingTiming{
				TTL: ttl,
				RTT: rtt,
			})
		} else {
			if !sentMap[seq] {
				res.Stats.Sent++
			}
			sentMap[seq] = true
		}
		// replace sequence number
		if startSequence != -1 {
			words[seqIndex] = "icmp_seq=" + strconv.Itoa(startSequence+seq+1)
			line = strings.Join(words, " ")
		}
	}
	return line, sentMap
}

func parseTCPLine(line string, sentMap []bool, startSequence int, res *ParsedPingOutput) (string, []bool) {
	seq := -1
	seqIndex := 0
	words := strings.Split(line, " ")
	for seqIndex < len(words) {
		if strings.HasPrefix(words[seqIndex], "tcp_conn=") {
			n, err := strconv.Atoi(words[seqIndex][9:])
			if err == nil {
				seq = n - 1 // seq starts at 1
			}
			break
		}
		seqIndex++
	}
	if seq >= len(sentMap) {
		sentMap = append(sentMap, false)
	}
	// Get timing
	if seq != -1 {
		if words[0] == "Reply" && words[1] == "from" {
			if !sentMap[seq] {
				res.Stats.Sent++
			}
			res.Stats.Rcv++
			rtt, _ := strconv.ParseFloat(words[seqIndex+1][5:], 64)
			res.Stats.Min = math.Min(res.Stats.Min, rtt)
			res.Stats.Max = math.Max(res.Stats.Max, rtt)
			res.Stats.Tsum += rtt
			res.Stats.Tsum2 += rtt * rtt
			res.Timings = append(res.Timings, globalping.PingTiming{
				RTT: rtt,
			})
		} else {
			if !sentMap[seq] {
				res.Stats.Sent++
			}
			sentMap[seq] = true
		}
		// replace sequence number
		if startSequence != -1 {
			words[seqIndex] = "tcp_conn=" + strconv.Itoa(startSequence+seq+1)
			line = strings.Join(words, " ")
		}
	}
	return line, sentMap
}

// https://github.com/iputils/iputils/tree/1c08152/ping/ping_common.c#L917
func computeMdev(tsum float64, tsum2 float64, rcv int, avg float64) float64 {
	if tsum < math.MaxInt32 {
		return math.Sqrt((tsum2 - ((tsum * tsum) / float64(rcv))) / float64(rcv))
	}
	return math.Sqrt(tsum2/float64(rcv) - avg*avg)
}

func (v *viewer) getAPICreditInfo() string {
	return v.printer.Color(apiCreditInfo, FGBrightYellow)
}

func (v *viewer) getAPICreditConsumptionInfo(width int) string {
	if v.ctx.MeasurementsCreated < 2 {
		return ""
	}
	if v.ctx.MeasurementsCreated == apiCreditLastMeasurementCount {
		return apiCreditLastConsumptionInfo
	}
	apiCreditLastMeasurementCount = v.ctx.MeasurementsCreated
	elapsedMinutes := v.utils.Now().Sub(v.ctx.RunSessionStartedAt).Minutes()
	consumption := int64(math.Ceil(float64((apiCreditLastMeasurementCount-1)*(len(v.ctx.AggregatedStats))) / elapsedMinutes))
	info := fmt.Sprintf(apiCreditConsumptionInfo, utils.Pluralize(consumption, "API credit"))
	if len(info) > width-4 {
		info = info[:max(width-5, 0)] + "..."
	}

	apiCreditLastConsumptionInfo = v.printer.Color(info, FGBrightYellow)
	return apiCreditLastConsumptionInfo
}
