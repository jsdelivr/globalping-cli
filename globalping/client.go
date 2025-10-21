package globalping

import (
	"net/http"
	"sync"
	"time"
)

const (
	GlobalpingAPIURL       = "https://api.globalping.io/v1"
	GlobalpingAuthURL      = "https://auth.globalping.io"
	GlobalpingDashboardURL = "https://dash.globalping.io"
)

type Client interface {
	// Creates a new measurement with parameters set in the request body. The measurement runs asynchronously and you can retrieve its current state at the URL returned in the Location header.
	//
	// https://globalping.io/docs/api.globalping.io#post-/v1/measurements
	CreateMeasurement(measurement *MeasurementCreate) (*MeasurementCreateResponse, error)

	// Returns the status and results of an existing measurement. Measurements are typically available for up to 7 days after creation.
	//
	// https://globalping.io/docs/api.globalping.io#get-/v1/measurements/-id-
	GetMeasurement(id string) (*Measurement, error)

	// Waits for the measurement to complete and returns the results.
	//
	// https://globalping.io/docs/api.globalping.io#get-/v1/measurements/-id-
	AwaitMeasurement(id string) (*Measurement, error)

	// Returns the status and results of an existing measurement. Measurements are typically available for up to 7 days after creation.
	//
	// https://globalping.io/docs/api.globalping.io#get-/v1/measurements/-id-
	GetMeasurementRaw(id string) ([]byte, error)

	// Returns a link to be used for authorization and listens for the authorization callback.
	//
	// onTokenRefresh will be called if the authorization is successful.
	Authorize(callback func(error)) (*AuthorizeResponse, error)

	// Returns the introspection response for the token.
	//
	// If the token is empty, the client's current token will be used.
	TokenIntrospection(token string) (*IntrospectionResponse, error)

	// Removes the current token from the client. It also revokes the tokens if the refresh token is available.
	//
	// onTokenRefresh will be called if the token is successfully removed.
	Logout() error

	// Revokes the token.
	RevokeToken(token string) error

	// Returns the rate limits for the current user or IP address.
	Limits() (*LimitsResponse, error)
}

type Config struct {
	HTTPClient *http.Client // If set, this client will be used for API requests and authorization

	APIURL       string // optional
	DashboardURL string // optional

	AuthURL          string // optional
	AuthClientID     string
	AuthClientSecret string
	AuthToken        *Token
	OnTokenRefresh   func(*Token)

	UserAgent string
}

type CacheEntry struct {
	ETag     string
	Data     []byte
	ExpireAt int64 // Unix timestamp
}

type client struct {
	mu    sync.RWMutex
	http  *http.Client
	cache map[string]*CacheEntry

	authClientId     string
	authClientSecret string
	token            *Token
	onTokenRefresh   func(*Token)

	apiURL                        string
	authURL                       string
	dashboardURL                  string
	apiResponseCacheExpireSeconds int64
	userAgent                     string
}

// NewClient creates a new client with the given configuration.
// The client will not have a cache cleanup goroutine, therefore cached responses will never be removed.
// If you want a cache cleanup goroutine, use NewClientWithCacheCleanup.
func NewClient(config Config) Client {
	c := &client{
		mu:               sync.RWMutex{},
		authClientId:     config.AuthClientID,
		authClientSecret: config.AuthClientSecret,
		onTokenRefresh:   config.OnTokenRefresh,
		apiURL:           config.APIURL,
		authURL:          config.AuthURL,
		dashboardURL:     config.DashboardURL,
		userAgent:        config.UserAgent,
		cache:            map[string]*CacheEntry{},
	}

	if config.APIURL == "" {
		c.apiURL = GlobalpingAPIURL
	}
	if config.AuthURL == "" {
		c.authURL = GlobalpingAuthURL
	}
	if config.DashboardURL == "" {
		c.dashboardURL = GlobalpingDashboardURL
	}

	if config.HTTPClient != nil {
		c.http = config.HTTPClient
	} else {
		c.http = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	if config.AuthToken != nil {
		c.token = &Token{
			AccessToken:  config.AuthToken.AccessToken,
			TokenType:    config.AuthToken.TokenType,
			RefreshToken: config.AuthToken.RefreshToken,
			ExpiresIn:    config.AuthToken.ExpiresIn,
			Expiry:       config.AuthToken.Expiry,
		}
		if c.token.TokenType == "" {
			c.token.TokenType = "Bearer"
		}
	}
	return c
}

// NewClientWithCacheCleanup creates a new client with a cache cleanup goroutine that runs every t.
// The cache cleanup goroutine will remove entries that have expired.
// If cacheExpireSeconds is 0, the cache entries will never expire.
func NewClientWithCacheCleanup(config Config, t *time.Ticker, cacheExpireSeconds int64) Client {
	c := NewClient(config).(*client)
	c.apiResponseCacheExpireSeconds = cacheExpireSeconds
	go func() {
		for range t.C {
			c.cleanupCache()
		}
	}()
	return c
}

func (c *client) getETag(id string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.cache[id]
	if !ok {
		return ""
	}
	return e.ETag
}

func (c *client) getCachedResponse(id string) []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.cache[id]
	if !ok {
		return nil
	}
	return e.Data
}

func (c *client) cacheResponse(id string, etag string, resp []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var expires int64
	if c.apiResponseCacheExpireSeconds != 0 {
		expires = time.Now().Unix() + c.apiResponseCacheExpireSeconds
	}
	e, ok := c.cache[id]
	if ok {
		e.ETag = etag
		e.Data = resp
		e.ExpireAt = expires
	} else {
		c.cache[id] = &CacheEntry{
			ETag:     etag,
			Data:     resp,
			ExpireAt: expires,
		}
	}
}

func (c *client) cleanupCache() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now().Unix()
	for k, v := range c.cache {
		if v.ExpireAt > 0 && v.ExpireAt < now {
			delete(c.cache, k)
		}
	}
}
