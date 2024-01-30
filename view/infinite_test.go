package view

import (
	"io"
	"math"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jsdelivr/globalping-cli/mocks"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
)

func TestOutputSingleLocationInProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	osStdOut := os.Stdout
	defer func() {
		os.Stdout = osStdOut
	}()

	rawOutput1 := `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.`
	rawOutput2 := `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=60 time=17.6 ms`
	rawOutput3 := `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=60 time=17.6 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=2 ttl=30 time=17.3 ms`
	rawOutput4 := `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=60 time=17.6 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=2 ttl=30 time=17.3 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=3 ttl=10 time=17.0 ms

---  ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 2002ms
rtt min/avg/max/mdev = 17.006/17.333/17.648/0.321 ms`

	fetcher := mocks.NewMockMeasurementsFetcher(ctrl)
	measurement := getPingGetMeasurement(measurementID1)

	callCount := 1 // 1st call is done in the caller.
	fetcher.EXPECT().GetMeasurement(measurementID1).DoAndReturn(func(id string) (*model.GetMeasurement, error) {
		callCount++
		switch callCount {
		case 2:
			measurement.Results[0].Result.RawOutput = rawOutput2
		case 3:
			measurement.Results[0].Result.RawOutput = rawOutput3
		case 4:
			measurement.Status = model.StatusFinished
			measurement.Results[0].Result.Status = model.StatusFinished
			measurement.Results[0].Result.RawOutput = rawOutput4
		}
		return measurement, nil
	}).Times(3)

	ctx := &model.Context{
		Cmd:            "ping",
		APIMinInterval: 0,
	}

	measurement.Status = model.StatusInProgress
	measurement.Results[0].Result.Status = model.StatusInProgress
	measurement.Results[0].Result.RawOutput = rawOutput1

	r, w, err := os.Pipe()
	assert.NoError(t, err)
	defer func() {
		w.Close()
		r.Close()
	}()
	os.Stdout = w

	err = outputSingleLocation(fetcher, measurement, ctx)
	w.Close()
	os.Stdout = osStdOut

	assert.NoError(t, err)
	output, err := io.ReadAll(r)
	r.Close()
	assert.NoError(t, err)
	assert.Equal(t,
		`> EU, DE, Berlin, ASN:3320, Deutsche Telekom AG
PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=60 time=17.6 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=2 ttl=30 time=17.3 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=3 ttl=10 time=17.0 ms
`,
		string(output),
	)

	assert.Equal(t,
		[]model.MeasurementStats{{Sent: 3, Rcv: 3, Lost: 0, Loss: 0, Last: 17, Min: 17.006, Avg: 17.333, Max: 17.648, Time: 2002}},
		ctx.CompletedStats,
	)
}

func TestOutputSingleLocationMultipleCalls(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	osStdOut := os.Stdout
	defer func() {
		os.Stdout = osStdOut
	}()

	fetcher := mocks.NewMockMeasurementsFetcher(ctrl)
	measurement := getPingGetMeasurement(measurementID1)
	fetcher.EXPECT().GetMeasurement(measurementID1).Times(0).Return(measurement, nil)

	ctx := &model.Context{
		Cmd:        "ping",
		MaxHistory: 3,
	}

	r, w, err := os.Pipe()
	assert.NoError(t, err)
	defer func() {
		w.Close()
		r.Close()
	}()
	os.Stdout = w

	err = outputSingleLocation(fetcher, measurement, ctx)
	assert.NoError(t, err)
	err = outputSingleLocation(fetcher, measurement, ctx)
	assert.NoError(t, err)
	err = outputSingleLocation(fetcher, measurement, ctx)
	assert.NoError(t, err)
	w.Close()
	os.Stdout = osStdOut

	output, err := io.ReadAll(r)
	r.Close()
	assert.NoError(t, err)
	assert.Equal(t,
		`> EU, DE, Berlin, ASN:3320, Deutsche Telekom AG
PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=60 time=17.6 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=2 ttl=60 time=17.6 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=3 ttl=60 time=17.6 ms
`,
		string(output))

	expectedStats := []model.MeasurementStats{{Sent: 3, Rcv: 3, Lost: 0, Loss: 0, Last: 17.6, Min: 17.639, Avg: 17.639, Max: 17.639, Time: 3000}}
	assert.Equal(t, expectedStats, ctx.InProgressStats)
	assert.Equal(t, expectedStats, ctx.CompletedStats)
}

func TestOutputMultipleLocationsInProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	osStdOut := os.Stdout
	defer func() {
		os.Stdout = osStdOut
	}()

	ctx := &model.Context{
		Cmd:            "ping",
		APIMinInterval: 0,
	}
	fetcher := mocks.NewMockMeasurementsFetcher(ctrl)
	res := getPingGetMeasurementMultipleLocations(measurementID1)

	rawOutput1 := `PING  (146.75.73.229) 56(84) bytes of data.`
	rawOutput2 := `PING  (146.75.73.229) 56(84) bytes of data.
64 bytes from 146.75.73.229 (146.75.73.229): icmp_seq=1 ttl=52 time=17.6 ms
no answer yet for icmp_seq=2`
	rawOutputFinal := `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=60 time=17.6 ms
no answer yet for icmp_seq=2
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=2 ttl=30 time=17.3 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=3 ttl=10 time=17.0 ms

---  ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 2002ms
rtt min/avg/max/mdev = 17.006/17.333/17.648/0.321 ms`
	expectedCtx := getDefaultPingCtx(len(res.Results))
	expectedTables := [6]*string{}

	callCount := 1 // 1st call is done in the caller.
	fetcher.EXPECT().GetMeasurement(measurementID1).DoAndReturn(func(id string) (*model.GetMeasurement, error) {
		callCount++
		switch callCount {
		case 2, 5:
			res.Results[0].Result.RawOutput = rawOutput2
			expectedTables[callCount-1], _ = generateTable(res, expectedCtx, 76)
		case 3, 6:
			res.Status = model.StatusFinished
			res.Results[0].Result.Status = model.StatusFinished
			res.Results[0].Result.RawOutput = rawOutputFinal
			expectedTables[callCount-1], _ = generateTable(res, expectedCtx, 76)
		}
		return res, nil
	}).Times(4)

	// 1st call
	res.Status = model.StatusInProgress
	res.Results[0].Result.Status = model.StatusInProgress
	res.Results[0].Result.RawOutput = rawOutput1
	expectedTables[0], _ = generateTable(res, expectedCtx, 76)

	r, w, err := os.Pipe()
	assert.NoError(t, err)
	defer func() {
		w.Close()
		r.Close()
	}()
	os.Stdout = w

	err = outputMultipleLocations(fetcher, res, ctx)
	assert.NoError(t, err)

	firstCallStats := []model.MeasurementStats{
		{Sent: 3, Rcv: 3, Lost: 0, Loss: 0, Last: 17, Min: 17.006, Avg: 17.333, Max: 17.648, Time: 2002},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.457, Avg: 5.457, Max: 5.457, Time: 2},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.069, Avg: 4.069, Max: 4.069, Time: 3},
	}
	assert.Equal(t, firstCallStats, ctx.InProgressStats)
	assert.Equal(t, firstCallStats, ctx.CompletedStats)

	// 2nd call
	res.Status = model.StatusInProgress
	res.Results[0].Result.Status = model.StatusInProgress
	res.Results[0].Result.RawOutput = rawOutput1
	expectedCtx.CompletedStats = firstCallStats

	callCount++
	expectedTables[3], _ = generateTable(res, expectedCtx, 76)
	err = outputMultipleLocations(fetcher, res, ctx)
	assert.NoError(t, err)
	w.Close()

	os.Stdout = osStdOut
	output, err := io.ReadAll(r)
	assert.NoError(t, err)

	secondCallStats := []model.MeasurementStats{
		{Sent: 6, Rcv: 6, Lost: 0, Loss: 0, Last: 17, Min: 17.006, Avg: 17.333, Max: 17.648, Time: 4004},
		{Sent: 2, Rcv: 2, Lost: 0, Loss: 0, Last: 5.46, Min: 5.457, Avg: 5.457, Max: 5.457, Time: 4},
		{Sent: 2, Rcv: 2, Lost: 0, Loss: 0, Last: 4.07, Min: 4.069, Avg: 4.069, Max: 4.069, Time: 6},
	}
	assert.Equal(t, secondCallStats, ctx.InProgressStats)
	assert.Equal(t, secondCallStats, ctx.CompletedStats)

	rr, ww, err := os.Pipe()
	assert.NoError(t, err)
	defer func() {
		ww.Close()
		rr.Close()
	}()

	os.Stdout = ww
	area, _ := pterm.DefaultArea.Start()
	for i := range expectedTables {
		area.Update(*expectedTables[i])
	}
	area.Stop()
	ww.Close()
	os.Stdout = osStdOut

	expectedOutput, err := io.ReadAll(rr)
	assert.NoError(t, err)
	rr.Close()

	assert.Equal(t, string(expectedOutput), string(output))
}

