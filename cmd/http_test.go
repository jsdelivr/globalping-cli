package cmd

import (
	"bytes"
	"os"
	"testing"

	apiMocks "github.com/jsdelivr/globalping-cli/mocks/api"
	utilsMocks "github.com/jsdelivr/globalping-cli/mocks/utils"
	viewMocks "github.com/jsdelivr/globalping-cli/mocks/view"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/jsdelivr/globalping-go"
	"github.com/spf13/cobra"
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

	gbMock := apiMocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(t.Context(), expectedOpts).Times(1).Return(expectedResponse, nil)

	expectedMeasurement := createDefaultMeasurement("http")
	gbMock.EXPECT().AwaitMeasurement(t.Context(), expectedResponse.ID).Times(1).Return(expectedMeasurement, nil)

	viewerMock := viewMocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().OutputDefault(measurementID1, expectedMeasurement, expectedOpts).Times(1)

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext()
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
	err := root.Cmd.ExecuteContext(t.Context())
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
	expectedOpts.Options.IPVersion = globalping.IPVersion4
	expectedOpts.Options.Request = &globalping.RequestOptions{
		Headers: map[string]string{},
	}

	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := apiMocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(t.Context(), expectedOpts).Times(1).Return(expectedResponse, nil)

	expectedMeasurement := createDefaultMeasurement("http")
	gbMock.EXPECT().AwaitMeasurement(t.Context(), expectedResponse.ID).Times(1).Return(expectedMeasurement, nil)

	viewerMock := viewMocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().OutputDefault(measurementID1, expectedMeasurement, expectedOpts).Times(1)

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext()
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "http", "jsdelivr.com",
		"from", "Berlin",
		"--ipv4",
	}
	err := root.Cmd.ExecuteContext(t.Context())
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
	expectedOpts.Options.IPVersion = globalping.IPVersion6
	expectedOpts.Options.Request = &globalping.RequestOptions{
		Headers: map[string]string{},
	}

	expectedResponse := createDefaultMeasurementCreateResponse()

	gbMock := apiMocks.NewMockClient(ctrl)
	gbMock.EXPECT().CreateMeasurement(t.Context(), expectedOpts).Times(1).Return(expectedResponse, nil)

	expectedMeasurement := createDefaultMeasurement("http")
	gbMock.EXPECT().AwaitMeasurement(t.Context(), expectedResponse.ID).Times(1).Return(expectedMeasurement, nil)

	viewerMock := viewMocks.NewMockViewer(ctrl)
	viewerMock.EXPECT().OutputDefault(measurementID1, expectedMeasurement, expectedOpts).Times(1)

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext()
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, viewerMock, utilsMock, gbMock, nil, _storage)
	os.Args = []string{"globalping", "http", "jsdelivr.com",
		"from", "Berlin",
		"--ipv6",
	}
	err := root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)

	assert.Equal(t, "", w.String())

	expectedCtx := createDefaultExpectedContext("http")
	expectedCtx.Ipv6 = true

	assert.Equal(t, expectedCtx, ctx)
}

func Test_Execute_HTTP_Invalid_Protocol(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	utilsMock := utilsMocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext()
	_storage := createDefaultTestStorage(t, utilsMock)
	root := NewRoot(printer, ctx, nil, utilsMock, nil, nil, _storage)

	os.Args = []string{"globalping", "http", "jsdelivr.com", "--protocol", "invalid"}
	err := root.Cmd.ExecuteContext(t.Context())
	assert.Error(t, err, "protocol INVALID is not supported")

	items, err := _storage.GetHistory(0)
	assert.NoError(t, err)
	assert.Empty(t, items)
}

func Test_ParseUrlData(t *testing.T) {
	urlData, err := parseUrlData("https://cdn.jsdelivr.net:8080/npm/react/?query=3")
	assert.NoError(t, err)
	assert.Equal(t, "cdn.jsdelivr.net", urlData.Host)
	assert.Equal(t, "/npm/react/", urlData.Path)
	assert.Equal(t, "HTTPS", urlData.Protocol)
	assert.Equal(t, uint16(8080), urlData.Port)
	assert.Equal(t, "query=3", urlData.Query)
}

func Test_ParseUrlData_NoScheme(t *testing.T) {
	urlData, err := parseUrlData("cdn.jsdelivr.net/npm/react/?query=3")
	assert.NoError(t, err)
	assert.Equal(t, "cdn.jsdelivr.net", urlData.Host)
	assert.Equal(t, "/npm/react/", urlData.Path)
	assert.Equal(t, "HTTPS", urlData.Protocol)
	assert.Equal(t, uint16(0), urlData.Port)
	assert.Equal(t, "query=3", urlData.Query)
}

