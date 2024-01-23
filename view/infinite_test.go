package view

import (
	"encoding/json"
	"io"
	"math"
	"os"
	"testing"

	"github.com/jsdelivr/globalping-cli/model"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
)

func TestOutputInfinite_SingleLocation(t *testing.T) {
	osStdErr := os.Stderr
	osStdOut := os.Stdout

	rErr, wErr, err := os.Pipe()
	assert.NoError(t, err)
	defer rErr.Close()

	rOut, wOut, err := os.Pipe()
	assert.NoError(t, err)
	defer rOut.Close()

	os.Stderr = wErr
	os.Stdout = wOut

	defer func() {
		os.Stderr = osStdErr
		os.Stdout = osStdOut
	}()

	ctx := &model.Context{
		Cmd: "ping",
	}
	measurement := getPingGetMeasurement(MeasurementID1)

	err = outputSingleLocation(measurement, ctx)
	assert.NoError(t, err)

	err = outputSingleLocation(measurement, ctx)
	assert.NoError(t, err)
	err = outputSingleLocation(measurement, ctx)
	assert.NoError(t, err)

	wErr.Close()
	wOut.Close()

	errOutput, err := io.ReadAll(rErr)
	assert.NoError(t, err)
	assert.Equal(t, "", string(errOutput))

	output, err := io.ReadAll(rOut)
	assert.NoError(t, err)
	assert.Equal(t,
		`> EU, DE, Berlin, ASN:3320, Deutsche Telekom AG
PING cdn.jsdelivr.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=60 time=17.64 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=2 ttl=60 time=17.64 ms
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=3 ttl=60 time=17.64 ms
`,
		string(output))
}

func TestOutputInfinite_MultipleLocations(t *testing.T) {
	osStdOut := os.Stdout
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	os.Stdout = w

	ctx := &model.Context{
		Cmd: "ping",
	}
	measurement := getPingGetMeasurementMultipleLocations(MeasurementID1)

	err = outputMultipleLocations(measurement, ctx)
	assert.NoError(t, err)

	w.Close()
	os.Stdout = osStdOut

	output, err := io.ReadAll(r)
	assert.NoError(t, err)
	r.Close()

	r, w, err = os.Pipe()
	assert.NoError(t, err)
	defer r.Close()
	os.Stdout = w
	defer func() {
		os.Stdout = osStdOut
	}()

	expectedCtx := getDefaultPingCtx(len(measurement.Results))
	expectedTable := generateTable(measurement, expectedCtx, 76) // 80 - 4. pterm defaults to 80 when terminal size is not detected.
	area, err := pterm.DefaultArea.Start()
	assert.NoError(t, err)
	area.Update(expectedTable)
	area.Stop()
	w.Close()
	os.Stdout = osStdOut

	expectedOutput, err := io.ReadAll(r)
	assert.NoError(t, err)
	r.Close()

	assert.Equal(t, string(expectedOutput), string(output))
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
	measurement := getPingGetMeasurementMultipleLocations(MeasurementID1)
	ctx := getDefaultPingCtx(len(measurement.Results))
	expectedTable := "\x1b[96m\x1b[96mLocation                                       \x1b[0m\x1b[0m | \x1b[96m\x1b[96mSent\x1b[0m\x1b[0m | \x1b[96m\x1b[96m   Loss\x1b[0m\x1b[0m | \x1b[96m\x1b[96m    Last\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Min\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Avg\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Max\x1b[0m\x1b[0m\n" +
		"EU, GB, London, ASN:0, OVH SAS                  |    1 |   0.00% |  0.77 ms |  0.77 ms |  0.77 ms |  0.77 ms\n" +
		"EU, DE, Falkenstein, ASN:0, Hetzner Online GmbH |    1 |   0.00% |  5.46 ms |  5.46 ms |  5.46 ms |  5.46 ms\n" +
		"EU, DE, Nuremberg, ASN:0, Hetzner Online GmbH   |    1 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms\n"
	table := generateTable(measurement, ctx, 500)
	assert.Equal(t, expectedTable, table)
}

func TestGenerateTableOneRowTruncated(t *testing.T) {
	measurement := getPingGetMeasurementMultipleLocations(MeasurementID1)
	measurement.Results[1].Probe.Network = "作者聚集的原创内容平台于201 1年1月正式上线让人们更"
	ctx := getDefaultPingCtx(len(measurement.Results))
	expectedTable := "\x1b[96m\x1b[96mLocation                                      \x1b[0m\x1b[0m | \x1b[96m\x1b[96mSent\x1b[0m\x1b[0m | \x1b[96m\x1b[96m   Loss\x1b[0m\x1b[0m | \x1b[96m\x1b[96m    Last\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Min\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Avg\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Max\x1b[0m\x1b[0m\n" +
		"EU, GB, London, ASN:0, OVH SAS                 |    1 |   0.00% |  0.77 ms |  0.77 ms |  0.77 ms |  0.77 ms\n" +
		"EU, DE, Falkenstein, ASN:0, 作者聚集的原创...  |    1 |   0.00% |  5.46 ms |  5.46 ms |  5.46 ms |  5.46 ms\n" +
		"EU, DE, Nuremberg, ASN:0, Hetzner Online GmbH  |    1 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms\n"
	table := generateTable(measurement, ctx, 106)
	assert.Equal(t, expectedTable, table)
}

