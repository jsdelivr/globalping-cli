package cmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHttpCmd(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T){
		"parseUrl":    testParseUrl,
		"overrideOpt": testOverrideOpt,
	} {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
		})
	}
}

func testParseUrl(t *testing.T) {
	flags, _ := parseURL("https://cdn.jsdelivr.net:8080/npm/react/?query=3")
	fmt.Printf("%+v", flags)
	assert.Equal(t, "/npm/react/", flags.Path)
	assert.Equal(t, "cdn.jsdelivr.net", flags.Host)
	assert.Equal(t, "https", flags.Protocol)
	assert.Equal(t, 8080, flags.Port)
	assert.Equal(t, "query=3", flags.Query)
}

func testOverrideOpt(t *testing.T) {
	assert.Equal(t, "new", overrideOpt("orig", "new"))
	assert.Equal(t, "orig", overrideOpt("orig", ""))
	assert.Equal(t, 10, overrideOptInt(0, 10))
	assert.Equal(t, 10, overrideOptInt(10, 0))
}
