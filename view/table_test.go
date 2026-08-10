package view

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/jsdelivr/globalping-go"
	"github.com/mattn/go-runewidth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_OutputTable_Ping_Success(t *testing.T) {
	measurement := createPingMeasurement_MultipleProbes(measurementID1)

	output, err := renderTableForTest(t, measurement, false)

	require.NoError(t, err)
	assertTableForTest(t, output, [][]string{
		{"Location", "Sent", "Loss", "Last", "Min", "Avg", "Max"},
		{"London, GB, EU, OVH SAS (AS0)", "1", "0.00%", "0.77 ms", "0.77 ms", "0.77 ms", "0.77 ms"},
		{"Falkenstein, DE, EU, Hetzner Online GmbH (AS0)", "1", "0.00%", "5.46 ms", "5.46 ms", "5.46 ms", "5.46 ms"},
		{"Nuremberg, DE, EU, Hetzner Online GmbH (AS0)", "1", "0.00%", "4.07 ms", "4.07 ms", "4.07 ms", "4.07 ms"},
	})
}

func Test_OutputTable_Ping_PartialOfflineUsesPlaceholders(t *testing.T) {
	measurement := createPingMeasurement(measurementID1)
	measurement.Results = append(measurement.Results, tableProbe("Paris", "FR", "Offline Network", globalping.TestStatusOffline))
	measurement.ProbesCount = len(measurement.Results)

	output, err := renderTableForTest(t, measurement, false)

	require.NoError(t, err)
	assertTableForTest(t, output, [][]string{
		{"Location", "Sent", "Loss", "Last", "Min", "Avg", "Max"},
		{"Berlin, DE, EU, Deutsche Telekom AG (AS3320)", "1", "0.00%", "17.6 ms", "17.6 ms", "17.6 ms", "17.6 ms"},
		{"Paris, FR, EU, Offline Network (AS64500)", "-", "-", "-", "-", "-", "-"},
	})
}

func Test_OutputTable_Ping_PartialFailureShowsFailureDetails(t *testing.T) {
	measurement := createPingMeasurement(measurementID1)
	failed := tableProbe("Paris", "FR", "Failed Network", globalping.TestStatusFailed)
	failed.Result.FailureSource = globalping.FailureSourceResolver
	failed.Result.RawOutput = "\r\n \r\n resolver lookup failed \r\nignored"
	measurement.Results = append(measurement.Results, failed)
	measurement.ProbesCount = len(measurement.Results)

	output, ctx, err := renderTableWithContextForTest(t, measurement, false, 0)

	require.NoError(t, err)
	assertTableWithFailureForTest(t, output, [][]string{
		{"Location", "Sent", "Loss", "Last", "Min", "Avg", "Max"},
		{"Berlin, DE, EU, Deutsche Telekom AG (AS3320)", "1", "0.00%", "17.6 ms", "17.6 ms", "17.6 ms", "17.6 ms"},
	}, "Paris, FR, EU, Failed Network (AS64500)", "Resolver error: resolver lookup failed")
	assert.Equal(t, NewMeasurementStats(), ctx.AggregatedStats[1])
	assert.Equal(t, NewMeasurementStats(), ctx.History.Find(measurement.ID).Stats[1])
}

func Test_FailureTableMessage_MapsFailureSources(t *testing.T) {
	for _, test := range []struct {
		name          string
		failureSource globalping.FailureSource
		expected      string
	}{
		{name: "target", failureSource: globalping.FailureSourceTarget, expected: "Target error: details"},
		{name: "resolver", failureSource: globalping.FailureSourceResolver, expected: "Resolver error: details"},
		{name: "internal", failureSource: globalping.FailureSourceInternal, expected: "Internal error: details"},
		{name: "null", expected: "Error: details"},
		{name: "unknown", failureSource: globalping.FailureSource("other"), expected: "Error: details"},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := globalping.ProbeResult{FailureSource: test.failureSource, RawOutput: "details"}

			assert.Equal(t, test.expected, failureTableMessage(&result))
		})
	}
}

func Test_FailureTableMessage_SelectsFirstNonEmptyLine(t *testing.T) {
	for _, test := range []struct {
		name      string
		rawOutput string
		expected  string
	}{
		{name: "LF", rawOutput: "\n \n first line \nsecond line", expected: "Error: first line"},
		{name: "CRLF", rawOutput: "\r\n\t\r\n first line \r\nsecond line", expected: "Error: first line"},
		{name: "empty", expected: "Error"},
		{name: "whitespace", rawOutput: " \r\n\t\n  ", expected: "Error"},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := globalping.ProbeResult{RawOutput: test.rawOutput}

			assert.Equal(t, test.expected, failureTableMessage(&result))
		})
	}
}

