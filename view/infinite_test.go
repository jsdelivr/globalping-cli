package view

import (
	"bytes"
	"math"
	"testing"
	"time"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_OutputInfinite_SingleProbe_InProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime.Add(500 * time.Millisecond)).Times(3)

	ctx := createDefaultContext("ping")
	ctx.CIMode = true
	hm := ctx.History.Find(measurementID1)
	w := new(bytes.Buffer)
	viewer := NewViewer(ctx, NewPrinter(nil, w, w), timeMock, nil)

	measurement := createPingMeasurement(measurementID1)
	measurement.Status = globalping.StatusInProgress
	measurement.Results[0].Result.Status = globalping.StatusInProgress
	measurement.Results[0].Result.RawOutput = `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.`

	err := viewer.OutputInfinite(measurement)
	assert.NoError(t, err)

	assert.Equal(t,
		`> Berlin, DE, EU, Deutsche Telekom AG (AS3320)
PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
`,
		w.String(),
	)

	measurement.Results[0].Result.RawOutput = `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=56 time=12.9 ms`

	err = viewer.OutputInfinite(measurement)
	assert.NoError(t, err)

	assert.Equal(t,
		`> Berlin, DE, EU, Deutsche Telekom AG (AS3320)
PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=56 time=12.9 ms
`,
		w.String(),
	)

	measurement.Results[0].Result.RawOutput = `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=56 time=12.9 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=2 ttl=56 time=12.7 ms`

	err = viewer.OutputInfinite(measurement)
	assert.NoError(t, err)

	assert.Equal(t,
		`> Berlin, DE, EU, Deutsche Telekom AG (AS3320)
PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=56 time=12.9 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=2 ttl=56 time=12.7 ms
`,
		w.String(),
	)

	expectedStats := &MeasurementStats{Sent: 2, Rcv: 2, Lost: 0, Loss: 0, Last: 12.7, Min: 12.7,
		Avg: 12.8, Max: 12.9, Time: 500, Tsum: 25.6, Tsum2: 327.7, Mdev: 0.0999}
	assertMeasurementStats(t, expectedStats, hm.Stats[0])

	measurement.Status = globalping.StatusFinished
	measurement.Results[0].Result.Status = globalping.StatusFinished
	measurement.Results[0].Result.RawOutput = `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=56 time=12.9 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=2 ttl=56 time=12.7 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=3 ttl=56 time=13.0 ms

---  ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 1001ms
rtt min/avg/max/mdev = 12.711/12.854/12.952/0.103 ms`

	err = viewer.OutputInfinite(measurement)
	assert.NoError(t, err)

	assert.Equal(t,
		`> Berlin, DE, EU, Deutsche Telekom AG (AS3320)
PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=56 time=12.9 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=2 ttl=56 time=12.7 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=3 ttl=56 time=13.0 ms
`,
		w.String(),
	)

	expectedStats = &MeasurementStats{Sent: 3, Rcv: 3, Lost: 0, Loss: 0, Last: 13, Min: 12.7,
		Avg: 12.8666, Max: 13, Time: 1001, Tsum: 38.6, Tsum2: 496.7, Mdev: 0.1247}
	assertMeasurementStats(t, expectedStats, ctx.AggregatedStats[0])

	ctx.MeasurementsCreated = 2
	ctx.History.Push(&HistoryItem{
		Id:        measurementID2,
		StartedAt: defaultCurrentTime.Add(1 * time.Millisecond),
	})
	measurement.ID = measurementID2
	err = viewer.OutputInfinite(measurement)
	assert.NoError(t, err)

	assert.Equal(t,
		`> Berlin, DE, EU, Deutsche Telekom AG (AS3320)
PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=56 time=12.9 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=2 ttl=56 time=12.7 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=3 ttl=56 time=13.0 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=4 ttl=56 time=12.9 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=5 ttl=56 time=12.7 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=6 ttl=56 time=13.0 ms
`,
		w.String(),
	)

	expectedStats = &MeasurementStats{Sent: 6, Rcv: 6, Lost: 0, Loss: 0, Last: 13, Min: 12.7,
		Avg: 12.8666, Max: 13, Time: 2002, Tsum: 77.2, Tsum2: 993.4, Mdev: 0.1247}
	assertMeasurementStats(t, expectedStats, ctx.AggregatedStats[0])
}

