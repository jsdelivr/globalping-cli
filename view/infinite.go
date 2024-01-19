package view

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/pterm/pterm"
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
	tableData := pterm.TableData{
		{
			"Location",
			formatValue("Sent", 4, pterm.FgLightCyan),
			formatValue("Loss", 7, pterm.FgLightCyan),
			formatValue("Last", 8, pterm.FgLightCyan),
			formatValue("Min", 8, pterm.FgLightCyan),
			formatValue("Avg", 8, pterm.FgLightCyan),
			formatValue("Max", 8, pterm.FgLightCyan),
		},
	}
	for i := range res.Results {
		result := &res.Results[i]
		localStats := &ctx.Stats[i]
		updateMeasurementStats(localStats, result)
		tableData = append(tableData, getRowValues(result, localStats))
	}
	t, err := pterm.DefaultTable.WithHasHeader().WithData(tableData).Srender()
	if err != nil {
		return err
	}
	ctx.Area.Update(t)

	return nil
}

func getRowValues(res *model.MeasurementResponse, stats *model.MeasurementStats) []string {
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
	return []string{
		getLocationText(res),
		formatValue(fmt.Sprintf("%d", stats.Sent), 4, pterm.FgDefault),
		formatValue(fmt.Sprintf("%.2f", stats.Loss)+"%", 7, pterm.FgDefault),
		formatValue(last, 8, pterm.FgDefault),
		formatValue(min, 8, pterm.FgDefault),
		formatValue(avg, 8, pterm.FgDefault),
		formatValue(max, 8, pterm.FgDefault),
	}
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

func formatValue(v string, width int, color pterm.Color) string {
	for len(v) < width {
		v = " " + v
	}
	return pterm.NewStyle(color).Sprint(v)
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
