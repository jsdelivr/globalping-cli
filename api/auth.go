package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jsdelivr/globalping-cli/storage"
)

var (
	ErrTypeExchangeFailed      = "exchange_failed"
	ErrTypeRefreshFailed       = "refresh_failed"
	ErrTypeRevokeFailed        = "revoke_failed"
	ErrTypeIntrospectionFailed = "introspection_failed"
	ErrTypeInvalidGrant        = "invalid_grant"
	ErrTypeNotAuthorized       = "not_authorized"
)

type AuthorizeError struct {
	Code        int    `json:"-"`
	ErrorType   string `json:"error"`
	Description string `json:"error_description"`
}

func (e *AuthorizeError) Error() string {
	return e.ErrorType + ": " + e.Description
}

type AuthorizeResponse struct {
	AuthorizeURL string
	CallbackURL  string
}

func (c *client) Authorize(ctx context.Context, callback func(error)) (*AuthorizeResponse, error) {
	verifier := generateVerifier()
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
	}
	callbackURL := ""
	mux.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		req.ParseForm()
		token, err := c.exchange(ctx, req.Form, verifier, callbackURL)
		if err != nil {
			http.Redirect(w, req, c.dashboardURL+"/authorize/error", http.StatusFound)
		} else {
			http.Redirect(w, req, c.dashboardURL+"/authorize/success", http.StatusFound)
		}
		go func() {
			server.Shutdown(context.Background())
			if err == nil {
				c.updateToken(token)
			}
			callback(err)
		}()
	})
	var err error
	var ln net.Listener
	ports := []int{60000, 60010, 60020, 60030, 60040, 60100, 60110, 60120, 60130, 60140}
	port := ""
	for i := range ports {
		port = strconv.Itoa(ports[i])
		ln, err = net.Listen("tcp", ":"+port)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	go func() {
		err := server.Serve(ln)
		if err != nil && err != http.ErrServerClosed {
			callback(&AuthorizeError{ErrorType: "failed to start server", Description: err.Error()})
		}
	}()
	callbackURL = "http://localhost:" + port + "/callback"
	q := url.Values{}
	q.Set("client_id", c.authClientId)
	q.Set("code_challenge", generateS256Challenge(verifier))
	q.Set("code_challenge_method", "S256")
	q.Set("response_type", "code")
	q.Set("scope", "measurements")
	q.Set("redirect_uri", callbackURL)

	return &AuthorizeResponse{
		AuthorizeURL: c.authURL + "/oauth/authorize?" + q.Encode(),
		CallbackURL:  callbackURL,
	}, nil
}

func (c *client) TokenIntrospection(ctx context.Context, token string) (*IntrospectionResponse, error) {
	if token == "" {
		t, err := c.getToken(ctx)
		if err != nil {
			return nil, &AuthorizeError{
				ErrorType:   ErrTypeNotAuthorized,
				Description: err.Error(),
			}
		}
		if t != nil {
			token = t.AccessToken
		}
	}
	if token == "" {
		return nil, &AuthorizeError{
			ErrorType:   ErrTypeNotAuthorized,
			Description: "client is not authorized",
		}
	}
	return c.introspection(ctx, token)
}

func (c *client) Logout(ctx context.Context) error {
	c.mu.RLock()
	t := c.token
	c.mu.RUnlock()
	if t == nil {
		return nil
	}
	err := c.RevokeToken(ctx, t.RefreshToken)
	if err != nil {
		return err
	}
	c.updateToken(nil)
	return nil
}

func (c *client) exchange(ctx context.Context, form url.Values, verifier string, redirect string) (*storage.Token, error) {
	if form.Get("error") != "" {
		return nil, &AuthorizeError{
			ErrorType:   form.Get("error"),
			Description: form.Get("error_description"),
		}
	}
	code := form.Get("code")
	if code == "" {
		return nil, &AuthorizeError{
			ErrorType:   "missing_code",
			Description: "missing code in response",
		}
	}
	q := url.Values{}
	q.Set("client_id", c.authClientId)
	q.Set("client_secret", c.authClientSecret)
	q.Set("code", code)
	q.Set("code_verifier", verifier)
	q.Set("grant_type", "authorization_code")
	q.Set("redirect_uri", redirect)
	req, err := http.NewRequestWithContext(ctx, "POST", c.authURL+"/oauth/token", strings.NewReader(q.Encode()))
	if err != nil {
		return nil, &AuthorizeError{
			ErrorType:   ErrTypeExchangeFailed,
			Description: err.Error(),
		}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(q.Encode())))
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, &AuthorizeError{
			ErrorType:   ErrTypeExchangeFailed,
			Description: err.Error(),
		}
	}
	if resp.StatusCode != http.StatusOK {
		err := &AuthorizeError{
			Code:        resp.StatusCode,
			ErrorType:   ErrTypeExchangeFailed,
			Description: resp.Status,
		}
		json.NewDecoder(resp.Body).Decode(err)
		return nil, err
	}
	t := &storage.Token{}
	err = json.NewDecoder(resp.Body).Decode(t)
	if err != nil {
		return nil, &AuthorizeError{
			ErrorType:   ErrTypeExchangeFailed,
			Description: err.Error(),
		}
	}
	if t.TokenType == "" {
		t.TokenType = "Bearer"
	}
	if t.ExpiresIn != 0 {
		t.Expiry = c.utils.Now().Add(time.Duration(t.ExpiresIn) * time.Second)
	}
	return t, nil
}

