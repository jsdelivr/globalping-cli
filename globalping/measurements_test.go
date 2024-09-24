package globalping

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/andybalholm/brotli"

	"github.com/stretchr/testify/assert"
)

func Test_CreateMeasurement_Valid(t *testing.T) {
	server := generateServer(`{"id":"abcd","probesCount":1}`, http.StatusAccepted)
	defer server.Close()
	client := NewClient(Config{APIURL: server.URL})

	opts := &MeasurementCreate{}
	res, err := client.CreateMeasurement(opts)

	assert.Equal(t, "abcd", res.ID)
	assert.Equal(t, 1, res.ProbesCount)
	assert.NoError(t, err)
}

func Test_CreateMeasurement_Authorized(t *testing.T) {
	server := generateServerAuthorized(`{"id":"abcd","probesCount":1}`)
	defer server.Close()
	client := NewClient(Config{
		AuthToken: &Token{
			AccessToken: "secret",
			Expiry:      time.Now().Add(1 * time.Hour),
		},
		APIURL: server.URL,
	})

	opts := &MeasurementCreate{}
	res, err := client.CreateMeasurement(opts)

	assert.Equal(t, "abcd", res.ID)
	assert.Equal(t, 1, res.ProbesCount)
	assert.NoError(t, err)
}

func Test_CreateMeasurement_AuthorizedError(t *testing.T) {
	server := generateServerAuthorized(`{"id":"abcd","probesCount":1}`)
	defer server.Close()
	client := NewClient(Config{
		APIURL: server.URL,
	})

	opts := &MeasurementCreate{}
	res, err := client.CreateMeasurement(opts)

	assert.Nil(t, res)
	assert.EqualError(t, err, "unauthorized: Unauthorized.")
}

func Test_CreateMeasurement_TokenRefreshed(t *testing.T) {
	now := time.Now()
	timeNow = func() time.Time {
		return now
	}
	defer func() {
		timeNow = time.Now
	}()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST request, got %s", r.Method)
			}
			err := r.ParseForm()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "<client_id>", r.Form.Get("client_id"))
			assert.Equal(t, "<client_secret>", r.Form.Get("client_secret"))
			assert.Equal(t, "refresh_token", r.Form.Get("grant_type"))
			assert.Equal(t, "refresh_tok3n", r.Form.Get("refresh_token"))

			w.Header().Set("Content-Type", "application/json")
			_, err = w.Write([]byte(`{"access_token":"new_token","token_type":"Bearer","refresh_token":"new_refresh_token","expires_in":3600}`))
			if err != nil {
				t.Fatal(err)
			}
			return
		}
		if r.URL.Path == "/measurements" {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST request, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			_, err := w.Write([]byte(`{"id":"abcd","probesCount":1}`))
			if err != nil {
				t.Fatal(err)
			}
			return
		}
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()
	client := NewClient(Config{
		APIURL:           server.URL,
		AuthURL:          server.URL,
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthToken: &Token{
			AccessToken:  "access_token",
			RefreshToken: "refresh_tok3n",
			Expiry:       time.Now().Add(-1 * time.Hour),
		},
		OnTokenRefresh: func(_t *Token) {
			assert.Equal(t, &Token{
				AccessToken:  "new_token",
				TokenType:    "Bearer",
				RefreshToken: "new_refresh_token",
				ExpiresIn:    3600,
				Expiry:       now.Add(3600 * time.Second),
			}, _t)
		},
	})

	opts := &MeasurementCreate{}
	res, err := client.CreateMeasurement(opts)
	assert.Nil(t, err)
	assert.Equal(t, "abcd", res.ID)
}

