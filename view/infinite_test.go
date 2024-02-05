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

func TestStreamingPacketsInProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	osStdOut := os.Stdout
	defer func() {
		os.Stdout = osStdOut
	}()

	rawOutput1 := `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.`
	rawOutput2 := `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=56 time=12.9 ms`
	rawOutput3 := `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=56 time=12.9 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=2 ttl=56 time=12.7 ms`
	rawOutput4 := `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=56 time=12.9 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=2 ttl=56 time=12.7 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=3 ttl=56 time=13.0 ms

---  ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 1001ms
rtt min/avg/max/mdev = 12.711/12.854/12.952/0.103 ms`

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

	err = outputStreamingPackets(fetcher, measurement, ctx)
	w.Close()
	os.Stdout = osStdOut

	assert.NoError(t, err)
	output, err := io.ReadAll(r)
	r.Close()
	assert.NoError(t, err)
	assert.Equal(t,
		`> EU, DE, Berlin, ASN:3320, Deutsche Telekom AG
PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=56 time=12.9 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=2 ttl=56 time=12.7 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=3 ttl=56 time=13.0 ms
`,
		string(output),
	)

	expectedStats := []model.MeasurementStats{{Sent: 3, Rcv: 3, Lost: 0, Loss: 0, Last: 13, Min: 12.7,
		Avg: 12.8666, Max: 13, Time: 1001, Tsum: 38.6, Tsum2: 496.7, Mdev: 0.1247}}
	assertMeasurementStats(t, &expectedStats[0], &ctx.InProgressStats[0])
	assertMeasurementStats(t, &expectedStats[0], &ctx.CompletedStats[0])
}

func TestStreamingPacketsMultipleCalls(t *testing.T) {
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

	err = outputStreamingPackets(fetcher, measurement, ctx)
	assert.NoError(t, err)
	err = outputStreamingPackets(fetcher, measurement, ctx)
	assert.NoError(t, err)
	err = outputStreamingPackets(fetcher, measurement, ctx)
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

	expectedStats := []model.MeasurementStats{{Sent: 3, Rcv: 3, Lost: 0, Loss: 0, Last: 17.6, Min: 17.6,
		Avg: 17.6, Max: 17.6, Time: 3000, Tsum: 52.8, Tsum2: 929.28, Mdev: 0}}
	assertMeasurementStats(t, &expectedStats[0], &ctx.InProgressStats[0])
	assertMeasurementStats(t, &expectedStats[0], &ctx.CompletedStats[0])
}