func (c *client) getToken(ctx context.Context) (*storage.Token, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token == nil {
		return nil, nil
	}
	if !c.token.Expiry.Before(c.utils.Now()) {
		return c.token, nil
	}
	if c.token.RefreshToken == "" {
		return nil, &AuthorizeError{
			ErrorType:   "refresh_failed",
			Description: "empty refresh token",
		}
	}
	t, err := c.refreshToken(ctx, c.token.RefreshToken)
	if err != nil {
		e, ok := err.(*AuthorizeError)
		if ok && e.ErrorType == ErrTypeInvalidGrant {
			c.saveToken(nil)
		}
		return nil, err
	}

	c.token = t
	c.saveToken(&storage.Token{
		AccessToken:  t.AccessToken,
		TokenType:    t.TokenType,
		RefreshToken: t.RefreshToken,
		ExpiresIn:    t.ExpiresIn,
		Expiry:       t.Expiry,
	})

	return t, nil
}

func (c *client) updateToken(t *storage.Token) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.token = t
	if t == nil {
		c.saveToken(nil)
		return
	}

	c.saveToken(&storage.Token{
		AccessToken:  t.AccessToken,
		TokenType:    t.TokenType,
		RefreshToken: t.RefreshToken,
		ExpiresIn:    t.ExpiresIn,
		Expiry:       t.Expiry,
	})
}

func (c *client) tryToRefreshToken(ctx context.Context, refreshToken string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token == nil {
		return false
	}
	// must have been called by a different goroutine
	if c.token.RefreshToken != refreshToken {
		return false
	}

	token, err := c.refreshToken(ctx, c.token.RefreshToken)
	if err != nil {
		e, ok := err.(*AuthorizeError)
		// If the refresh token is invalid, clear the token
		if ok && e.ErrorType == ErrTypeInvalidGrant {
			c.token = nil
			c.saveToken(nil)
		}
		return false
	}

	c.token = token
	c.saveToken(&storage.Token{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.ExpiresIn,
		Expiry:       token.Expiry,
	})

	return true
}

func (c *client) refreshToken(ctx context.Context, token string) (*storage.Token, error) {
	q := url.Values{}
	q.Set("client_id", c.authClientId)
	q.Set("client_secret", c.authClientSecret)
	q.Set("refresh_token", token)
	q.Set("grant_type", "refresh_token")
	req, err := http.NewRequestWithContext(ctx, "POST", c.authURL+"/oauth/token", strings.NewReader(q.Encode()))
	if err != nil {
		return nil, &AuthorizeError{
			ErrorType:   ErrTypeRefreshFailed,
			Description: err.Error(),
		}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(q.Encode())))
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, &AuthorizeError{
			ErrorType:   ErrTypeRefreshFailed,
			Description: err.Error(),
		}
	}
	if resp.StatusCode != http.StatusOK {
		err := &AuthorizeError{
			Code:        resp.StatusCode,
			ErrorType:   ErrTypeRefreshFailed,
			Description: resp.Status,
		}
		json.NewDecoder(resp.Body).Decode(err)
		return nil, err
	}
	t := &storage.Token{}
	err = json.NewDecoder(resp.Body).Decode(t)
	if err != nil {
		return nil, &AuthorizeError{
			ErrorType:   ErrTypeRefreshFailed,
			Description: err.Error(),
		}
	}
	if t.TokenType == "" {
		t.TokenType = "Bearer"
	}
	if t.ExpiresIn != 0 {
		t.Expiry = c.utils.Now().Add(time.Duration(t.ExpiresIn) * time.Second)
	}
	return t, nil
}

func (c *client) saveToken(token *storage.Token) {
	c.storage.GetProfile().Token = token
	err := c.storage.SaveConfig()
	if err != nil {
		c.printer.ErrPrintf("Error: Token was refreshed but failed to save to storage: %v\n", err)
	}
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

func (c *client) introspection(ctx context.Context, token string) (*IntrospectionResponse, error) {
	form := url.Values{"token": {token}}.Encode()
	req, err := http.NewRequestWithContext(ctx, "POST", c.authURL+"/oauth/token/introspect", strings.NewReader(form))
	if err != nil {
		return nil, &AuthorizeError{
			ErrorType:   ErrTypeIntrospectionFailed,
			Description: err.Error(),
		}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(form)))
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, &AuthorizeError{
			ErrorType:   ErrTypeIntrospectionFailed,
			Description: err.Error(),
		}
	}
	if resp.StatusCode != http.StatusOK {
		err := &AuthorizeError{
			Code:        resp.StatusCode,
			ErrorType:   ErrTypeIntrospectionFailed,
			Description: resp.Status,
		}
		json.NewDecoder(resp.Body).Decode(err)
		return nil, err
	}
	ires := &IntrospectionResponse{}
	err = json.NewDecoder(resp.Body).Decode(ires)
	if err != nil {
		return nil, &AuthorizeError{
			ErrorType:   ErrTypeIntrospectionFailed,
			Description: err.Error(),
		}
	}
	return ires, nil
}

func (c *client) RevokeToken(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	form := url.Values{"token": {token}}.Encode()
	req, err := http.NewRequestWithContext(ctx, "POST", c.authURL+"/oauth/token/revoke", strings.NewReader(form))
	if err != nil {
		return &AuthorizeError{
			ErrorType:   ErrTypeRevokeFailed,
			Description: err.Error(),
		}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(form)))
	resp, err := c.http.Do(req)
	if err != nil {
		return &AuthorizeError{
			ErrorType:   ErrTypeRevokeFailed,
			Description: err.Error(),
		}
	}
	if resp.StatusCode != http.StatusOK {
		err := &AuthorizeError{
			Code:        resp.StatusCode,
			ErrorType:   ErrTypeRevokeFailed,
			Description: resp.Status,
		}
		json.NewDecoder(resp.Body).Decode(err)
		return err
	}
	return nil
}

func generateVerifier() string {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(data)
}

func generateS256Challenge(verifier string) string {
	sha := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sha[:])
}