func Test_CreateMeasurement_Unauthorized_TokenRefreshed(t *testing.T) {
	now := time.Now()
	timeNow = func() time.Time {
		return now
	}
	defer func() {
		timeNow = time.Now
	}()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST request, got %s", r.Method)
			}
			err := r.ParseForm()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "<client_id>", r.Form.Get("client_id"))
			assert.Equal(t, "<client_secret>", r.Form.Get("client_secret"))
			assert.Equal(t, "refresh_token", r.Form.Get("grant_type"))
			assert.Equal(t, "refresh_tok3n", r.Form.Get("refresh_token"))

			w.Header().Set("Content-Type", "application/json")
			_, err = w.Write([]byte(`{"access_token":"new_token","token_type":"Bearer","refresh_token":"new_refresh_token","expires_in":3600}`))
			if err != nil {
				t.Fatal(err)
			}
			return
		}
		if r.URL.Path == "/measurements" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": {"type": "unauthorized", "message": "Unauthorized."}}`))
			return
		}
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()
	client := NewClient(Config{
		APIURL:           server.URL,
		AuthURL:          server.URL,
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthToken: &Token{
			AccessToken:  "access_token",
			RefreshToken: "refresh_tok3n",
			Expiry:       time.Now().Add(1 * time.Hour),
		},
		OnTokenRefresh: func(_t *Token) {
			assert.Equal(t, &Token{
				AccessToken:  "new_token",
				TokenType:    "Bearer",
				RefreshToken: "new_refresh_token",
				ExpiresIn:    3600,
				Expiry:       now.Add(3600 * time.Second),
			}, _t)
		},
	})

	opts := &MeasurementCreate{}
	res, err := client.CreateMeasurement(opts)
	assert.Nil(t, res)
	e, ok := err.(*MeasurementError)
	assert.True(t, ok)
	assert.Equal(t, StatusUnauthorizedWithTokenRefreshed, e.Code)
	assert.Equal(t, "unauthorized: Unauthorized.", e.Message)
}

func Test_CreateMeasurement_Unauthorized_Token_Not_Refreshed(t *testing.T) {
	now := time.Now()
	timeNow = func() time.Time {
		return now
	}
	defer func() {
		timeNow = time.Now
	}()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "invalid_grant", "error_description": "Invalid refresh token."}`))
			return
		}
		if r.URL.Path == "/measurements" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": {"type": "unauthorized", "message": "Unauthorized."}}`))
			return
		}
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()
	isOnTokenRefreshCalled := false
	client := NewClient(Config{
		APIURL:           server.URL,
		AuthURL:          server.URL,
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthToken: &Token{
			AccessToken:  "access_token",
			RefreshToken: "refresh_tok3n",
			Expiry:       time.Now().Add(1 * time.Hour),
		},
		OnTokenRefresh: func(_t *Token) {
			isOnTokenRefreshCalled = true
			assert.Nil(t, _t)
		},
	})

	opts := &MeasurementCreate{}
	res, err := client.CreateMeasurement(opts)
	assert.Nil(t, res)
	assert.EqualError(t, err, "unauthorized: Unauthorized.")
	assert.True(t, isOnTokenRefreshCalled)
}
func Test_CreateMeasurement_Unauthorized_NoRefreshToken(t *testing.T) {
	now := time.Now()
	timeNow = func() time.Time {
		return now
	}
	defer func() {
		timeNow = time.Now
	}()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": {"type": "unauthorized", "message": "Unauthorized."}}`))
	}))
	defer server.Close()
	client := NewClient(Config{
		APIURL:           server.URL,
		AuthURL:          server.URL,
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthToken: &Token{
			AccessToken: "access_token",
			Expiry:      time.Now().Add(1 * time.Hour),
		},
		OnTokenRefresh: func(_t *Token) {
			t.Fatal("should not be called")
		},
	})

	opts := &MeasurementCreate{}
	res, err := client.CreateMeasurement(opts)
	assert.Nil(t, res)
	assert.EqualError(t, err, invalidTokenErr)
}

