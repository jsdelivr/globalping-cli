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

func testLocationsMultipleWhitespace(t *testing.T) {
	locations := createLocations("New York, Los Angeles ")
	assert.Equal(t, []model.Locations{{Magic: "New York"}, {Magic: "Los Angeles"}}, locations)
}
