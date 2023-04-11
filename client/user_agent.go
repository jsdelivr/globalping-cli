package client

import (
	"fmt"

	"github.com/jsdelivr/globalping-cli/version"
)

func userAgent() string {
	return fmt.Sprintf("globalping-cli/v%s (https://github.com/jsdelivr/globalping-cli)", version.Version)
}