func Test_OutputTable_DNS_MixedSuccessAndFailure(t *testing.T) {
	success := tableProbe("Berlin", "DE", "DNS Network", globalping.TestStatusFinished)
	success.Result.StatusCodeName = "NOERROR"
	success.Result.AnswersRaw = json.RawMessage(`[]`)
	success.Result.TimingsRaw = json.RawMessage(`{"total":1}`)
	success.Result.Resolver = "192.0.2.53"
	failed := tableProbe("Paris", "FR", "Failed Network", globalping.TestStatusFailed)
	failed.Result.FailureSource = globalping.FailureSourceTarget
	failed.Result.RawOutput = "connection refused"

	output, err := renderTableForTest(t, tableMeasurement("dns", success, failed), false)

	require.NoError(t, err)
	assertTableWithFailureForTest(t, output, [][]string{
		{"Location", "Status", "Answers", "Time", "Resolver"},
		{"Berlin, DE, EU, DNS Network (AS64500)", "NOERROR", "0", "1 ms", "192.0.2.53"},
	}, "Paris, FR, EU, Failed Network (AS64500)", "Target error: connection refused")
}

func Test_GenerateMeasurementTable_TruncatesFailureToWidth(t *testing.T) {
	success := tableProbe("Berlin", "DE", "DNS Network", globalping.TestStatusFinished)
	success.Result.StatusCodeName = "NOERROR"
	failed := tableProbe("Paris", "FR", "Failed Network", globalping.TestStatusFailed)
	failed.Result.FailureSource = globalping.FailureSourceInternal
	failed.Result.RawOutput = "世界世界世界世界世界世界世界世界世界世界 long failure details"
	measurement := tableMeasurement("dns", success, failed)
	printer := NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer))
	printer.DisableStyling()
	v := &viewer{ctx: createDefaultContext("dns"), printer: printer}

	output := v.generateMeasurementTable(measurement, 56)
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")

	require.Len(t, lines, 3)
	assert.LessOrEqual(t, runewidth.StringWidth(lines[2]), 56)
	assert.Equal(t, 1, strings.Count(lines[2], colSeparator))
	assert.Contains(t, lines[2], "Internal error:")
	assert.True(t, strings.HasSuffix(strings.TrimSpace(lines[2]), "..."))
}

func Test_OutputTable_Traceroute(t *testing.T) {
	success := tableProbe("Berlin", "DE", "Trace Network", globalping.TestStatusFinished)
	success.Result.HopsRaw = json.RawMessage(`[
		{"resolvedAddress":"192.0.2.1","timings":[{"rtt":9}]},
		{"resolvedAddress":"192.0.2.2","timings":[{"rtt":8}]},
		{"resolvedAddress":"192.0.2.3","timings":[{"rtt":1.2},{"rtt":2.345},{"rtt":100.4}]}
	]`)
	offline := tableProbe("Paris", "FR", "Offline Network", globalping.TestStatusOffline)
	measurement := tableMeasurement("traceroute", success, offline)

	output, err := renderTableForTest(t, measurement, false)

	require.NoError(t, err)
	assertTableForTest(t, output, [][]string{
		{"Location", "Hops", "Last", "Min", "Avg", "Max"},
		{"Berlin, DE, EU, Trace Network (AS64500)", "3", "100 ms", "1.20 ms", "34.6 ms", "100 ms"},
		{"Paris, FR, EU, Offline Network (AS64500)", "-", "-", "-", "-", "-"},
	})
}

func Test_OutputTable_Traceroute_LastSkipsTrailingTimeouts(t *testing.T) {
	result := tableProbe("Berlin", "DE", "Trace Network", globalping.TestStatusFinished)
	result.Result.HopsRaw = json.RawMessage(`[
		{"timings":[{"rtt":1.5},{"rtt":8.25},{"rtt":null}]}
	]`)

	output, err := renderTableForTest(t, tableMeasurement("traceroute", result), false)

	require.NoError(t, err)
	assertTableForTest(t, output, [][]string{
		{"Location", "Hops", "Last", "Min", "Avg", "Max"},
		{"Berlin, DE, EU, Trace Network (AS64500)", "1", "8.25 ms", "1.50 ms", "4.88 ms", "8.25 ms"},
	})
}

