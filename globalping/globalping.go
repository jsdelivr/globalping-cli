package globalping

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/andybalholm/brotli"
	"github.com/jsdelivr/globalping-cli/version"
)

// boolean indicates whether to print CLI help on error
func (c *client) CreateMeasurement(measurement *MeasurementCreate) (*MeasurementCreateResponse, bool, error) {
	postData, err := json.Marshal(measurement)
	if err != nil {
		return nil, false, errors.New("failed to marshal post data - please report this bug")
	}

	// Create a new request
	req, err := http.NewRequest("POST", c.config.GlobalpingAPIURL+"/measurements", bytes.NewBuffer(postData))
	if err != nil {
		return nil, false, errors.New("failed to create request - please report this bug")
	}
	req.Header.Set("User-Agent", userAgent())
	req.Header.Set("Accept-Encoding", "br")
	req.Header.Set("Content-Type", "application/json")

	if c.config.GlobalpingToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.GlobalpingToken)
	}

	// Make the request
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, false, errors.New("request failed - please try again later")
	}
	defer resp.Body.Close()

	// If an error is returned
	if resp.StatusCode != http.StatusAccepted {
		// Decode the response body as JSON
		var data MeasurementCreateError

		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return nil, false, errors.New("invalid error format returned - please report this bug")
		}

		// 422 error
		if data.Error.Type == "no_probes_found" {
			return nil, true, errors.New("no suitable probes found - please choose a different location")
		}

		// 400 error
		if data.Error.Type == "validation_error" {
			resErr := ""
			for _, v := range data.Error.Params {
				resErr += fmt.Sprintf(" - %s\n", v)
			}
			return nil, true, fmt.Errorf("invalid parameters\n%sPlease check the help for more information", resErr)
		}

		// 401 error
		if data.Error.Type == "unauthorized" {
			return nil, false, fmt.Errorf("unauthorized: %s", data.Error.Message)
		}

		// 500 error
		if data.Error.Type == "api_error" {
			return nil, false, errors.New("internal server error - please try again later")
		}

		// If the error type is unknown
		return nil, false, fmt.Errorf("unknown error response: %s", data.Error.Type)
	}

	// Read the response body

	var bodyReader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "br" {
		bodyReader = brotli.NewReader(bodyReader)
	}

	res := &MeasurementCreateResponse{}
	err = json.NewDecoder(bodyReader).Decode(res)
	if err != nil {
		return nil, false, fmt.Errorf("invalid post measurement format returned - please report this bug: %s", err)
	}

	return res, false, nil
}

// GetRawMeasurement returns API response as a GetMeasurement object
func (c *client) GetMeasurement(id string) (*Measurement, error) {
	respBytes, err := c.GetMeasurementRaw(id)
	if err != nil {
		return nil, err
	}
	m := &Measurement{}
	err = json.Unmarshal(respBytes, m)
	if err != nil {
		return nil, fmt.Errorf("invalid get measurement format returned: %v %s", err, string(respBytes))
	}
	return m, nil
}

// GetMeasurementRaw returns the API response's raw json response
func (c *client) GetMeasurementRaw(id string) ([]byte, error) {
	// Create a new request
	req, err := http.NewRequest("GET", c.config.GlobalpingAPIURL+"/measurements/"+id, nil)
	if err != nil {
		return nil, errors.New("err: failed to create request")
	}

	req.Header.Set("User-Agent", userAgent())
	req.Header.Set("Accept-Encoding", "br")

	etag := c.etags[id]
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	// Make the request
	resp, err := c.http.Do(req)
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
		respBytes := c.measurements[etag]
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
	c.etags[id] = etag
	c.measurements[etag] = respBytes

	return respBytes, nil
}

func DecodeDNSTimings(timings json.RawMessage) (*DNSTimings, error) {
	t := &DNSTimings{}
	err := json.Unmarshal(timings, t)
	if err != nil {
		return nil, errors.New("invalid timings format returned (other)")
	}
	return t, nil
}

func DecodeHTTPTimings(timings json.RawMessage) (*HTTPTimings, error) {
	t := &HTTPTimings{}
	err := json.Unmarshal(timings, t)
	if err != nil {
		return nil, errors.New("invalid timings format returned (other)")
	}
	return t, nil
}

func DecodePingTimings(timings json.RawMessage) ([]PingTiming, error) {
	t := []PingTiming{}
	err := json.Unmarshal(timings, &t)
	if err != nil {
		return nil, errors.New("invalid timings format returned (ping)")
	}
	return t, nil
}

func DecodePingStats(stats json.RawMessage) (*PingStats, error) {
	s := &PingStats{}
	err := json.Unmarshal(stats, s)
	if err != nil {
		return nil, errors.New("invalid stats format returned")
	}
	return s, nil
}

func userAgent() string {
	return fmt.Sprintf("globalping-cli/v%s (https://github.com/jsdelivr/globalping-cli)", version.Version)
}
