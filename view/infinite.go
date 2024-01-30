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
	if ctx.History == nil {
		ctx.History = model.NewRbuffer(ctx.MaxHistory)
	}
	ctx.History.Push(id)

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

func OutputSummary(ctx *model.Context) {
	if len(ctx.InProgressStats) == 0 {
		return
	}

	if len(ctx.InProgressStats) == 1 {
		stats := ctx.InProgressStats[0]

		fmt.Printf("\n--- %s ping statistics ---\n", ctx.Hostname)
		fmt.Printf("%d packets transmitted, %d received, %.2f%% packet loss, time %.0fms\n",
			stats.Sent,
			stats.Rcv,
			stats.Loss,
			stats.Time,
		)
		// TODO: Add mdev
		min := "-"
		avg := "-"
		max := "-"
		if stats.Min != math.MaxFloat64 {
			min = fmt.Sprintf("%.3f", stats.Min)
		}
		if stats.Avg != -1 {
			avg = fmt.Sprintf("%.3f", stats.Avg)
		}
		if stats.Max != -1 {
			max = fmt.Sprintf("%.3f", stats.Max)
		}
		fmt.Printf("rtt min/avg/max = %s/%s/%s ms\n", min, avg, max)
	}

	if ctx.Share && ctx.History != nil {
		if len(ctx.InProgressStats) > 1 {
			fmt.Println()
		}
		ids := ctx.History.ToString("+")
		if ids != "" {
			fmt.Println(formatWithLeadingArrow(shareMessage(ids), !ctx.CI))
		}
		if ctx.CallCount > ctx.MaxHistory {
			fmt.Printf("For long-running continuous mode measurements, only the last %d packets are shared.\n", ctx.Packets*ctx.MaxHistory)
		}
	}
}

