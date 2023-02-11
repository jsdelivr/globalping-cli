package client_test

import (
	"net/http"
	"net/http/httptest"
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
    "createdAt": "2022-07-17T16:19:52.909Z",
    "updatedAt": "2022-07-17T16:19:52.909Z",
    "results": [
        {
            "probe": {
                "continent": "AF",
                "country": "ZA",
                "state": null,
                "city": "cape town",
                "asn": 16509,
                "network": "amazon.com inc.",
                "tags": []
            },
            "result": {
                "timings": [
                    {
                        "ttl": 108,
                        "rtt": 16.5
                    },
                    {
                        "ttl": 108,
                        "rtt": 16.5
                    },
                    {
                        "ttl": 108,
                        "rtt": 16.5
                    }
                ],
                "stats": {
                  "min": 16.474,
                  "avg": 16.504,
                  "max": 16.543,
                  "loss": 0,
                },
                "rawOutput": "PING"
            }
        }
	]}`)
	defer server.Close()
	client.ApiUrl = server.URL

	res, err := client.GetAPI("abcd")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "abcd", res.ID)
	assert.Equal(t, "ping", res.Type)
	assert.Equal(t, "finished", res.Status)

}
