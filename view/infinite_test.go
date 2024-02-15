package view

import (
	"io"
	"math"
	"os"
	"testing"
	"time"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/mocks"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_OutputInfinite_SingleProbe_InProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	gbMock := mocks.NewMockClient(ctrl)
	measurement := getPingGetMeasurement(measurementID1)

	callCount := 0 // 1st call is done in the caller.
	gbMock.EXPECT().GetMeasurement(measurementID1).DoAndReturn(func(id string) (*globalping.Measurement, error) {
		callCount++
		switch callCount {
		case 1:
			measurement.Results[0].Result.RawOutput = rawOutput1
		case 2:
			measurement.Results[0].Result.RawOutput = rawOutput2
		case 3:
			measurement.Results[0].Result.RawOutput = rawOutput3
		case 4:
			measurement.Status = globalping.StatusFinished
			measurement.Results[0].Result.Status = globalping.StatusFinished
			measurement.Results[0].Result.RawOutput = rawOutput4
		}
		return measurement, nil
	}).Times(4)

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime.Add(1 * time.Millisecond)).Times(3)

	r, w, err := os.Pipe()
	assert.NoError(t, err)
	defer r.Close()
	defer w.Close()

	ctx := &Context{
		Cmd:        "ping",
		MaxHistory: 3,
		MStartedAt: defaultCurrentTime,
	}
	viewer := NewViewer(ctx, NewPrinter(w), timeMock, gbMock)

	measurement.Status = globalping.StatusInProgress
	measurement.Results[0].Result.Status = globalping.StatusInProgress
	measurement.Results[0].Result.RawOutput = rawOutput1

	err = viewer.OutputInfinite(measurement.ID)
	assert.NoError(t, err)
	w.Close()

	output, err := io.ReadAll(r)
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

	expectedStats := []MeasurementStats{{Sent: 3, Rcv: 3, Lost: 0, Loss: 0, Last: 13, Min: 12.7,
		Avg: 12.8666, Max: 13, Time: 1001, Tsum: 38.6, Tsum2: 496.7, Mdev: 0.1247}}
	assertMeasurementStats(t, &expectedStats[0], &ctx.InProgressStats[0])
	assertMeasurementStats(t, &expectedStats[0], &ctx.CompletedStats[0])
}

func Test_OutputInfinite_SingleProbe_Failed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gbMock := mocks.NewMockClient(ctrl)
	measurement := getPingGetMeasurement(measurementID1)
	measurement.Status = globalping.StatusFailed
	measurement.Results[0].Result.Status = globalping.StatusFailed
	measurement.Results[0].Result.RawOutput = `ping: cdn.jsdelivr.net.xc: Name or service not known`
	gbMock.EXPECT().GetMeasurement(measurementID1).Times(1).Return(measurement, nil)

	r, w, err := os.Pipe()
	assert.NoError(t, err)
	defer r.Close()
	defer w.Close()

	ctx := &Context{
		Cmd:        "ping",
		MaxHistory: 3,
	}
	viewer := NewViewer(ctx, NewPrinter(w), nil, gbMock)
	err = viewer.OutputInfinite(measurement.ID)
	assert.Equal(t, "all probes failed", err.Error())
	w.Close()

	output, err := io.ReadAll(r)
	assert.NoError(t, err)
	assert.Equal(t,
		`> EU, DE, Berlin, ASN:3320, Deutsche Telekom AG
ping: cdn.jsdelivr.net.xc: Name or service not known
`,
		string(output),
	)

	expectedStats := []MeasurementStats{NewMeasurementStats()}
	assertMeasurementStats(t, &expectedStats[0], &ctx.InProgressStats[0])
	assertMeasurementStats(t, &expectedStats[0], &ctx.CompletedStats[0])
}