func Test_ParseUrlData_HostOnly(t *testing.T) {
	urlData, err := parseUrlData("cdn.jsdelivr.net")
	assert.NoError(t, err)
	assert.Equal(t, "cdn.jsdelivr.net", urlData.Host)
	assert.Equal(t, "", urlData.Path)
	assert.Equal(t, "HTTPS", urlData.Protocol)
	assert.Equal(t, uint16(0), urlData.Port)
	assert.Equal(t, "", urlData.Query)
}

func Test_ParseHttpHeaders_None(t *testing.T) {
	headerStrings := []string{}

	m, err := parseHttpHeaders(headerStrings)
	assert.NoError(t, err)

	assert.Empty(t, m)
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
	ctx := createDefaultContext()
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil, nil)

	ctx.Target = "https://example.com/my/path?x=123&yz=abc"
	ctx.From = "london"
	ctx.Full = true

	cmd := &cobra.Command{}

	m, err := root.buildHttpMeasurementRequest(cmd, ctx.Target)
	assert.NoError(t, err)

	expectedM := &globalping.MeasurementCreate{
		Limit:             1,
		Type:              "http",
		Target:            "example.com",
		InProgressUpdates: true,
		Options: &globalping.MeasurementOptions{
			Protocol: "HTTPS",
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
	ctx := createDefaultContext()
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil, nil)

	ctx.Target = "https://example.com/my/path?x=123&yz=abc"
	ctx.From = "london"
	ctx.Full = true
	ctx.Method = "HEAD"

	cmd := &cobra.Command{}

	m, err := root.buildHttpMeasurementRequest(cmd, ctx.Target)
	assert.NoError(t, err)

	expectedM := &globalping.MeasurementCreate{
		Limit:             1,
		Type:              "http",
		Target:            "example.com",
		InProgressUpdates: true,
		Options: &globalping.MeasurementOptions{
			Protocol: "HTTPS",
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
	ctx := createDefaultContext()
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil, nil)

	ctx.Target = "https://example.com/my/path?x=123&yz=abc"
	ctx.From = "london"

	cmd := &cobra.Command{}

	m, err := root.buildHttpMeasurementRequest(cmd, ctx.Target)
	assert.NoError(t, err)

	expectedM := &globalping.MeasurementCreate{
		Limit:             1,
		Type:              "http",
		Target:            "example.com",
		InProgressUpdates: true,
		Options: &globalping.MeasurementOptions{
			Protocol: "HTTPS",
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

func Test_BuildHttpMeasurementRequest_UsesTargetSpecificURLDefaults(t *testing.T) {
	ctx := createDefaultContext()
	ctx.Port = 443
	root := NewRoot(view.NewPrinter(nil, nil, nil), ctx, nil, nil, nil, nil, nil)
	cmd, _, err := root.Cmd.Find([]string{"http"})
	assert.NoError(t, err)

	first, err := root.buildHttpMeasurementRequest(cmd, "http://one.example/first?x=1")
	assert.NoError(t, err)
	second, err := root.buildHttpMeasurementRequest(cmd, "https://two.example/second?y=2")
	assert.NoError(t, err)

	assert.Equal(t, "one.example", first.Target)
	assert.Equal(t, "HTTP", first.Options.Protocol)
	assert.Equal(t, uint16(80), first.Options.Port)
	assert.Equal(t, "/first", first.Options.Request.Path)
	assert.Equal(t, "x=1", first.Options.Request.Query)
	assert.Equal(t, "two.example", second.Target)
	assert.Equal(t, "HTTPS", second.Options.Protocol)
	assert.Equal(t, uint16(443), second.Options.Port)
	assert.Equal(t, "/second", second.Options.Request.Path)
	assert.Equal(t, "y=2", second.Options.Request.Query)
}

func Test_BuildHttpMeasurementRequest_ExplicitFlagsOverrideEachURL(t *testing.T) {
	ctx := createDefaultContext()
	ctx.Protocol = "HTTP2"
	ctx.Port = 9443
	ctx.Path = "/override"
	ctx.Query = "override=1"
	root := NewRoot(view.NewPrinter(nil, nil, nil), ctx, nil, nil, nil, nil, nil)
	cmd, _, err := root.Cmd.Find([]string{"http"})
	assert.NoError(t, err)
	assert.NoError(t, cmd.Flags().Set("protocol", "HTTP2"))
	assert.NoError(t, cmd.Flags().Set("port", "9443"))

	for _, target := range []string{"http://one.example:8080/first?x=1", "https://two.example:8443/second?y=2"} {
		measurement, err := root.buildHttpMeasurementRequest(cmd, target)

		assert.NoError(t, err)
		assert.Equal(t, "HTTP2", measurement.Options.Protocol)
		assert.Equal(t, uint16(9443), measurement.Options.Port)
		assert.Equal(t, "/override", measurement.Options.Request.Path)
		assert.Equal(t, "override=1", measurement.Options.Request.Query)
	}
}
