package globalping

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Authorize(t *testing.T) {
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
	client := NewClient(Config{
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthURL:          server.URL,
		DashboardURL:     server.URL,
		OnTokenRefresh: func(_token *Token) {
			assert.Equal(t, &Token{
				AccessToken:  "token",
				TokenType:    "bearer",
				RefreshToken: "refresh",
				ExpiresIn:    3600,
				Expiry:       _token.Expiry,
			}, _token)
		},
	})
	res, err := client.Authorize(func(err error) {
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
	now := time.Now()
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

	onTokenRefreshCalled := false
	client := NewClient(Config{
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthURL:          server.URL,
		DashboardURL:     server.URL,
		AuthToken: &Token{
			AccessToken: "tok3n",
			Expiry:      now.Add(time.Hour),
		},
		OnTokenRefresh: func(_ *Token) {
			onTokenRefreshCalled = true
		},
	})
	res, err := client.TokenIntrospection("")
	assert.Nil(t, err)
	assert.Equal(t, introspectionRes, res)

	assert.False(t, onTokenRefreshCalled)
}

func Test_TokenIntrospection_Token_Refreshed(t *testing.T) {
	now := time.Now()
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

	var token *Token
	client := NewClient(Config{
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthURL:          server.URL,
		DashboardURL:     server.URL,
		AuthToken: &Token{
			AccessToken:  "tok3n",
			RefreshToken: "refresh_tok3n",
			Expiry:       now.Add(-time.Hour),
		},
		OnTokenRefresh: func(_t *Token) {
			token = _t
		},
	})
	res, err := client.TokenIntrospection("")
	assert.Nil(t, err)
	assert.Equal(t, introspectionRes, res)

	assert.Equal(t, &Token{
		AccessToken:  "new_token",
		TokenType:    "bearer",
		RefreshToken: "new_refresh_token",
		ExpiresIn:    3600,
		Expiry:       token.Expiry,
	}, token)
}

func Test_TokenIntrospection_With_Token(t *testing.T) {
	now := time.Now()
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

	onTokenRefreshCalled := false
	client := NewClient(Config{
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthURL:          server.URL,
		DashboardURL:     server.URL,
		AuthToken: &Token{
			AccessToken: "local_token",
			Expiry:      now.Add(time.Hour),
		},
		OnTokenRefresh: func(_ *Token) {
			onTokenRefreshCalled = true
		},
	})
	res, err := client.TokenIntrospection("tok3n")
	assert.Nil(t, err)
	assert.Equal(t, introspectionRes, res)

	assert.False(t, onTokenRefreshCalled)
}

func Test_TokenIntrospection_No_Token(t *testing.T) {
	client := NewClient(Config{})
	res, err := client.TokenIntrospection("")
	assert.Nil(t, res)
	e, ok := err.(*AuthorizeError)
	assert.True(t, ok)
	assert.Equal(t, ErrTypeNotAuthorized, e.ErrorType)
	assert.Equal(t, "client is not authorized", e.Description)
}

func Test_Logout(t *testing.T) {
	isCalled := false
	now := time.Now()
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

	onTokenRefreshCalled := false
	client := NewClient(Config{
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthURL:          server.URL,
		DashboardURL:     server.URL,
		AuthToken: &Token{
			AccessToken:  "tok3n",
			RefreshToken: "refresh_tok3n",
			Expiry:       now.Add(time.Hour),
		},
		OnTokenRefresh: func(token *Token) {
			onTokenRefreshCalled = true
			assert.Nil(t, token)
		},
	})
	err := client.Logout()
	assert.Nil(t, err)
	assert.True(t, isCalled)
	assert.True(t, onTokenRefreshCalled)
}

func Test_RevokeToken(t *testing.T) {
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

	client := NewClient(Config{
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthURL:          server.URL,
		DashboardURL:     server.URL,
	})
	err := client.RevokeToken("refresh_tok3n")
	assert.Nil(t, err)
	assert.True(t, isCalled)
}

func Test_Logout_No_RefreshToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()

	onTokenRefreshCalled := false
	client := NewClient(Config{
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthURL:          server.URL,
		DashboardURL:     server.URL,
		AuthToken: &Token{
			AccessToken: "tok3n",
		},
		OnTokenRefresh: func(token *Token) {
			onTokenRefreshCalled = true
			assert.Nil(t, token)
		},
	})
	err := client.Logout()
	assert.Nil(t, err)
	assert.True(t, onTokenRefreshCalled)
}

func Test_Logout_AccessToken_Is_Set(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()

	onTokenRefreshCalled := false
	client := NewClient(Config{
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthURL:          server.URL,
		DashboardURL:     server.URL,
		AuthToken: &Token{
			AccessToken: "tok3n",
			Expiry:      time.Now().Add(time.Hour),
		},
		OnTokenRefresh: func(token *Token) {
			assert.Nil(t, token)
		},
	})
	err := client.Logout()
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, onTokenRefreshCalled)
}