func Test_OutputInfinite_SingleProbe_MultipleCalls(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gbMock := mocks.NewMockClient(ctrl)
	measurement := getPingGetMeasurement(measurementID1)
	gbMock.EXPECT().GetMeasurement(measurementID1).Times(3).Return(measurement, nil)

	r, w, err := os.Pipe()
	assert.NoError(t, err)
	defer r.Close()
	defer w.Close()

	ctx := &Context{
		Cmd:        "ping",
		MaxHistory: 3,
	}
	viewer := NewViewer(ctx, NewPrinter(w), nil, gbMock)

	err = viewer.OutputInfinite(measurement.ID)
	assert.NoError(t, err)
	err = viewer.OutputInfinite(measurement.ID)
	assert.NoError(t, err)
	err = viewer.OutputInfinite(measurement.ID)
	assert.NoError(t, err)
	w.Close()

	output, err := io.ReadAll(r)
	assert.NoError(t, err)
	assert.Equal(t,
		`> EU, DE, Berlin, ASN:3320, Deutsche Telekom AG
PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=60 time=17.6 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=2 ttl=60 time=17.6 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=3 ttl=60 time=17.6 ms
`,
		string(output))

	expectedStats := []MeasurementStats{{Sent: 3, Rcv: 3, Lost: 0, Loss: 0, Last: 17.6, Min: 17.6,
		Avg: 17.6, Max: 17.6, Time: 3000, Tsum: 52.8, Tsum2: 929.28, Mdev: 0}}
	assertMeasurementStats(t, &expectedStats[0], &ctx.InProgressStats[0])
	assertMeasurementStats(t, &expectedStats[0], &ctx.CompletedStats[0])
}

func Test_OutputInfinite_MultipleProbes_MultipleCalls(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	osStdout := os.Stdout

	gbMock := mocks.NewMockClient(ctrl)
	res := getPingGetMeasurementMultipleLocations(measurementID1)

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime.Add(1 * time.Millisecond)).AnyTimes()

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

	expectedViewer := &viewer{
		ctx:  getDefaultPingCtx(len(res.Results)),
		time: timeMock,
	}
	expectedViewer.ctx.MStartedAt = defaultCurrentTime
	expectedTables := [6]*string{}

	callCount := 0 // 1st call is done in the caller.
	gbMock.EXPECT().GetMeasurement(measurementID1).DoAndReturn(func(id string) (*globalping.Measurement, error) {
		callCount++
		switch callCount {
		case 1, 4:
			res.Results[0].Result.RawOutput = rawOutput1
		case 2, 5:
			res.Results[0].Result.RawOutput = rawOutput2
			expectedTables[callCount-1], _ = expectedViewer.generateTable(res, 78)
		case 3, 6:
			res.Status = globalping.StatusFinished
			res.Results[0].Result.Status = globalping.StatusFinished
			res.Results[0].Result.RawOutput = rawOutputFinal
			expectedTables[callCount-1], _ = expectedViewer.generateTable(res, 78)
		}
		return res, nil
	}).Times(6)

	// 1st call
	res.Status = globalping.StatusInProgress
	res.Results[0].Result.Status = globalping.StatusInProgress
	res.Results[0].Result.RawOutput = rawOutput1
	expectedTables[0], _ = expectedViewer.generateTable(res, 78)

	r, w, err := os.Pipe()
	assert.NoError(t, err)
	defer r.Close()
	defer w.Close()

	ctx := &Context{
		Cmd:        "ping",
		MaxHistory: 3,
		MStartedAt: defaultCurrentTime,
	}
	viewer := NewViewer(ctx, NewPrinter(w), timeMock, gbMock)
	os.Stdout = w
	err = viewer.OutputInfinite(measurementID1)
	assert.NoError(t, err)

	firstCallStats := []MeasurementStats{
		{Sent: 3, Rcv: 3, Lost: 0, Loss: 0, Last: 17, Min: 17, Avg: 17.3, Max: 17.6, Time: 2002, Tsum: 51.9, Tsum2: 898.05, Mdev: 0.2449},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.46, Avg: 5.46, Max: 5.46, Time: 200, Tsum: 5.46, Tsum2: 29.8116},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.07, Avg: 4.07, Max: 4.07, Time: 300, Tsum: 4.07, Tsum2: 16.5649},
	}
	for i := range firstCallStats {
		assertMeasurementStats(t, &firstCallStats[i], &ctx.InProgressStats[i])
		assertMeasurementStats(t, &firstCallStats[i], &ctx.CompletedStats[i])
	}

	// 2nd call
	res.Status = globalping.StatusInProgress
	res.Results[0].Result.Status = globalping.StatusInProgress
	res.Results[0].Result.RawOutput = rawOutput1
	expectedViewer.ctx.CompletedStats = firstCallStats

	expectedTables[3], _ = expectedViewer.generateTable(res, 78)
	err = viewer.OutputInfinite(measurementID1)
	assert.NoError(t, err)
	w.Close()

	output, err := io.ReadAll(r)
	assert.NoError(t, err)

	secondCallStats := []MeasurementStats{
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
	defer rr.Close()
	defer ww.Close()
	os.Stdout = ww
	// TODO: fix writer for AreaPrinter
	// pterm.DefaultArea.SetWriter(ww)
	area, _ := pterm.DefaultArea.Start()
	for i := range expectedTables {
		area.Update(*expectedTables[i])
	}
	area.Stop()
	os.Stdout = osStdout
	ww.Close()

	expectedOutput, err := io.ReadAll(rr)
	assert.NoError(t, err)
	assert.Equal(t, string(expectedOutput), string(output))
}

