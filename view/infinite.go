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
		fmt.Printf("PING %s (%s)\n", res.Target, measurement.Result.ResolvedAddress)
	}
	timings, err := client.DecodePingTimings(measurement.Result.TimingsRaw)
	if err != nil {
		return err
	}
	for i := range timings {
		ctx.Stats[0].Sent++
		t := timings[i]
		fmt.Printf("%s: icmp_seq=%d ttl=%d time=%.2f ms\n",
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
			ctx.Stats[i].Min = math.MaxFloat64
		}
		// Create new writer
		ctx.Area, err = pterm.DefaultArea.Start()
		if err != nil {
			return errors.New("failed to start writer: " + err.Error())
		}
	}
	tableData := pterm.TableData{
		{"Location", "Loss", "Sent", "Last", "Avg", "Min", "Max"},
	}
	for i := range res.Results {
		result := &res.Results[i]
		localStats := &ctx.Stats[i]
		updateMeasurementStats(localStats, result)
		tableData = append(tableData, []string{
			getLocationText(result),
			fmt.Sprintf("%.2f", localStats.Loss) + "%",
			fmt.Sprintf("%d", localStats.Sent),
			formatDuration(localStats.Last),
			formatDuration(localStats.Avg),
			formatDuration(localStats.Min),
			formatDuration(localStats.Max),
		})
	}
	t, err := pterm.DefaultTable.WithHasHeader().WithData(tableData).Srender()
	if err != nil {
		return err
	}
	ctx.Area.Update(t)

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

func updateMeasurementStats(localStats *model.MeasurementStats, result *model.MeasurementResponse) error {
	stats, err := client.DecodePingStats(result.Result.StatsRaw)
	if err != nil {
		return err
	}
	timings, err := client.DecodePingTimings(result.Result.TimingsRaw)
	if err != nil {
		return err
	}
	if stats.Min < localStats.Min && stats.Min != 0 {
		localStats.Min = stats.Min
	}
	if stats.Max > localStats.Max {
		localStats.Max = stats.Max
	}
	if stats.Avg != 0 {
		localStats.Avg = (localStats.Avg*float64(localStats.Sent) + stats.Avg*float64(stats.Total)) / float64(localStats.Sent+stats.Total)
	}
	if len(timings) != 0 {
		localStats.Last = timings[len(timings)-1].RTT
	}
	localStats.Sent += stats.Total
	localStats.Lost += stats.Drop
	localStats.Loss = float64(localStats.Lost) / float64(localStats.Sent) * 100
	return nil
}