func TestOutputTableViewMultipleCalls(t *testing.T) {
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
			expectedTables[callCount-1], _ = generateTable(res, expectedCtx, 78)
		case 3, 6:
			res.Status = model.StatusFinished
			res.Results[0].Result.Status = model.StatusFinished
			res.Results[0].Result.RawOutput = rawOutputFinal
			expectedTables[callCount-1], _ = generateTable(res, expectedCtx, 78)
		}
		return res, nil
	}).Times(4)

	// 1st call
	res.Status = model.StatusInProgress
	res.Results[0].Result.Status = model.StatusInProgress
	res.Results[0].Result.RawOutput = rawOutput1
	expectedTables[0], _ = generateTable(res, expectedCtx, 78)

	r, w, err := os.Pipe()
	assert.NoError(t, err)
	defer func() {
		w.Close()
		r.Close()
	}()
	os.Stdout = w

	err = outputTableView(fetcher, res, ctx)
	assert.NoError(t, err)

	firstCallStats := []model.MeasurementStats{
		{Sent: 3, Rcv: 3, Lost: 0, Loss: 0, Last: 17, Min: 17, Avg: 17.3, Max: 17.6, Time: 2002, Tsum: 51.9, Tsum2: 898.05, Mdev: 0.2449},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.46, Avg: 5.46, Max: 5.46, Time: 200, Tsum: 5.46, Tsum2: 29.8116},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.07, Avg: 4.07, Max: 4.07, Time: 300, Tsum: 4.07, Tsum2: 16.5649},
	}
	for i := range firstCallStats {
		assertMeasurementStats(t, &firstCallStats[i], &ctx.InProgressStats[i])
		assertMeasurementStats(t, &firstCallStats[i], &ctx.CompletedStats[i])
	}

	// 2nd call
	res.Status = model.StatusInProgress
	res.Results[0].Result.Status = model.StatusInProgress
	res.Results[0].Result.RawOutput = rawOutput1
	expectedCtx.CompletedStats = firstCallStats

	callCount++
	expectedTables[3], _ = generateTable(res, expectedCtx, 78)
	err = outputTableView(fetcher, res, ctx)
	assert.NoError(t, err)
	w.Close()

	os.Stdout = osStdOut
	output, err := io.ReadAll(r)
	assert.NoError(t, err)

	secondCallStats := []model.MeasurementStats{
		{Sent: 6, Rcv: 6, Lost: 0, Loss: 0, Last: 17, Min: 17, Avg: 17.3, Max: 17.6, Time: 4004, Tsum: 103.8, Tsum2: 1796.1, Mdev: 0.2449},
		{Sent: 2, Rcv: 2, Lost: 0, Loss: 0, Last: 5.46, Min: 5.46, Avg: 5.46, Max: 5.46, Time: 400, Tsum: 10.92, Tsum2: 59.6232},
		{Sent: 2, Rcv: 2, Lost: 0, Loss: 0, Last: 4.07, Min: 4.07, Avg: 4.07, Max: 4.07, Time: 600, Tsum: 8.14, Tsum2: 33.1298},
	}
	for i := range secondCallStats {
		assertMeasurementStats(t, &secondCallStats[i], &ctx.InProgressStats[i])
		assertMeasurementStats(t, &secondCallStats[i], &ctx.CompletedStats[i])
	}

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

func TestOutputTableView(t *testing.T) {
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

	err = outputTableView(fetcher, measurement, ctx)
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
	expectedTable, _ := generateTable(measurement, expectedCtx, 78) // 80 - 2. pterm defaults to 80 when terminal size is not detected.
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
			{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 100, Tsum: 0.77, Tsum2: 0.5929},
			{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.46, Avg: 5.46, Max: 5.46, Time: 200, Tsum: 5.46, Tsum2: 29.8116},
			{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.07, Avg: 4.07, Max: 4.07, Time: 300, Tsum: 4.07, Tsum2: 16.5649},
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
				{Sent: 10, Rcv: 9, Lost: 1, Loss: 10, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 1000, Mdev: 0.001},
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
rtt min/avg/max/mdev = 0.770/0.770/0.770/0.001 ms
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
rtt min/avg/max/mdev = -/-/-/- ms
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
rtt min/avg/max/mdev = -/-/-/- ms
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
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 100, Tsum: 0.77, Tsum2: 0.5929},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.46, Avg: 5.46, Max: 5.46, Time: 200, Tsum: 5.46, Tsum2: 29.8116},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.07, Avg: 4.07, Max: 4.07, Time: 300, Tsum: 4.07, Tsum2: 16.5649},
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
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 100, Tsum: 0.77, Tsum2: 0.5929},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.46, Avg: 5.46, Max: 5.46, Time: 200, Tsum: 5.46, Tsum2: 29.8116},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.07, Avg: 4.07, Max: 4.07, Time: 300, Tsum: 4.07, Tsum2: 16.5649},
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
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 100, Tsum: 0.77, Tsum2: 0.5929},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.46, Avg: 5.46, Max: 5.46, Time: 200, Tsum: 5.46, Tsum2: 29.8116},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.07, Avg: 4.07, Max: 4.07, Time: 300, Tsum: 4.07, Tsum2: 16.5649},
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
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 100, Tsum: 0.77, Tsum2: 0.5929},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.46, Avg: 5.46, Max: 5.46, Time: 200, Tsum: 5.46, Tsum2: 29.8116},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.07, Avg: 4.07, Max: 4.07, Time: 300, Tsum: 4.07, Tsum2: 16.5649},
	}, stats)
}