func Test_OutputInfinite_MultipleProbes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	osStdout := os.Stdout

	measurement := getPingGetMeasurementMultipleLocations(measurementID1)
	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().GetMeasurement(measurementID1).Times(1).Return(measurement, nil)

	r, w, err := os.Pipe()
	assert.NoError(t, err)
	defer r.Close()
	defer w.Close()

	ctx := &Context{
		Cmd:        "ping",
		MaxHistory: 3,
	}
	v := NewViewer(ctx, NewPrinter(w), nil, gbMock)
	os.Stdout = w
	err = v.OutputInfinite(measurementID1)
	assert.NoError(t, err)
	os.Stdout = osStdout
	w.Close()

	output, err := io.ReadAll(r)
	assert.NoError(t, err)

	expectedViewer := &viewer{
		ctx: getDefaultPingCtx(len(measurement.Results)),
	}

	rr, ww, err := os.Pipe()
	assert.NoError(t, err)
	defer rr.Close()
	defer ww.Close()
	os.Stdout = ww
	// TODO: fix writer for AreaPrinter
	// pterm.DefaultArea.SetWriter(ww)
	area, _ := pterm.DefaultArea.Start()
	expectedTable, _ := expectedViewer.generateTable(measurement, 78)
	area.Update(*expectedTable)
	area.Stop()
	os.Stdout = osStdout
	ww.Close()

	expectedOutput, err := io.ReadAll(rr)
	assert.NoError(t, err)
	assert.Equal(t, string(expectedOutput), string(output))
	assert.Equal(t,
		[]MeasurementStats{
			{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 100, Tsum: 0.77, Tsum2: 0.5929},
			{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.46, Avg: 5.46, Max: 5.46, Time: 200, Tsum: 5.46, Tsum2: 29.8116},
			{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.07, Avg: 4.07, Max: 4.07, Time: 300, Tsum: 4.07, Tsum2: 16.5649},
		},
		ctx.CompletedStats,
	)
}

func Test_OutputInfinite_MultipleProbes_All_Failed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	osStdout := os.Stdout

	measurement := getPingGetMeasurementMultipleLocations(measurementID1)
	measurement.Status = globalping.StatusFinished
	for i := range measurement.Results {
		measurement.Results[i].Result.Status = globalping.StatusFailed
		measurement.Results[i].Result.RawOutput = `ping: cdn.jsdelivr.net.xc: Name or service not known`
	}
	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().GetMeasurement(measurementID1).Times(1).Return(measurement, nil)

	r, w, err := os.Pipe()
	assert.NoError(t, err)
	defer r.Close()
	defer w.Close()

	ctx := &Context{
		Cmd:        "ping",
		MaxHistory: 3,
	}
	v := NewViewer(ctx, NewPrinter(w), nil, gbMock)
	os.Stdout = w
	err = v.OutputInfinite(measurementID1)
	os.Stdout = osStdout
	assert.Equal(t, "all probes failed", err.Error())
	w.Close()

	output, err := io.ReadAll(r)
	assert.NoError(t, err)

	assert.Equal(t, `> EU, GB, London, ASN:0, OVH SAS
ping: cdn.jsdelivr.net.xc: Name or service not known
> EU, DE, Falkenstein, ASN:0, Hetzner Online GmbH
ping: cdn.jsdelivr.net.xc: Name or service not known
> EU, DE, Nuremberg, ASN:0, Hetzner Online GmbH
ping: cdn.jsdelivr.net.xc: Name or service not known
`, string(output))
	expectedStats := []MeasurementStats{
		NewMeasurementStats(),
		NewMeasurementStats(),
		NewMeasurementStats(),
	}
	assert.Equal(t,
		expectedStats,
		ctx.CompletedStats,
	)
}

