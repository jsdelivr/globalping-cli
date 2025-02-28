package cmd

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/mocks"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_Execute_HTTP_Default(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("http")
	expectedOpts.Options.Protocol = "HTTPS"
	expectedOpts.Options.Port = 99
	expectedOpts.Options.Resolver = "1.1.1.1"
	expectedOpts.Options.Request = &globalping.RequestOptions{
		Host:    "example.com",
		Path:    "/robots.txt",
		Query:   "test=1",
		Method:  "GET",
		Headers: map[string]string{"X-Test": "1"},
	}

	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(expectedOpts).Times(1).Return(expectedResponse, nil)

	viewerMock := mocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().Output(measurementID1, expectedOpts).Times(1).Return(nil)

	utilsMock := mocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("http")
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "http", "jsdelivr.com",
		"from", "Berlin",
		"--protocol", "HTTPS",
		"--method", "GET",
		"--host", "example.com",
		"--path", "/robots.txt",
		"--query", "test=1",
		"--header", "X-Test: 1",
		"--resolver", "1.1.1.1",
		"--port", "99",
		"--full",
	}
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("http")
	expectedCtx.Protocol = "HTTPS"
	expectedCtx.Method = "GET"
	expectedCtx.Host = "example.com"
	expectedCtx.Path = "/robots.txt"
	expectedCtx.Query = "test=1"
	expectedCtx.Headers = []string{"X-Test: 1"}
	expectedCtx.Resolver = "1.1.1.1"
	expectedCtx.Port = 99
	expectedCtx.Full = true

	assert.Equal(t, expectedCtx, ctx)

	b, err := _storage.GetMeasurements()
	assert.NoError(t, err)
	expectedHistory := measurementID1 + "\n"
	assert.Equal(t, expectedHistory, string(b))

	items, err := _storage.GetHistory(0)
	assert.NoError(t, err)
	expectedHistoryItems := []string{createDefaultExpectedHistoryItem(
		"1",
		"http jsdelivr.com from Berlin --protocol HTTPS --method GET --host example.com --path /robots.txt --query test=1 --header X-Test: 1 --resolver 1.1.1.1 --port 99 --full",
		measurementID1,
	)}
	assert.Equal(t, expectedHistoryItems, items)
}

func Test_Execute_HTTP_IPv4(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("http")
	expectedOpts.Options.Protocol = "https"
	expectedOpts.Options.IPVersion = globalping.IPVersion4
	expectedOpts.Options.Request = &globalping.RequestOptions{
		Headers: map[string]string{},
	}

	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(expectedOpts).Times(1).Return(expectedResponse, nil)

	viewerMock := mocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().Output(measurementID1, expectedOpts).Times(1).Return(nil)

	utilsMock := mocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("http")
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "http", "jsdelivr.com",
		"from", "Berlin",
		"--ipv4",
	}
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("http")
	expectedCtx.Ipv4 = true

	assert.Equal(t, expectedCtx, ctx)
}

func Test_Execute_HTTP_IPv6(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := createDefaultMeasurementCreate("http")
	expectedOpts.Options.Protocol = "https"
	expectedOpts.Options.IPVersion = globalping.IPVersion6
	expectedOpts.Options.Request = &globalping.RequestOptions{
		Headers: map[string]string{},
	}

	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(expectedOpts).Times(1).Return(expectedResponse, nil)

	viewerMock := mocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().Output(measurementID1, expectedOpts).Times(1).Return(nil)

	utilsMock := mocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("http")
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "http", "jsdelivr.com",
		"from", "Berlin",
		"--ipv6",
	}
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("http")
	expectedCtx.Ipv6 = true

	assert.Equal(t, expectedCtx, ctx)
}

func Test_ParseUrlData(t *testing.T) {
	urlData, err := parseUrlData("https://cdn.jsdelivr.net:8080/npm/react/?query=3")
	assert.NoError(t, err)
	assert.Equal(t, "cdn.jsdelivr.net", urlData.Host)
	assert.Equal(t, "/npm/react/", urlData.Path)
	assert.Equal(t, "https", urlData.Protocol)
	assert.Equal(t, 8080, urlData.Port)
	assert.Equal(t, "query=3", urlData.Query)
}

func Test_ParseUrlData_NoScheme(t *testing.T) {
	urlData, err := parseUrlData("cdn.jsdelivr.net/npm/react/?query=3")
	assert.NoError(t, err)
	assert.Equal(t, "cdn.jsdelivr.net", urlData.Host)
	assert.Equal(t, "/npm/react/", urlData.Path)
	assert.Equal(t, "https", urlData.Protocol)
	assert.Equal(t, 0, urlData.Port)
	assert.Equal(t, "query=3", urlData.Query)
}

