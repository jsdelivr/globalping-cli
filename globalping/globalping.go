package globalping

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/andybalholm/brotli"
	"github.com/jsdelivr/globalping-cli/utils"
	"github.com/jsdelivr/globalping-cli/version"
)

var (
	moreCreditsRequiredNoAuthErr = "You only have %s remaining, and %d were required. Try requesting fewer probes or wait %s for the rate limit to reset. You can get higher limits by creating an account. Sign up at https://globalping.io"
	moreCreditsRequiredAuthErr   = "You only have %s remaining, and %d were required. Try requesting fewer probes or wait %s for the rate limit to reset. You can get higher limits by sponsoring us or hosting probes."
	noCreditsNoAuthErr           = "You have run out of credits for this session. You can wait %s for the rate limit to reset or get higher limits by creating an account. Sign up at https://globalping.io"
	noCreditsAuthErr             = "You have run out of credits for this session. You can wait %s for the rate limit to reset or get higher limits by sponsoring us or hosting probes."
)

func (c *client) CreateMeasurement(measurement *MeasurementCreate) (*MeasurementCreateResponse, error) {
	postData, err := json.Marshal(measurement)
	if err != nil {
		return nil, &MeasurementError{Message: "failed to marshal post data - please report this bug"}
	}

	req, err := http.NewRequest("POST", c.config.GlobalpingAPIURL+"/measurements", bytes.NewBuffer(postData))
	if err != nil {
		return nil, &MeasurementError{Message: "failed to create request - please report this bug"}
	}
	req.Header.Set("User-Agent", userAgent())
	req.Header.Set("Accept-Encoding", "br")
	req.Header.Set("Content-Type", "application/json")

	if c.config.GlobalpingToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.GlobalpingToken)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, &MeasurementError{Message: "request failed - please try again later"}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		var data MeasurementCreateError
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return nil, &MeasurementError{Message: "invalid error format returned - please report this bug"}
		}
		err := &MeasurementError{
			Code: resp.StatusCode,
		}
		if resp.StatusCode == http.StatusBadRequest {
			resErr := ""
			for _, v := range data.Error.Params {
				resErr += fmt.Sprintf(" - %s\n", v)
			}
			// Remove the last \n
			if len(resErr) > 0 {
				resErr = resErr[:len(resErr)-1]
			}
			err.Message = fmt.Sprintf("invalid parameters\n%s", resErr)
			return nil, err
		}

		if resp.StatusCode == http.StatusUnauthorized {
			err.Message = fmt.Sprintf("unauthorized: %s", data.Error.Message)
			return nil, err
		}

		if resp.StatusCode == http.StatusUnprocessableEntity {
			err.Message = "no suitable probes found - please choose a different location"
			return nil, err
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitRemaining, _ := strconv.ParseInt(resp.Header.Get("X-RateLimit-Remaining"), 10, 64)
			rateLimitReset, _ := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64)
			creditsRemaining, _ := strconv.ParseInt(resp.Header.Get("X-Credits-Remaining"), 10, 64)
			creditsRequired, _ := strconv.ParseInt(resp.Header.Get("X-Credits-Required"), 10, 64)
			remaining := rateLimitRemaining + creditsRemaining
			required := rateLimitRemaining + creditsRequired
			if c.config.GlobalpingToken == "" {
				if remaining > 0 {
					err.Message = fmt.Sprintf(moreCreditsRequiredNoAuthErr, utils.Pluralize(remaining, "credit"), required, utils.FormatSeconds(rateLimitReset))
					return nil, err
				}
				err.Message = fmt.Sprintf(noCreditsNoAuthErr, utils.FormatSeconds(rateLimitReset))
				return nil, err

			} else {
				if remaining > 0 {
					err.Message = fmt.Sprintf(moreCreditsRequiredAuthErr, utils.Pluralize(remaining, "credit"), required, utils.FormatSeconds(rateLimitReset))
					return nil, err
				}
				err.Message = fmt.Sprintf(noCreditsAuthErr, utils.FormatSeconds(rateLimitReset))
				return nil, err
			}
		}

		if resp.StatusCode == http.StatusInternalServerError {
			err.Message = "internal server error - please try again later"
			return nil, err
		}

		err.Message = fmt.Sprintf("unknown error response: %s", data.Error.Type)
		return nil, err
	}

	var bodyReader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "br" {
		bodyReader = brotli.NewReader(bodyReader)
	}

	res := &MeasurementCreateResponse{}
	err = json.NewDecoder(bodyReader).Decode(res)
	if err != nil {
		return nil, &MeasurementError{
			Message: fmt.Sprintf("invalid post measurement format returned - please report this bug: %s", err),
		}
	}

	return res, nil
}

func (c *client) GetMeasurement(id string) (*Measurement, error) {
	respBytes, err := c.GetMeasurementRaw(id)
	if err != nil {
		return nil, err
	}
	m := &Measurement{}
	err = json.Unmarshal(respBytes, m)
	if err != nil {
		return nil, &MeasurementError{
			Message: fmt.Sprintf("invalid get measurement format returned: %v %s", err, string(respBytes)),
		}
	}
	return m, nil
}

func (c *client) GetMeasurementRaw(id string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.config.GlobalpingAPIURL+"/measurements/"+id, nil)
	if err != nil {
		return nil, &MeasurementError{Message: "failed to create request"}
	}

	req.Header.Set("User-Agent", userAgent())
	req.Header.Set("Accept-Encoding", "br")

	etag := c.etags[id]
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, &MeasurementError{Message: "request failed"}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotModified {
		err := &MeasurementError{
			Code: resp.StatusCode,
		}
		if resp.StatusCode == http.StatusNotFound {
			err.Message = "measurement not found"
			return nil, err
		}

		if resp.StatusCode == http.StatusInternalServerError {
			err.Message = "internal server error - please try again later"
			return nil, err
		}
		err.Message = fmt.Sprintf("response code %d", resp.StatusCode)
		return nil, err
	}

	if resp.StatusCode == http.StatusNotModified {
		respBytes := c.measurements[etag]
		if respBytes == nil {
			return nil, &MeasurementError{Message: "response not found in etags cache"}
		}
		return respBytes, nil
	}

	var bodyReader io.Reader = resp.Body

	if resp.Header.Get("Content-Encoding") == "br" {
		bodyReader = brotli.NewReader(bodyReader)
	}

	// Read the response body
	respBytes, err := io.ReadAll(bodyReader)
	if err != nil {
		return nil, &MeasurementError{Message: "failed to read response body"}
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
		return nil, &MeasurementError{Message: "invalid timings format returned (other)"}
	}
	return t, nil
}

func DecodeHTTPTimings(timings json.RawMessage) (*HTTPTimings, error) {
	t := &HTTPTimings{}
	err := json.Unmarshal(timings, t)
	if err != nil {
		return nil, &MeasurementError{Message: "invalid timings format returned (other)"}
	}
	return t, nil
}

func DecodePingTimings(timings json.RawMessage) ([]PingTiming, error) {
	t := []PingTiming{}
	err := json.Unmarshal(timings, &t)
	if err != nil {
		return nil, &MeasurementError{Message: "invalid timings format returned (ping)"}
	}
	return t, nil
}

func DecodePingStats(stats json.RawMessage) (*PingStats, error) {
	s := &PingStats{}
	err := json.Unmarshal(stats, s)
	if err != nil {
		return nil, &MeasurementError{Message: "invalid stats format returned"}
	}
	return s, nil
}

func userAgent() string {
	return fmt.Sprintf("globalping-cli/v%s (https://github.com/jsdelivr/globalping-cli)", version.Version)
}