func Test_OutputTable_Traceroute_EmptyHops(t *testing.T) {
	result := tableProbe("Berlin", "DE", "Trace Network", globalping.TestStatusFinished)
	result.Result.HopsRaw = json.RawMessage(`[]`)

	output, err := renderTableForTest(t, tableMeasurement("traceroute", result), false)

	require.NoError(t, err)
	assertTableForTest(t, output, [][]string{
		{"Location", "Hops", "Last", "Min", "Avg", "Max"},
		{"Berlin, DE, EU, Trace Network (AS64500)", "0", "-", "-", "-", "-"},
	})
}

func Test_OutputTable_Traceroute_MalformedDecodedFieldsUsePlaceholders(t *testing.T) {
	result := tableProbe("Berlin", "DE", "Trace Network", globalping.TestStatusFinished)
	result.Result.HopsRaw = json.RawMessage(`{"not":"an array"}`)

	output, err := renderTableForTest(t, tableMeasurement("traceroute", result), false)

	require.NoError(t, err)
	assertTableForTest(t, output, [][]string{
		{"Location", "Hops", "Last", "Min", "Avg", "Max"},
		{"Berlin, DE, EU, Trace Network (AS64500)", "-", "-", "-", "-", "-"},
	})
}

func Test_OutputTable_MTR(t *testing.T) {
	success := tableProbe("Berlin", "DE", "MTR Network", globalping.TestStatusFinished)
	success.Result.HopsRaw = json.RawMessage(`[
		{"resolvedAddress":"192.0.2.1","stats":{"min":2,"avg":3,"max":4},"timings":[{"rtt":3}]},
		{"resolvedAddress":"192.0.2.2","stats":{"min":1.2345,"avg":8.3456,"max":123.4567},"timings":[{"rtt":5.5},{"rtt":12.345}]}
	]`)
	offline := tableProbe("Paris", "FR", "Offline Network", globalping.TestStatusOffline)
	measurement := tableMeasurement("mtr", success, offline)

	output, err := renderTableForTest(t, measurement, false)

	require.NoError(t, err)
	assertTableForTest(t, output, [][]string{
		{"Location", "Hops", "Last", "Min", "Avg", "Max"},
		{"Berlin, DE, EU, MTR Network (AS64500)", "2", "12.3 ms", "1.23 ms", "8.35 ms", "123 ms"},
		{"Paris, FR, EU, Offline Network (AS64500)", "-", "-", "-", "-", "-"},
	})
}

func Test_OutputTable_MTR_LastSkipsTrailingTimeouts(t *testing.T) {
	result := tableProbe("Berlin", "DE", "MTR Network", globalping.TestStatusFinished)
	result.Result.HopsRaw = json.RawMessage(`[
		{"stats":{"min":1,"avg":2,"max":3},"timings":[{"rtt":7.5},{"rtt":null}]}
	]`)

	output, err := renderTableForTest(t, tableMeasurement("mtr", result), false)

	require.NoError(t, err)
	assertTableForTest(t, output, [][]string{
		{"Location", "Hops", "Last", "Min", "Avg", "Max"},
		{"Berlin, DE, EU, MTR Network (AS64500)", "1", "7.50 ms", "1.00 ms", "2.00 ms", "3.00 ms"},
	})
}

func Test_OutputTable_MTR_EmptyHops(t *testing.T) {
	result := tableProbe("Berlin", "DE", "MTR Network", globalping.TestStatusFinished)
	result.Result.HopsRaw = json.RawMessage(`[]`)

	output, err := renderTableForTest(t, tableMeasurement("mtr", result), false)

	require.NoError(t, err)
	assertTableForTest(t, output, [][]string{
		{"Location", "Hops", "Last", "Min", "Avg", "Max"},
		{"Berlin, DE, EU, MTR Network (AS64500)", "0", "-", "-", "-", "-"},
	})
}

func Test_OutputTable_MTR_MissingDecodedFieldsUsePlaceholders(t *testing.T) {
	result := tableProbe("Berlin", "DE", "MTR Network", globalping.TestStatusFinished)

	output, err := renderTableForTest(t, tableMeasurement("mtr", result), false)

	require.NoError(t, err)
	assertTableForTest(t, output, [][]string{
		{"Location", "Hops", "Last", "Min", "Avg", "Max"},
		{"Berlin, DE, EU, MTR Network (AS64500)", "-", "-", "-", "-", "-"},
	})
}

