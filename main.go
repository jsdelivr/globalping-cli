package main

import (
	"github.com/jsdelivr/globalping-cli/cmd"
	pkgversion "github.com/jsdelivr/globalping-cli/version"
)

var (
	// https://goreleaser.com/cookbooks/using-main.version/
	version = "1.5.2"
)

func main() {
	pkgversion.Version = version
	cmd.Execute()
}