func TestOutputMultipleLocations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	osStdOut := os.Stdout
	defer func() {
		os.Stdout = osStdOut
	}()

	ctx := &model.Context{
		Cmd: "ping",
	}
	measurement := getPingGetMeasurementMultipleLocations(measurementID1)
	fetcher := mocks.NewMockMeasurementsFetcher(ctrl)
	fetcher.EXPECT().GetMeasurement(measurementID1).Times(0).Return(measurement, nil)

	r, w, err := os.Pipe()
	assert.NoError(t, err)
	defer func() {
		w.Close()
		r.Close()
	}()
	os.Stdout = w

	err = outputMultipleLocations(fetcher, measurement, ctx)
	assert.NoError(t, err)
	w.Close()

	os.Stdout = osStdOut

	output, err := io.ReadAll(r)
	assert.NoError(t, err)
	r.Close()

	rr, ww, err := os.Pipe()
	assert.NoError(t, err)
	defer func() {
		ww.Close()
		rr.Close()
	}()
	os.Stdout = ww

	expectedCtx := getDefaultPingCtx(len(measurement.Results))
	expectedTable, _ := generateTable(measurement, expectedCtx, 76) // 80 - 4. pterm defaults to 80 when terminal size is not detected.
	area, _ := pterm.DefaultArea.Start()
	area.Update(*expectedTable)
	area.Stop()
	ww.Close()
	os.Stdout = osStdOut

	expectedOutput, err := io.ReadAll(rr)
	assert.NoError(t, err)
	rr.Close()

	assert.Equal(t, string(expectedOutput), string(output))
	assert.Equal(t,
		[]model.MeasurementStats{
			{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 1},
			{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.457, Avg: 5.457, Max: 5.457, Time: 2},
			{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.069, Avg: 4.069, Max: 4.069, Time: 3},
		},
		ctx.CompletedStats,
	)
}

