package cmd

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/mocks"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_Execute_HTTP_Default(t *testing.T) {
	t.Cleanup(sessionCleanup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedOpts := &globalping.MeasurementCreate{
		Type:   "http",
		Target: "jsdelivr.com",
		Limit:  1,
		Options: &globalping.MeasurementOptions{
			Protocol: "HTTPS",
			Port:     99,
			Resolver: "1.1.1.1",
			Request: &globalping.RequestOptions{
				Host:    "example.com",
				Path:    "/robots.txt",
				Query:   "test=1",
				Method:  "GET",
				Headers: map[string]string{"X-Test": "1"},
			},
		},
		Locations: []globalping.Locations{
			{Magic: "Berlin"},
		},
	}
	expectedResponse := getMeasurementCreateResponse(measurementID1)

	gbMock := mocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(expectedOpts).Times(1).Return(expectedResponse, false, nil)

	viewerMock := mocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().Output(measurementID1, expectedOpts).Times(1).Return(nil)

	ctx := &view.Context{
		MaxHistory: 1,
	}
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	defer r.Close()
	defer w.Close()

	printer := view.NewPrinter(nil, w, w)
	root := NewRoot(printer, ctx, viewerMock, nil, gbMock, nil)
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
	err = root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)
	w.Close()

	output, err := io.ReadAll(r)
	assert.NoError(t, err)
	assert.Equal(t, "", string(output))

	expectedCtx := &view.Context{
		Cmd:        "http",
		Target:     "jsdelivr.com",
		From:       "Berlin",
		Limit:      1,
		Host:       "example.com",
		Resolver:   "1.1.1.1",
		Protocol:   "HTTPS",
		Method:     "GET",
		Query:      "test=1",
		Path:       "/robots.txt",
		Headers:    []string{"X-Test: 1"},
		Full:       true,
		Port:       99,
		CIMode:     true,
		MaxHistory: 1,
	}
	assert.Equal(t, expectedCtx, ctx)

	b, err := os.ReadFile(getMeasurementsPath())
	assert.NoError(t, err)
	expectedHistory := []byte(measurementID1 + "\n")
	assert.Equal(t, expectedHistory, b)
}

func TestParseUrlData(t *testing.T) {
	urlData, err := parseUrlData("https://cdn.jsdelivr.net:8080/npm/react/?query=3")
	assert.NoError(t, err)
	assert.Equal(t, "cdn.jsdelivr.net", urlData.Host)
	assert.Equal(t, "/npm/react/", urlData.Path)
	assert.Equal(t, "https", urlData.Protocol)
	assert.Equal(t, 8080, urlData.Port)
	assert.Equal(t, "query=3", urlData.Query)
}

func TestParseUrlDataNoScheme(t *testing.T) {
	urlData, err := parseUrlData("cdn.jsdelivr.net/npm/react/?query=3")
	assert.NoError(t, err)
	assert.Equal(t, "cdn.jsdelivr.net", urlData.Host)
	assert.Equal(t, "/npm/react/", urlData.Path)
	assert.Equal(t, "http", urlData.Protocol)
	assert.Equal(t, 0, urlData.Port)
	assert.Equal(t, "query=3", urlData.Query)
}

func TestParseUrlDataHostOnly(t *testing.T) {
	urlData, err := parseUrlData("cdn.jsdelivr.net")
	assert.NoError(t, err)
	assert.Equal(t, "cdn.jsdelivr.net", urlData.Host)
	assert.Equal(t, "", urlData.Path)
	assert.Equal(t, "http", urlData.Protocol)
	assert.Equal(t, 0, urlData.Port)
	assert.Equal(t, "", urlData.Query)
}

func TestOverrideOpt(t *testing.T) {
	assert.Equal(t, "new", overrideOpt("orig", "new"))
	assert.Equal(t, "orig", overrideOpt("orig", ""))
	assert.Equal(t, 10, overrideOptInt(0, 10))
	assert.Equal(t, 10, overrideOptInt(10, 0))
}

func TestParseHttpHeaders_None(t *testing.T) {
	headerStrings := []string{}

	m, err := parseHttpHeaders(headerStrings)
	assert.NoError(t, err)

	assert.Nil(t, nil, m)
}

func TestParseHttpHeaders_Single(t *testing.T) {
	headerStrings := []string{"ABC: 123x"}

	m, err := parseHttpHeaders(headerStrings)
	assert.NoError(t, err)

	assert.Equal(t, map[string]string{"ABC": "123x"}, m)
}

func TestParseHttpHeaders_Multiple(t *testing.T) {
	headerStrings := []string{"ABC: 123x", "DEF: 456y,789z"}

	m, err := parseHttpHeaders(headerStrings)
	assert.NoError(t, err)

	assert.Equal(t, map[string]string{"ABC": "123x", "DEF": "456y,789z"}, m)
}

func TestParseHttpHeaders_Invalid(t *testing.T) {
	headerStrings := []string{"ABC=123x"}

	_, err := parseHttpHeaders(headerStrings)
	assert.ErrorContains(t, err, "invalid header")
}

func Test_BuildHttpMeasurementRequest_FULL(t *testing.T) {
	ctx := &view.Context{}
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil)

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
				Host:    "example.com",
				Query:   "x=123&yz=abc",
				Method:  "GET",
			},
		},
	}

	assert.Equal(t, expectedM, m)
}

func TestBuildHttpMeasurementRequest_HEAD(t *testing.T) {
	ctx := &view.Context{}
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil)

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
				Host:    "example.com",
				Query:   "x=123&yz=abc",
			},
		},
	}

	assert.Equal(t, expectedM, m)
}
