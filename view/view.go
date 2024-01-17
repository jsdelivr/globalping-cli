package view

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/mattn/go-runewidth"
	"github.com/pterm/pterm"
)

var (
	// UI styles
	terminalLayoutHighlight = lipgloss.NewStyle().
				Bold(true).Foreground(lipgloss.Color("#17D4A7"))

	terminalLayoutArrow = lipgloss.NewStyle().SetString(">").Bold(true).Foreground(lipgloss.Color("#17D4A7")).PaddingRight(1).String()

	terminalLayoutBold = lipgloss.NewStyle().Bold(true)
)

var apiPollInterval = 500 * time.Millisecond

// Used to trim the output to fit the terminal in live view
func trimOutput(output string, terminalW, terminalH int) string {
	maxW := terminalW - 4 // 4 extra chars to be safe from overflow
	maxH := terminalH - 4 // 4 extra lines to be safe from overflow

	if maxW <= 0 || maxH <= 0 {
		panic("terminal width / height too limited to display results")
	}

	text := strings.ReplaceAll(output, "\t", "  ")

	// Split output into lines
	lines := strings.Split(text, "\n")

	if len(lines) > maxH {
		//  too many lines, trim first lines
		lines = lines[len(lines)-maxH:]
	}

	for i := 0; i < len(lines); i++ {
		rWidth := runewidth.StringWidth(lines[i])
		if rWidth > maxW {
			line := lines[i]
			trimmedLine := string(lines[i][:len(line)-rWidth+maxW])
			lines[i] = trimmedLine
		}
	}

	// Join lines back into a string
	txt := strings.Join(lines, "\n")

	return txt
}

// Generate header that also checks if the probe has a state in it in the form %s, %s, (%s), %s, ASN:%d
func generateHeader(result *model.MeasurementResponse, useStyling bool) string {
	var output strings.Builder

	// Continent + Country + (State) + City + ASN + Network + (Region Tag)
	output.WriteString(getLocationText(result))

	// Check tags to see if there's a region code
	if len(result.Probe.Tags) > 0 {
		for _, tag := range result.Probe.Tags {
			// If tag ends in a number, it's likely a region code and should be displayed
			if _, err := strconv.Atoi(tag[len(tag)-1:]); err == nil {
				output.WriteString(" (" + tag + ")")
				break
			}
		}
	}

	headerWithFormat := formatWithLeadingArrow(output.String(), useStyling)
	return headerWithFormat
}

func formatWithLeadingArrow(text string, useStyling bool) string {
	if useStyling {
		return terminalLayoutArrow + terminalLayoutHighlight.Render(text)
	}
	return "> " + text
}

func LiveView(id string, data *model.GetMeasurement, ctx model.Context, m model.PostMeasurement) {
	var err error

	// Create new writer
	areaPrinter, err := pterm.DefaultArea.Start()
	if err != nil {
		fmt.Printf("failed to start writer: %v\n", err)
		return
	}
	areaPrinter.RemoveWhenDone = true

	defer func() {
		// Stop area printer and clear area if not already done
		err := areaPrinter.Stop()
		if err != nil {
			fmt.Printf("failed to stop writer: %v\n", err)
		}
	}()

	w, h, err := pterm.GetTerminalSize()
	if err != nil {
		fmt.Printf("failed to get terminal size: %v\n", err)
		return
	}

	// String builder for output
	var output strings.Builder

	fetcher := client.NewMeasurementsFetcher(client.ApiUrl)

	// Poll API until the measurement is complete
	for data.Status == "in-progress" {
		time.Sleep(apiPollInterval)
		data, err = fetcher.GetMeasurement(id)
		if err != nil {
			fmt.Printf("failed to get data: %v\n", err)
			return
		}

		// Reset string builder
		output.Reset()

		// Output every result in case of multiple probes
		for i := range data.Results {
			result := &data.Results[i]
			// Output slightly different format if state is available
			output.WriteString(generateHeader(result, !ctx.CI) + "\n")

			if isBodyOnlyHttpGet(ctx, m) {
				output.WriteString(strings.TrimSpace(result.Result.RawBody) + "\n\n")
			} else {
				output.WriteString(strings.TrimSpace(result.Result.RawOutput) + "\n\n")
			}
		}

		areaPrinter.Update(trimOutput(output.String(), w, h))
	}

	// Stop area printer and clear area
	err = areaPrinter.Stop()
	if err != nil {
		fmt.Printf("failed to stop writer: %v\n", err)
	}

	if os.Getenv("LIVE_DEBUG") != "1" {
		PrintStandardResults(id, data, ctx, m)
	}
}

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
		OutputLatency(id, res, *ctx)
		return nil
	}

	if ctx.JsonOutput {
		OutputJson(id, fetcher, *ctx)
		return nil
	}

	// One location view
	if len(res.Results) == 1 {
		if len(ctx.Stats) == 0 {
			// Initialize state
			ctx.Stats = make([]model.MeasurementStats, 1)
			ctx.Packets = client.PacketsMax
			// Print header
			fmt.Println(generateHeader(&res.Results[0], !ctx.CI))
			fmt.Printf("PING %s (%s)\n", res.Target, res.Results[0].Result.ResolvedAddress)
		}
		timings, err := client.DecodePingTimings(res.Results[0].Result.TimingsRaw)
		if err != nil {
			return err
		}
		for i := range timings {
			ctx.Stats[0].Sent++
			t := timings[i]
			fmt.Printf("%s: icmp_seq=%d ttl=%d time=%.2f ms\n",
				res.Results[0].Result.ResolvedAddress,
				ctx.Stats[0].Sent,
				t.TTL,
				t.RTT)
		}
		return nil
	}

	// Multiple location view
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
			fmt.Sprintf("%.2f ms", localStats.Last),
			fmt.Sprintf("%.2f ms", localStats.Avg),
			fmt.Sprintf("%.2f ms", localStats.Min),
			fmt.Sprintf("%.2f ms", localStats.Max),
		})
	}
	t, err := pterm.DefaultTable.WithHasHeader().WithData(tableData).Srender()
	if err != nil {
		return err
	}
	ctx.Area.Update(t)
	return err
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
	localStats.Lost += stats.Drop
	if stats.Min < localStats.Min {
		localStats.Min = stats.Min
	}
	if stats.Max > localStats.Max {
		localStats.Max = stats.Max
	}
	localStats.Avg = (localStats.Avg*float64(localStats.Sent) + stats.Avg*float64(stats.Total)) / float64(localStats.Sent+stats.Total)
	localStats.Sent += stats.Total
	if len(timings) != 0 {
		localStats.Last = timings[len(timings)-1].RTT
	}
	localStats.Loss = float64(localStats.Lost) / float64(localStats.Sent) * 100
	return nil
}

