package globalping

import (
	"encoding/json"
	"net/http"
)

// https://www.jsdelivr.com/docs/api.globalping.io#get-/v1/limits
type LimitsResponse struct {
	RateLimits RateLimits   `json:"rateLimit"`
	Credits    CreditLimits `json:"credits"` // Only for authenticated requests
}

type RateLimits struct {
	Measurements MeasurementsLimits `json:"measurements"`
}

type MeasurementsLimits struct {
	Create MeasurementsCreateLimits `json:"create"`
}

type CreateLimitType string

const (
	CreateLimitTypeIP   CreateLimitType = "ip"
	CreateLimitTypeUser CreateLimitType = "user"
)

type MeasurementsCreateLimits struct {
	Type      CreateLimitType `json:"type"`
	Limit     int64           `json:"limit"`
	Remaining int64           `json:"remaining"`
	Reset     int64           `json:"reset"`
}

type CreditLimits struct {
	Remaining int64 `json:"remaining"`
}

type LimitsError struct {
	Code    int    `json:"-"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (e *LimitsError) Error() string {
	return e.Message
}

type LimitsErrorResponse struct {
	Error *LimitsError `json:"error"`
}

func (c *client) Limits() (*LimitsResponse, error) {
	req, err := http.NewRequest("GET", c.apiURL+"/limits", nil)
	if err != nil {
		return nil, &LimitsError{Message: "failed to create request - please report this bug"}
	}
	token, err := c.getToken()
	if err != nil {
		return nil, &LimitsError{Message: "failed to get token: " + err.Error()}
	}
	if token != nil {
		req.Header.Set("Authorization", token.TokenType+" "+token.AccessToken)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, &LimitsError{Message: "request failed - please try again later"}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errResp := &LimitsErrorResponse{
			Error: &LimitsError{
				Code:    resp.StatusCode,
				Type:    "unexpected_status_code",
				Message: "unexpected status code: " + resp.Status,
			},
		}
		json.NewDecoder(resp.Body).Decode(errResp)
		return nil, errResp.Error
	}
	limits := &LimitsResponse{}
	err = json.NewDecoder(resp.Body).Decode(limits)
	if err != nil {
		return nil, &LimitsError{Message: "invalid format returned - please report this bug"}
	}
	return limits, nil
}