func Test_OutputInfinite_SingleProbe_Failed(t *testing.T) {
	measurement := createPingMeasurement(measurementID1)
	measurement.Status = globalping.StatusFailed
	measurement.Results[0].Result.Status = globalping.StatusFailed
	measurement.Results[0].Result.RawOutput = `ping: cdn.jsdelivr.net.xc: Name or service not known`

	ctx := createDefaultContext("ping")
	ctx.CIMode = true
	w := new(bytes.Buffer)
	viewer := NewViewer(ctx, NewPrinter(nil, w, w), nil, nil)
	err := viewer.OutputInfinite(measurement)
	assert.Equal(t, "all probes failed", err.Error())

	assert.Equal(t,
		`> Berlin, DE, EU, Deutsche Telekom AG (AS3320)
ping: cdn.jsdelivr.net.xc: Name or service not known
`,
		w.String(),
	)

	assert.Nil(t, ctx.AggregatedStats)
}

func Test_OutputInfinite_MultipleProbes_MultipleCalls(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime.Add(1 * time.Millisecond)).AnyTimes()

	measurement := createPingMeasurement_MultipleProbes(measurementID1)
	measurement.Status = globalping.StatusInProgress
	measurement.Results[0].Result.Status = globalping.StatusInProgress
	measurement.Results[0].Result.RawOutput = `PING  (146.75.73.229) 56(84) bytes of data.`

	ctx := createDefaultContext("ping")
	ctx.CIMode = true
	w := new(bytes.Buffer)
	viewer := NewViewer(ctx, NewPrinter(nil, w, w), timeMock, nil)

	// Call 1
	expectedOutput := `Location                                       | Sent |    Loss |     Last |      Min |      Avg |      Max
London, GB, EU, OVH SAS (AS0)                  |    0 |   0.00% |        - |        - |        - |        -
Falkenstein, DE, EU, Hetzner Online GmbH (AS0) |    1 |   0.00% |  5.46 ms |  5.46 ms |  5.46 ms |  5.46 ms
Nuremberg, DE, EU, Hetzner Online GmbH (AS0)   |    1 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms
`
	err := viewer.OutputInfinite(measurement)
	assert.NoError(t, err)

	expectedStats := []*MeasurementStats{
		NewMeasurementStats(),
		NewMeasurementStats(),
		NewMeasurementStats(),
	}
	assertMeasurementStats(t, expectedStats[0], ctx.AggregatedStats[0])
	assertMeasurementStats(t, expectedStats[1], ctx.AggregatedStats[1])
	assertMeasurementStats(t, expectedStats[2], ctx.AggregatedStats[2])

	measurement.Results[0].Result.RawOutput = `PING  (146.75.73.229) 56(84) bytes of data.
64 bytes from 146.75.73.229 (146.75.73.229): icmp_seq=1 ttl=52 time=17.6 ms
no answer yet for icmp_seq=2`

	// Call 2
	expectedOutput += "\033[4A\033[0J" +
		`Location                                       | Sent |    Loss |     Last |      Min |      Avg |      Max
London, GB, EU, OVH SAS (AS0)                  |    2 |  50.00% |  17.6 ms |  17.6 ms |  17.6 ms |  17.6 ms
Falkenstein, DE, EU, Hetzner Online GmbH (AS0) |    1 |   0.00% |  5.46 ms |  5.46 ms |  5.46 ms |  5.46 ms
Nuremberg, DE, EU, Hetzner Online GmbH (AS0)   |    1 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms
`

	err = viewer.OutputInfinite(measurement)
	assert.NoError(t, err)

	assertMeasurementStats(t, expectedStats[0], ctx.AggregatedStats[0])
	assertMeasurementStats(t, expectedStats[1], ctx.AggregatedStats[1])
	assertMeasurementStats(t, expectedStats[2], ctx.AggregatedStats[2])

	measurement.Status = globalping.StatusFinished
	measurement.Results[0].Result.Status = globalping.StatusFinished
	measurement.Results[0].Result.RawOutput = `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=60 time=17.6 ms
no answer yet for icmp_seq=2
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=2 ttl=30 time=17.3 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=3 ttl=10 time=17.0 ms

---  ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 2002ms
rtt min/avg/max/mdev = 17.006/17.333/17.648/0.321 ms`

	// Call 3
	expectedOutput += "\033[4A\033[0J" +
		`Location                                       | Sent |    Loss |     Last |      Min |      Avg |      Max
London, GB, EU, OVH SAS (AS0)                  |    3 |   0.00% |  17.0 ms |  17.0 ms |  17.3 ms |  17.6 ms
Falkenstein, DE, EU, Hetzner Online GmbH (AS0) |    1 |   0.00% |  5.46 ms |  5.46 ms |  5.46 ms |  5.46 ms
Nuremberg, DE, EU, Hetzner Online GmbH (AS0)   |    1 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms
`

	err = viewer.OutputInfinite(measurement)
	assert.NoError(t, err)

	expectedStats = []*MeasurementStats{
		{Sent: 3, Rcv: 3, Lost: 0, Loss: 0, Last: 17, Min: 17, Avg: 17.3, Max: 17.6, Time: 2002, Tsum: 51.9, Tsum2: 898.05, Mdev: 0.2449},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.46, Avg: 5.46, Max: 5.46, Time: 200, Tsum: 5.46, Tsum2: 29.8116},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.07, Avg: 4.07, Max: 4.07, Time: 300, Tsum: 4.07, Tsum2: 16.5649},
	}
	assertMeasurementStats(t, expectedStats[0], ctx.AggregatedStats[0])
	assertMeasurementStats(t, expectedStats[1], ctx.AggregatedStats[1])
	assertMeasurementStats(t, expectedStats[2], ctx.AggregatedStats[2])

	// Call 4
	measurement2 := createPingMeasurement_MultipleProbes(measurementID2)
	measurement2.Results[0].Result.RawOutput = measurement.Results[0].Result.RawOutput

	ctx.History.Push(&HistoryItem{
		Id:        measurementID2,
		StartedAt: defaultCurrentTime.Add(1 * time.Millisecond),
	})

	err = viewer.OutputInfinite(measurement2)
	assert.NoError(t, err)

	expectedStats = []*MeasurementStats{
		{Sent: 6, Rcv: 6, Lost: 0, Loss: 0, Last: 17, Min: 17, Avg: 17.3, Max: 17.6, Time: 4004, Tsum: 103.8, Tsum2: 1796.1, Mdev: 0.2449},
		{Sent: 2, Rcv: 2, Lost: 0, Loss: 0, Last: 5.46, Min: 5.46, Avg: 5.46, Max: 5.46, Time: 400, Tsum: 10.92, Tsum2: 59.6232},
		{Sent: 2, Rcv: 2, Lost: 0, Loss: 0, Last: 4.07, Min: 4.07, Avg: 4.07, Max: 4.07, Time: 600, Tsum: 8.14, Tsum2: 33.1298},
	}
	assertMeasurementStats(t, expectedStats[0], ctx.AggregatedStats[0])
	assertMeasurementStats(t, expectedStats[1], ctx.AggregatedStats[1])
	assertMeasurementStats(t, expectedStats[2], ctx.AggregatedStats[2])

	expectedOutput += "\033[4A\033[0J" +
		`Location                                       | Sent |    Loss |     Last |      Min |      Avg |      Max
London, GB, EU, OVH SAS (AS0)                  |    6 |   0.00% |  17.0 ms |  17.0 ms |  17.3 ms |  17.6 ms
Falkenstein, DE, EU, Hetzner Online GmbH (AS0) |    2 |   0.00% |  5.46 ms |  5.46 ms |  5.46 ms |  5.46 ms
Nuremberg, DE, EU, Hetzner Online GmbH (AS0)   |    2 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms
`
	assert.Equal(t, expectedOutput, w.String())
}