func Test_FormatDuration(t *testing.T) {
	d := formatDuration(1.2345)
	assert.Equal(t, "1.23 ms", d)
	d = formatDuration(12.345)
	assert.Equal(t, "12.3 ms", d)
	d = formatDuration(123.4567)
	assert.Equal(t, "123 ms", d)
}

func Test_GenerateTable_Full(t *testing.T) {
	measurement := getPingGetMeasurementMultipleLocations(measurementID1)
	expectedTable := "\x1b[96m\x1b[96mLocation                                       \x1b[0m\x1b[0m | \x1b[96m\x1b[96mSent\x1b[0m\x1b[0m | \x1b[96m\x1b[96m   Loss\x1b[0m\x1b[0m | \x1b[96m\x1b[96m    Last\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Min\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Avg\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Max\x1b[0m\x1b[0m\n" +
		"EU, GB, London, ASN:0, OVH SAS                  |    1 |   0.00% |  0.77 ms |  0.77 ms |  0.77 ms |  0.77 ms\n" +
		"EU, DE, Falkenstein, ASN:0, Hetzner Online GmbH |    1 |   0.00% |  5.46 ms |  5.46 ms |  5.46 ms |  5.46 ms\n" +
		"EU, DE, Nuremberg, ASN:0, Hetzner Online GmbH   |    1 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms\n"

	expectedViewer := &viewer{
		ctx: getDefaultPingCtx(len(measurement.Results)),
	}
	table, stats := expectedViewer.generateTable(measurement, 500)
	assert.Equal(t, expectedTable, *table)
	assert.Equal(t, []MeasurementStats{
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 100, Tsum: 0.77, Tsum2: 0.5929},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.46, Avg: 5.46, Max: 5.46, Time: 200, Tsum: 5.46, Tsum2: 29.8116},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.07, Avg: 4.07, Max: 4.07, Time: 300, Tsum: 4.07, Tsum2: 16.5649},
	}, stats)
}

func Test_GenerateTable_OneRow_Truncated(t *testing.T) {
	measurement := getPingGetMeasurementMultipleLocations(measurementID1)
	measurement.Results[1].Probe.Network = "作者聚集的原创内容平台于201 1年1月正式上线让人们更"
	expectedTable := "\x1b[96m\x1b[96mLocation                                      \x1b[0m\x1b[0m | \x1b[96m\x1b[96mSent\x1b[0m\x1b[0m | \x1b[96m\x1b[96m   Loss\x1b[0m\x1b[0m | \x1b[96m\x1b[96m    Last\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Min\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Avg\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Max\x1b[0m\x1b[0m\n" +
		"EU, GB, London, ASN:0, OVH SAS                 |    1 |   0.00% |  0.77 ms |  0.77 ms |  0.77 ms |  0.77 ms\n" +
		"EU, DE, Falkenstein, ASN:0, 作者聚集的原创...  |    1 |   0.00% |  5.46 ms |  5.46 ms |  5.46 ms |  5.46 ms\n" +
		"EU, DE, Nuremberg, ASN:0, Hetzner Online GmbH  |    1 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms\n"
	expectedViewer := &viewer{
		ctx: getDefaultPingCtx(len(measurement.Results)),
	}
	table, stats := expectedViewer.generateTable(measurement, 106)
	assert.Equal(t, expectedTable, *table)
	assert.Equal(t, []MeasurementStats{
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 100, Tsum: 0.77, Tsum2: 0.5929},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.46, Avg: 5.46, Max: 5.46, Time: 200, Tsum: 5.46, Tsum2: 29.8116},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.07, Avg: 4.07, Max: 4.07, Time: 300, Tsum: 4.07, Tsum2: 16.5649},
	}, stats)
}

