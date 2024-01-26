package view

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/mattn/go-runewidth"
	"github.com/pterm/pterm"
)

// Table defaults
var (
	colSeparator = " | "
)

func OutputInfinite(id string, ctx *model.Context) error {
	fetcher := client.NewMeasurementsFetcher(client.ApiUrl)
	res, err := fetcher.GetMeasurement(id)
	if err != nil {
		return err
	}
	// Probe may not have started yet
	for len(res.Results) == 0 {
		time.Sleep(ctx.APIMinInterval)
		res, err = fetcher.GetMeasurement(id)
		if err != nil {
			return err
		}
	}

	if ctx.Latency || ctx.JsonOutput {
		for res.Status == model.StatusInProgress {
			time.Sleep(ctx.APIMinInterval)
			res, err = fetcher.GetMeasurement(res.ID)
			if err != nil {
				return err
			}
		}
		if ctx.Latency {
			return OutputLatency(id, res, *ctx)
		}
		if ctx.JsonOutput {
			return OutputJson(id, fetcher, *ctx)
		}
	}

	if len(res.Results) == 1 {
		return outputSingleLocation(fetcher, res, ctx)
	}
	return outputMultipleLocations(fetcher, res, ctx)
}

func outputSingleLocation(
	fetcher client.MeasurementsFetcher,
	res *model.GetMeasurement,
	ctx *model.Context,
) error {
	if len(ctx.Stats) == 0 {
		ctx.Stats = make([]model.MeasurementStats, 1)
	}
	printHeader := true
	linesPrinted := 0
	var err error
	for {
		measurement := &res.Results[0]
		if measurement.Result.RawOutput != "" {
			parsedOutput, err := parsePingRawOutput(measurement, ctx.Stats[0].Sent)
			if err != nil {
				return err
			}
			if printHeader && ctx.Stats[0].Sent == 0 {
				fmt.Println(generateProbeInfo(measurement, !ctx.CI))
				fmt.Printf("PING %s (%s) 56(84) bytes of data.\n",
					measurement.Result.ResolvedHostname,
					measurement.Result.ResolvedAddress,
				)
				printHeader = false
			}
			for linesPrinted < len(parsedOutput.RawPacketLines) {
				fmt.Println(parsedOutput.RawPacketLines[linesPrinted])
				linesPrinted++
			}
			if res.Status != model.StatusInProgress {
				ctx.Stats[0].Sent += parsedOutput.Stats.Total
			}
		}
		if res.Status != model.StatusInProgress {
			break
		}
		time.Sleep(ctx.APIMinInterval)
		res, err = fetcher.GetMeasurement(res.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func outputMultipleLocations(
	fetcher client.MeasurementsFetcher,
	res *model.GetMeasurement,
	ctx *model.Context) error {
	var err error
	if len(ctx.Stats) == 0 {
		// Initialize state
		ctx.Stats = make([]model.MeasurementStats, len(res.Results))
		for i := range ctx.Stats {
			ctx.Stats[i].Last = -1
			ctx.Stats[i].Min = math.MaxFloat64
			ctx.Stats[i].Avg = -1
			ctx.Stats[i].Max = -1
		}
		// Create new writer
		ctx.Area, err = pterm.DefaultArea.Start()
		if err != nil {
			return errors.New("failed to start writer: " + err.Error())
		}
	}
	for {
		o := generateTable(res, ctx, pterm.GetTerminalWidth()-4)
		if o != nil {
			ctx.Area.Update(*o)
		}
		if res.Status != model.StatusInProgress {
			break
		}
		time.Sleep(ctx.APIMinInterval)
		res, err = fetcher.GetMeasurement(res.ID)
		if err != nil {
			return err
		}
	}
	return nil
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

func generateTable(res *model.GetMeasurement, ctx *model.Context, areaWidth int) *string {
	table := [][7]string{{"Location", "Sent", "Loss", "Last", "Min", "Avg", "Max"}}
	// Calculate max column width and max line width
	// We handle multi-line values only for the first column
	maxLineWidth := 0
	colMax := [7]int{len(table[0][0]), 4, 7, 8, 8, 8, 8}
	for i := 1; i < len(table[0]); i++ {
		maxLineWidth += len(table[i]) + len(colSeparator)
	}
	skip := false
	for i := range res.Results {
		measurement := &res.Results[i]
		if measurement.Result.RawOutput == "" {
			skip = true
			break
		}
		stats, _ := mergeMeasurementStats(ctx.Stats[i], measurement)
		if measurement.Result.Status != model.StatusInProgress {
			ctx.Stats[i] = *stats
		}
		row := getRowValues(stats)
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
		return nil
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
	return &output
}

func mergeMeasurementStats(mStats model.MeasurementStats, measurement *model.MeasurementResponse) (*model.MeasurementStats, error) {
	var pStats *model.PingStats
	var timings []model.PingTiming
	var err error
	if measurement.Result.Status == model.StatusInProgress {
		o, err := parsePingRawOutput(measurement, mStats.Sent)
		if err != nil {
			return nil, err
		}
		pStats = o.Stats
		timings = o.Timings
	} else {
		pStats, err = client.DecodePingStats(measurement.Result.StatsRaw)
		if err != nil {
			return nil, err
		}
		timings, err = client.DecodePingTimings(measurement.Result.TimingsRaw)
		if err != nil {
			return nil, err
		}
	}
	if pStats.Rcv > 0 {
		if pStats.Min < mStats.Min && pStats.Min != 0 {
			mStats.Min = pStats.Min
		}
		if pStats.Max > mStats.Max {
			mStats.Max = pStats.Max
		}
		mStats.Avg = (mStats.Avg*float64(mStats.Sent) + pStats.Avg*float64(pStats.Total)) / float64(mStats.Sent+pStats.Total)
		mStats.Last = timings[len(timings)-1].RTT
	}
	mStats.Sent += pStats.Total
	mStats.Lost += pStats.Drop
	if mStats.Sent > 0 {
		mStats.Loss = float64(mStats.Lost) / float64(mStats.Sent) * 100
	}
	return &mStats, nil
}

func getRowValues(stats *model.MeasurementStats) [7]string {
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
	RawPacketLines []string
	Timings        []model.PingTiming
	Stats          *model.PingStats
}

// If startIncmpSeq is -1, RawPacketLines will be empty
func parsePingRawOutput(m *model.MeasurementResponse, startIncmpSeq int) (*ParsedPingOutput, error) {
	scanner := bufio.NewScanner(strings.NewReader(m.Result.RawOutput))
	scanner.Scan()
	header := scanner.Text()
	words := strings.Split(header, " ")
	if len(words) > 2 {
		m.Result.ResolvedHostname = words[1]
		if len(words[2]) < 2 {
			return nil, errors.New("could not parse ping header")
		}
		m.Result.ResolvedAddress = words[2][1 : len(words[2])-1]
	} else {
		return nil, errors.New("could not parse ping header")
	}

	res := &ParsedPingOutput{
		Timings: make([]model.PingTiming, 0),
		Stats: &model.PingStats{
			Min: math.MaxFloat64,
			Max: -1,
			Avg: -1,
		},
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
		var err error
		words := strings.Split(line, " ")
		for icmp_seq_index < len(words) {
			if strings.HasPrefix(words[icmp_seq_index], "icmp_seq=") {
				icmp_seq, err = strconv.Atoi(words[icmp_seq_index][9:])
				icmp_seq-- // icmp_seq starts at 1
				if err != nil {
					return nil, errors.New("could not parse ping header: " + err.Error())
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
					res.Stats.Total++
				}
				res.Stats.Rcv++
				ttl, _ := strconv.Atoi(words[icmp_seq_index+1][4:])
				rtt, _ := strconv.ParseFloat(words[icmp_seq_index+2][5:], 64)
				res.Stats.Min = math.Min(res.Stats.Min, rtt)
				res.Stats.Max = math.Max(res.Stats.Max, rtt)
				if res.Stats.Rcv == 1 {
					res.Stats.Avg = rtt
				} else {
					res.Stats.Avg = (res.Stats.Avg*float64(res.Stats.Rcv-1) + rtt) / float64(res.Stats.Rcv)
				}
				res.Timings = append(res.Timings, model.PingTiming{
					TTL: ttl,
					RTT: rtt,
				})
			} else {
				if !sentMap[icmp_seq] {
					res.Stats.Total++
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
	// Parse summary
	hasSummary := scanner.Scan()
	if !hasSummary {
		res.Stats.Drop = res.Stats.Total - res.Stats.Rcv
		res.Stats.Loss = float64(res.Stats.Drop) / float64(res.Stats.Total) * 100
		return res, nil
	}
	scanner.Scan() // skip ---  ping statistics ---
	line := scanner.Text()
	words = strings.Split(line, " ")
	if len(words) < 3 {
		return res, nil
	}
	if words[1] == "packets" && words[2] == "transmitted," {
		res.Stats.Total, _ = strconv.Atoi(words[0])
		res.Stats.Rcv, _ = strconv.Atoi(words[3])
		res.Stats.Loss, _ = strconv.ParseFloat(words[5][:len(words[5])-1], 64)
		res.Stats.Drop = res.Stats.Total - res.Stats.Rcv
	}
	hasSummary = scanner.Scan()
	if !hasSummary {
		return res, nil
	}
	line = scanner.Text()
	words = strings.Split(line, " ")
	if len(words) < 2 {
		return res, nil
	}
	if words[0] == "rtt" && words[1] == "min/avg/max/mdev" {
		words = strings.Split(words[3], "/")
		res.Stats.Min, _ = strconv.ParseFloat(words[0], 64)
		res.Stats.Avg, _ = strconv.ParseFloat(words[1], 64)
		res.Stats.Max, _ = strconv.ParseFloat(words[2], 64)
	}
	return res, nil
}