func Test_ParseUrlData_HostOnly(t *testing.T) {
	urlData, err := parseUrlData("cdn.jsdelivr.net")
	assert.NoError(t, err)
	assert.Equal(t, "cdn.jsdelivr.net", urlData.Host)
	assert.Equal(t, "", urlData.Path)
	assert.Equal(t, "https", urlData.Protocol)
	assert.Equal(t, 0, urlData.Port)
	assert.Equal(t, "", urlData.Query)
}

func Test_OverrideOpt(t *testing.T) {
	assert.Equal(t, "new", overrideOpt("orig", "new"))
	assert.Equal(t, "orig", overrideOpt("orig", ""))
	assert.Equal(t, 10, overrideOptInt(0, 10))
	assert.Equal(t, 10, overrideOptInt(10, 0))
}

func Test_ParseHttpHeaders_None(t *testing.T) {
	headerStrings := []string{}

	m, err := parseHttpHeaders(headerStrings)
	assert.NoError(t, err)

	assert.Nil(t, nil, m)
}

func Test_ParseHttpHeaders_Single(t *testing.T) {
	headerStrings := []string{"ABC: 123x"}

	m, err := parseHttpHeaders(headerStrings)
	assert.NoError(t, err)

	assert.Equal(t, map[string]string{"ABC": "123x"}, m)
}

func Test_ParseHttpHeaders_Multiple(t *testing.T) {
	headerStrings := []string{"ABC: 123x", "DEF: 456y,789z"}

	m, err := parseHttpHeaders(headerStrings)
	assert.NoError(t, err)

	assert.Equal(t, map[string]string{"ABC": "123x", "DEF": "456y,789z"}, m)
}

func Test_ParseHttpHeaders_Invalid(t *testing.T) {
	headerStrings := []string{"ABC=123x"}

	_, err := parseHttpHeaders(headerStrings)
	assert.ErrorContains(t, err, "invalid header")
}

func Test_BuildHttpMeasurementRequest_Full(t *testing.T) {
	ctx := createDefaultContext("http")
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil, nil)

	ctx.Target = "https://example.com/my/path?x=123&yz=abc"
	ctx.From = "london"
	ctx.Full = true

	m, err := root.buildHttpMeasurementRequest()
	assert.NoError(t, err)

	expectedM := &globalping.MeasurementCreate{
		Limit:             1,
		Type:              "http",
		Target:            "example.com",
		InProgressUpdates: true,
		Options: &globalping.MeasurementOptions{
			Protocol: "https",
			Request: &globalping.RequestOptions{
				Headers: map[string]string{},
				Path:    "/my/path",
				Host:    "",
				Query:   "x=123&yz=abc",
				Method:  "GET",
			},
		},
	}

	assert.Equal(t, expectedM, m)
}

func Test_BuildHttpMeasurementRequest_FullHead(t *testing.T) {
	ctx := createDefaultContext("http")
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil, nil)

	ctx.Target = "https://example.com/my/path?x=123&yz=abc"
	ctx.From = "london"
	ctx.Full = true
	ctx.Method = "HEAD"

	m, err := root.buildHttpMeasurementRequest()
	assert.NoError(t, err)

	expectedM := &globalping.MeasurementCreate{
		Limit:             1,
		Type:              "http",
		Target:            "example.com",
		InProgressUpdates: true,
		Options: &globalping.MeasurementOptions{
			Protocol: "https",
			Request: &globalping.RequestOptions{
				Headers: map[string]string{},
				Path:    "/my/path",
				Host:    "",
				Query:   "x=123&yz=abc",
				Method:  "HEAD",
			},
		},
	}

	assert.Equal(t, expectedM, m)
}

func Test_BuildHttpMeasurementRequest_HEAD(t *testing.T) {
	ctx := createDefaultContext("http")
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil, nil)

	ctx.Target = "https://example.com/my/path?x=123&yz=abc"
	ctx.From = "london"

	m, err := root.buildHttpMeasurementRequest()
	assert.NoError(t, err)

	expectedM := &globalping.MeasurementCreate{
		Limit:             1,
		Type:              "http",
		Target:            "example.com",
		InProgressUpdates: true,
		Options: &globalping.MeasurementOptions{
			Protocol: "https",
			Request: &globalping.RequestOptions{
				Headers: map[string]string{},
				Path:    "/my/path",
				Host:    "",
				Query:   "x=123&yz=abc",
			},
		},
	}

	assert.Equal(t, expectedM, m)
}
