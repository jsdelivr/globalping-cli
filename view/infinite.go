package view

import (
	"errors"
	"fmt"
	"math"
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
		time.Sleep(apiPollInterval)
		res, err = fetcher.GetMeasurement(id)
		if err != nil {
			return err
		}
	}
	// Wait for results to be complete
	for res.Status == "in-progress" {
		time.Sleep(apiPollInterval)
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

	if len(res.Results) == 1 {
		return outputSingleLocation(res, ctx)
	}
	return outputMultipleLocations(res, ctx)
}

func outputSingleLocation(res *model.GetMeasurement, ctx *model.Context) error {
	measurement := &res.Results[0]
	if len(ctx.Stats) == 0 {
		// Initialize state
		ctx.Stats = make([]model.MeasurementStats, 1)
		// Print header
		fmt.Println(generateHeader(measurement, !ctx.CI))
		fmt.Printf("PING %s (%s) 56(84) bytes of data.\n", res.Target, measurement.Result.ResolvedAddress)
	}
	timings, err := client.DecodePingTimings(measurement.Result.TimingsRaw)
	if err != nil {
		return err
	}
	for i := range timings {
		ctx.Stats[0].Sent++
		t := timings[i]
		fmt.Printf("64 bytes from %s (%s): icmp_seq=%d ttl=%d time=%.2f ms\n",
			measurement.Result.ResolvedHostname,
			measurement.Result.ResolvedAddress,
			ctx.Stats[0].Sent,
			t.TTL,
			t.RTT)
	}
	return nil
}

func outputMultipleLocations(res *model.GetMeasurement, ctx *model.Context) error {
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
	ctx.Area.Update(generateTable(res, ctx, pterm.GetTerminalWidth()-4))

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

func generateTable(res *model.GetMeasurement, ctx *model.Context, areaWidth int) string {
	table := [][7]string{{"Location", "Sent", "Loss", "Last", "Min", "Avg", "Max"}}
	// Calculate max column width and max line width
	// We handle multi-line values only for the first column
	maxLineWidth := 0
	colMax := [7]int{
		len(table[0][0]),
		4,
		7,
		8,
		8,
		8,
		8,
	}
	for i := 1; i < len(table[0]); i++ {
		maxLineWidth += len(table[i]) + len(colSeparator)
	}
	for i := range res.Results {
		result := &res.Results[i]
		stats := &ctx.Stats[i]
		updateMeasurementStats(stats, result)
		row := getRowValues(stats)
		rowWidth := 0
		for j := 1; j < len(row); j++ {
			rowWidth += len(row[j]) + len(colSeparator)
			colMax[j] = max(colMax[j], len(row[j]))
		}
		maxLineWidth = max(maxLineWidth, rowWidth)
		row[0] = getLocationText(result)
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
	return output
}

func updateMeasurementStats(localStats *model.MeasurementStats, result *model.MeasurementResponse) error {
	stats, err := client.DecodePingStats(result.Result.StatsRaw)
	if err != nil {
		return err
	}
	timings, err := client.DecodePingTimings(result.Result.TimingsRaw)
	if err != nil {
		return err
	}
	if stats.Rcv > 0 {
		if stats.Min < localStats.Min && stats.Min != 0 {
			localStats.Min = stats.Min
		}
		if stats.Max > localStats.Max {
			localStats.Max = stats.Max
		}
		localStats.Avg = (localStats.Avg*float64(localStats.Sent) + stats.Avg*float64(stats.Total)) / float64(localStats.Sent+stats.Total)
		localStats.Last = timings[len(timings)-1].RTT
	}
	localStats.Sent += stats.Total
	localStats.Lost += stats.Drop
	localStats.Loss = float64(localStats.Lost) / float64(localStats.Sent) * 100
	return nil
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
