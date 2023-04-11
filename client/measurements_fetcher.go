package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/andybalholm/brotli"
	"github.com/jsdelivr/globalping-cli/model"
)

type MeasurementsFetcher interface {
	GetMeasurement(id string) (*model.GetMeasurement, error)
	GetRawMeasurement(id string) ([]byte, error)
}

type measurementsFetcher struct {
	// The api url endpoint
	apiUrl string

	// http client
	cl *http.Client

	// caches Etags by measurement id
	etags map[string]string

	// caches Measurements by ETag
	measurements map[string][]byte
}

func NewMeasurementsFetcher(apiUrl string) *measurementsFetcher {
	return &measurementsFetcher{
		apiUrl:       apiUrl,
		cl:           &http.Client{},
		etags:        map[string]string{},
		measurements: map[string][]byte{},
	}
}

// GetRawMeasurement returns API response as a GetMeasurement object
func (f *measurementsFetcher) GetMeasurement(id string) (*model.GetMeasurement, error) {
	respBytes, err := f.GetRawMeasurement(id)
	if err != nil {
		return nil, err
	}

	var m model.GetMeasurement
	err = json.Unmarshal(respBytes, &m)
	if err != nil {
		return nil, fmt.Errorf("invalid get measurement format returned: %v %s", err, string(respBytes))
	}

	return &m, nil
}

// GetRawMeasurement returns the API response's raw json response
func (f *measurementsFetcher) GetRawMeasurement(id string) ([]byte, error) {
	// Create a new request
	req, err := http.NewRequest("GET", f.apiUrl+"/"+id, nil)
	if err != nil {
		return nil, errors.New("err: failed to create request")
	}

	req.Header.Set("User-Agent", userAgent())
	req.Header.Set("Accept-Encoding", "br")

	etag := f.etags[id]
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	// Make the request
	resp, err := f.cl.Do(req)
	if err != nil {
		return nil, errors.New("err: request failed")
	}
	defer resp.Body.Close()

	// 404 not found
	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("err: measurement not found")
	}

	// 500 error
	if resp.StatusCode == http.StatusInternalServerError {
		return nil, errors.New("err: internal server error - please try again later")
	}

	// 304 not modified
	if resp.StatusCode == http.StatusNotModified {
		// get response bytes from cache
		respBytes := f.measurements[etag]
		if respBytes == nil {
			return nil, errors.New("err: response not found in etags cache")
		}

		return respBytes, nil
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("err: response code %d", resp.StatusCode)
	}

	var bodyReader io.Reader = resp.Body

	if resp.Header.Get("Content-Encoding") == "br" {
		bodyReader = brotli.NewReader(bodyReader)
	}

	// Read the response body
	respBytes, err := io.ReadAll(bodyReader)
	if err != nil {
		return nil, errors.New("err: failed to read response body")
	}

	// save etag and response to cache
	etag = resp.Header.Get("ETag")
	f.etags[id] = etag
	f.measurements[etag] = respBytes

	return respBytes, nil
}
