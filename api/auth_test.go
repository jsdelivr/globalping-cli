package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	utilsMock "github.com/jsdelivr/globalping-cli/mocks/utils"
	"github.com/jsdelivr/globalping-cli/storage"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_Authorize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	utilsMock := utilsMock.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	succesCalled := false
	expectedRedirectURI := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/authorize/error" {
			t.Fatalf("unexpected request to %s", r.URL.Path)
			return
		}
		if r.URL.Path == "/authorize/success" {
			succesCalled = true
			return
		}
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
			assert.Equal(t, "authorization_code", r.Form.Get("grant_type"))
			assert.Equal(t, "cod3", r.Form.Get("code"))
			assert.Equal(t, expectedRedirectURI, r.Form.Get("redirect_uri"))
			assert.Equal(t, 43, len(r.Form.Get("code_verifier")))

			w.Header().Set("Content-Type", "application/json")
			_, err = w.Write(getTokenJSON())
			assert.Nil(t, err)
			return
		}
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()

	_storage := createDefaultTestStorage(t, utilsMock)

	client := NewClient(Config{
		Utils:            utilsMock,
		Storage:          _storage,
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthURL:          server.URL,
		DashboardURL:     server.URL,
	})
	res, err := client.Authorize(t.Context(), func(err error) {
		assert.Nil(t, err)
	})
	assert.Nil(t, err)
	expectedRedirectURI = res.CallbackURL
	u, err := url.Parse(res.AuthorizeURL)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, server.URL+"/oauth/authorize", u.Scheme+"://"+u.Host+u.Path)
	assert.Equal(t, "<client_id>", u.Query().Get("client_id"))
	assert.Equal(t, 43, len(u.Query().Get("code_challenge")))
	assert.Equal(t, "S256", u.Query().Get("code_challenge_method"))
	assert.Equal(t, "code", u.Query().Get("response_type"))
	assert.Equal(t, "measurements", u.Query().Get("scope"))
	assert.Equal(t, expectedRedirectURI, u.Query().Get("redirect_uri"))

	_, err = http.Post(res.CallbackURL+"?code=cod3", "application/x-www-form-urlencoded", nil)
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, succesCalled, "/authorize/success not called")
}

func Test_TokenIntrospection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	utilsMock := utilsMock.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	introspectionRes := &IntrospectionResponse{
		Active:    true,
		Scope:     "measurements",
		ClientID:  "<client_id>",
		Username:  "user",
		TokenType: "bearer",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token/introspect" {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST request, got %s", r.Method)
			}
			err := r.ParseForm()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "tok3n", r.Form.Get("token"))

			w.Header().Set("Content-Type", "application/json")
			b, _ := json.Marshal(introspectionRes)
			_, err = w.Write(b)
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
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthURL:          server.URL,
		DashboardURL:     server.URL,
		AuthToken: &storage.Token{
			AccessToken: "tok3n",
			Expiry:      defaultCurrentTime.Add(time.Hour),
		},
	})
	res, err := client.TokenIntrospection(t.Context(), "")
	assert.Nil(t, err)
	assert.Equal(t, introspectionRes, res)
}

func Test_TokenIntrospection_Token_Refreshed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	utilsMock := utilsMock.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	introspectionRes := &IntrospectionResponse{
		Active:    true,
		Scope:     "measurements",
		ClientID:  "<client_id>",
		Username:  "user",
		TokenType: "bearer",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token/introspect" {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST request, got %s", r.Method)
			}
			err := r.ParseForm()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "new_token", r.Form.Get("token"))

			w.Header().Set("Content-Type", "application/json")
			b, _ := json.Marshal(introspectionRes)
			_, err = w.Write(b)
			if err != nil {
				t.Fatal(err)
			}
			return
		}
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
			_, err = w.Write([]byte(`{"access_token":"new_token","token_type":"bearer","refresh_token":"new_refresh_token","expires_in":3600}`))
			if err != nil {
				t.Fatal(err)
			}
			return
		}
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()

	_storage := createDefaultTestStorage(t, utilsMock)
	_storage.GetProfile().Token = &storage.Token{
		AccessToken:  "token",
		RefreshToken: "refresh",
	}

	client := NewClient(Config{
		Utils:            utilsMock,
		Storage:          _storage,
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthURL:          server.URL,
		DashboardURL:     server.URL,
		AuthToken: &storage.Token{
			AccessToken:  "tok3n",
			RefreshToken: "refresh_tok3n",
			Expiry:       defaultCurrentTime.Add(-time.Hour),
		},
	})
	res, err := client.TokenIntrospection(t.Context(), "")
	assert.Nil(t, err)
	assert.Equal(t, introspectionRes, res)

	assert.Equal(t, &storage.Token{
		AccessToken:  "new_token",
		TokenType:    "bearer",
		RefreshToken: "new_refresh_token",
		ExpiresIn:    3600,
		Expiry:       defaultCurrentTime.Add(3600 * time.Second),
	}, _storage.GetProfile().Token)
}