func outputSingleLocation(
	fetcher client.MeasurementsFetcher,
	res *model.GetMeasurement,
	ctx *model.Context,
) error {
	if len(ctx.CompletedStats) == 0 {
		ctx.CompletedStats = []model.MeasurementStats{model.NewMeasurementStats()}
		ctx.InProgressStats = []model.MeasurementStats{model.NewMeasurementStats()}
	}
	printHeader := true
	linesPrinted := 0
	var err error
	for {
		measurement := &res.Results[0]
		if measurement.Result.RawOutput != "" {
			parsedOutput := parsePingRawOutput(measurement, ctx.CompletedStats[0].Sent)
			if printHeader && ctx.CompletedStats[0].Sent == 0 {
				ctx.Hostname = parsedOutput.Hostname
				fmt.Println(generateProbeInfo(measurement, !ctx.CI))
				fmt.Printf("PING %s (%s) %s bytes of data.\n",
					parsedOutput.Hostname,
					parsedOutput.Address,
					parsedOutput.BytesOfData,
				)
				printHeader = false
			}
			for linesPrinted < len(parsedOutput.RawPacketLines) {
				fmt.Println(parsedOutput.RawPacketLines[linesPrinted])
				linesPrinted++
			}
			ctx.InProgressStats[0] = mergeMeasurementStats(ctx.CompletedStats[0], measurement)
			if res.Status != model.StatusInProgress {
				ctx.CompletedStats[0] = ctx.InProgressStats[0]
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
	if len(ctx.CompletedStats) == 0 {
		// Initialize state
		ctx.CompletedStats = make([]model.MeasurementStats, len(res.Results))
		for i := range ctx.CompletedStats {
			ctx.CompletedStats[i].Last = -1
			ctx.CompletedStats[i].Min = math.MaxFloat64
			ctx.CompletedStats[i].Avg = -1
			ctx.CompletedStats[i].Max = -1
		}
		// Create new writer
		ctx.Area, err = pterm.DefaultArea.Start()
		if err != nil {
			return errors.New("failed to start writer: " + err.Error())
		}
	}
	for {
		o, stats := generateTable(res, ctx, pterm.GetTerminalWidth()-4)
		if o != nil {
			ctx.Area.Update(*o)
		}
		if stats != nil {
			ctx.InProgressStats = stats
		}
		if res.Status != model.StatusInProgress {
			if stats != nil {
				ctx.CompletedStats = stats
			}
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

func generateTable(res *model.GetMeasurement, ctx *model.Context, areaWidth int) (*string, []model.MeasurementStats) {
	table := [][7]string{{"Location", "Sent", "Loss", "Last", "Min", "Avg", "Max"}}
	// Calculate max column width and max line width
	// We handle multi-line values only for the first column
	maxLineWidth := 0
	colMax := [7]int{len(table[0][0]), 4, 7, 8, 8, 8, 8}
	for i := 1; i < len(table[0]); i++ {
		maxLineWidth += len(table[i]) + len(colSeparator)
	}
	skip := false
	newStats := make([]model.MeasurementStats, len(res.Results))
	for i := range res.Results {
		measurement := &res.Results[i]
		if measurement.Result.RawOutput == "" {
			skip = true
			break
		}
		newStats[i] = mergeMeasurementStats(ctx.CompletedStats[i], measurement)
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

func mergeMeasurementStats(mStats model.MeasurementStats, measurement *model.MeasurementResponse) model.MeasurementStats {
	o := parsePingRawOutput(measurement, mStats.Sent)
	if o.Stats.Rcv > 0 {
		if o.Stats.Min < mStats.Min && o.Stats.Min != 0 {
			mStats.Min = o.Stats.Min
		}
		if o.Stats.Max > mStats.Max {
			mStats.Max = o.Stats.Max
		}
		mStats.Avg = (mStats.Avg*float64(mStats.Sent) + o.Stats.Avg*float64(o.Stats.Total)) / float64(mStats.Sent+o.Stats.Total)
		mStats.Last = o.Timings[len(o.Timings)-1].RTT
	}
	mStats.Sent += o.Stats.Total
	mStats.Lost += o.Stats.Drop
	mStats.Time += o.Time
	mStats.Rcv += o.Stats.Rcv
	if mStats.Sent > 0 {
		mStats.Loss = float64(mStats.Lost) / float64(mStats.Sent) * 100
	}
	// TODO: Add mdev
	return mStats
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
	Hostname       string
	Address        string
	BytesOfData    string
	RawPacketLines []string
	Timings        []model.PingTiming
	Stats          *model.PingStats
	Time           float64
}

// If startIncmpSeq is -1, RawPacketLines will be empty
func parsePingRawOutput(m *model.MeasurementResponse, startIncmpSeq int) *ParsedPingOutput {
	res := &ParsedPingOutput{
		Timings: make([]model.PingTiming, 0),
		Stats: &model.PingStats{
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
		return res
	}
	scanner.Scan() // skip ---  ping statistics ---
	line := scanner.Text()
	words = strings.Split(line, " ")
	if len(words) < 3 {
		return res
	}
	if words[1] == "packets" && words[2] == "transmitted," {
		res.Stats.Total, _ = strconv.Atoi(words[0])
		res.Stats.Rcv, _ = strconv.Atoi(words[3])
		res.Stats.Loss, _ = strconv.ParseFloat(words[5][:len(words[5])-1], 64)
		res.Stats.Drop = res.Stats.Total - res.Stats.Rcv
		res.Time, _ = strconv.ParseFloat(words[9][:len(words[9])-2], 64)
	}
	hasSummary = scanner.Scan()
	if !hasSummary {
		return res
	}
	line = scanner.Text()
	words = strings.Split(line, " ")
	if len(words) < 2 {
		return res
	}
	if words[0] == "rtt" && words[1] == "min/avg/max/mdev" {
		words = strings.Split(words[3], "/")
		res.Stats.Min, _ = strconv.ParseFloat(words[0], 64)
		res.Stats.Avg, _ = strconv.ParseFloat(words[1], 64)
		res.Stats.Max, _ = strconv.ParseFloat(words[2], 64)
		res.Stats.Mdev, _ = strconv.ParseFloat(words[3], 64)
	}
	return res
}