func TestOutputSummary(t *testing.T) {

	t.Run("No_stats", func(t *testing.T) {
		osStdOut := os.Stdout
		defer func() {
			os.Stdout = osStdOut
		}()

		r, w, err := os.Pipe()
		assert.NoError(t, err)
		defer func() {
			w.Close()
			r.Close()
		}()

		ctx := &model.Context{}
		os.Stdout = w
		OutputSummary(ctx)
		w.Close()
		os.Stdout = osStdOut

		output, err := io.ReadAll(r)
		assert.NoError(t, err)
		r.Close()
		assert.Equal(t, "", string(output))
	})

	t.Run("With_stats_Single_location", func(t *testing.T) {
		osStdOut := os.Stdout
		defer func() {
			os.Stdout = osStdOut
		}()

		r, w, err := os.Pipe()
		assert.NoError(t, err)
		defer func() {
			w.Close()
			r.Close()
		}()

		ctx := &model.Context{
			InProgressStats: []model.MeasurementStats{
				{Sent: 10, Rcv: 9, Lost: 1, Loss: 10, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 1000},
			},
		}
		os.Stdout = w
		OutputSummary(ctx)
		w.Close()
		os.Stdout = osStdOut

		output, err := io.ReadAll(r)
		assert.NoError(t, err)
		r.Close()
		assert.Equal(t, `
---  ping statistics ---
10 packets transmitted, 9 received, 10.00% packet loss, time 1000ms
rtt min/avg/max = 0.770/0.770/0.770 ms
`,
			string(output))
	})

	t.Run("With_stats_In_progress", func(t *testing.T) {
		osStdOut := os.Stdout
		defer func() {
			os.Stdout = osStdOut
		}()

		r, w, err := os.Pipe()
		assert.NoError(t, err)
		defer func() {
			w.Close()
			r.Close()
		}()

		ctx := &model.Context{
			InProgressStats: []model.MeasurementStats{
				{Sent: 1, Rcv: 0, Lost: 1, Loss: 100, Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1, Time: 0},
			},
		}
		os.Stdout = w
		OutputSummary(ctx)
		w.Close()
		os.Stdout = osStdOut

		output, err := io.ReadAll(r)
		assert.NoError(t, err)
		r.Close()
		assert.Equal(t, `
---  ping statistics ---
1 packets transmitted, 0 received, 100.00% packet loss, time 0ms
rtt min/avg/max = -/-/- ms
`,
			string(output))
	})

	t.Run("Multiple_locations", func(t *testing.T) {
		osStdOut := os.Stdout
		defer func() {
			os.Stdout = osStdOut
		}()

		r, w, err := os.Pipe()
		assert.NoError(t, err)
		defer func() {
			w.Close()
			r.Close()
		}()

		ctx := &model.Context{
			InProgressStats: []model.MeasurementStats{
				model.NewMeasurementStats(),
				model.NewMeasurementStats(),
			},
		}
		os.Stdout = w
		OutputSummary(ctx)
		w.Close()
		os.Stdout = osStdOut

		output, err := io.ReadAll(r)
		assert.NoError(t, err)
		r.Close()
		assert.Equal(t, "", string(output))
	})

	t.Run("Single_location_Share", func(t *testing.T) {
		osStdOut := os.Stdout
		defer func() {
			os.Stdout = osStdOut
		}()

		r, w, err := os.Pipe()
		assert.NoError(t, err)
		defer func() {
			w.Close()
			r.Close()
		}()

		ctx := &model.Context{
			History: &model.Rbuffer{
				Index: 0,
				Slice: []string{measurementID1},
			},
			InProgressStats: []model.MeasurementStats{
				{Sent: 1, Rcv: 0, Lost: 1, Loss: 100, Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1, Time: 0},
			},
			Share: true,
		}
		os.Stdout = w
		OutputSummary(ctx)
		w.Close()
		os.Stdout = osStdOut

		output, err := io.ReadAll(r)
		assert.NoError(t, err)
		r.Close()

		expectedOutput := `
---  ping statistics ---
1 packets transmitted, 0 received, 100.00% packet loss, time 0ms
rtt min/avg/max = -/-/- ms
` + formatWithLeadingArrow(shareMessage(measurementID1), true) + "\n"

		assert.Equal(t, expectedOutput, string(output))
	})

	t.Run("Multiple_locations_Share", func(t *testing.T) {
		osStdOut := os.Stdout
		defer func() {
			os.Stdout = osStdOut
		}()

		r, w, err := os.Pipe()
		assert.NoError(t, err)
		defer func() {
			w.Close()
			r.Close()
		}()

		ctx := &model.Context{
			History: &model.Rbuffer{
				Index: 0,
				Slice: []string{measurementID1, measurementID2},
			},
			InProgressStats: []model.MeasurementStats{
				model.NewMeasurementStats(),
				model.NewMeasurementStats(),
			},
			Share: true,
		}
		os.Stdout = w
		OutputSummary(ctx)
		w.Close()
		os.Stdout = osStdOut

		output, err := io.ReadAll(r)
		assert.NoError(t, err)
		r.Close()

		expectedOutput := "\n" + formatWithLeadingArrow(shareMessage(measurementID1+"+"+measurementID2), true) + "\n"

		assert.Equal(t, expectedOutput, string(output))
	})

	t.Run("Multiple_locations_Share_More_calls_than_MaxHistory", func(t *testing.T) {
		osStdOut := os.Stdout
		defer func() {
			os.Stdout = osStdOut
		}()

		r, w, err := os.Pipe()
		assert.NoError(t, err)
		defer func() {
			w.Close()
			r.Close()
		}()

		ctx := &model.Context{
			History: &model.Rbuffer{
				Index: 0,
				Slice: []string{measurementID2},
			},
			InProgressStats: []model.MeasurementStats{
				model.NewMeasurementStats(),
				model.NewMeasurementStats(),
			},
			Share:      true,
			CallCount:  2,
			MaxHistory: 1,
			Packets:    16,
		}
		os.Stdout = w
		OutputSummary(ctx)
		w.Close()
		os.Stdout = osStdOut

		output, err := io.ReadAll(r)
		assert.NoError(t, err)
		r.Close()

		expectedOutput := "\n" + formatWithLeadingArrow(shareMessage(measurementID2), true) +
			"\nFor long-running continuous mode measurements, only the last 16 packets are shared.\n"

		assert.Equal(t, expectedOutput, string(output))
	})
}