func Test_GenerateTable_MultiLine_Truncated(t *testing.T) {
	measurement := getPingGetMeasurementMultipleLocations(measurementID1)
	measurement.Results[1].Probe.Network = "Hetzner Online GmbH\nLorem ipsum\nLorem ipsum dolor sit amet"
	expectedTable := "\x1b[96m\x1b[96mLocation                                      \x1b[0m\x1b[0m | \x1b[96m\x1b[96mSent\x1b[0m\x1b[0m | \x1b[96m\x1b[96m   Loss\x1b[0m\x1b[0m | \x1b[96m\x1b[96m    Last\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Min\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Avg\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Max\x1b[0m\x1b[0m\n" +
		"EU, GB, London, ASN:0, OVH SAS                 |    1 |   0.00% |  0.77 ms |  0.77 ms |  0.77 ms |  0.77 ms\n" +
		"EU, DE, Falkenstein, ASN:0, Hetzner Online ... |    1 |   0.00% |  5.46 ms |  5.46 ms |  5.46 ms |  5.46 ms\n" +
		"Lorem ipsum                                    |      |         |          |          |          |         \n" +
		"Lorem ipsum dolor sit amet                     |      |         |          |          |          |         \n" +
		"EU, DE, Nuremberg, ASN:0, Hetzner Online GmbH  |    1 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms\n"
	expectedViewer := &viewer{
		ctx: getDefaultPingCtx(len(measurement.Results)),
	}
	table, stats := expectedViewer.generateTable(measurement, 106)
	assert.Equal(t, expectedTable, *table)
	assert.Equal(t, []MeasurementStats{
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 100, Tsum: 0.77, Tsum2: 0.5929},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.46, Avg: 5.46, Max: 5.46, Time: 200, Tsum: 5.46, Tsum2: 29.8116},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.07, Avg: 4.07, Max: 4.07, Time: 300, Tsum: 4.07, Tsum2: 16.5649},
	}, stats)
}

func Test_GenerateTable_MaxTruncated(t *testing.T) {
	measurement := getPingGetMeasurementMultipleLocations(measurementID1)
	expectedTable := "\x1b[96m\x1b[96mLoc...\x1b[0m\x1b[0m | \x1b[96m\x1b[96mSent\x1b[0m\x1b[0m | \x1b[96m\x1b[96m   Loss\x1b[0m\x1b[0m | \x1b[96m\x1b[96m    Last\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Min\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Avg\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Max\x1b[0m\x1b[0m\n" +
		"EU,... |    1 |   0.00% |  0.77 ms |  0.77 ms |  0.77 ms |  0.77 ms\n" +
		"EU,... |    1 |   0.00% |  5.46 ms |  5.46 ms |  5.46 ms |  5.46 ms\n" +
		"EU,... |    1 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms\n"
	expectedViewer := &viewer{
		ctx: getDefaultPingCtx(len(measurement.Results)),
	}
	table, stats := expectedViewer.generateTable(measurement, 0)
	assert.Equal(t, expectedTable, *table)
	assert.Equal(t, []MeasurementStats{
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 100, Tsum: 0.77, Tsum2: 0.5929},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.46, Avg: 5.46, Max: 5.46, Time: 200, Tsum: 5.46, Tsum2: 29.8116},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.07, Avg: 4.07, Max: 4.07, Time: 300, Tsum: 4.07, Tsum2: 16.5649},
	}, stats)
}

