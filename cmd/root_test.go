package cmd

import (
	"testing"

	"github.com/jsdelivr/globalping-cli/model"

	"github.com/stretchr/testify/assert"
)

func TestCreateLocations(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T){
		"valid_single":              testLocationsSingle,
		"valid_multiple":            testLocationsMultiple,
		"valid_multiple_whitespace": testLocationsMultipleWhitespace,
	} {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
		})
	}
}

func testLocationsSingle(t *testing.T) {
	locations := createLocations("New York")
	assert.Equal(t, []model.Locations{{Magic: "New York"}}, locations)
}

func testLocationsMultiple(t *testing.T) {
	locations := createLocations("New York,Los Angeles")
	assert.Equal(t, []model.Locations{{Magic: "New York"}, {Magic: "Los Angeles"}}, locations)
}

// Check if multiple locations with whitespace are parsed correctly
func testLocationsMultipleWhitespace(t *testing.T) {
	locations := createLocations("New York, Los Angeles ")
	assert.Equal(t, []model.Locations{{Magic: "New York"}, {Magic: "Los Angeles"}}, locations)
}

func TestCreateContext(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T){
		"no_arg":             testContextNoArg,
		"country":            testContextCountry,
		"country_whitespace": testContextCountryWhitespace,
		"no_target":          testContextNoTarget,
		"ci_env":             testContextCIEnv,
	} {
		t.Run(scenario, func(t *testing.T) {
			ctx = model.Context{}
			fn(t)
		})
	}
}

func testContextNoArg(t *testing.T) {
	err := createContext("test", []string{"1.1.1.1"})
	assert.Equal(t, "test", ctx.Cmd)
	assert.Equal(t, "1.1.1.1", ctx.Target)
	assert.Equal(t, "world", ctx.From)
	assert.False(t, ctx.CI)
	assert.NoError(t, err)
}

func testContextCountry(t *testing.T) {
	err := createContext("test", []string{"1.1.1.1", "from", "Germany"})
	assert.Equal(t, "test", ctx.Cmd)
	assert.Equal(t, "1.1.1.1", ctx.Target)
	assert.Equal(t, "Germany", ctx.From)
	assert.False(t, ctx.CI)
	assert.NoError(t, err)
}

// Check if country with whitespace is parsed correctly
func testContextCountryWhitespace(t *testing.T) {
	err := createContext("test", []string{"1.1.1.1", "from", " Germany, France"})
	assert.Equal(t, "test", ctx.Cmd)
	assert.Equal(t, "1.1.1.1", ctx.Target)
	assert.Equal(t, "Germany, France", ctx.From)
	assert.False(t, ctx.CI)
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
	assert.Equal(t, "world", ctx.From)
	assert.True(t, ctx.CI)
	assert.NoError(t, err)
}