func TestFormatDuration(t *testing.T) {
	d := formatDuration(1.2345)
	assert.Equal(t, "1.23 ms", d)
	d = formatDuration(12.345)
	assert.Equal(t, "12.3 ms", d)
	d = formatDuration(123.4567)
	assert.Equal(t, "123 ms", d)
}

func TestGenerateTableFull(t *testing.T) {
	measurement := getPingGetMeasurementMultipleLocations(measurementID1)
	ctx := getDefaultPingCtx(len(measurement.Results))
	expectedTable := "\x1b[96m\x1b[96mLocation                                       \x1b[0m\x1b[0m | \x1b[96m\x1b[96mSent\x1b[0m\x1b[0m | \x1b[96m\x1b[96m   Loss\x1b[0m\x1b[0m | \x1b[96m\x1b[96m    Last\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Min\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Avg\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Max\x1b[0m\x1b[0m\n" +
		"EU, GB, London, ASN:0, OVH SAS                  |    1 |   0.00% |  0.77 ms |  0.77 ms |  0.77 ms |  0.77 ms\n" +
		"EU, DE, Falkenstein, ASN:0, Hetzner Online GmbH |    1 |   0.00% |  5.46 ms |  5.46 ms |  5.46 ms |  5.46 ms\n" +
		"EU, DE, Nuremberg, ASN:0, Hetzner Online GmbH   |    1 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms\n"
	table, stats := generateTable(measurement, ctx, 500)
	assert.Equal(t, expectedTable, *table)
	assert.Equal(t, []model.MeasurementStats{
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 1},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.457, Avg: 5.457, Max: 5.457, Time: 2},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.069, Avg: 4.069, Max: 4.069, Time: 3},
	}, stats)
}

func TestGenerateTableOneRowTruncated(t *testing.T) {
	measurement := getPingGetMeasurementMultipleLocations(measurementID1)
	measurement.Results[1].Probe.Network = "作者聚集的原创内容平台于201 1年1月正式上线让人们更"
	ctx := getDefaultPingCtx(len(measurement.Results))
	expectedTable := "\x1b[96m\x1b[96mLocation                                      \x1b[0m\x1b[0m | \x1b[96m\x1b[96mSent\x1b[0m\x1b[0m | \x1b[96m\x1b[96m   Loss\x1b[0m\x1b[0m | \x1b[96m\x1b[96m    Last\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Min\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Avg\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Max\x1b[0m\x1b[0m\n" +
		"EU, GB, London, ASN:0, OVH SAS                 |    1 |   0.00% |  0.77 ms |  0.77 ms |  0.77 ms |  0.77 ms\n" +
		"EU, DE, Falkenstein, ASN:0, 作者聚集的原创...  |    1 |   0.00% |  5.46 ms |  5.46 ms |  5.46 ms |  5.46 ms\n" +
		"EU, DE, Nuremberg, ASN:0, Hetzner Online GmbH  |    1 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms\n"
	table, stats := generateTable(measurement, ctx, 106)
	assert.Equal(t, expectedTable, *table)
	assert.Equal(t, []model.MeasurementStats{
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 1},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.457, Avg: 5.457, Max: 5.457, Time: 2},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.069, Avg: 4.069, Max: 4.069, Time: 3},
	}, stats)
}

func TestGenerateTableMultiLineTruncated(t *testing.T) {
	measurement := getPingGetMeasurementMultipleLocations(measurementID1)
	measurement.Results[1].Probe.Network = "Hetzner Online GmbH\nLorem ipsum\nLorem ipsum dolor sit amet"
	ctx := getDefaultPingCtx(len(measurement.Results))
	expectedTable := "\x1b[96m\x1b[96mLocation                                      \x1b[0m\x1b[0m | \x1b[96m\x1b[96mSent\x1b[0m\x1b[0m | \x1b[96m\x1b[96m   Loss\x1b[0m\x1b[0m | \x1b[96m\x1b[96m    Last\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Min\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Avg\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Max\x1b[0m\x1b[0m\n" +
		"EU, GB, London, ASN:0, OVH SAS                 |    1 |   0.00% |  0.77 ms |  0.77 ms |  0.77 ms |  0.77 ms\n" +
		"EU, DE, Falkenstein, ASN:0, Hetzner Online ... |    1 |   0.00% |  5.46 ms |  5.46 ms |  5.46 ms |  5.46 ms\n" +
		"Lorem ipsum                                    |      |         |          |          |          |         \n" +
		"Lorem ipsum dolor sit amet                     |      |         |          |          |          |         \n" +
		"EU, DE, Nuremberg, ASN:0, Hetzner Online GmbH  |    1 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms\n"
	table, stats := generateTable(measurement, ctx, 106)
	assert.Equal(t, expectedTable, *table)
	assert.Equal(t, []model.MeasurementStats{
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 1},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.457, Avg: 5.457, Max: 5.457, Time: 2},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.069, Avg: 4.069, Max: 4.069, Time: 3},
	}, stats)
}

