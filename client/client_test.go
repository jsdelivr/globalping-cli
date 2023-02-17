package client_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"

	"github.com/stretchr/testify/assert"
)

// Generate server for testing
func generateServer(json string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(json))
	}))
	return server
}

func generateServerError(json string, statusCode int) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(json))
	}))
	return server
}

// Dummy interface since we have mock responses
var opts = model.PostMeasurement{}

// PostAPI tests
func TestPostAPI(t *testing.T) {
	// Suppress error outputs
	os.Stdout, _ = os.Open(os.DevNull)
	for scenario, fn := range map[string]func(t *testing.T){
		"valid":      testPostValid,
		"no_probes":  testPostNoProbes,
		"validation": testPostValidation,
		"api_error":  testPostInternalError,
	} {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
		})
	}
}

// Test a valid call of PostAPI
func testPostValid(t *testing.T) {
	server := generateServer(`{"id":"abcd","probesCount":1}`)
	defer server.Close()
	client.ApiUrl = server.URL

	res, err := client.PostAPI(opts)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "abcd", res.ID)
	assert.Equal(t, 1, res.ProbesCount)
}

func testPostNoProbes(t *testing.T) {
	server := generateServerError(`{
    "error": {
      "message": "No suitable probes found",
      "type": "no_probes_found"
    }}`, 422)
	defer server.Close()
	client.ApiUrl = server.URL

	_, err := client.PostAPI(opts)
	assert.EqualError(t, err, "no suitable probes found")
}

func testPostValidation(t *testing.T) {
	server := generateServerError(`{
    "error": {
        "message": "Validation Failed",
        "type": "validation_error",
        "params": {
            "measurement": "\"measurement\" does not match any of the allowed types",
			"target": "\"target\" does not match any of the allowed types"
        }
    }}`, 400)
	defer server.Close()
	client.ApiUrl = server.URL

	_, err := client.PostAPI(opts)
	assert.EqualError(t, err, "validation error")
}

func testPostInternalError(t *testing.T) {
	server := generateServerError(`{
    "error": {
      "message": "Internal Server Error",
      "type": "api_error"
    }}`, 500)
	defer server.Close()
	client.ApiUrl = server.URL

	_, err := client.PostAPI(opts)
	assert.EqualError(t, err, "internal server error - please try again later")
}

// GetAPI tests
func TestGetAPI(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T){
		"valid": testGetValid,
		"json":  testGetJson,
		"ping":  testGetPing,
	} {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
		})
	}
}

func testGetValid(t *testing.T) {
	server := generateServer(`{"id":"abcd"}`)
	defer server.Close()
	client.ApiUrl = server.URL

	res, err := client.GetAPI("abcd")
	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v", res)

	assert.Equal(t, "abcd", res.ID)
}

func testGetJson(t *testing.T) {
	server := generateServer(`{"id":"abcd"}`)
	defer server.Close()
	client.ApiUrl = server.URL

	res, err := client.GetApiJson("abcd")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, `{"id":"abcd"}`, res)
}

func testGetPing(t *testing.T) {
	server := generateServer(`{
	"id": "abcd",
	"type": "ping",
	"status": "finished",
	"createdAt": "2023-02-17T18:11:52.825Z",
	"updatedAt": "2023-02-17T18:11:53.969Z",
	"probesCount": 1,
	"results": [
		{
		"probe": {
			"continent": "NA",
			"region": "Northern America",
			"country": "CA",
			"state": null,
			"city": "City",
			"asn": 7794,
			"longitude": -80.2222,
			"latitude": 43.3662,
			"network": "Network",
			"tags": [],
			"resolvers": [
			"1.1.1.1",
			"8.8.4.4"
			]
		},
		"result": {
			"status": "finished",
			"rawOutput": "PING",
			"resolvedAddress": "1.1.1.1",
			"resolvedHostname": "1.1.1.1:",
			"timings": [],
			"stats": {
				"min": 24.891,
				"max": 28.193,
				"avg": 27.088,
				"total": 3,
				"loss": 0,
				"rcv": 3,
				"drop": 0
			}
		}
	}]}`)
	defer server.Close()
	client.ApiUrl = server.URL

	res, err := client.GetAPI("abcd")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "abcd", res.ID)
	assert.Equal(t, "ping", res.Type)
	assert.Equal(t, "finished", res.Status)
	assert.Equal(t, "2023-02-17T18:11:52.825Z", res.CreatedAt)
	assert.Equal(t, "2023-02-17T18:11:53.969Z", res.UpdatedAt)
	assert.Equal(t, 1, res.ProbesCount)
	assert.Equal(t, 1, len(res.Results))

	assert.Equal(t, "NA", res.Results[0].Probe.Continent)
	assert.Equal(t, "Northern America", res.Results[0].Probe.Region)
	assert.Equal(t, "CA", res.Results[0].Probe.Country)
	assert.Equal(t, "", res.Results[0].Probe.State)
	assert.Equal(t, "City", res.Results[0].Probe.City)
	assert.Equal(t, 7794, res.Results[0].Probe.ASN)
	assert.Equal(t, "Network", res.Results[0].Probe.Network)
	assert.Equal(t, 0, len(res.Results[0].Probe.Tags))

	assert.Equal(t, "PING", res.Results[0].Result.RawOutput)
	assert.Equal(t, "1.1.1.1", res.Results[0].Result.ResolvedAddress)
	assert.Equal(t, 27.088, res.Results[0].Result.Stats["avg"])
	assert.Equal(t, 28.193, res.Results[0].Result.Stats["max"])
	assert.Equal(t, 24.891, res.Results[0].Result.Stats["min"])
	assert.Equal(t, float64(3), res.Results[0].Result.Stats["total"])
	assert.Equal(t, float64(3), res.Results[0].Result.Stats["rcv"])
	assert.Equal(t, float64(0), res.Results[0].Result.Stats["loss"])
	assert.Equal(t, float64(0), res.Results[0].Result.Stats["drop"])
}