func Test_OutputInfinite_MultipleProbes_MultipleConcurrentCalls(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime.Add(1 * time.Millisecond)).AnyTimes()

	// Call 1
	measurement1 := createPingMeasurement_MultipleProbes(measurementID1)
	measurement1.Status = globalping.StatusInProgress
	measurement1.Results[0].Result.Status = globalping.StatusInProgress
	measurement1.Results[0].Result.RawOutput = `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=60 time=10 ms`
	measurement1.Results[1].Result.Status = globalping.StatusInProgress
	measurement1.Results[1].Result.RawOutput = `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.`

	ctx := createDefaultContext("ping")
	hm1 := ctx.History.Find(measurementID1)
	hm1.Status = globalping.StatusInProgress
	ctx.CIMode = true
	w := new(bytes.Buffer)
	viewer := NewViewer(ctx, NewPrinter(nil, w, w), timeMock, nil)

	expectedOutput := `Location                                       | Sent |    Loss |     Last |      Min |      Avg |      Max
London, GB, EU, OVH SAS (AS0)                  |    1 |   0.00% |  10.0 ms |  10.0 ms |  10.0 ms |  10.0 ms
Falkenstein, DE, EU, Hetzner Online GmbH (AS0) |    0 |   0.00% |        - |        - |        - |        -
Nuremberg, DE, EU, Hetzner Online GmbH (AS0)   |    1 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms
`

	err := viewer.OutputInfinite(measurement1)
	assert.NoError(t, err)

	// Call 2
	measurement2 := createPingMeasurement_MultipleProbes(measurementID2)
	measurement2.Status = globalping.StatusInProgress
	measurement2.Results[0].Result.Status = globalping.StatusInProgress
	measurement2.Results[0].Result.RawOutput = `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.`
	measurement2.Results[1].Result.Status = globalping.StatusInProgress
	measurement2.Results[1].Result.RawOutput = `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=60 time=20 ms`
	ctx.History.Push(&HistoryItem{
		Id:        measurementID2,
		Status:    globalping.StatusInProgress,
		StartedAt: defaultCurrentTime.Add(1 * time.Millisecond),
	})

	expectedOutput += "\033[4A\033[0J" +
		`Location                                       | Sent |    Loss |     Last |      Min |      Avg |      Max
London, GB, EU, OVH SAS (AS0)                  |    1 |   0.00% |  10.0 ms |  10.0 ms |  10.0 ms |  10.0 ms
Falkenstein, DE, EU, Hetzner Online GmbH (AS0) |    1 |   0.00% |  20.0 ms |  20.0 ms |  20.0 ms |  20.0 ms
Nuremberg, DE, EU, Hetzner Online GmbH (AS0)   |    2 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms
`

	err = viewer.OutputInfinite(measurement2)
	assert.NoError(t, err)

	// Call 3
	measurement1.Results[1].Result.RawOutput = `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=60 time=20 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=2 ttl=30 time=25 ms`

	expectedOutput += "\033[4A\033[0J" +
		`Location                                       | Sent |    Loss |     Last |      Min |      Avg |      Max
London, GB, EU, OVH SAS (AS0)                  |    1 |   0.00% |  10.0 ms |  10.0 ms |  10.0 ms |  10.0 ms
Falkenstein, DE, EU, Hetzner Online GmbH (AS0) |    3 |   0.00% |  20.0 ms |  20.0 ms |  21.7 ms |  25.0 ms
Nuremberg, DE, EU, Hetzner Online GmbH (AS0)   |    2 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms
`

	err = viewer.OutputInfinite(measurement1)
	assert.NoError(t, err)

	// Call 4
	measurement1.Status = globalping.StatusFinished
	measurement1.Results[0].Result.Status = globalping.StatusFinished
	measurement1.Results[0].Result.RawOutput = `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=60 time=10 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=2 ttl=30 time=15 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=3 ttl=30 time=25 ms

---  ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 100ms
rtt min/avg/max/mdev = 10/15/25/5 ms`
	measurement1.Results[1].Result.Status = globalping.StatusFinished
	measurement1.Results[1].Result.RawOutput = `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=60 time=20 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=2 ttl=30 time=25 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=3 ttl=30 time=30 ms

---  ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 200ms
rtt min/avg/max/mdev = 20/25/30/5 ms`
	hm1.Status = globalping.StatusFinished

	expectedOutput += "\033[4A\033[0J" +
		`Location                                       | Sent |    Loss |     Last |      Min |      Avg |      Max
London, GB, EU, OVH SAS (AS0)                  |    3 |   0.00% |  25.0 ms |  10.0 ms |  16.7 ms |  25.0 ms
Falkenstein, DE, EU, Hetzner Online GmbH (AS0) |    4 |   0.00% |  20.0 ms |  20.0 ms |  23.8 ms |  30.0 ms
Nuremberg, DE, EU, Hetzner Online GmbH (AS0)   |    2 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms
`

	err = viewer.OutputInfinite(measurement1)
	assert.NoError(t, err)

	// Call 5
	measurement2.Results[0].Result.Status = globalping.StatusFinished
	measurement2.Results[0].Result.RawOutput = `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=60 time=10 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=2 ttl=30 time=15 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=3 ttl=30 time=25 ms

---  ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 100ms
rtt min/avg/max/mdev = 10/15/25/5 ms`

	err = viewer.OutputInfinite(measurement2)
	assert.NoError(t, err)

	expectedOutput += "\033[4A\033[0J" +
		`Location                                       | Sent |    Loss |     Last |      Min |      Avg |      Max
London, GB, EU, OVH SAS (AS0)                  |    6 |   0.00% |  25.0 ms |  10.0 ms |  16.7 ms |  25.0 ms
Falkenstein, DE, EU, Hetzner Online GmbH (AS0) |    4 |   0.00% |  20.0 ms |  20.0 ms |  23.8 ms |  30.0 ms
Nuremberg, DE, EU, Hetzner Online GmbH (AS0)   |    2 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms
`
	assert.Equal(t, expectedOutput, w.String())
}