func Test_CreateMeasurement_MoreCreditsRequiredNoAuthError(t *testing.T) {
	rateLimitReset := "61"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "1")
		w.Header().Set("X-RateLimit-Reset", rateLimitReset)
		w.Header().Set("X-Credits-Remaining", "1")
		w.Header().Set("X-Request-Cost", "3")
		w.WriteHeader(429)
		_, err := w.Write([]byte(`{
			"error": {
			  "message": "API rate limit exceeded.",
			  "type": "rate_limit_exceeded"
			}}`))
		if err != nil {
			panic(err)
		}
	}))
	defer server.Close()

	client := NewClient(Config{APIURL: server.URL})
	opts := &MeasurementCreate{}
	_, err := client.CreateMeasurement(opts)
	assert.EqualError(t, err, fmt.Sprintf(moreCreditsRequiredNoAuthErr, "2 credits", 3, "1 minute"))

	rateLimitReset = "100"
	_, err = client.CreateMeasurement(opts)
	assert.EqualError(t, err, fmt.Sprintf(moreCreditsRequiredNoAuthErr, "2 credits", 3, "2 minutes"))
}

func Test_CreateMeasurement_MoreCreditsRequiredAuthError(t *testing.T) {
	rateLimitReset := "40"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", rateLimitReset)
		w.Header().Set("X-Credits-Remaining", "1")
		w.Header().Set("X-Request-Cost", "2")
		w.WriteHeader(429)
		_, err := w.Write([]byte(`{
			"error": {
			  "message": "API rate limit exceeded.",
			  "type": "rate_limit_exceeded"
			}}`))
		if err != nil {
			panic(err)
		}
	}))
	defer server.Close()

	client := NewClient(Config{
		AuthToken: &Token{
			AccessToken: "secret",
			Expiry:      time.Now().Add(1 * time.Hour),
		},
		APIURL: server.URL,
	})
	opts := &MeasurementCreate{}

	_, err := client.CreateMeasurement(opts)
	assert.EqualError(t, err, fmt.Sprintf(moreCreditsRequiredAuthErr, "1 credit", 2, "40 seconds"))

	rateLimitReset = "1"
	_, err = client.CreateMeasurement(opts)
	assert.EqualError(t, err, fmt.Sprintf(moreCreditsRequiredAuthErr, "1 credit", 2, "1 second"))
}

func Test_CreateMeasurement_NoCreditsNoAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", "5")
		w.Header().Set("X-Credits-Remaining", "0")
		w.WriteHeader(429)
		_, err := w.Write([]byte(`{
			"error": {
			  "message": "API rate limit exceeded.",
			  "type": "rate_limit_exceeded"
			}}`))
		if err != nil {
			panic(err)
		}
	}))
	defer server.Close()

	client := NewClient(Config{APIURL: server.URL})
	opts := &MeasurementCreate{}
	_, err := client.CreateMeasurement(opts)

	assert.EqualError(t, err, fmt.Sprintf(noCreditsNoAuthErr, "5 seconds"))
}

func Test_CreateMeasurement_NoCreditsAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", "5")
		w.Header().Set("X-Credits-Remaining", "0")
		w.WriteHeader(429)
		_, err := w.Write([]byte(`{
			"error": {
			  "message": "API rate limit exceeded.",
			  "type": "rate_limit_exceeded"
			}}`))
		if err != nil {
			panic(err)
		}
	}))
	defer server.Close()

	client := NewClient(Config{
		AuthToken: &Token{
			AccessToken: "secret",
			Expiry:      time.Now().Add(1 * time.Hour),
		},
		APIURL: server.URL,
	})
	opts := &MeasurementCreate{}
	_, err := client.CreateMeasurement(opts)

	assert.EqualError(t, err, fmt.Sprintf(noCreditsAuthErr, "5 seconds"))
}