func TestMergeMeasurementStats(t *testing.T) {
	o := parsePingRawOutput(&model.MeasurementResponse{
		Result: model.ResultData{
			RawOutput: `PING  (142.250.65.174) 56(84) bytes of data.`,
		},
	}, 0)
	newStats := mergeMeasurementStats(
		model.MeasurementStats{Sent: 0, Lost: 0, Loss: 0, Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1},
		o,
	)
	assert.Equal(t,
		model.MeasurementStats{Sent: 0, Lost: 0, Loss: 0, Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1},
		newStats,
	)
	o = parsePingRawOutput(&model.MeasurementResponse{
		Result: model.ResultData{
			RawOutput: `PING  (142.250.65.174) 56(84) bytes of data.
no answer yet for icmp_seq=1
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=1 ttl=59 time=10 ms
no answer yet for icmp_seq=2
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=2 ttl=59 time=20 ms
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=3 ttl=59 time=30 ms
no answer yet for icmp_seq=4`,
		},
	}, 0)
	newStats = mergeMeasurementStats(
		model.MeasurementStats{Sent: 0, Lost: 0, Loss: 0, Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1},
		o)
	assertMeasurementStats(t, &model.MeasurementStats{Sent: 4, Rcv: 3, Lost: 1, Loss: 25, Last: 30, Min: 10,
		Avg: 20, Max: 30, Tsum: 60, Tsum2: 1400, Mdev: 8.1649},
		&newStats)
	o = parsePingRawOutput(&model.MeasurementResponse{
		Result: model.ResultData{
			RawOutput: `PING  (142.250.65.174) 56(84) bytes of data.
no answer yet for icmp_seq=1
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=1 ttl=59 time=10 ms
no answer yet for icmp_seq=2
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=2 ttl=59 time=10 ms
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=3 ttl=59 time=30 ms
no answer yet for icmp_seq=4
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=3 ttl=59 time=30 ms

---  ping statistics ---
4 packets transmitted, 4 received, 0% packet loss, time 1000ms
rtt min/avg/max/mdev = 10/20/30/0 ms`,
		},
	}, 0)
	newStats = mergeMeasurementStats(
		model.MeasurementStats{Sent: 0, Lost: 0, Loss: 0, Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1},
		o)
	assertMeasurementStats(t, &model.MeasurementStats{Sent: 4, Rcv: 4, Lost: 0, Loss: 0, Last: 30, Min: 10,
		Avg: 20, Max: 30, Time: 1000, Tsum: 80, Tsum2: 2000, Mdev: 10}, &newStats)
	o = parsePingRawOutput(&model.MeasurementResponse{
		Result: model.ResultData{
			RawOutput: `PING  (142.250.65.174) 56(84) bytes of data.
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=1 ttl=59 time=10 ms
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=1 ttl=59 time=20 ms
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=1 ttl=59 time=30 ms

---  ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 1000ms
rtt min/avg/max/mdev = 10/20/30/0 ms`,
		},
	}, 0)
	newStats = mergeMeasurementStats(
		model.MeasurementStats{Sent: 5, Rcv: 4, Lost: 1, Loss: 20, Last: 30, Min: 10, Avg: 20, Max: 30,
			Time: 1000, Tsum: 80, Tsum2: 2000, Mdev: 10},
		o)
	assertMeasurementStats(t, &model.MeasurementStats{
		Sent:  8,
		Rcv:   7,
		Lost:  1,
		Loss:  12.5,
		Last:  30,
		Min:   10,
		Avg:   20,
		Max:   30,
		Time:  2000,
		Tsum:  140,
		Tsum2: 3400,
		Mdev:  9.2582,
	}, &newStats)
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
	assert.Equal(t, "142.250.65.174", res.Address)
	assert.Equal(t, "56(84)", res.BytesOfData)
	assert.Nil(t, res.RawPacketLines)
	assert.Equal(t, []model.PingTiming{
		{RTT: 1.06, TTL: 59},
		{RTT: 1.10, TTL: 59},
		{RTT: 1.11, TTL: 59},
	}, res.Timings)
	assertMeasurementStats(t, &model.MeasurementStats{
		Sent:  3,
		Rcv:   3,
		Lost:  0,
		Loss:  0,
		Last:  1.11,
		Min:   1.06,
		Avg:   1.09,
		Max:   1.11,
		Tsum:  3.2700,
		Tsum2: 3.5657,
		Mdev:  0.0216,
	}, res.Stats)
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
	assert.Equal(t, "142.250.65.174", res.Address)
	assert.Equal(t, "56(84)", res.BytesOfData)
	assert.Nil(t, res.RawPacketLines)
	assert.Equal(t, []model.PingTiming{
		{RTT: 1.06, TTL: 59},
		{RTT: 1.10, TTL: 59},
		{RTT: 1.11, TTL: 59},
	}, res.Timings)
	assertMeasurementStats(t, &model.MeasurementStats{
		Sent:  4,
		Rcv:   3,
		Lost:  1,
		Loss:  25,
		Last:  1.11,
		Min:   1.06,
		Avg:   1.09,
		Max:   1.11,
		Tsum:  3.2700,
		Tsum2: 3.5657,
		Mdev:  0.0216,
	}, res.Stats)
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
	assert.Equal(t, "142.250.65.174", res.Address)
	assert.Equal(t, "56(84)", res.BytesOfData)
	assert.Equal(t, []string{
		"no answer yet for icmp_seq=5",
		"64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=5 ttl=59 time=1.06 ms",
		"no answer yet for icmp_seq=6",
		"64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=6 ttl=59 time=1.10 ms",
		"64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=7 ttl=59 time=1.11 ms",
		"no answer yet for icmp_seq=8",
	}, res.RawPacketLines)
	assert.Equal(t, []model.PingTiming{
		{RTT: 1.06, TTL: 59},
		{RTT: 1.10, TTL: 59},
		{RTT: 1.11, TTL: 59},
	}, res.Timings)
	assertMeasurementStats(t, &model.MeasurementStats{
		Sent:  4,
		Rcv:   3,
		Lost:  1,
		Loss:  25,
		Last:  1.11,
		Min:   1.06,
		Avg:   1.09,
		Max:   1.11,
		Tsum:  3.27,
		Tsum2: 3.5657,
		Mdev:  0.0216,
	}, res.Stats)
}