func Test_OutputInfinite_MultipleProbes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	measurement := createPingMeasurement_MultipleProbes(measurementID1)

	ctx := createDefaultContext("ping")
	w := new(bytes.Buffer)
	v := NewViewer(ctx, NewPrinter(nil, w, w), nil, nil)
	err := v.OutputInfinite(measurement)
	assert.NoError(t, err)

	expectedOutput := "\033[96mLocation                                      \033[0m | \033[96mSent\033[0m | \033[96m   Loss\033[0m | \033[96m    Last\033[0m | \033[96m     Min\033[0m | \033[96m     Avg\033[0m | \033[96m     Max\033[0m" +
		`
London, GB, EU, OVH SAS (AS0)                  |    1 |   0.00% |  0.77 ms |  0.77 ms |  0.77 ms |  0.77 ms
Falkenstein, DE, EU, Hetzner Online GmbH (AS0) |    1 |   0.00% |  5.46 ms |  5.46 ms |  5.46 ms |  5.46 ms
Nuremberg, DE, EU, Hetzner Online GmbH (AS0)   |    1 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms
`
	assert.Equal(t, expectedOutput, w.String())
	assert.Equal(t,
		[]*MeasurementStats{
			{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 100, Tsum: 0.77, Tsum2: 0.5929},
			{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.46, Avg: 5.46, Max: 5.46, Time: 200, Tsum: 5.46, Tsum2: 29.8116},
			{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.07, Avg: 4.07, Max: 4.07, Time: 300, Tsum: 4.07, Tsum2: 16.5649},
		},
		ctx.AggregatedStats,
	)
}

