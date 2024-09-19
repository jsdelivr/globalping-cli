package globalping

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type Token struct {
	// AccessToken is the token that authorizes and authenticates
	// the requests.
	AccessToken string `json:"access_token"`

	// TokenType is the type of token.
	// The Type method returns either this or "Bearer", the default.
	TokenType string `json:"token_type,omitempty"`

	// RefreshToken is a token that's used by the application
	// (as opposed to the user) to refresh the access token
	// if it expires.
	RefreshToken string `json:"refresh_token,omitempty"`

	// Expiry is the optional expiration time of the access token.
	//
	// If zero, TokenSource implementations will reuse the same
	// token forever and RefreshToken or equivalent
	// mechanisms for that TokenSource will not be used.
	Expiry time.Time `json:"expiry,omitempty"`
}

type AuthorizeError struct {
	Code        int
	ErrorCode   string
	Description string
}

func (e *AuthorizeError) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode, e.Description)
}

func (c *client) Authorize(callback func(error)) string {
	pkce := oauth2.GenerateVerifier()
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
	}
	mux.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		req.ParseForm()
		token, err := c.exchange(req.Form, pkce, "http://"+server.Addr+"/callback")
		if err != nil {
			http.Redirect(w, req, c.dashboardURL+"/authorize/error", http.StatusFound)
		} else {
			http.Redirect(w, req, c.dashboardURL+"/authorize/success", http.StatusFound)
		}
		go func() {
			server.Shutdown(req.Context())
			if err == nil {
				c.token.Store(token)
				if c.onTokenRefresh != nil {
					c.onTokenRefresh(mapToken(token))
				}
			}
			callback(err)
		}()
	})
	go func() {
		ports := []int{60000, 60010, 60020, 60030, 60040, 60100, 60110, 60120, 60130, 60140}
		var err error
		for _, port := range ports {
			server.Addr = fmt.Sprintf("localhost:%d", port)
			err = server.ListenAndServe()
			if err == nil {
				break
			}
		}
		if err != nil {
			if err == http.ErrServerClosed {
				return
			}
			callback(&AuthorizeError{ErrorCode: "failed to start server", Description: err.Error()})
		}
	}()
	return c.oauth2.AuthCodeURL("", oauth2.S256ChallengeOption(pkce))
}

func (c *client) TokenIntrospection(token string) (*IntrospectionResponse, error) {
	if token == "" {
		var err error
		token, _, err = c.accessToken()
		if err != nil {
			return nil, &AuthorizeError{
				ErrorCode:   "not_authorized",
				Description: "client is not authorized: " + err.Error(),
			}
		}
	}
	if token == "" {
		return nil, &AuthorizeError{
			ErrorCode:   "not_authorized",
			Description: "client is not authorized",
		}
	}
	return c.introspection(token)
}

func (c *client) Logout() error {
	t := c.token.Load()
	if t == nil {
		return nil
	}
	if t.RefreshToken != "" {
		err := c.revoke(t.RefreshToken)
		if err != nil {
			return err
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tokenSource = nil
	c.token.Store(nil)
	if c.onTokenRefresh != nil {
		c.onTokenRefresh(nil)
	}
	return nil
}

func (c *client) exchange(form url.Values, pkce string, redirect string) (*oauth2.Token, error) {
	if form.Get("error") != "" {
		return nil, &AuthorizeError{
			ErrorCode:   form.Get("error"),
			Description: form.Get("error_description"),
		}
	}
	code := form.Get("code")
	if code == "" {
		return nil, &AuthorizeError{
			ErrorCode:   "missing_code",
			Description: "missing code in response",
		}
	}
	return c.oauth2.Exchange(
		context.Background(),
		code,
		oauth2.VerifierOption(pkce),
		oauth2.SetAuthURLParam("redirect_uri", redirect),
	)
}

func (c *client) accessToken() (string, string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.tokenSource == nil {
		return "", "", nil
	}
	token, err := c.tokenSource.Token()
	if err != nil {
		e, ok := err.(*oauth2.RetrieveError)
		if ok && e.ErrorCode == "invalid_grant" && c.onTokenRefresh != nil {
			c.onTokenRefresh(nil)
		}
		return "", "", err
	}
	curr := c.token.Load()
	if curr != nil && token.AccessToken != curr.AccessToken {
		c.token.Store(token)
		if c.onTokenRefresh != nil {
			c.onTokenRefresh(mapToken(token))
		}
	}
	return token.AccessToken, token.Type(), nil
}

// https://datatracker.ietf.org/doc/html/rfc7662#section-2.1
type IntrospectionResponse struct {
	// Required fields
	Active bool `json:"active"`

	// Optional fields
	Scope     string `json:"scope"`
	ClientID  string `json:"client_id"`
	Username  string `json:"username"`
	TokenType string `json:"token_type"`
	Exp       int64  `json:"exp"` // Expiration Time. Unix timestamp
	Iat       int64  `json:"iat"` // Issued At. Unix timestamp
	Nbf       int64  `json:"nbf"` // Not to be used before. Unix timestamp
	Sub       string `json:"sub"` // Subject
	Aud       string `json:"aud"` // Audience
	Iss       string `json:"iss"` // Issuer
	Jti       string `json:"jti"` // JWT ID
}

func (c *client) introspection(token string) (*IntrospectionResponse, error) {
	form := url.Values{"token": {token}}.Encode()
	req, err := http.NewRequest("POST", c.authURL+"/oauth/token/introspect", strings.NewReader(form))
	if err != nil {
		return nil, &AuthorizeError{
			ErrorCode:   "introspection_failed",
			Description: err.Error(),
		}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(form)))
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, &AuthorizeError{
			ErrorCode:   "introspection_failed",
			Description: err.Error(),
		}
	}
	if resp.StatusCode != http.StatusOK {
		err := &AuthorizeError{
			Code:        resp.StatusCode,
			ErrorCode:   "introspection_failed",
			Description: resp.Status,
		}
		json.NewDecoder(resp.Body).Decode(err)
		return nil, err
	}
	ires := &IntrospectionResponse{}
	err = json.NewDecoder(resp.Body).Decode(ires)
	return ires, err
}

func (c *client) revoke(refreshToken string) error {
	form := url.Values{"token": {refreshToken}}.Encode()
	req, err := http.NewRequest("POST", c.authURL+"/oauth/token/revoke", strings.NewReader(form))
	if err != nil {
		return &AuthorizeError{
			ErrorCode:   "revoke_failed",
			Description: err.Error(),
		}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(form)))
	resp, err := c.http.Do(req)
	if err != nil {
		return &AuthorizeError{
			ErrorCode:   "revoke_failed",
			Description: err.Error(),
		}
	}
	if resp.StatusCode != http.StatusOK {
		err := &AuthorizeError{
			Code:        resp.StatusCode,
			ErrorCode:   "revoke_failed",
			Description: resp.Status,
		}
		json.NewDecoder(resp.Body).Decode(err)
		return err
	}
	return nil
}

func mapToken(t *oauth2.Token) *Token {
	return &Token{
		AccessToken:  t.AccessToken,
		TokenType:    t.TokenType,
		RefreshToken: t.RefreshToken,
		Expiry:       t.Expiry,
	}
}