func TestComputeMdev(t *testing.T) {
	rtt1 := 10.0
	rtt2 := 10.0
	rtt3 := 30.0
	rtt4 := 30.0
	tsum := rtt1 + rtt2 + rtt3 + rtt4
	tsum2 := rtt1*rtt1 + rtt2*rtt2 + rtt3*rtt3 + rtt4*rtt4
	avg := tsum / 4
	mdev := computeMdev(tsum, tsum2, 4, avg)
	assert.InDelta(t, 10.0, mdev, 0.0001)
}

func assertMeasurementStats(t *testing.T, expected *model.MeasurementStats, actual *model.MeasurementStats) {
	assert.Equal(t, expected.Sent, actual.Sent)
	assert.Equal(t, expected.Rcv, actual.Rcv)
	assert.Equal(t, expected.Lost, actual.Lost)
	assert.InDelta(t, expected.Loss, actual.Loss, 0.0001)
	assert.Equal(t, expected.Last, actual.Last)
	assert.Equal(t, expected.Min, actual.Min)
	assert.InDelta(t, expected.Avg, actual.Avg, 0.0001)
	assert.Equal(t, expected.Max, actual.Max)
	assert.Equal(t, expected.Time, actual.Time)
	assert.InDelta(t, expected.Tsum, actual.Tsum, 0.0001)
	assert.InDelta(t, expected.Tsum2, actual.Tsum2, 0.0001)
	assert.InDelta(t, expected.Mdev, actual.Mdev, 0.0001)
}