func Test_OutputInfinite_MultipleProbes_All_Failed(t *testing.T) {
	measurement := createPingMeasurement_MultipleProbes(measurementID1)
	measurement.Status = globalping.StatusFinished
	for i := range measurement.Results {
		measurement.Results[i].Result.Status = globalping.StatusFailed
		measurement.Results[i].Result.RawOutput = `ping: cdn.jsdelivr.net.xc: Name or service not known`
	}

	ctx := createDefaultContext("ping")
	ctx.CIMode = true
	w := new(bytes.Buffer)
	v := NewViewer(ctx, NewPrinter(nil, w, w), nil, nil)
	err := v.OutputInfinite(measurement)

	assert.Equal(t, "all probes failed", err.Error())
	assert.Equal(t, `> London, GB, EU, OVH SAS (AS0)
ping: cdn.jsdelivr.net.xc: Name or service not known
> Falkenstein, DE, EU, Hetzner Online GmbH (AS0)
ping: cdn.jsdelivr.net.xc: Name or service not known
> Nuremberg, DE, EU, Hetzner Online GmbH (AS0)
ping: cdn.jsdelivr.net.xc: Name or service not known
`, w.String())

	assert.Nil(t, ctx.AggregatedStats)
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
	ctx := createDefaultContext("ping")
	ctx.AggregatedStats = []*MeasurementStats{
		NewMeasurementStats(),
		NewMeasurementStats(),
		NewMeasurementStats(),
	}
	hm := ctx.History.Find(measurementID1)
	viewer := &viewer{ctx: ctx}
	measurement := createPingMeasurement_MultipleProbes(measurementID1)
	table, _, stats := viewer.generateTable(hm, measurement, 500)

	expectedTable := "\033[96mLocation                                      \033[0m | \033[96mSent\033[0m | \033[96m   Loss\033[0m | \033[96m    Last\033[0m | \033[96m     Min\033[0m | \033[96m     Avg\033[0m | \033[96m     Max\033[0m\n" +
		"London, GB, EU, OVH SAS (AS0)                  |    1 |   0.00% |  0.77 ms |  0.77 ms |  0.77 ms |  0.77 ms\n" +
		"Falkenstein, DE, EU, Hetzner Online GmbH (AS0) |    1 |   0.00% |  5.46 ms |  5.46 ms |  5.46 ms |  5.46 ms\n" +
		"Nuremberg, DE, EU, Hetzner Online GmbH (AS0)   |    1 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms\n"
	assert.Equal(t, expectedTable, *table)

	assert.Equal(t, []*MeasurementStats{
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 100, Tsum: 0.77, Tsum2: 0.5929},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.46, Avg: 5.46, Max: 5.46, Time: 200, Tsum: 5.46, Tsum2: 29.8116},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.07, Avg: 4.07, Max: 4.07, Time: 300, Tsum: 4.07, Tsum2: 16.5649},
	}, stats)
}