func Test_OutputTable_DNS(t *testing.T) {
	success := tableProbe("Berlin", "DE", "DNS Network", globalping.TestStatusFinished)
	success.Result.StatusCode = 0
	success.Result.StatusCodeName = "NOERROR"
	success.Result.AnswersRaw = json.RawMessage(`[
		{"name":"example.com.","type":"A","ttl":60,"class":"IN","value":"192.0.2.1"},
		{"name":"example.com.","type":"A","ttl":60,"class":"IN","value":"192.0.2.2"}
	]`)
	success.Result.TimingsRaw = json.RawMessage(`{"total":4.567}`)
	success.Result.Resolver = "1.1.1.1"
	empty := tableProbe("London", "GB", "Empty DNS Network", globalping.TestStatusFinished)
	empty.Result.StatusCode = 3
	empty.Result.AnswersRaw = json.RawMessage(`[]`)
	empty.Result.TimingsRaw = json.RawMessage(`{"total":0}`)
	offline := tableProbe("Paris", "FR", "Offline Network", globalping.TestStatusOffline)
	measurement := tableMeasurement("dns", success, empty, offline)

	output, err := renderTableForTest(t, measurement, false)

	require.NoError(t, err)
	assertTableForTest(t, output, [][]string{
		{"Location", "Status", "Answers", "Time", "Resolver"},
		{"Berlin, DE, EU, DNS Network (AS64500)", "NOERROR", "2", "4.57 ms", "1.1.1.1"},
		{"London, GB, EU, Empty DNS Network (AS64500)", "3", "0", "0 ms", "-"},
		{"Paris, FR, EU, Offline Network (AS64500)", "-", "-", "-", "-"},
	})
}

func Test_OutputTable_DNS_MalformedDecodedFieldsUsePlaceholders(t *testing.T) {
	result := tableProbe("Berlin", "DE", "DNS Network", globalping.TestStatusFinished)
	result.Result.StatusCodeName = "SERVFAIL"
	result.Result.AnswersRaw = json.RawMessage(`{"not":"an array"}`)
	result.Result.TimingsRaw = json.RawMessage(`[]`)

	output, err := renderTableForTest(t, tableMeasurement("dns", result), false)

	require.NoError(t, err)
	assertTableForTest(t, output, [][]string{
		{"Location", "Status", "Answers", "Time", "Resolver"},
		{"Berlin, DE, EU, DNS Network (AS64500)", "SERVFAIL", "-", "-", "-"},
	})
}

func Test_OutputTable_DNSTraceUsesFinalHopAndOmitsStatus(t *testing.T) {
	success := tableProbe("Berlin", "DE", "DNS Trace Network", globalping.TestStatusFinished)
	success.Result.HopsRaw = json.RawMessage(`[
		{"resolver":"root.example","answers":[],"timings":{"total":1}},
		{"resolver":"final.example","answers":[{"name":"example.com.","type":"A","ttl":60,"class":"IN","value":"192.0.2.1"}],"timings":{"total":9.876}}
	]`)
	malformed := tableProbe("Paris", "FR", "Malformed Trace Network", globalping.TestStatusFinished)
	malformed.Result.HopsRaw = json.RawMessage(`{"not":"an array"}`)
	measurement := tableMeasurement("dns", success, malformed)

	output, err := renderTableForTest(t, measurement, true)

	require.NoError(t, err)
	assertTableForTest(t, output, [][]string{
		{"Location", "Answers", "Time", "Resolver"},
		{"Berlin, DE, EU, DNS Trace Network (AS64500)", "1", "9.88 ms", "final.example"},
		{"Paris, FR, EU, Malformed Trace Network (AS64500)", "-", "-", "-"},
	})
}