func TestGenerateTableMultiLineTruncated(t *testing.T) {
	measurement := getPingGetMeasurementMultipleLocations(MeasurementID1)
	measurement.Results[1].Probe.Network = "Hetzner Online GmbH\nLorem ipsum\nLorem ipsum dolor sit amet"
	ctx := getDefaultPingCtx(len(measurement.Results))
	expectedTable := "\x1b[96m\x1b[96mLocation                                      \x1b[0m\x1b[0m | \x1b[96m\x1b[96mSent\x1b[0m\x1b[0m | \x1b[96m\x1b[96m   Loss\x1b[0m\x1b[0m | \x1b[96m\x1b[96m    Last\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Min\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Avg\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Max\x1b[0m\x1b[0m\n" +
		"EU, GB, London, ASN:0, OVH SAS                 |    1 |   0.00% |  0.77 ms |  0.77 ms |  0.77 ms |  0.77 ms\n" +
		"EU, DE, Falkenstein, ASN:0, Hetzner Online ... |    1 |   0.00% |  5.46 ms |  5.46 ms |  5.46 ms |  5.46 ms\n" +
		"Lorem ipsum                                    |      |         |          |          |          |         \n" +
		"Lorem ipsum dolor sit amet                     |      |         |          |          |          |         \n" +
		"EU, DE, Nuremberg, ASN:0, Hetzner Online GmbH  |    1 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms\n"
	table := generateTable(measurement, ctx, 106)
	assert.Equal(t, expectedTable, table)
}

func TestGenerateTableMaxTruncated(t *testing.T) {
	measurement := getPingGetMeasurementMultipleLocations(MeasurementID1)
	ctx := getDefaultPingCtx(len(measurement.Results))
	expectedTable := "\x1b[96m\x1b[96mLoc...\x1b[0m\x1b[0m | \x1b[96m\x1b[96mSent\x1b[0m\x1b[0m | \x1b[96m\x1b[96m   Loss\x1b[0m\x1b[0m | \x1b[96m\x1b[96m    Last\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Min\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Avg\x1b[0m\x1b[0m | \x1b[96m\x1b[96m     Max\x1b[0m\x1b[0m\n" +
		"EU,... |    1 |   0.00% |  0.77 ms |  0.77 ms |  0.77 ms |  0.77 ms\n" +
		"EU,... |    1 |   0.00% |  5.46 ms |  5.46 ms |  5.46 ms |  5.46 ms\n" +
		"EU,... |    1 |   0.00% |  4.07 ms |  4.07 ms |  4.07 ms |  4.07 ms\n"
	table := generateTable(measurement, ctx, 0)
	assert.Equal(t, expectedTable, table)
}

func TestUpdateMeasurementStats(t *testing.T) {
	stats := model.MeasurementStats{
		Sent: 2,
		Lost: 0,
		Loss: 0,
		Last: 1,
		Min:  1,
		Avg:  1.5,
		Max:  2,
	}
	result := model.MeasurementResponse{
		Result: model.ResultData{
			StatsRaw:   json.RawMessage(`{"min":6,"avg":6,"max":6,"total":1,"rcv":1,"drop":0,"loss":0}`),
			TimingsRaw: json.RawMessage(`[{"ttl":60,"rtt":6}]`),
		},
	}
	err := updateMeasurementStats(&stats, &result)
	assert.NoError(t, err)
	assert.Equal(t, model.MeasurementStats{
		Sent: 3,
		Lost: 0,
		Loss: 0,
		Last: 6,
		Min:  1,
		Avg:  3,
		Max:  6,
	}, stats)
	result = model.MeasurementResponse{
		Result: model.ResultData{
			StatsRaw:   json.RawMessage(`{"min":0,"avg":0,"max":0,"total":1,"rcv":0,"drop":1,"loss":100}`),
			TimingsRaw: json.RawMessage(`[]`),
		},
	}
	err = updateMeasurementStats(&stats, &result)
	assert.NoError(t, err)
	assert.Equal(t, model.MeasurementStats{
		Sent: 4,
		Lost: 1,
		Loss: 25,
		Last: 6,
		Min:  1,
		Avg:  3,
		Max:  6,
	}, stats)
}

func TestGetRowValuesNoPacketsRcv(t *testing.T) {
	stats := model.MeasurementStats{
		Sent: 1,
		Lost: -1,
		Loss: 0,
		Last: -1,
		Min:  math.MaxFloat64,
		Avg:  -1,
		Max:  -1,
	}
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