func Test_TokenIntrospection_With_Token(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	utilsMock := utilsMock.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	introspectionRes := &IntrospectionResponse{
		Active:    true,
		Scope:     "measurements",
		ClientID:  "<client_id>",
		Username:  "user",
		TokenType: "bearer",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token/introspect" {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST request, got %s", r.Method)
			}
			err := r.ParseForm()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "tok3n", r.Form.Get("token"))

			w.Header().Set("Content-Type", "application/json")
			b, _ := json.Marshal(introspectionRes)
			_, err = w.Write(b)
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
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthURL:          server.URL,
		DashboardURL:     server.URL,
		AuthToken: &storage.Token{
			AccessToken: "local_token",
			Expiry:      defaultCurrentTime.Add(time.Hour),
		},
	})
	res, err := client.TokenIntrospection(t.Context(), "tok3n")
	assert.Nil(t, err)
	assert.Equal(t, introspectionRes, res)
}

func Test_TokenIntrospection_No_Token(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	utilsMock := utilsMock.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	_storage := createDefaultTestStorage(t, utilsMock)
	client := NewClient(Config{
		Utils:   utilsMock,
		Storage: _storage,
	})

	res, err := client.TokenIntrospection(t.Context(), "")
	assert.Nil(t, res)

	e, ok := err.(*AuthorizeError)
	assert.True(t, ok)
	assert.Equal(t, ErrTypeNotAuthorized, e.ErrorType)
	assert.Equal(t, "client is not authorized", e.Description)
}

func Test_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	utilsMock := utilsMock.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	isCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isCalled = true
		if r.URL.Path == "/oauth/token/revoke" {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST request, got %s", r.Method)
			}
			err := r.ParseForm()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "refresh_tok3n", r.Form.Get("token"))
			return
		}
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()

	_storage := createDefaultTestStorage(t, utilsMock)

	client := NewClient(Config{
		Utils:            utilsMock,
		Storage:          _storage,
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthURL:          server.URL,
		DashboardURL:     server.URL,
		AuthToken: &storage.Token{
			AccessToken:  "tok3n",
			RefreshToken: "refresh_tok3n",
			Expiry:       defaultCurrentTime.Add(time.Hour),
		},
	})
	err := client.Logout(t.Context())
	assert.Nil(t, err)
	assert.True(t, isCalled)
}

func Test_RevokeToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	utilsMock := utilsMock.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	isCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isCalled = true
		if r.URL.Path == "/oauth/token/revoke" {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST request, got %s", r.Method)
			}
			err := r.ParseForm()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "refresh_tok3n", r.Form.Get("token"))
			return
		}
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()

	_storage := createDefaultTestStorage(t, utilsMock)

	client := NewClient(Config{
		Utils:            utilsMock,
		Storage:          _storage,
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthURL:          server.URL,
		DashboardURL:     server.URL,
	})
	err := client.RevokeToken(t.Context(), "refresh_tok3n")
	assert.Nil(t, err)
	assert.True(t, isCalled)
}

func Test_Logout_No_RefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	utilsMock := utilsMock.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()

	_storage := createDefaultTestStorage(t, utilsMock)

	client := NewClient(Config{
		Utils:            utilsMock,
		Storage:          _storage,
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthURL:          server.URL,
		DashboardURL:     server.URL,
		AuthToken: &storage.Token{
			AccessToken: "tok3n",
		},
	})
	err := client.Logout(t.Context())
	assert.Nil(t, err)
}

func Test_Logout_AccessToken_Is_Set(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	utilsMock := utilsMock.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()

	_storage := createDefaultTestStorage(t, utilsMock)

	client := NewClient(Config{
		Utils:            utilsMock,
		Storage:          _storage,
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthURL:          server.URL,
		DashboardURL:     server.URL,
		AuthToken: &storage.Token{
			AccessToken: "tok3n",
			Expiry:      defaultCurrentTime.Add(time.Hour),
		},
	})
	err := client.Logout(t.Context())
	if err != nil {
		t.Fatal(err)
	}
}