func Test_OutputTable_HTTP(t *testing.T) {
	validString := tableProbe("Berlin", "DE", "HTTP Network", globalping.TestStatusFinished)
	validString.Result.StatusCode = 200
	validString.Result.StatusCodeName = "OK"
	validString.Result.HeadersRaw = json.RawMessage(`{"CoNtEnT-LeNgTh":" 00123 "}`)
	validString.Result.TimingsRaw = json.RawMessage(`{"total":44}`)
	validString.Result.ResolvedAddress = "192.0.2.10"

	validNumber := tableProbe("London", "GB", "Numeric Header Network", globalping.TestStatusFinished)
	validNumber.Result.StatusCode = 204
	validNumber.Result.StatusCodeName = "No Content"
	validNumber.Result.HeadersRaw = json.RawMessage(`{"content-length":42}`)
	validNumber.Result.TimingsRaw = json.RawMessage(`{"total":5}`)
	validNumber.Result.ResolvedAddress = "192.0.2.11"

	missing := tableProbe("Paris", "FR", "Missing Header Network", globalping.TestStatusFinished)
	missing.Result.StatusCode = 304
	missing.Result.HeadersRaw = json.RawMessage(`{"server":"example"}`)
	missing.Result.RawHeaders = "content-length: 999"
	missing.Result.RawBody = "this body length must not be used"
	missing.Result.TimingsRaw = json.RawMessage(`{"total":100}`)
	missing.Result.ResolvedAddress = "192.0.2.12"

	invalid := tableProbe("Prague", "CZ", "Invalid Header Network", globalping.TestStatusFinished)
	invalid.Result.StatusCode = 500
	invalid.Result.StatusCodeName = "Internal Server Error"
	invalid.Result.HeadersRaw = json.RawMessage(`{"CONTENT-LENGTH":"-4"}`)
	invalid.Result.TimingsRaw = json.RawMessage(`{"total":1}`)
	invalid.Result.ResolvedAddress = "192.0.2.13"

	malformed := tableProbe("Rome", "IT", "Malformed Header Network", globalping.TestStatusFinished)
	malformed.Result.StatusCode = 200
	malformed.Result.StatusCodeName = "OK"
	malformed.Result.HeadersRaw = json.RawMessage(`{"content-length":["12"]}`)
	malformed.Result.TimingsRaw = json.RawMessage(`not-json`)

	offline := tableProbe("Madrid", "ES", "Offline Network", globalping.TestStatusOffline)
	measurement := tableMeasurement("http", validString, validNumber, missing, invalid, malformed, offline)

	output, err := renderTableForTest(t, measurement, false)

	require.NoError(t, err)
	assertTableForTest(t, output, [][]string{
		{"Location", "Status", "Content-Length", "Total", "Resolved IP"},
		{"Berlin, DE, EU, HTTP Network (AS64500)", "200 OK", "123 B", "44 ms", "192.0.2.10"},
		{"London, GB, EU, Numeric Header Network (AS64500)", "204 No Content", "42 B", "5 ms", "192.0.2.11"},
		{"Paris, FR, EU, Missing Header Network (AS64500)", "304", "-", "100 ms", "192.0.2.12"},
		{"Prague, CZ, EU, Invalid Header Network (AS64500)", "500 Internal Server Error", "-", "1 ms", "192.0.2.13"},
		{"Rome, IT, EU, Malformed Header Network (AS64500)", "200 OK", "-", "-", "-"},
		{"Madrid, ES, EU, Offline Network (AS64500)", "-", "-", "-", "-"},
	})
}

func Test_OutputTable_HTTP_SizeColumnWithoutContentLength(t *testing.T) {
	crlf := tableProbe("Berlin", "DE", "HTTP Network", globalping.TestStatusFinished)
	crlf.Result.StatusCodeName = "Connection failed"
	crlf.Result.RawOutput = "HTTP/1.1 200 OK\r\nServer: example\r\n\r\nhello"

	lf := tableProbe("Paris", "FR", "Unicode Body Network", globalping.TestStatusFinished)
	lf.Result.StatusCode = 200
	lf.Result.StatusCodeName = "OK"
	lf.Result.RawOutput = "HTTP/1.1 200 OK\nServer: example\n\nžluťoučký"
	lf.Result.Truncated = true

	empty := tableProbe("Rome", "IT", "Empty Body Network", globalping.TestStatusFinished)
	empty.Result.StatusCode = 204
	empty.Result.RawOutput = "HTTP/1.1 204 No Content\r\nServer: example\r\n\r\n"

	measurement := tableMeasurement("http", crlf, lf, empty)
	output, err := renderTableForTest(t, measurement, false)

	require.NoError(t, err)
	assertTableForTest(t, output, [][]string{
		{"Location", "Status", "Bytes", "Total", "Resolved IP"},
		{"Berlin, DE, EU, HTTP Network (AS64500)", "Connection failed", "5 B", "-", "-"},
		{"Paris, FR, EU, Unicode Body Network (AS64500)", "200 OK", "13+ B", "-", "-"},
		{"Rome, IT, EU, Empty Body Network (AS64500)", "204", "0 B", "-", "-"},
	})

	noBody := tableProbe("London", "GB", "No Body Network", globalping.TestStatusFinished)
	noBody.Result.StatusCode = 204
	noBody.Result.RawOutput = "HTTP/1.1 204 No Content\r\nServer: example"
	output, err = renderTableForTest(t, tableMeasurement("http", noBody), false)

	require.NoError(t, err)
	assertTableForTest(t, output, [][]string{
		{"Location", "Status", "Total", "Resolved IP"},
		{"London, GB, EU, No Body Network (AS64500)", "204", "-", "-"},
	})
}

