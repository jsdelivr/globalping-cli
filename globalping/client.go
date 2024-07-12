package globalping

import (
	"net/http"
	"sync"
	"time"
)

type Client interface {
	// Creates a new measurement with parameters set in the request body. The measurement runs asynchronously and you can retrieve its current state at the URL returned in the Location header.
	//
	// https://www.jsdelivr.com/docs/api.globalping.io#post-/v1/measurements
	CreateMeasurement(measurement *MeasurementCreate) (*MeasurementCreateResponse, error)
	// Returns the status and results of an existing measurement. Measurements are typically available for up to 7 days after creation.
	//
	// https://www.jsdelivr.com/docs/api.globalping.io#get-/v1/measurements/-id-
	GetMeasurement(id string) (*Measurement, error)
	// Returns the status and results of an existing measurement. Measurements are typically available for up to 7 days after creation.
	//
	// https://www.jsdelivr.com/docs/api.globalping.io#get-/v1/measurements/-id-
	GetMeasurementRaw(id string) ([]byte, error)
}

type Config struct {
	APIURL   string
	APIToken string
}

type CacheEntry struct {
	ETag     string
	Data     []byte
	ExpireAt int64 // Unix timestamp
}

type client struct {
	sync.RWMutex
	http  *http.Client
	cache map[string]*CacheEntry

	apiURL                        string
	apiToken                      string
	apiResponseCacheExpireSeconds int64
}

// NewClient creates a new client with the given configuration.
// The client will not have a cache cleanup goroutine, therefore cached responses will never be removed.
// If you want a cache cleanup goroutine, use NewClientWithCacheCleanup.
func NewClient(config Config) Client {
	return &client{
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiURL:   config.APIURL,
		apiToken: config.APIToken,
		cache:    map[string]*CacheEntry{},
	}
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
	c.RLock()
	defer c.RUnlock()
	e, ok := c.cache[id]
	if !ok {
		return ""
	}
	return e.ETag
}

func (c *client) getCachedResponse(id string) []byte {
	c.RLock()
	defer c.RUnlock()
	e, ok := c.cache[id]
	if !ok {
		return nil
	}
	return e.Data
}

func (c *client) cacheResponse(id string, etag string, resp []byte) {
	c.Lock()
	defer c.Unlock()
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
	c.Lock()
	defer c.Unlock()
	now := time.Now().Unix()
	for k, v := range c.cache {
		if v.ExpireAt > 0 && v.ExpireAt < now {
			delete(c.cache, k)
		}
	}
}
