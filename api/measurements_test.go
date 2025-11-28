package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	globalpingMock "github.com/jsdelivr/globalping-cli/mocks/globalping"
	utilsMock "github.com/jsdelivr/globalping-cli/mocks/utils"
	"github.com/jsdelivr/globalping-cli/storage"
	"github.com/jsdelivr/globalping-go"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_CreateMeasurement_TokenRefreshed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	utilsMock := utilsMock.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	opts := &globalping.MeasurementCreate{}
	expectedMeasurement := &globalping.MeasurementCreateResponse{ID: "abcd"}

	globalpingMock := globalpingMock.NewMockClient(ctrl)
	globalpingMock.EXPECT().CreateMeasurement(t.Context(), opts).Return(expectedMeasurement, nil).Times(1)
	globalpingMock.EXPECT().SetToken("new_token").Times(1)

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
	}))
	defer server.Close()

	_storage := createDefaultTestStorage(t, utilsMock)
	_storage.GetProfile().Token = &storage.Token{
		AccessToken:  "token",
		RefreshToken: "refresh_tok3n",
		Expiry:       defaultCurrentTime.Add(-1 * time.Hour),
	}

	client := NewClient(Config{
		Utils:            utilsMock,
		Storage:          _storage,
		Globalping:       globalpingMock,
		AuthURL:          server.URL,
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>"})

	res, err := client.CreateMeasurement(t.Context(), opts)
	assert.Nil(t, err)
	assert.Equal(t, expectedMeasurement, res)

	assert.Equal(t, &storage.Token{
		AccessToken:  "new_token",
		TokenType:    "Bearer",
		RefreshToken: "new_refresh_token",
		ExpiresIn:    3600,
		Expiry:       defaultCurrentTime.Add(3600 * time.Second),
	}, _storage.GetProfile().Token)
}

func Test_CreateMeasurement_Unauthorized_TokenRefreshed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	utilsMock := utilsMock.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	opts := &globalping.MeasurementCreate{}

	globalpingMock := globalpingMock.NewMockClient(ctrl)
	globalpingMock.EXPECT().CreateMeasurement(t.Context(), opts).Return(nil, &globalping.MeasurementError{
		StatusCode: http.StatusUnauthorized,
		Type:       "unauthorized",
		Message:    "Unauthorized.",
	}).Times(1)
	globalpingMock.EXPECT().SetToken("access_token")

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
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()

	_storage := createDefaultTestStorage(t, utilsMock)

	client := NewClient(Config{
		Utils:            utilsMock,
		Storage:          _storage,
		Globalping:       globalpingMock,
		AuthURL:          server.URL,
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthToken: &storage.Token{
			AccessToken:  "access_token",
			RefreshToken: "refresh_tok3n",
			Expiry:       time.Now().Add(1 * time.Hour),
		},
	})

	res, err := client.CreateMeasurement(t.Context(), opts)
	assert.Nil(t, res)
	e, ok := err.(*globalping.MeasurementError)
	assert.True(t, ok)
	assert.Equal(t, StatusUnauthorizedWithTokenRefreshed, e.StatusCode)
	assert.Equal(t, "Unauthorized.", e.Message)

	assert.Equal(t, &storage.Token{
		AccessToken:  "new_token",
		TokenType:    "Bearer",
		RefreshToken: "new_refresh_token",
		ExpiresIn:    3600,
		Expiry:       defaultCurrentTime.Add(3600 * time.Second),
	}, _storage.GetProfile().Token)
}

func Test_CreateMeasurement_Unauthorized_Token_Not_Refreshed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	utilsMock := utilsMock.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	opts := &globalping.MeasurementCreate{}

	globalpingMock := globalpingMock.NewMockClient(ctrl)
	globalpingMock.EXPECT().CreateMeasurement(t.Context(), opts).Return(nil, &globalping.MeasurementError{
		StatusCode: http.StatusUnauthorized,
		Type:       "unauthorized",
		Message:    "Unauthorized.",
	}).Times(1)
	globalpingMock.EXPECT().SetToken("access_token")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "invalid_grant", "error_description": "Invalid refresh token."}`))
			return
		}
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()

	_storage := createDefaultTestStorage(t, utilsMock)

	client := NewClient(Config{
		Utils:            utilsMock,
		Storage:          _storage,
		Globalping:       globalpingMock,
		AuthURL:          server.URL,
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthToken: &storage.Token{
			AccessToken:  "access_token",
			RefreshToken: "refresh_tok3n",
			Expiry:       time.Now().Add(1 * time.Hour),
		},
	})

	res, err := client.CreateMeasurement(t.Context(), opts)
	assert.Nil(t, res)
	assert.EqualError(t, err, "unauthorized: You have been signed out by the API. Please try signing in again.")

	assert.Nil(t, _storage.GetProfile().Token)
}

func Test_CreateMeasurement_Unauthorized_NoRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	utilsMock := utilsMock.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	opts := &globalping.MeasurementCreate{}

	globalpingMock := globalpingMock.NewMockClient(ctrl)
	globalpingMock.EXPECT().CreateMeasurement(t.Context(), opts).Return(nil, &globalping.MeasurementError{
		StatusCode: http.StatusUnauthorized,
		Type:       "unauthorized",
		Message:    "Unauthorized.",
	}).Times(1)
	globalpingMock.EXPECT().SetToken("access_token")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": {"type": "unauthorized", "message": "Unauthorized."}}`))
	}))
	defer server.Close()

	_storage := createDefaultTestStorage(t, utilsMock)

	client := NewClient(Config{
		Utils:            utilsMock,
		Storage:          _storage,
		Globalping:       globalpingMock,
		AuthURL:          server.URL,
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthToken: &storage.Token{
			AccessToken: "access_token",
			Expiry:      time.Now().Add(1 * time.Hour),
		},
	})

	res, err := client.CreateMeasurement(t.Context(), opts)
	assert.Nil(t, res)
	assert.EqualError(t, err, "unauthorized: "+invalidTokenErr)

	assert.Nil(t, _storage.GetProfile().Token)
}