func Test_OutputTable_HTTP_EmptyBodyKeepsBytesColumn(t *testing.T) {
	empty := tableProbe("Rome", "IT", "Empty Body Network", globalping.TestStatusFinished)
	empty.Result.StatusCode = 204
	empty.Result.RawOutput = "HTTP/1.1 204 No Content\r\nServer: example\r\n\r\n"

	output, err := renderTableForTest(t, tableMeasurement("http", empty), false)

	require.NoError(t, err)
	assertTableForTest(t, output, [][]string{
		{"Location", "Status", "Bytes", "Total", "Resolved IP"},
		{"Rome, IT, EU, Empty Body Network (AS64500)", "204", "0 B", "-", "-"},
	})
}

func Test_RenderMeasurementTable_UnicodeCellsAlignByDisplayWidth(t *testing.T) {
	rows := [][]string{
		{"Location", "Status", "Resolver"},
		{"Tokyo, JP, AS64500, 東京網", "NOERROR", "一.example"},
		{"Berlin, DE, AS64500, Network", "SERVFAIL", "dns.example"},
	}
	printer := NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer))
	printer.DisableStyling()
	v := &viewer{printer: printer}

	output := v.renderMeasurementTable(rows, 120, "dns")
	assertDisplayTableLayoutForTest(t, output, 120)
}

func Test_RenderMeasurementTable_TruncatesOnlyAllowedFields(t *testing.T) {
	rows := [][]string{
		{"Location", "Status", "Content-Length", "Total", "Resolved IP"},
		{"A very long probe location in a very long network", "599 An unusually long status name", "123456789 B", "123 ms", "2001:db8:ffff:ffff:ffff:ffff:ffff:ffff"},
		{"Tokyo 東京", "200 OK", "42 B", "1.00 ms", "192.0.2.1"},
	}
	printer := NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer))
	v := &viewer{printer: printer}

	statusOnly := v.renderMeasurementTable(rows, 130, "http")
	statusOnlyLines := strings.Split(strings.TrimSpace(stripANSIForTest(statusOnly)), "\n")
	statusOnlyCells := trimCellsForTest(strings.Split(statusOnlyLines[1], colSeparator))
	assert.Equal(t, rows[1][0], statusOnlyCells[0])
	assert.Equal(t, "599", statusOnlyCells[1])
	assert.Equal(t, rows[1][4], statusOnlyCells[4])

	wide := v.renderMeasurementTable(rows, 80, "http")
	assert.Contains(t, wide, "\x1b[96m")
	wideLines := strings.Split(strings.TrimSpace(stripANSIForTest(wide)), "\n")
	wantOffsets := displaySeparatorOffsetsForTest(wideLines[0])
	for _, line := range wideLines {
		assert.Equal(t, wantOffsets, displaySeparatorOffsetsForTest(line))
	}
	assert.LessOrEqual(t, runewidth.StringWidth(wideLines[1]), 80)
	wideHeader := strings.Split(wideLines[0], colSeparator)
	assert.Equal(t, []string{"Location", "Status", "Content-Length", "Total", "Resolved IP"}, trimCellsForTest(wideHeader))
	assert.Contains(t, wideLines[1], "599")
	assert.NotContains(t, wideLines[1], "An unusually long status name")
	assert.Contains(t, wideLines[1], "123456789 B")
	assert.Contains(t, wideLines[1], "123 ms")
	wideCells := trimCellsForTest(strings.Split(wideLines[1], colSeparator))
	assert.Equal(t, "599", wideCells[1])
	assert.True(t, strings.HasPrefix(wideCells[4], "2001:db8"))
	assert.True(t, strings.HasSuffix(wideCells[4], "..."))
	assert.NotContains(t, wideLines[1], "2001:db8:ffff:ffff:ffff:ffff:ffff:ffff")

	narrow := v.renderMeasurementTable(rows, 32, "http")
	narrowLines := strings.Split(strings.TrimSpace(stripANSIForTest(narrow)), "\n")
	assert.Greater(t, runewidth.StringWidth(narrowLines[1]), 32)
	assert.Contains(t, narrowLines[1], "599")
	assert.NotContains(t, narrowLines[1], "An unusually long status name")
	assert.Contains(t, narrowLines[1], "123456789 B")
	assert.Contains(t, narrowLines[1], "123 ms")
}

