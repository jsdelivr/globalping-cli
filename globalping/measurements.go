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
)

var (
	moreCreditsRequiredNoAuthErr = "You only have %s remaining, and %d were required. Try requesting fewer probes or wait %s for the rate limit to reset. You can get higher limits by creating an account. Sign up at https://globalping.io"
	moreCreditsRequiredAuthErr   = "You only have %s remaining, and %d were required. Try requesting fewer probes or wait %s for the rate limit to reset. You can get higher limits by sponsoring us or hosting probes."
	noCreditsNoAuthErr           = "You have run out of credits for this session. You can wait %s for the rate limit to reset or get higher limits by creating an account. Sign up at https://globalping.io"
	noCreditsAuthErr             = "You have run out of credits for this session. You can wait %s for the rate limit to reset or get higher limits by sponsoring us or hosting probes."
)

var (
	StatusUnauthorizedWithTokenRefreshed = 1000
)

func (c *client) CreateMeasurement(measurement *MeasurementCreate) (*MeasurementCreateResponse, error) {
	postData, err := json.Marshal(measurement)
	if err != nil {
		return nil, &MeasurementError{Message: "failed to marshal post data - please report this bug"}
	}

	req, err := http.NewRequest("POST", c.apiURL+"/measurements", bytes.NewBuffer(postData))
	if err != nil {
		return nil, &MeasurementError{Message: "failed to create request - please report this bug"}
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept-Encoding", "br")
	req.Header.Set("Content-Type", "application/json")

	token, tokenType, err := c.accessToken()
	if err != nil {
		return nil, &MeasurementError{Message: "failed to get token: " + err.Error()}
	}
	if token != "" {
		req.Header.Set("Authorization", tokenType+" "+token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, &MeasurementError{Message: "request failed - please try again later"}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		var data MeasurementErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return nil, &MeasurementError{Message: "invalid error format returned - please report this bug"}
		}
		err := data.Error
		err.Code = resp.StatusCode
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

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			token, _, e := c.accessToken()
			if e == nil && token != "" {
				err.Code = StatusUnauthorizedWithTokenRefreshed
			}
			err.Message = "unauthorized: " + data.Error.Message
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
			requestCost, _ := strconv.ParseInt(resp.Header.Get("X-Request-Cost"), 10, 64)
			remaining := rateLimitRemaining + creditsRemaining
			if token == "" {
				if remaining > 0 {
					err.Message = fmt.Sprintf(moreCreditsRequiredNoAuthErr, utils.Pluralize(remaining, "credit"), requestCost, utils.FormatSeconds(rateLimitReset))
					return nil, err
				}
				err.Message = fmt.Sprintf(noCreditsNoAuthErr, utils.FormatSeconds(rateLimitReset))
				return nil, err

			} else {
				if remaining > 0 {
					err.Message = fmt.Sprintf(moreCreditsRequiredAuthErr, utils.Pluralize(remaining, "credit"), requestCost, utils.FormatSeconds(rateLimitReset))
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
	req, err := http.NewRequest("GET", c.apiURL+"/measurements/"+id, nil)
	if err != nil {
		return nil, &MeasurementError{Message: "failed to create request"}
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept-Encoding", "br")

	etag := c.getETag(id)
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
		respBytes := c.getCachedResponse(id)
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
	c.cacheResponse(id, resp.Header.Get("ETag"), respBytes)

	return respBytes, nil
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

func DecodeTracerouteHops(hops json.RawMessage) ([]TracerouteHop, error) {
	t := []TracerouteHop{}
	err := json.Unmarshal(hops, &t)
	if err != nil {
		return nil, &MeasurementError{Message: "invalid hops format returned"}
	}
	return t, nil
}

func DecodeDNSAnswers(answers json.RawMessage) ([]DNSAnswer, error) {
	a := []DNSAnswer{}
	err := json.Unmarshal(answers, &a)
	if err != nil {
		return nil, &MeasurementError{Message: "invalid answers format returned"}
	}
	return a, nil
}

func DecodeTraceDNSHops(hops json.RawMessage) ([]TraceDNSHop, error) {
	t := []TraceDNSHop{}
	err := json.Unmarshal(hops, &t)
	if err != nil {
		return nil, &MeasurementError{Message: "invalid hops format returned"}
	}
	return t, nil
}

func DecodeDNSTimings(timings json.RawMessage) (*DNSTimings, error) {
	t := &DNSTimings{}
	err := json.Unmarshal(timings, t)
	if err != nil {
		return nil, &MeasurementError{Message: "invalid timings format returned (other)"}
	}
	return t, nil
}

func DecodeMTRHops(hops json.RawMessage) ([]MTRHop, error) {
	t := []MTRHop{}
	err := json.Unmarshal(hops, &t)
	if err != nil {
		return nil, &MeasurementError{Message: "invalid hops format returned"}
	}
	return t, nil
}

func DecodeHTTPHeaders(headers json.RawMessage) (map[string]string, error) {
	h := map[string]string{}
	err := json.Unmarshal(headers, &h)
	if err != nil {
		return nil, &MeasurementError{Message: "invalid headers format returned"}
	}
	return h, nil
}

func DecodeHTTPTimings(timings json.RawMessage) (*HTTPTimings, error) {
	t := &HTTPTimings{}
	err := json.Unmarshal(timings, t)
	if err != nil {
		return nil, &MeasurementError{Message: "invalid timings format returned (other)"}
	}
	return t, nil
}

func DecodeHTTPTLS(tls json.RawMessage) (*HTTPTLSCertificate, error) {
	t := &HTTPTLSCertificate{}
	err := json.Unmarshal(tls, t)
	if err != nil {
		return nil, &MeasurementError{Message: "invalid tls format returned"}
	}
	return t, nil
}
