package main

import "github.com/jsdelivr/globalping-cli/cmd"

var (
	// https://goreleaser.com/cookbooks/using-main.version/
	version = "dev"
)

func main() {
	cmd.Execute(version)
}
