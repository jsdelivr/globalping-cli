package cmd

import (
	"testing"

	"github.com/jsdelivr/globalping-cli/view"
	"github.com/stretchr/testify/assert"
)

func TestCreateContext(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T){
		"no_arg":             testContextNoArg,
		"country":            testContextCountry,
		"country_whitespace": testContextCountryWhitespace,
		"no_target":          testContextNoTarget,
		"ci_env":             testContextCIEnv,
	} {
		t.Run(scenario, func(t *testing.T) {
			ctx = &view.Context{}
			fn(t)
		})
	}
}

func testContextNoArg(t *testing.T) {
	err := createContext("test", []string{"1.1.1.1"})
	assert.Equal(t, "test", ctx.Cmd)
	assert.Equal(t, "1.1.1.1", ctx.Target)
	assert.Equal(t, "", ctx.From)
	assert.NoError(t, err)
}

func testContextCountry(t *testing.T) {
	err := createContext("test", []string{"1.1.1.1", "from", "Germany"})
	assert.Equal(t, "test", ctx.Cmd)
	assert.Equal(t, "1.1.1.1", ctx.Target)
	assert.Equal(t, "Germany", ctx.From)
	assert.NoError(t, err)
}

// Check if country with whitespace is parsed correctly
func testContextCountryWhitespace(t *testing.T) {
	err := createContext("test", []string{"1.1.1.1", "from", " Germany, France"})
	assert.Equal(t, "test", ctx.Cmd)
	assert.Equal(t, "1.1.1.1", ctx.Target)
	assert.Equal(t, "Germany, France", ctx.From)
	assert.NoError(t, err)
}

func testContextNoTarget(t *testing.T) {
	err := createContext("test", []string{})
	assert.Error(t, err)
}

func testContextCIEnv(t *testing.T) {
	t.Setenv("CI", "true")
	err := createContext("test", []string{"1.1.1.1"})
	assert.Equal(t, "test", ctx.Cmd)
	assert.Equal(t, "1.1.1.1", ctx.Target)
	assert.Equal(t, "", ctx.From)
	assert.True(t, ctx.CI)
	assert.NoError(t, err)
}