func TestGenerateTableMaxTruncated(t *testing.T) {
	measurement := getPingGetMeasurementMultipleLocations(measurementID1)
	ctx := getDefaultPingCtx(len(measurement.Results))
	expectedTable := "\x1b[96m\x1b[96mLoc...\x1b[0m\x1b[0m | \x1b[96m\x1b[96mSent\x1b[0m\x1b[0m | \x1b[96m\x1b[96m   Loss\x1b[0m\x1b[0m | \x1b[96m\x1b[96m    Last\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Min\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Avg\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Max\x1b[0m\x1b[0m\n" +
		"EU,... |    1 |   0.00% |  0.77 ms |  0.77 ms |  0.77 ms |  0.77 ms\n" +
		"EU,... |    1 |   0.00% |  5.46 ms |  5.46 ms |  5.46 ms |  5.46 ms\n" +
		"EU,... |    1 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms\n"
	table, stats := generateTable(measurement, ctx, 0)
	assert.Equal(t, expectedTable, *table)
	assert.Equal(t, []model.MeasurementStats{
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 1},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.457, Avg: 5.457, Max: 5.457, Time: 2},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.069, Avg: 4.069, Max: 4.069, Time: 3},
	}, stats)
}

func TestMergeMeasurementStats(t *testing.T) {
	result := model.MeasurementResponse{
		Result: model.ResultData{
			RawOutput: `PING  (142.250.65.174) 56(84) bytes of data.`,
		},
	}
	newStats := mergeMeasurementStats(
		model.MeasurementStats{Sent: 0, Lost: 0, Loss: 0, Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1},
		&result,
	)
	assert.Equal(t,
		model.MeasurementStats{Sent: 0, Lost: 0, Loss: 0, Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1},
		newStats,
	)
	result = model.MeasurementResponse{
		Result: model.ResultData{
			RawOutput: `PING  (142.250.65.174) 56(84) bytes of data.
no answer yet for icmp_seq=1
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=1 ttl=59 time=1.06 ms
no answer yet for icmp_seq=2
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=2 ttl=59 time=1.10 ms
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=3 ttl=59 time=1.11 ms
no answer yet for icmp_seq=4`,
		},
	}
	newStats = mergeMeasurementStats(
		model.MeasurementStats{Sent: 0, Lost: 0, Loss: 0, Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1},
		&result)
	assert.Equal(t,
		model.MeasurementStats{Sent: 4, Rcv: 3, Lost: 1, Loss: 25, Last: 1.11, Min: 1.06, Avg: 1.09, Max: 1.11},
		newStats,
	)
	result = model.MeasurementResponse{
		Result: model.ResultData{
			RawOutput: `PING  (142.250.65.174) 56(84) bytes of data.
no answer yet for icmp_seq=1
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=1 ttl=59 time=1.06 ms
no answer yet for icmp_seq=2
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=2 ttl=59 time=1.10 ms
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=3 ttl=59 time=1.11 ms
no answer yet for icmp_seq=4

---  ping statistics ---
4 packets transmitted, 4 received, 0% packet loss, time 1002ms
rtt min/avg/max/mdev = 1.061/1.090/1.108/0.020 ms`,
		},
	}
	newStats = mergeMeasurementStats(
		model.MeasurementStats{Sent: 0, Lost: 0, Loss: 0, Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1},
		&result)
	assert.Equal(t,
		model.MeasurementStats{Sent: 4, Rcv: 4, Lost: 0, Loss: 0, Last: 1.11, Min: 1.061, Avg: 1.09, Max: 1.108, Time: 1002},
		newStats,
	)
}

func TestGetRowValuesNoPacketsRcv(t *testing.T) {
	stats := model.MeasurementStats{Sent: 1, Lost: 0, Loss: 0, Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1}
	rowValues := getRowValues(&stats)
	assert.Equal(t, [7]string{
		"",
		"1",
		"0.00%",
		"-",
		"-",
		"-",
		"-",
	},
		rowValues)
}

