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

	res, showHelp, err := client.PostAPI(opts)

	assert.Equal(t, "abcd", res.ID)
	assert.Equal(t, 1, res.ProbesCount)
	assert.False(t, showHelp)
	assert.NoError(t, err)
}

func testPostNoProbes(t *testing.T) {
	server := generateServerError(`{
    "error": {
      "message": "No suitable probes found",
      "type": "no_probes_found"
    }}`, 422)
	defer server.Close()
	client.ApiUrl = server.URL

	_, showHelp, err := client.PostAPI(opts)
	assert.EqualError(t, err, "no suitable probes found - please choose a different location")
	assert.True(t, showHelp)
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

	_, showHelp, err := client.PostAPI(opts)
	assert.EqualError(t, err, "invalid parameters - please check the help for more information")
	assert.True(t, showHelp)
}

func testPostInternalError(t *testing.T) {
	server := generateServerError(`{
    "error": {
      "message": "Internal Server Error",
      "type": "api_error"
    }}`, 500)
	defer server.Close()
	client.ApiUrl = server.URL

	_, showHelp, err := client.PostAPI(opts)
	assert.EqualError(t, err, "err: internal server error - please try again later")
	assert.False(t, showHelp)
}

// GetAPI tests
func TestGetAPI(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T){
		"valid": testGetValid,
	} {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
		})
	}
}

// Test a valid call of GetAPI
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
