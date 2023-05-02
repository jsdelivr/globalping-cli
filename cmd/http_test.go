package cmd

import (
	"testing"

	"github.com/jsdelivr/globalping-cli/model"
	"github.com/stretchr/testify/assert"
)

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

func TestBuildHttpMeasurementRequest_FULL(t *testing.T) {
	ctx = model.Context{
		Target: "https://example.com/my/path?x=123&yz=abc",
		From:   "london",
		Full:   true,
	}

	httpCmdOpts = &HttpCmdOpts{
		Method: "HEAD",
	}

	m, err := buildHttpMeasurementRequest()
	assert.NoError(t, err)

	expectedM := model.PostMeasurement{Limit: 0,
		Locations: []model.Locations{
			{Magic: "london"}},
		Type:              "http",
		Target:            "example.com",
		InProgressUpdates: true,
		Options: &model.MeasurementOptions{
			Protocol: "https",
			Request: &model.RequestOptions{
				Headers: map[string]string{},
				Path:    "/my/path",
				Host:    "example.com",
				Query:   "x=123&yz=abc",
				Method:  "GET",
			},
		},
	}

	assert.Equal(t, expectedM, m)

	// restore
	httpCmdOpts = &HttpCmdOpts{}
	ctx = model.Context{}
}

func TestBuildHttpMeasurementRequest_HEAD(t *testing.T) {
	ctx = model.Context{
		Target: "https://example.com/my/path?x=123&yz=abc",
		From:   "london",
	}

	httpCmdOpts = &HttpCmdOpts{
		Method: "HEAD",
	}

	m, err := buildHttpMeasurementRequest()
	assert.NoError(t, err)

	expectedM := model.PostMeasurement{Limit: 0,
		Locations: []model.Locations{
			{Magic: "london"}},
		Type:              "http",
		Target:            "example.com",
		InProgressUpdates: true,
		Options: &model.MeasurementOptions{
			Protocol: "https",
			Request: &model.RequestOptions{
				Headers: map[string]string{},
				Path:    "/my/path",
				Host:    "example.com",
				Query:   "x=123&yz=abc",
				Method:  "HEAD",
			},
		},
	}

	assert.Equal(t, expectedM, m)

	// restore
	httpCmdOpts = &HttpCmdOpts{}
	ctx = model.Context{}
}
