package view

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/jsdelivr/globalping-cli/model"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
)

func TestOutputInfinite_SingleLocation(t *testing.T) {
	osStdErr := os.Stderr
	osStdOut := os.Stdout

	rStdErr, myStdErr, err := os.Pipe()
	assert.NoError(t, err)
	defer rStdErr.Close()

	rStdOut, myStdOut, err := os.Pipe()
	assert.NoError(t, err)
	defer rStdOut.Close()

	os.Stderr = myStdErr
	os.Stdout = myStdOut

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

	myStdErr.Close()
	myStdOut.Close()

	errOutput, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "", string(errOutput))

	output, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t,
		`> EU, DE, Berlin, ASN:3320, Deutsche Telekom AG
PING cdn.jsdelivr.net (151.101.1.229)
151.101.1.229: icmp_seq=1 ttl=60 time=17.64 ms
151.101.1.229: icmp_seq=2 ttl=60 time=17.64 ms
151.101.1.229: icmp_seq=3 ttl=60 time=17.64 ms
`,
		string(output))
}

func TestOutputInfinite_MultipleLocations(t *testing.T) {
	osStdErr := os.Stderr
	osStdOut := os.Stdout

	rStdErr, myStdErr, err := os.Pipe()
	assert.NoError(t, err)
	defer rStdErr.Close()

	rStdOut, myStdOut, err := os.Pipe()
	assert.NoError(t, err)
	defer rStdOut.Close()

	os.Stderr = myStdErr
	os.Stdout = myStdOut

	defer func() {
		os.Stderr = osStdErr
		os.Stdout = osStdOut
	}()

	ctx := &model.Context{
		Cmd: "ping",
	}
	measurement := getPingGetMeasurementMultipleLocations(MeasurementID1)

	err = outputMultipleLocations(measurement, ctx)
	assert.NoError(t, err)

	myStdErr.Close()
	myStdOut.Close()

	errOutput, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "", string(errOutput))

	output, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)

	expectedTableData := pterm.TableData{
		{"Location", "Loss", "Sent", "Last", "Avg", "Min", "Max"},
		{"EU, GB, London, ASN:0, OVH SAS", "0.00%", "1", "0.77 ms", "0.77 ms", "0.77 ms", "0.77 ms"},
		{"EU, DE, Falkenstein, ASN:0, Hetzner Online GmbH", "0.00%", "1", "5.46 ms", "5.46 ms", "5.46 ms", "5.46 ms"},
		{"EU, DE, Nuremberg, ASN:0, Hetzner Online GmbH", "0.00%", "1", "4.07 ms", "4.07 ms", "4.07 ms", "4.07 ms"},
	}
	expectedTable, _ := pterm.DefaultTable.WithHasHeader().WithData(expectedTableData).Srender()
	assert.Equal(t, "\n\n\n\n\n\n\n"+expectedTable+"\n", string(output))
}

func TestFormatDuration(t *testing.T) {
	d := formatDuration(1.2345)
	assert.Equal(t, "1.23 ms", d)
	d = formatDuration(12.345)
	assert.Equal(t, "12.3 ms", d)
	d = formatDuration(123.4567)
	assert.Equal(t, "123 ms", d)
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