func Test_RenderMeasurementTable_TruncatesLocationBeforeTimings(t *testing.T) {
	rows := [][]string{
		{"Location", "Hops", "Last", "Min", "Avg", "Max"},
		{"Falkenstein, DE, EU, Hetzner Online (AS24940)", "8", "3.65 ms", "3.48 ms", "3.57 ms", "3.65 ms"},
	}
	printer := NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer))
	printer.DisableStyling()
	v := &viewer{printer: printer}

	output := v.renderMeasurementTable(rows, 88, "traceroute")
	lines := strings.Split(strings.TrimSpace(output), "\n")

	require.Len(t, lines, 2)
	assert.LessOrEqual(t, runewidth.StringWidth(lines[1]), 88)
	assert.Contains(t, lines[1], "Falkenstein, DE, EU, Hetzner Online")
	assert.Contains(t, lines[1], "...")
	assert.NotContains(t, lines[1], "(AS24940)")
	assert.Contains(t, lines[1], "3.65 ms")
	assert.Contains(t, lines[1], "3.48 ms")
	assert.Contains(t, lines[1], "3.57 ms")
}

func Test_OutputTableThenShare_SeparatesMultiRowTable(t *testing.T) {
	first := tableProbe("Berlin", "DE", "DNS Network", globalping.TestStatusFinished)
	first.Result.StatusCodeName = "NOERROR"
	first.Result.AnswersRaw = json.RawMessage(`[]`)
	first.Result.TimingsRaw = json.RawMessage(`{"total":1}`)
	second := tableProbe("Paris", "FR", "DNS Network", globalping.TestStatusFinished)
	second.Result.StatusCodeName = "NOERROR"
	second.Result.AnswersRaw = json.RawMessage(`[]`)
	second.Result.TimingsRaw = json.RawMessage(`{"total":2}`)
	measurement := tableMeasurement("dns", first, second)
	ctx := createDefaultContext("dns")
	ctx.Cmd = "dns"
	ctx.Table = true
	ctx.Share = true
	ctx.History.Push(&HistoryItem{Id: measurement.ID, StartedAt: defaultCurrentTime})
	w := new(bytes.Buffer)
	printer := NewPrinter(nil, w, w)
	printer.DisableStyling()
	v := NewViewer(ctx, printer, nil)

	require.NoError(t, v.OutputTable(measurement))
	v.OutputShare()

	assert.Contains(t, w.String(), "\n\n> View the results online:")
}

func Test_OutputTable_AllFailedPreservesProbeErrors(t *testing.T) {
	for _, measurementType := range []globalping.MeasurementType{"ping", "traceroute", "mtr", "dns", "http"} {
		t.Run(string(measurementType), func(t *testing.T) {
			failed := tableProbe("Berlin", "DE", "Failed Network", globalping.TestStatusFailed)
			failed.Result.RawOutput = "measurement failed"
			offline := tableProbe("Paris", "FR", "Offline Network", globalping.TestStatusOffline)
			offline.Result.RawOutput = "This probe is currently offline. Please try again later."
			measurement := tableMeasurement(measurementType, failed, offline)

			output, ctx, err := renderTableWithContextForTest(t, measurement, measurementType == "dns", 2)

			require.EqualError(t, err, "all probes failed")
			assert.Equal(t, 0, ctx.TableOutputRows)
			assert.Equal(t, `> Berlin, DE, EU, Failed Network (AS64500)
measurement failed
> Paris, FR, EU, Offline Network (AS64500)
This probe is currently offline. Please try again later.
`, output)
		})
	}
}

func tableMeasurement(measurementType globalping.MeasurementType, results ...globalping.ProbeMeasurement) *globalping.Measurement {
	return &globalping.Measurement{
		ID:          measurementID1,
		Type:        measurementType,
		Status:      globalping.MeasurementStatusFinished,
		ProbesCount: len(results),
		Results:     results,
	}
}

func tableProbe(city, country, network string, status globalping.TestStatus) globalping.ProbeMeasurement {
	return globalping.ProbeMeasurement{
		Probe: globalping.ProbeDetails{
			City:      city,
			Country:   country,
			Continent: "EU",
			Network:   network,
			ASN:       64500,
		},
		Result: globalping.ProbeResult{Status: status},
	}
}