func Test_MergeMeasurementStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime.Add(1 * time.Millisecond)).Times(2)

	v := viewer{
		ctx: &Context{
			MStartedAt: defaultCurrentTime,
		},
		time: timeMock,
	}

	o := v.parsePingRawOutput(&globalping.ProbeMeasurement{
		Result: globalping.ProbeResult{
			RawOutput: `PING  (142.250.65.174) 56(84) bytes of data.`,
		},
	}, 0)
	newStats := mergeMeasurementStats(
		MeasurementStats{Sent: 0, Lost: 0, Loss: 0, Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1},
		o,
	)
	assert.Equal(t,
		MeasurementStats{Sent: 0, Lost: 0, Loss: 0, Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1, Time: 1},
		newStats,
	)
	o = v.parsePingRawOutput(&globalping.ProbeMeasurement{
		Result: globalping.ProbeResult{
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
		MeasurementStats{Sent: 0, Lost: 0, Loss: 0, Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1},
		o)
	assertMeasurementStats(t, &MeasurementStats{Sent: 4, Rcv: 3, Lost: 1, Loss: 25, Last: 30, Min: 10,
		Avg: 20, Max: 30, Time: 1, Tsum: 60, Tsum2: 1400, Mdev: 8.1649},
		&newStats)
	o = v.parsePingRawOutput(&globalping.ProbeMeasurement{
		Result: globalping.ProbeResult{
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
		MeasurementStats{Sent: 0, Lost: 0, Loss: 0, Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1},
		o)
	assertMeasurementStats(t, &MeasurementStats{Sent: 4, Rcv: 4, Lost: 0, Loss: 0, Last: 30, Min: 10,
		Avg: 20, Max: 30, Time: 1000, Tsum: 80, Tsum2: 2000, Mdev: 10}, &newStats)
	o = v.parsePingRawOutput(&globalping.ProbeMeasurement{
		Result: globalping.ProbeResult{
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
		MeasurementStats{Sent: 5, Rcv: 4, Lost: 1, Loss: 20, Last: 30, Min: 10, Avg: 20, Max: 30,
			Time: 1000, Tsum: 80, Tsum2: 2000, Mdev: 10},
		o)
	assertMeasurementStats(t, &MeasurementStats{
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

func Test_GetRowValues_NoPacketsRcv(t *testing.T) {
	stats := MeasurementStats{Sent: 1, Lost: 0, Loss: 0, Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1}
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

func Test_GetRowValues(t *testing.T) {
	stats := MeasurementStats{
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

func Test_ParsePingRawOutput_Full(t *testing.T) {
	v := viewer{
		ctx: &Context{},
	}

	m := &globalping.ProbeMeasurement{
		Result: globalping.ProbeResult{
			RawOutput: `PING cdn.jsdelivr.net (142.250.65.174) 56(84) bytes of data.
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=1 ttl=59 time=1.06 ms
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=2 ttl=59 time=1.10 ms
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=3 ttl=59 time=1.11 ms

---  ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 1002ms
rtt min/avg/max/mdev = 1.061/1.090/1.108/0.020 ms`,
		},
	}

	res := v.parsePingRawOutput(m, -1)
	assert.Equal(t, "142.250.65.174", res.Address)
	assert.Equal(t, "56(84)", res.BytesOfData)
	assert.Nil(t, res.RawPacketLines)
	assert.Equal(t, []globalping.PingTiming{
		{RTT: 1.06, TTL: 59},
		{RTT: 1.10, TTL: 59},
		{RTT: 1.11, TTL: 59},
	}, res.Timings)
	assertMeasurementStats(t, &MeasurementStats{
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

func Test_ParsePingRawOutput_NoStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime.Add(1 * time.Millisecond))

	v := viewer{
		ctx: &Context{
			MStartedAt: defaultCurrentTime,
		},
		time: timeMock,
	}

	m := &globalping.ProbeMeasurement{
		Result: globalping.ProbeResult{
			RawOutput: `PING  (142.250.65.174) 56(84) bytes of data.
no answer yet for icmp_seq=1
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=1 ttl=59 time=1.06 ms
no answer yet for icmp_seq=2
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=2 ttl=59 time=1.10 ms
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=3 ttl=59 time=1.11 ms
no answer yet for icmp_seq=4`,
		},
	}
	res := v.parsePingRawOutput(m, -1)
	assert.Equal(t, "142.250.65.174", res.Address)
	assert.Equal(t, "56(84)", res.BytesOfData)
	assert.Nil(t, res.RawPacketLines)
	assert.Equal(t, []globalping.PingTiming{
		{RTT: 1.06, TTL: 59},
		{RTT: 1.10, TTL: 59},
		{RTT: 1.11, TTL: 59},
	}, res.Timings)
	assertMeasurementStats(t, &MeasurementStats{
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

func Test_ParsePingRawOutput_NoStats_WithStartIncmpSeq(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime.Add(1 * time.Millisecond))

	v := viewer{
		ctx: &Context{
			MStartedAt: defaultCurrentTime,
		},
		time: timeMock,
	}

	m := &globalping.ProbeMeasurement{
		Result: globalping.ProbeResult{
			RawOutput: `PING  (142.250.65.174) 56(84) bytes of data.
no answer yet for icmp_seq=1
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=1 ttl=59 time=1.06 ms
no answer yet for icmp_seq=2
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=2 ttl=59 time=1.10 ms
64 bytes from lga25s71-in-f14.1e100.net (142.250.65.174): icmp_seq=3 ttl=59 time=1.11 ms
no answer yet for icmp_seq=4`,
		},
	}
	res := v.parsePingRawOutput(m, 4)
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
	assert.Equal(t, []globalping.PingTiming{
		{RTT: 1.06, TTL: 59},
		{RTT: 1.10, TTL: 59},
		{RTT: 1.11, TTL: 59},
	}, res.Timings)
	assertMeasurementStats(t, &MeasurementStats{
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

func Test_ComputeMdev(t *testing.T) {
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

func assertMeasurementStats(t *testing.T, expected *MeasurementStats, actual *MeasurementStats) {
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
