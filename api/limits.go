package api

import (
	"context"

	"github.com/jsdelivr/globalping-go"
)

func (c *client) Limits(ctx context.Context) (*globalping.LimitsResponse, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, err
	}
	if token != nil {
		c.globalping.SetToken(token.AccessToken)
	} else {
		c.globalping.SetToken("")
	}
	return c.globalping.Limits(ctx)
}