func Test_GenerateTable_CIMode(t *testing.T) {
	ctx := createDefaultContext("ping")
	ctx.CIMode = true
	ctx.AggregatedStats = []*MeasurementStats{
		NewMeasurementStats(),
		NewMeasurementStats(),
		NewMeasurementStats(),
	}
	hm := ctx.History.Find(measurementID1)
	viewer := &viewer{ctx: ctx}

	measurement := createPingMeasurement_MultipleProbes(measurementID1)
	table, _, stats := viewer.generateTable(hm, measurement, 500)

	expectedTable := `Location                                       | Sent |    Loss |     Last |      Min |      Avg |      Max
London, GB, EU, OVH SAS (AS0)                  |    1 |   0.00% |  0.77 ms |  0.77 ms |  0.77 ms |  0.77 ms
Falkenstein, DE, EU, Hetzner Online GmbH (AS0) |    1 |   0.00% |  5.46 ms |  5.46 ms |  5.46 ms |  5.46 ms
Nuremberg, DE, EU, Hetzner Online GmbH (AS0)   |    1 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms
`
	assert.Equal(t, expectedTable, *table)

	assert.Equal(t, []*MeasurementStats{
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 100, Tsum: 0.77, Tsum2: 0.5929},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.46, Avg: 5.46, Max: 5.46, Time: 200, Tsum: 5.46, Tsum2: 29.8116},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.07, Avg: 4.07, Max: 4.07, Time: 300, Tsum: 4.07, Tsum2: 16.5649},
	}, stats)
}

func Test_GenerateTable_OneRow_Truncated(t *testing.T) {
	ctx := createDefaultContext("ping")
	ctx.AggregatedStats = []*MeasurementStats{
		NewMeasurementStats(),
		NewMeasurementStats(),
		NewMeasurementStats(),
	}
	hm := ctx.History.Find(measurementID1)
	viewer := &viewer{ctx: ctx}

	measurement := createPingMeasurement_MultipleProbes(measurementID1)
	measurement.Results[1].Probe.Network = "作者聚集的原创内容平台于201 1年1月正式上线让人们更"
	table, _, stats := viewer.generateTable(hm, measurement, 104)

	expectedTable := "\033[96mLocation                                    \033[0m | \033[96mSent\033[0m | \033[96m   Loss\033[0m | \033[96m    Last\033[0m | \033[96m     Min\033[0m | \033[96m     Avg\033[0m | \033[96m     Max\033[0m\n" +
		"London, GB, EU, OVH SAS (AS0)                |    1 |   0.00% |  0.77 ms |  0.77 ms |  0.77 ms |  0.77 ms\n" +
		"Falkenstein, DE, EU, 作者聚集的原创内容平... |    1 |   0.00% |  5.46 ms |  5.46 ms |  5.46 ms |  5.46 ms\n" +
		"Nuremberg, DE, EU, Hetzner Online GmbH (AS0) |    1 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms\n"
	assert.Equal(t, expectedTable, *table)

	assert.Equal(t, []*MeasurementStats{
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 100, Tsum: 0.77, Tsum2: 0.5929},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.46, Avg: 5.46, Max: 5.46, Time: 200, Tsum: 5.46, Tsum2: 29.8116},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.07, Avg: 4.07, Max: 4.07, Time: 300, Tsum: 4.07, Tsum2: 16.5649},
	}, stats)
}