func Test_CreateMeasurement_NoProbes(t *testing.T) {
	server := generateServer(`{
    "error": {
      "message": "No suitable probes found",
      "type": "no_probes_found"
    }}`, 422)
	defer server.Close()

	client := NewClient(Config{APIURL: server.URL})
	opts := &MeasurementCreate{}
	_, err := client.CreateMeasurement(opts)

	assert.Equal(t, &MeasurementError{
		Code:    422,
		Type:    "no_probes_found",
		Message: "no suitable probes found - please choose a different location",
	}, err)
}

func Test_CreateMeasurement_Validation(t *testing.T) {
	server := generateServer(`{
    "error": {
        "message": "Validation Failed",
        "type": "validation_error",
        "params": {
			"target": "\"target\" does not match any of the allowed types"
        }
    }}`, 400)
	defer server.Close()
	client := NewClient(Config{APIURL: server.URL})

	opts := &MeasurementCreate{}
	_, err := client.CreateMeasurement(opts)

	assert.Equal(t, &MeasurementError{
		Code: 400,
		Type: "validation_error",
		Message: `invalid parameters
 - "target" does not match any of the allowed types`,
		Params: map[string]interface{}{
			"target": "\"target\" does not match any of the allowed types",
		},
	}, err)
}

func Test_CreateMeasurement_InternalError(t *testing.T) {
	server := generateServer(`{
    "error": {
      "message": "Internal Server Error",
      "type": "api_error"
    }}`, 500)
	defer server.Close()
	client := NewClient(Config{APIURL: server.URL})

	opts := &MeasurementCreate{}
	_, err := client.CreateMeasurement(opts)
	assert.EqualError(t, err, "internal server error - please try again later")
}

func Test_GetMeasurement_Valid(t *testing.T) {
	server := generateServer(`{"id":"abcd"}`, http.StatusOK)
	defer server.Close()
	client := NewClient(Config{APIURL: server.URL})
	res, err := client.GetMeasurement("abcd")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "abcd", res.ID)
}

