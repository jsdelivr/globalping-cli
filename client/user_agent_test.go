package client

import (
	"testing"

	"github.com/jsdelivr/globalping-cli/version"
	"github.com/stretchr/testify/assert"
)

func TestUserAgent(t *testing.T) {
	version.Version = "x.y.z"
	assert.Equal(t, "globalping-cli/vx.y.z (https://github.com/jsdelivr/globalping-cli)", userAgent())
}