func Test_GenerateTable_MultiLine_Truncated(t *testing.T) {
	ctx := createDefaultContext("ping")
	ctx.AggregatedStats = []*MeasurementStats{
		NewMeasurementStats(),
		NewMeasurementStats(),
		NewMeasurementStats(),
	}
	hm := ctx.History.Find(measurementID1)
	viewer := &viewer{ctx: ctx}

	measurement := createPingMeasurement_MultipleProbes(measurementID1)
	measurement.Results[1].Probe.Network = "Hetzner Online GmbH\nLorem ipsum\nLorem ipsum dolor sit amet"
	table, _, stats := viewer.generateTable(hm, measurement, 99)

	expectedTable := "\033[96mLocation                               \033[0m | \033[96mSent\033[0m | \033[96m   Loss\033[0m | \033[96m    Last\033[0m | \033[96m     Min\033[0m | \033[96m     Avg\033[0m | \033[96m     Max\033[0m\n" +
		"London, GB, EU, OVH SAS (AS0)           |    1 |   0.00% |  0.77 ms |  0.77 ms |  0.77 ms |  0.77 ms\n" +
		"Falkenstein, DE, EU, Hetzner Online ... |    1 |   0.00% |  5.46 ms |  5.46 ms |  5.46 ms |  5.46 ms\n" +
		"Lorem ipsum                             |      |         |          |          |          |         \n" +
		"Lorem ipsum dolor sit amet (AS0)        |      |         |          |          |          |         \n" +
		"Nuremberg, DE, EU, Hetzner Online Gm... |    1 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms\n"
	assert.Equal(t, expectedTable, *table)

	assert.Equal(t, []*MeasurementStats{
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 100, Tsum: 0.77, Tsum2: 0.5929},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.46, Avg: 5.46, Max: 5.46, Time: 200, Tsum: 5.46, Tsum2: 29.8116},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.07, Avg: 4.07, Max: 4.07, Time: 300, Tsum: 4.07, Tsum2: 16.5649},
	}, stats)
}

func Test_GenerateTable_MaxTruncated(t *testing.T) {
	ctx := createDefaultContext("ping")
	ctx.AggregatedStats = []*MeasurementStats{
		NewMeasurementStats(),
		NewMeasurementStats(),
		NewMeasurementStats(),
	}
	hm := ctx.History.Find(measurementID1)
	viewer := &viewer{ctx: ctx}

	measurement := createPingMeasurement_MultipleProbes(measurementID1)
	table, _, stats := viewer.generateTable(hm, measurement, 0)

	expectedTable := "\033[96mLoc...\033[0m | \033[96mSent\033[0m | \033[96m   Loss\033[0m | \033[96m    Last\033[0m | \033[96m     Min\033[0m | \033[96m     Avg\033[0m | \033[96m     Max\033[0m\n" +
		"Lon... |    1 |   0.00% |  0.77 ms |  0.77 ms |  0.77 ms |  0.77 ms\n" +
		"Fal... |    1 |   0.00% |  5.46 ms |  5.46 ms |  5.46 ms |  5.46 ms\n" +
		"Nur... |    1 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms\n"
	assert.Equal(t, expectedTable, *table)

	assert.Equal(t, []*MeasurementStats{
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 0.77, Min: 0.77, Avg: 0.77, Max: 0.77, Time: 100, Tsum: 0.77, Tsum2: 0.5929},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 5.46, Min: 5.46, Avg: 5.46, Max: 5.46, Time: 200, Tsum: 5.46, Tsum2: 29.8116},
		{Sent: 1, Rcv: 1, Lost: 0, Loss: 0, Last: 4.07, Min: 4.07, Avg: 4.07, Max: 4.07, Time: 300, Tsum: 4.07, Tsum2: 16.5649},
	}, stats)
}

func Test_GetRowValues_NoPacketsRcv(t *testing.T) {
	stats := &MeasurementStats{Sent: 1, Lost: 0, Loss: 0, Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1}
	rowValues := getRowValues(stats)
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
	stats := &MeasurementStats{
		Sent: 100,
		Lost: 10,
		Loss: 10,
		Last: 12.345,
		Min:  1.2345,
		Avg:  8.3456,
		Max:  123.4567,
	}
	rowValues := getRowValues(stats)
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
	ctx := createDefaultContext("ping")
	v := viewer{ctx: ctx}

	hm := ctx.History.Find(measurementID1)
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

	res := v.parsePingRawOutput(hm, m, -1)
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
		Time:  1002,
	}, res.Stats)
}