func Test_GetMeasurement_Ping(t *testing.T) {
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
	}]}`, http.StatusOK)
	defer server.Close()
	client := NewClient(Config{APIURL: server.URL})

	res, err := client.GetMeasurement("abcd")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "abcd", res.ID)
	assert.Equal(t, "ping", res.Type)
	assert.Equal(t, StatusFinished, res.Status)
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
	stats, err := DecodePingStats(res.Results[0].Result.StatsRaw)
	assert.NoError(t, err)
	assert.Equal(t, float64(27.088), stats.Avg)
	assert.Equal(t, float64(28.193), stats.Max)
	assert.Equal(t, float64(24.891), stats.Min)
	assert.Equal(t, 3, stats.Total)
	assert.Equal(t, 3, stats.Rcv)
	assert.Equal(t, 0, stats.Drop)
	assert.Equal(t, float64(0), stats.Loss)
}

func Test_GetMeasurement_Traceroute(t *testing.T) {
	server := generateServer(`{
	"id": "abcd",
	"type": "traceroute",
	"status": "finished",
	"createdAt": "2023-02-23T07:55:23.414Z",
	"updatedAt": "2023-02-23T07:55:25.496Z",
	"probesCount": 1,
	"results": [
		{
		"probe": {
			"continent": "EU",
			"region": "Northern Europe",
			"country": "GB",
			"state": null,
			"city": "London",
			"asn": 16276,
			"longitude": -0.1257,
			"latitude": 51.5085,
			"network": "OVH SAS",
			"tags": [],
			"resolvers": [
			"private"
			]
		},
		"result": {
			"rawOutput": "TRACEROUTE",
			"status": "finished",
			"resolvedAddress": "1.1.1.1",
			"resolvedHostname": "1.1.1.1",
			"hops": [
			{
				"resolvedHostname": "54.37.244.252",
				"resolvedAddress": "54.37.244.252",
				"timings": [
				{
					"rtt": 0.408
				},
				{
					"rtt": 0.502
				}
				]
			},
			{
				"resolvedHostname": "93.123.11.62",
				"resolvedAddress": "93.123.11.62",
				"timings": [
				{
					"rtt": 0.507
				},
				{
					"rtt": 0.524
				}
				]
			}
			]
	}}]}`, http.StatusOK)
	defer server.Close()

	client := NewClient(Config{APIURL: server.URL})

	res, err := client.GetMeasurement("abcd")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "abcd", res.ID)
	assert.Equal(t, "traceroute", res.Type)
	assert.Equal(t, StatusFinished, res.Status)
	assert.Equal(t, "2023-02-23T07:55:23.414Z", res.CreatedAt)
	assert.Equal(t, "2023-02-23T07:55:25.496Z", res.UpdatedAt)
	assert.Equal(t, 1, res.ProbesCount)
	assert.Equal(t, 1, len(res.Results))

	assert.Equal(t, "EU", res.Results[0].Probe.Continent)
	assert.Equal(t, "Northern Europe", res.Results[0].Probe.Region)
	assert.Equal(t, "GB", res.Results[0].Probe.Country)
	assert.Equal(t, "", res.Results[0].Probe.State)
	assert.Equal(t, "London", res.Results[0].Probe.City)
	assert.Equal(t, 16276, res.Results[0].Probe.ASN)
	assert.Equal(t, "OVH SAS", res.Results[0].Probe.Network)
	assert.Equal(t, 0, len(res.Results[0].Probe.Tags))

	assert.Equal(t, "TRACEROUTE", res.Results[0].Result.RawOutput)
	assert.Equal(t, "1.1.1.1", res.Results[0].Result.ResolvedAddress)
	assert.Equal(t, "1.1.1.1", res.Results[0].Result.ResolvedHostname)
}

func Test_GetMeasurement_Dns(t *testing.T) {
	server := generateServer(`{
	"id": "abcd",
	"type": "dns",
	"status": "finished",
	"createdAt": "2023-02-23T08:00:37.431Z",
	"updatedAt": "2023-02-23T08:00:37.640Z",
	"probesCount": 1,
	"results": [
		{
		"probe": {
			"continent": "EU",
			"region": "Western Europe",
			"country": "NL",
			"state": null,
			"city": "Amsterdam",
			"asn": 60404,
			"longitude": 4.8897,
			"latitude": 52.374,
			"network": "Liteserver",
			"tags": [],
			"resolvers": [
			"185.31.172.240",
			"89.188.29.4"
			]
		},
		"result": {
			"status": "finished",
			"statusCodeName": "NOERROR",
			"statusCode": 0,
			"rawOutput": "DNS",
			"answers": [
			{
				"name": "jsdelivr.com.",
				"type": "A",
				"ttl": 30,
				"class": "IN",
				"value": "92.223.84.84"
			}
			],
			"timings": {
			"total": 15
			},
			"resolver": "185.31.172.240"
		}
	}]}`, http.StatusOK)
	defer server.Close()
	client := NewClient(Config{APIURL: server.URL})

	res, err := client.GetMeasurement("abcd")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "abcd", res.ID)
	assert.Equal(t, "dns", res.Type)
	assert.Equal(t, StatusFinished, res.Status)
	assert.Equal(t, "2023-02-23T08:00:37.431Z", res.CreatedAt)
	assert.Equal(t, "2023-02-23T08:00:37.640Z", res.UpdatedAt)
	assert.Equal(t, 1, res.ProbesCount)
	assert.Equal(t, 1, len(res.Results))

	assert.Equal(t, "EU", res.Results[0].Probe.Continent)
	assert.Equal(t, "Western Europe", res.Results[0].Probe.Region)
	assert.Equal(t, "NL", res.Results[0].Probe.Country)
	assert.Equal(t, "", res.Results[0].Probe.State)
	assert.Equal(t, "Amsterdam", res.Results[0].Probe.City)
	assert.Equal(t, 60404, res.Results[0].Probe.ASN)
	assert.Equal(t, "Liteserver", res.Results[0].Probe.Network)
	assert.Equal(t, 0, len(res.Results[0].Probe.Tags))

	assert.Equal(t, "DNS", res.Results[0].Result.RawOutput)
	assert.Equal(t, StatusFinished, res.Results[0].Result.Status)
	assert.IsType(t, json.RawMessage{}, res.Results[0].Result.TimingsRaw)

	// Test timings
	timings, _ := DecodeDNSTimings(res.Results[0].Result.TimingsRaw)
	assert.Equal(t, float64(15), timings.Total)
}

func Test_GetMeasurement_Mtr(t *testing.T) {
	server := generateServer(`{
	"id": "abcd",
	"type": "mtr",
	"status": "finished",
	"createdAt": "2023-02-23T08:08:25.187Z",
	"updatedAt": "2023-02-23T08:08:29.829Z",
	"probesCount": 1,
	"results": [
		{
		"probe": {
			"continent": "EU",
			"region": "Western Europe",
			"country": "NL",
			"state": null,
			"city": "Amsterdam",
			"asn": 54825,
			"longitude": 4.8897,
			"latitude": 52.374,
			"network": "Packet Host, Inc.",
			"tags": [],
			"resolvers": []
		},
		"result": {
			"status": "finished",
			"rawOutput": "MTR",
			"resolvedAddress": "92.223.84.84",
			"resolvedHostname": "92.223.84.84",
			"hops": [
			{
				"stats": {
				"min": 0.176,
				"max": 0.226,
				"avg": 0.2,
				"total": 3,
				"loss": 0,
				"rcv": 3,
				"drop": 0,
				"stDev": 0,
				"jMin": 0,
				"jMax": 0.2,
				"jAvg": 0.1
				},
				"asn": [],
				"timings": [
				{
					"rtt": 0.176
				},
				{
					"rtt": 0.216
				},
				{
					"rtt": 0.226
				}
				],
				"resolvedAddress": "172.19.66.225",
				"duplicate": false,
				"resolvedHostname": "172.19.66.225"
			},
			{
				"stats": {
				"min": 0.894,
				"max": 0.894,
				"avg": 0.9,
				"total": 1,
				"loss": 0,
				"rcv": 1,
				"drop": 0,
				"stDev": 0,
				"jMin": 0.9,
				"jMax": 0.9,
				"jAvg": 0.9
				},
				"asn": [
				199524
				],
				"timings": [
				{
					"rtt": 0.894
				}
				],
				"resolvedAddress": "92.223.84.84",
				"duplicate": true,
				"resolvedHostname": "92.223.84.84"
			}
			]
		}
	}]}`, http.StatusOK)
	defer server.Close()
	client := NewClient(Config{APIURL: server.URL})

	res, err := client.GetMeasurement("abcd")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "abcd", res.ID)
	assert.Equal(t, "mtr", res.Type)
	assert.Equal(t, StatusFinished, res.Status)
	assert.Equal(t, "2023-02-23T08:08:25.187Z", res.CreatedAt)
	assert.Equal(t, "2023-02-23T08:08:29.829Z", res.UpdatedAt)
	assert.Equal(t, 1, res.ProbesCount)
	assert.Equal(t, 1, len(res.Results))

	assert.Equal(t, "EU", res.Results[0].Probe.Continent)
	assert.Equal(t, "Western Europe", res.Results[0].Probe.Region)
	assert.Equal(t, "NL", res.Results[0].Probe.Country)
	assert.Equal(t, "", res.Results[0].Probe.State)
	assert.Equal(t, "Amsterdam", res.Results[0].Probe.City)
	assert.Equal(t, 54825, res.Results[0].Probe.ASN)
	assert.Equal(t, "Packet Host, Inc.", res.Results[0].Probe.Network)
	assert.Equal(t, 0, len(res.Results[0].Probe.Tags))

	assert.Equal(t, "MTR", res.Results[0].Result.RawOutput)
	assert.Equal(t, StatusFinished, res.Results[0].Result.Status)
	assert.IsType(t, json.RawMessage{}, res.Results[0].Result.TimingsRaw)
}

func Test_GetMeasurement_Http(t *testing.T) {
	server := generateServer(`{
	"id": "abcd",
	"type": "http",
	"status": "finished",
	"createdAt": "2023-02-23T08:16:11.335Z",
	"updatedAt": "2023-02-23T08:16:12.548Z",
	"probesCount": 1,
	"results": [
		{
		"probe": {
			"continent": "NA",
			"region": "Northern America",
			"country": "CA",
			"state": null,
			"city": "Pembroke",
			"asn": 577,
			"longitude": -77.1162,
			"latitude": 45.8168,
			"network": "Bell Canada",
			"tags": [],
			"resolvers": [
			"private",
			"private"
			]
		},
		"result": {
			"status": "finished",
			"resolvedAddress": "5.101.222.14",
			"headers": {
			"server": "nginx",
			"date": "Thu, 23 Feb 2023 08:16:12 GMT",
			"content-type": "text/html; charset=utf-8",
			"connection": "close",
			"location": "/",
			"cf-ray": "79de849d3fa30c33-AMS",
			"vary": "Accept-Encoding",
			"cf-cache-status": "DYNAMIC",
			"x-render-origin-server": "Render",
			"x-response-time": "1ms",
			"cache": "MISS, MISS",
			"x-id": "am3-up-gc88, td2-up-gc10",
			"x-nginx": "nginx-be, nginx-be"
			},
			"rawHeaders": "Server: nginx\nDate: Thu, 23 Feb 2023 08:16:12 GMT\nContent-Type: text/html; charset=utf-8\nConnection: close\nLocation: /\nCF-Ray: 79de849d3fa30c33-AMS\nVary: Accept-Encoding\nCF-Cache-Status: DYNAMIC\nx-render-origin-server: Render\nx-response-time: 1ms\nCache: MISS\nX-ID: am3-up-gc88\nX-NGINX: nginx-be\nCache: MISS\nX-ID: td2-up-gc10\nX-NGINX: nginx-be",
			"rawBody": null,
			"statusCode": 301,
			"statusCodeName": "Moved Permanently",
			"timings": {
			"total": 583,
			"download": 18,
			"firstByte": 450,
			"dns": 24,
			"tls": 70,
			"tcp": 19
			},
			"tls": {
			"authorized": true,
			"createdAt": "2023-02-18T00:00:00.000Z",
			"expiresAt": "2024-02-18T23:59:59.000Z",
			"issuer": {
				"C": "GB",
				"ST": "Greater Manchester",
				"L": "Salford",
				"O": "Sectigo Limited",
				"CN": "Sectigo RSA Domain Validation Secure Server CA"
			},
			"subject": {
				"CN": "jsdelivr.com",
				"alt": "DNS:jsdelivr.com, DNS:data.jsdelivr.com, DNS:www.jsdelivr.com"
			}
			},
			"rawOutput": "HTTP"
		}
	}]}`, http.StatusOK)
	defer server.Close()
	client := NewClient(Config{APIURL: server.URL})

	res, err := client.GetMeasurement("abcd")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "abcd", res.ID)
	assert.Equal(t, "http", res.Type)
	assert.Equal(t, StatusFinished, res.Status)
	assert.Equal(t, "2023-02-23T08:16:11.335Z", res.CreatedAt)
	assert.Equal(t, "2023-02-23T08:16:12.548Z", res.UpdatedAt)
	assert.Equal(t, 1, res.ProbesCount)
	assert.Equal(t, 1, len(res.Results))

	assert.Equal(t, "NA", res.Results[0].Probe.Continent)
	assert.Equal(t, "Northern America", res.Results[0].Probe.Region)
	assert.Equal(t, "CA", res.Results[0].Probe.Country)
	assert.Equal(t, "", res.Results[0].Probe.State)
	assert.Equal(t, "Pembroke", res.Results[0].Probe.City)
	assert.Equal(t, 577, res.Results[0].Probe.ASN)
	assert.Equal(t, "Bell Canada", res.Results[0].Probe.Network)
	assert.Equal(t, 0, len(res.Results[0].Probe.Tags))

	assert.Equal(t, "HTTP", res.Results[0].Result.RawOutput)
	assert.Equal(t, StatusFinished, res.Results[0].Result.Status)
	assert.IsType(t, json.RawMessage{}, res.Results[0].Result.TimingsRaw)

	// Test timings
	timings, _ := DecodeHTTPTimings(res.Results[0].Result.TimingsRaw)
	assert.Equal(t, 583, timings.Total)
	assert.Equal(t, 18, timings.Download)
	assert.Equal(t, 450, timings.FirstByte)
	assert.Equal(t, 24, timings.DNS)
	assert.Equal(t, 70, timings.TLS)
	assert.Equal(t, 19, timings.TCP)
}

func Test_GetMeasurement_WithEtag(t *testing.T) {
	id1 := "123abc"
	id2 := "567xyz"

	cacheMissCount := 0
	cacheHitCount := 0

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		id := parts[len(parts)-1]

		etag := func(id string) string {
			return "etag-" + id
		}

		if r.Header.Get("If-None-Match") == etag(id) {
			// cache hit
			cacheHitCount++
			w.Header().Set("ETag", etag(id))
			w.WriteHeader(http.StatusNotModified)

			return
		}

		// cache miss, return full response
		cacheMissCount++
		m := &Measurement{
			ID: id,
		}

		w.Header().Set("ETag", etag(id))

		err := json.NewEncoder(w).Encode(m)
		assert.NoError(t, err)
	}))

	client := NewClient(Config{APIURL: s.URL})

	// first request for id1
	m, err := client.GetMeasurement(id1)
	assert.NoError(t, err)

	assert.Equal(t, id1, m.ID)

	// first request for id1
	m, err = client.GetMeasurement(id2)
	assert.NoError(t, err)

	assert.Equal(t, id2, m.ID)

	// second request for id1
	m, err = client.GetMeasurement(id2)
	assert.NoError(t, err)

	assert.Equal(t, id2, m.ID)

	assert.Equal(t, 1, cacheHitCount)
	assert.Equal(t, 2, cacheMissCount)
}

func Test_GetMeasurement_WithBrotli(t *testing.T) {
	id := "123abc"

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		id := parts[len(parts)-1]

		assert.Equal(t, "br", r.Header.Get("Accept-Encoding"))

		m := &Measurement{
			ID: id,
		}

		w.Header().Set("Content-Encoding", "br")

		rW := brotli.NewWriter(w)
		defer rW.Close()

		err := json.NewEncoder(rW).Encode(m)
		assert.NoError(t, err)
	}))

	client := NewClient(Config{APIURL: s.URL})

	m, err := client.GetMeasurement(id)
	assert.NoError(t, err)

	assert.Equal(t, id, m.ID)
}

func Test_GetMeasurementRaw_Json(t *testing.T) {
	server := generateServer(`{"id":"abcd"}`, http.StatusOK)
	defer server.Close()
	client := NewClient(Config{APIURL: server.URL})
	res, err := client.GetMeasurementRaw("abcd")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, `{"id":"abcd"}`, string(res))
}

func generateServer(json string, statusCode int) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		_, err := w.Write([]byte(json))
		if err != nil {
			panic(err)
		}
	}))
	return server
}

func generateServerAuthorized(json string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": {"type": "unauthorized", "message": "Unauthorized."}}`))
			return
		}
		w.WriteHeader(http.StatusAccepted)
		_, err := w.Write([]byte(json))
		if err != nil {
			panic(err)
		}
	}))
	return server
}