func Test_CreateMeasurement_MoreCreditsRequiredNoAuthError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	utilsMock := utilsMock.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	opts := &globalping.MeasurementCreate{}

	globalpingMock := globalpingMock.NewMockClient(ctrl)
	header := http.Header{}
	header.Set("X-RateLimit-Remaining", "1")
	header.Set("X-RateLimit-Reset", "61")
	header.Set("X-Credits-Remaining", "1")
	header.Set("X-Request-Cost", "2")
	globalpingMock.EXPECT().CreateMeasurement(t.Context(), opts).Return(nil, &globalping.MeasurementError{
		StatusCode: http.StatusTooManyRequests,
		Header:     header,
		Type:       "rate_limit_exceeded",
		Message:    "API rate limit exceeded.",
	})
	globalpingMock.EXPECT().SetToken("").Times(1)

	_storage := createDefaultTestStorage(t, utilsMock)

	client := NewClient(Config{
		Utils:      utilsMock,
		Storage:    _storage,
		Globalping: globalpingMock,
	})
	_, err := client.CreateMeasurement(t.Context(), opts)
	assert.EqualError(t, err, "rate_limit_exceeded: "+fmt.Sprintf(moreCreditsRequiredNoAuthErr, "2 credits", 2, "1 minute"))

	assert.Nil(t, _storage.GetProfile().Token)
}

func Test_CreateMeasurement_MoreCreditsRequiredAuthError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	utilsMock := utilsMock.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	opts := &globalping.MeasurementCreate{}

	globalpingMock := globalpingMock.NewMockClient(ctrl)
	header := http.Header{}
	header.Set("X-RateLimit-Remaining", "0")
	header.Set("X-RateLimit-Reset", "40")
	header.Set("X-Credits-Remaining", "1")
	header.Set("X-Request-Cost", "2")
	globalpingMock.EXPECT().CreateMeasurement(t.Context(), opts).Return(nil, &globalping.MeasurementError{
		StatusCode: http.StatusTooManyRequests,
		Header:     header,
		Type:       "rate_limit_exceeded",
		Message:    "API rate limit exceeded.",
	})
	globalpingMock.EXPECT().SetToken("secret").Times(1)

	_storage := createDefaultTestStorage(t, utilsMock)

	client := NewClient(Config{
		Utils:      utilsMock,
		Storage:    _storage,
		Globalping: globalpingMock,
		AuthToken: &storage.Token{
			AccessToken: "secret",
			Expiry:      time.Now().Add(1 * time.Hour),
		},
	})

	_, err := client.CreateMeasurement(t.Context(), opts)
	assert.EqualError(t, err, "rate_limit_exceeded: "+fmt.Sprintf(moreCreditsRequiredAuthErr, "1 credit", 2, "40 seconds"))
}

func Test_CreateMeasurement_NoCreditsNoAuthError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	utilsMock := utilsMock.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	opts := &globalping.MeasurementCreate{}

	globalpingMock := globalpingMock.NewMockClient(ctrl)
	header := http.Header{}
	header.Set("X-RateLimit-Remaining", "0")
	header.Set("X-RateLimit-Reset", "5")
	header.Set("X-Credits-Remaining", "0")
	globalpingMock.EXPECT().CreateMeasurement(t.Context(), opts).Return(nil, &globalping.MeasurementError{
		StatusCode: http.StatusTooManyRequests,
		Header:     header,
		Type:       "rate_limit_exceeded",
		Message:    "API rate limit exceeded.",
	})
	globalpingMock.EXPECT().SetToken("").Times(1)

	_storage := createDefaultTestStorage(t, utilsMock)

	client := NewClient(Config{
		Utils:      utilsMock,
		Storage:    _storage,
		Globalping: globalpingMock,
	})

	_, err := client.CreateMeasurement(t.Context(), opts)

	assert.EqualError(t, err, "rate_limit_exceeded: "+fmt.Sprintf(noCreditsNoAuthErr, "5 seconds"))
}

func Test_CreateMeasurement_NoCreditsAuthError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	utilsMock := utilsMock.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	opts := &globalping.MeasurementCreate{}

	globalpingMock := globalpingMock.NewMockClient(ctrl)
	header := http.Header{}
	header.Set("X-RateLimit-Remaining", "0")
	header.Set("X-RateLimit-Reset", "5")
	header.Set("X-Credits-Remaining", "0")
	globalpingMock.EXPECT().CreateMeasurement(t.Context(), opts).Return(nil, &globalping.MeasurementError{
		StatusCode: http.StatusTooManyRequests,
		Header:     header,
		Type:       "rate_limit_exceeded",
		Message:    "API rate limit exceeded.",
	})
	globalpingMock.EXPECT().SetToken("secret").Times(1)

	_storage := createDefaultTestStorage(t, utilsMock)

	client := NewClient(Config{
		Utils:      utilsMock,
		Storage:    _storage,
		Globalping: globalpingMock,
		AuthToken: &storage.Token{
			AccessToken: "secret",
			Expiry:      time.Now().Add(1 * time.Hour),
		},
	})

	_, err := client.CreateMeasurement(t.Context(), opts)
	assert.EqualError(t, err, "rate_limit_exceeded: "+fmt.Sprintf(noCreditsAuthErr, "5 seconds"))
}