func renderTableForTest(t *testing.T, measurement *globalping.Measurement, trace bool) (string, error) {
	output, _, err := renderTableWithContextForTest(t, measurement, trace, 0)
	return output, err
}

func renderTableWithContextForTest(t *testing.T, measurement *globalping.Measurement, trace bool, initialTableOutputRows int) (string, *Context, error) {
	t.Helper()
	ctx := createDefaultContext(string(measurement.Type))
	ctx.Cmd = string(measurement.Type)
	ctx.Table = true
	ctx.Trace = trace
	ctx.TableOutputRows = initialTableOutputRows
	if ctx.History.Find(measurement.ID) == nil {
		ctx.History.Push(&HistoryItem{Id: measurement.ID, StartedAt: defaultCurrentTime})
	}
	w := new(bytes.Buffer)
	printer := NewPrinter(nil, w, w)
	printer.DisableStyling()
	viewer := NewViewer(ctx, printer, nil)
	err := viewer.OutputTable(measurement)
	return w.String(), ctx, err
}

func assertTableForTest(t *testing.T, output string, expected [][]string) {
	t.Helper()
	require.NotContains(t, output, "\x1b", "CI-style table output must not contain terminal styling")
	require.NotContains(t, output, "\t", "table output must not contain tabs")
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
	require.Len(t, lines, len(expected))

	separatorOffsets := tableSeparatorOffsetsForTest(lines[0])
	for i, line := range lines {
		assert.Equal(t, separatorOffsets, tableSeparatorOffsetsForTest(line), "columns must align on output line %d", i+1)
		parts := strings.Split(line, colSeparator)
		require.Len(t, parts, len(expected[i]), "unexpected column count on output line %d", i+1)
		for j := range parts {
			parts[j] = strings.TrimSpace(parts[j])
		}
		assert.Equal(t, expected[i], parts, "unexpected cells on output line %d", i+1)
	}
}

func assertTableWithFailureForTest(t *testing.T, output string, expectedRows [][]string, failureLocation, failureMessage string) {
	t.Helper()
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
	require.Len(t, lines, len(expectedRows)+1)
	assertTableForTest(t, strings.Join(lines[:len(expectedRows)], "\n")+"\n", expectedRows)

	headerSeparatorOffsets := tableSeparatorOffsetsForTest(lines[0])
	failureParts := strings.Split(lines[len(lines)-1], colSeparator)
	require.Len(t, failureParts, 2)
	assert.Equal(t, headerSeparatorOffsets[0], tableSeparatorOffsetsForTest(lines[len(lines)-1])[0])
	assert.Equal(t, failureLocation, strings.TrimSpace(failureParts[0]))
	assert.Equal(t, failureMessage, strings.TrimSpace(failureParts[1]))
}

func tableSeparatorOffsetsForTest(line string) []int {
	offsets := []int{}
	for offset := 0; ; {
		i := strings.Index(line[offset:], colSeparator)
		if i == -1 {
			return offsets
		}
		offset += i
		offsets = append(offsets, offset)
		offset += len(colSeparator)
	}
}

var ansiPatternForTest = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSIForTest(value string) string {
	return ansiPatternForTest.ReplaceAllString(value, "")
}

func trimCellsForTest(cells []string) []string {
	for i := range cells {
		cells[i] = strings.TrimSpace(cells[i])
	}
	return cells
}

func assertDisplayTableLayoutForTest(t *testing.T, output string, maxWidth int) {
	t.Helper()
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
	require.NotEmpty(t, lines)
	wantOffsets := displaySeparatorOffsetsForTest(stripANSIForTest(lines[0]))
	for i, line := range lines {
		visible := stripANSIForTest(line)
		assert.LessOrEqual(t, runewidth.StringWidth(visible), maxWidth, "line %d exceeds the requested width", i+1)
		assert.Equal(t, wantOffsets, displaySeparatorOffsetsForTest(visible), "display columns must align on line %d", i+1)
	}
}

func displaySeparatorOffsetsForTest(line string) []int {
	parts := strings.Split(line, colSeparator)
	offsets := make([]int, 0, len(parts)-1)
	offset := 0
	for i := 0; i < len(parts)-1; i++ {
		offset += runewidth.StringWidth(parts[i])
		offsets = append(offsets, offset)
		offset += runewidth.StringWidth(colSeparator)
	}
	return offsets
}