func TestGetRowValues(t *testing.T) {
	stats := model.MeasurementStats{
		Sent: 100,
		Lost: 10,
		Loss: 10,
		Last: 12.345,
		Min:  1.2345,
		Avg:  8.3456,
		Max:  123.4567,
	}
	rowValues := getRowValues(&stats)
	assert.Equal(t, [7]string{
		"",
		"100",
		"10.00%",
		"12.3 ms",
		"1.23 ms",
		"8.35 ms",
		"123 ms",
	},
		rowValues)
}

func TestParsePingRawOutputFull(t *testing.T) {
	m := &model.MeasurementResponse{
		Result: model.ResultData{
			RawOutput: `PING cdn.jsdelivr.net (142.250.65.174) 56(84) bytes of data.
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=1 ttl=59 time=1.06 ms
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=2 ttl=59 time=1.10 ms
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=3 ttl=59 time=1.11 ms

---  ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 1002ms
rtt min/avg/max/mdev = 1.061/1.090/1.108/0.020 ms`,
		},
	}
	res := parsePingRawOutput(m, -1)
	assert.Equal(t, &ParsedPingOutput{
		Hostname:    "cdn.jsdelivr.net",
		Address:     "142.250.65.174",
		BytesOfData: "56(84)",
		Timings: []model.PingTiming{
			{RTT: 1.06, TTL: 59},
			{RTT: 1.10, TTL: 59},
			{RTT: 1.11, TTL: 59},
		},
		Stats: &model.PingStats{
			Min: 1.061, Avg: 1.090, Max: 1.108, Total: 3, Rcv: 3, Drop: 0, Loss: 0, Mdev: 0.020,
		},
		Time: 1002,
	}, res)
}

func TestParsePingRawOutputNoStats(t *testing.T) {
	m := &model.MeasurementResponse{
		Result: model.ResultData{
			RawOutput: `PING  (142.250.65.174) 56(84) bytes of data.
no answer yet for icmp_seq=1
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=1 ttl=59 time=1.06 ms
no answer yet for icmp_seq=2
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=2 ttl=59 time=1.10 ms
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=3 ttl=59 time=1.11 ms
no answer yet for icmp_seq=4`,
		},
	}
	res := parsePingRawOutput(m, -1)
	assert.Equal(t, &ParsedPingOutput{
		Address:     "142.250.65.174",
		BytesOfData: "56(84)",
		Timings: []model.PingTiming{
			{RTT: 1.06, TTL: 59},
			{RTT: 1.10, TTL: 59},
			{RTT: 1.11, TTL: 59},
		},
		Stats: &model.PingStats{
			Min: 1.06, Avg: 1.09, Max: 1.11, Total: 4, Rcv: 3, Drop: 1, Loss: 25,
		},
	}, res)
}

func TestParsePingRawOutputNoStatsWithStartIncmpSeq(t *testing.T) {
	m := &model.MeasurementResponse{
		Result: model.ResultData{
			RawOutput: `PING  (142.250.65.174) 56(84) bytes of data.
no answer yet for icmp_seq=1
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=1 ttl=59 time=1.06 ms
no answer yet for icmp_seq=2
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=2 ttl=59 time=1.10 ms
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=3 ttl=59 time=1.11 ms
no answer yet for icmp_seq=4`,
		},
	}
	res := parsePingRawOutput(m, 4)
	assert.Equal(t, &ParsedPingOutput{
		Address:     "142.250.65.174",
		BytesOfData: "56(84)",
		RawPacketLines: []string{
			"no answer yet for icmp_seq=5",
			"64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=5 ttl=59 time=1.06 ms",
			"no answer yet for icmp_seq=6",
			"64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=6 ttl=59 time=1.10 ms",
			"64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=7 ttl=59 time=1.11 ms",
			"no answer yet for icmp_seq=8",
		},
		Timings: []model.PingTiming{
			{RTT: 1.06, TTL: 59},
			{RTT: 1.10, TTL: 59},
			{RTT: 1.11, TTL: 59},
		},
		Stats: &model.PingStats{
			Min: 1.06, Avg: 1.09, Max: 1.11, Total: 4, Rcv: 3, Drop: 1, Loss: 25,
		},
	}, res)
}
