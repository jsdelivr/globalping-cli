package api

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/jsdelivr/globalping-cli/storage"
	"github.com/jsdelivr/globalping-cli/utils"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/jsdelivr/globalping-go"
)

const (
	GlobalpingAuthURL      = "https://auth.globalping.io"
	GlobalpingDashboardURL = "https://dash.globalping.io"
)

type Client interface {
	// Creates a new measurement with parameters set in the request body. The measurement runs asynchronously and you can retrieve its current state at the URL returned in the Location header.
	//
	// https://globalping.io/docs/api.globalping.io#post-/v1/measurements
	CreateMeasurement(ctx context.Context, measurement *globalping.MeasurementCreate) (*globalping.MeasurementCreateResponse, error)

	// Returns the status and results of an existing measurement. Measurements are typically available for up to 7 days after creation.
	//
	// https://globalping.io/docs/api.globalping.io#get-/v1/measurements/-id-
	GetMeasurement(ctx context.Context, id string) (*globalping.Measurement, error)

	// Waits for the measurement to complete and returns the results.
	//
	// https://globalping.io/docs/api.globalping.io#get-/v1/measurements/-id-
	AwaitMeasurement(ctx context.Context, id string) (*globalping.Measurement, error)

	// Returns the status and results of an existing measurement. Measurements are typically available for up to 7 days after creation.
	//
	// https://globalping.io/docs/api.globalping.io#get-/v1/measurements/-id-
	GetMeasurementRaw(ctx context.Context, id string) ([]byte, error)

	// Returns a link to be used for authorization and listens for the authorization callback.
	Authorize(ctx context.Context, callback func(error)) (*AuthorizeResponse, error)

	// Returns the introspection response for the token.
	//
	// If the token is empty, the client's current token will be used.
	TokenIntrospection(ctx context.Context, token string) (*IntrospectionResponse, error)

	// Removes the current token from the client. It also revokes the tokens if the refresh token is available.
	Logout(ctx context.Context) error

	// Revokes the token.
	RevokeToken(ctx context.Context, token string) error

	// Returns the rate limits for the current user or IP address.
	Limits(ctx context.Context) (*globalping.LimitsResponse, error)

	// Closes the client.
	Close()
}

type Config struct {
	Utils      utils.Utils
	Storage    *storage.LocalStorage
	Printer    *view.Printer
	Globalping globalping.Client
	HTTPClient *http.Client // If set, this client will be used for API requests and authorization

	DashboardURL string // optional

	AuthURL          string // optional
	AuthClientID     string
	AuthClientSecret string
	AuthToken        *storage.Token

	UserAgent string
}

type client struct {
	utils      utils.Utils
	storage    *storage.LocalStorage
	printer    *view.Printer
	http       *http.Client
	globalping globalping.Client
	mu         sync.RWMutex
	done       chan struct{}

	authClientId     string
	authClientSecret string
	token            *storage.Token

	authURL      string
	dashboardURL string
}

func NewClient(config Config) Client {
	c := &client{
		utils:            config.Utils,
		storage:          config.Storage,
		printer:          config.Printer,
		globalping:       config.Globalping,
		mu:               sync.RWMutex{},
		done:             make(chan struct{}),
		authClientId:     config.AuthClientID,
		authClientSecret: config.AuthClientSecret,
		authURL:          config.AuthURL,
		dashboardURL:     config.DashboardURL,
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
		c.token = &storage.Token{
			AccessToken:  config.AuthToken.AccessToken,
			TokenType:    config.AuthToken.TokenType,
			RefreshToken: config.AuthToken.RefreshToken,
			ExpiresIn:    config.AuthToken.ExpiresIn,
			Expiry:       config.AuthToken.Expiry,
		}
		if c.token.TokenType == "" {
			c.token.TokenType = "Bearer"
		}
	} else {
		profile := c.storage.GetProfile()
		if profile.Token != nil {
			c.token = &storage.Token{
				AccessToken:  profile.Token.AccessToken,
				TokenType:    profile.Token.TokenType,
				RefreshToken: profile.Token.RefreshToken,
				ExpiresIn:    profile.Token.ExpiresIn,
				Expiry:       profile.Token.Expiry,
			}
		}
	}

	t := time.NewTicker(10 * time.Second)
	go func() {
		defer t.Stop()

		for {
			select {
			case <-t.C:
				c.globalping.CacheClean()
			case <-c.done:
				return
			}
		}
	}()

	return c
}

func (c *client) Close() {
	close(c.done)
}