// If json flag is used, only output json
// TODO: Return errors instead of printing them
func OutputJson(id string, fetcher client.MeasurementsFetcher, ctx model.Context) {
	output, err := fetcher.GetRawMeasurement(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(output))

	if ctx.Share {
		fmt.Fprintln(os.Stderr, formatWithLeadingArrow(shareMessage(id), !ctx.CI))
	}
	fmt.Println()
}

// Prints non-json non-latency results to the screen
func PrintStandardResults(id string, data *model.GetMeasurement, ctx model.Context, m model.PostMeasurement) {
	for i := range data.Results {
		result := &data.Results[i]
		if i > 0 {
			// new line as separator if more than 1 result
			fmt.Println()
		}

		// Output slightly different format if state is available
		fmt.Fprintln(os.Stderr, generateHeader(result, !ctx.CI))

		if isBodyOnlyHttpGet(ctx, m) {
			fmt.Println(strings.TrimSpace(result.Result.RawBody))
		} else {
			fmt.Println(strings.TrimSpace(result.Result.RawOutput))
		}
	}

	if ctx.Share {
		fmt.Fprintln(os.Stderr, formatWithLeadingArrow(shareMessage(id), !ctx.CI))
	}
}

func isBodyOnlyHttpGet(ctx model.Context, m model.PostMeasurement) bool {
	return ctx.Cmd == "http" && m.Options != nil && m.Options.Request != nil && m.Options.Request.Method == "GET" && !ctx.Full
}

// TODO: Return errors instead of printing them
func OutputResults(id string, ctx model.Context, m model.PostMeasurement) {
	fetcher := client.NewMeasurementsFetcher(client.ApiUrl)

	// Wait for first result to arrive from a probe before starting display (can be in-progress)
	data, err := fetcher.GetMeasurement(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Probe may not have started yet
	for len(data.Results) == 0 {
		time.Sleep(apiPollInterval)
		data, err = fetcher.GetMeasurement(id)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	if ctx.CI || ctx.JsonOutput || ctx.Latency {
		// Poll API until the measurement is complete
		for data.Status == "in-progress" {
			time.Sleep(apiPollInterval)
			data, err = fetcher.GetMeasurement(id)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		if ctx.Latency {
			OutputLatency(id, data, ctx)
			return
		}

		if ctx.JsonOutput {
			OutputJson(id, fetcher, ctx)
			return
		}

		if ctx.CI {
			PrintStandardResults(id, data, ctx, m)
			return
		}

		panic(fmt.Sprintf("case not handled. %+v", ctx))
	}

	LiveView(id, data, ctx, m)
}

func shareMessage(id string) string {
	return fmt.Sprintf("View the results online: https://www.jsdelivr.com/globalping?measurement=%s", id)
}

func getLocationText(m *model.MeasurementResponse) string {
	state := ""
	if m.Probe.State != "" {
		state = "(" + m.Probe.State + "), "
	}
	return m.Probe.Continent +
		", " + m.Probe.Country +
		", " + state + m.Probe.City +
		", ASN:" + fmt.Sprint(m.Probe.ASN) +
		", " + m.Probe.Network
}
