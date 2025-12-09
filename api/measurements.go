package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jsdelivr/globalping-cli/utils"
	"github.com/jsdelivr/globalping-go"
)

var (
	moreCreditsRequiredNoAuthErr = "You only have %s remaining, and %d were required. Try requesting fewer probes or wait %s for the rate limit to reset. You can get higher limits by creating an account. Sign up at https://dash.globalping.io?view=add-credits"
	moreCreditsRequiredAuthErr   = "You only have %s remaining, and %d were required. Try requesting fewer probes or wait %s for the rate limit to reset. You can get higher limits by sponsoring us or hosting probes. Learn more at https://dash.globalping.io?view=add-credits"
	noCreditsNoAuthErr           = "You have run out of credits for this session. You can wait %s for the rate limit to reset or get higher limits by creating an account. Sign up at https://dash.globalping.io?view=add-credits"
	noCreditsAuthErr             = "You have run out of credits for this session. You can wait %s for the rate limit to reset or get higher limits by sponsoring us or hosting probes. Learn more at https://dash.globalping.io?view=add-credits"
	invalidRefreshTokenErr       = "You have been signed out by the API. Please try signing in again."
	invalidTokenErr              = "Your access token has been rejected by the API. Try signing in with a new token."
)

var (
	StatusUnauthorizedWithTokenRefreshed = 1000
)

func (c *client) CreateMeasurement(ctx context.Context, measurement *globalping.MeasurementCreate) (*globalping.MeasurementCreateResponse, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, err
	}
	if token != nil {
		c.globalping.SetToken(token.AccessToken)
	} else {
		c.globalping.SetToken("")
	}

	res, err := c.globalping.CreateMeasurement(ctx, measurement)
	if err == nil {
		return res, nil
	}

	apiErr, ok := err.(*globalping.MeasurementError)
	if !ok {
		return nil, err
	}

	if apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden {
		if token != nil {
			if token.RefreshToken == "" {
				apiErr.Message = invalidTokenErr
				return nil, apiErr
			}
			if c.tryToRefreshToken(ctx, token.RefreshToken) {
				apiErr.StatusCode = StatusUnauthorizedWithTokenRefreshed
				return nil, apiErr
			}
			apiErr.Message = invalidRefreshTokenErr
			return nil, apiErr
		}
		return nil, apiErr
	}

	if apiErr.StatusCode == http.StatusUnprocessableEntity {
		apiErr.Message = fmt.Sprintf("%s - please try a different location", utils.TextFromSentence(apiErr.Message))
		return nil, apiErr
	}

	if apiErr.StatusCode == http.StatusTooManyRequests {
		rateLimitRemaining, _ := strconv.ParseInt(apiErr.Header.Get("X-RateLimit-Remaining"), 10, 64)
		rateLimitReset, _ := strconv.ParseInt(apiErr.Header.Get("X-RateLimit-Reset"), 10, 64)
		creditsRemaining, _ := strconv.ParseInt(apiErr.Header.Get("X-Credits-Remaining"), 10, 64)
		requestCost, _ := strconv.ParseInt(apiErr.Header.Get("X-Request-Cost"), 10, 64)
		remaining := rateLimitRemaining + creditsRemaining
		if token == nil {
			if remaining > 0 {
				apiErr.Message = fmt.Sprintf(moreCreditsRequiredNoAuthErr, utils.Pluralize(remaining, "credit"), requestCost, utils.FormatSeconds(rateLimitReset))
				return nil, apiErr
			}
			apiErr.Message = fmt.Sprintf(noCreditsNoAuthErr, utils.FormatSeconds(rateLimitReset))
			return nil, apiErr

		} else {
			if remaining > 0 {
				apiErr.Message = fmt.Sprintf(moreCreditsRequiredAuthErr, utils.Pluralize(remaining, "credit"), requestCost, utils.FormatSeconds(rateLimitReset))
				return nil, apiErr
			}
			apiErr.Message = fmt.Sprintf(noCreditsAuthErr, utils.FormatSeconds(rateLimitReset))
			return nil, apiErr
		}
	}

	return nil, apiErr
}

func (c *client) GetMeasurement(ctx context.Context, id string) (*globalping.Measurement, error) {
	return c.globalping.GetMeasurement(ctx, id)
}

func (c *client) AwaitMeasurement(ctx context.Context, id string) (*globalping.Measurement, error) {
	return c.globalping.AwaitMeasurement(ctx, id)
}

func (c *client) GetMeasurementRaw(ctx context.Context, id string) ([]byte, error) {
	return c.globalping.GetMeasurementRaw(ctx, id)
}
