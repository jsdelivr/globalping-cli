package api

import (
	"runtime"
)

const UserAgent = "Globalping API Go Client"

const (
	apiRoot    = "https://api.perfops.net"
	basePath   = apiRoot
	libVersion = "v1.0.0"
	userAgent  = UserAgent + "/" + libVersion + " (" + runtime.GOOS + "/" + runtime.GOARCH + ")"
)