func Test_ParsePingRawOutput_NoStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime.Add(100 * time.Millisecond))

	ctx := createDefaultContext("ping")
	v := viewer{ctx: ctx, time: timeMock}

	hm := ctx.History.Find(measurementID1)

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
	res := v.parsePingRawOutput(hm, m, -1)
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
		Time:  100,
	}, res.Stats)
}

func Test_ParsePingRawOutput_NoStats_WithStartIncmpSeq(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime.Add(100 * time.Millisecond))

	ctx := createDefaultContext("ping")
	v := viewer{ctx: ctx, time: timeMock}

	hm := ctx.History.Find(measurementID1)

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
	res := v.parsePingRawOutput(hm, m, 4)
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
		Time:  100,
	}, res.Stats)
}

// Happens using --from AS58404
func Test_ParsePingRawOutput_WithRedirect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime.Add(100 * time.Millisecond))

	ctx := createDefaultContext("ping")
	v := viewer{ctx: ctx, time: timeMock}

	hm := ctx.History.Find(measurementID1)

	m := &globalping.ProbeMeasurement{
		Result: globalping.ProbeResult{
			RawOutput: `PING  (104.18.187.31) 56(84) bytes of data.
From goldenfast.net (103.102.153.1): icmp_seq=1 Redirect Host(New nexthop: goldenfast.net (43.252.139.2))
64 bytes from 104.18.187.31 (104.18.187.31): icmp_seq=1 ttl=59 time=0.558 ms
From goldenfast.net (103.102.153.1): icmp_seq=2 Redirect Host(New nexthop: goldenfast.net (43.252.139.2))
64 bytes from 104.18.187.31 (104.18.187.31): icmp_seq=2 ttl=59 time=0.705 ms`,
		},
	}
	res := v.parsePingRawOutput(hm, m, 0)
	assert.Equal(t, "104.18.187.31", res.Address)
	assert.Equal(t, "56(84)", res.BytesOfData)
	assert.Equal(t, []string{
		"From goldenfast.net (103.102.153.1): icmp_seq=1 Redirect Host(New nexthop: goldenfast.net (43.252.139.2))",
		"64 bytes from 104.18.187.31 (104.18.187.31): icmp_seq=1 ttl=59 time=0.558 ms",
		"From goldenfast.net (103.102.153.1): icmp_seq=2 Redirect Host(New nexthop: goldenfast.net (43.252.139.2))",
		"64 bytes from 104.18.187.31 (104.18.187.31): icmp_seq=2 ttl=59 time=0.705 ms",
	}, res.RawPacketLines)
	assert.Equal(t, []globalping.PingTiming{
		{RTT: 0.558, TTL: 59},
		{RTT: 0.705, TTL: 59},
	}, res.Timings)
	assertMeasurementStats(t, &MeasurementStats{
		Sent:  2,
		Rcv:   2,
		Lost:  0,
		Loss:  0,
		Last:  0.705,
		Min:   0.558,
		Avg:   0.6315,
		Max:   0.705,
		Tsum:  1.263,
		Tsum2: 0.8083,
		Mdev:  0.0735,
		Time:  100,
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
	assert.Equal(t, expected.Sent, actual.Sent, "Sent")
	assert.Equal(t, expected.Rcv, actual.Rcv, "Rcv")
	assert.Equal(t, expected.Lost, actual.Lost, "Lost")
	assert.InDelta(t, expected.Loss, actual.Loss, 0.0001, "Loss")
	assert.Equal(t, expected.Last, actual.Last, "Last")
	assert.Equal(t, expected.Min, actual.Min, "Min")
	assert.InDelta(t, expected.Avg, actual.Avg, 0.0001, "Avg")
	assert.Equal(t, expected.Max, actual.Max, "Max")
	assert.Equal(t, expected.Time, actual.Time, "Time")
	assert.InDelta(t, expected.Tsum, actual.Tsum, 0.0001, "Tsum")
	assert.InDelta(t, expected.Tsum2, actual.Tsum2, 0.0001, "Tsum2")
	assert.InDelta(t, expected.Mdev, actual.Mdev, 0.0001, "Mdev")
}
