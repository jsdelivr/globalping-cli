package cmd

import (
	"testing"

	"github.com/jsdelivr/globalping-cli/model"
	"github.com/stretchr/testify/assert"
)

func TestInProgressUpdates_CI(t *testing.T) {
	ci := true
	assert.Equal(t, false, inProgressUpdates(ci))
}

func TestInProgressUpdates_NotCI(t *testing.T) {
	ci := false
	assert.Equal(t, true, inProgressUpdates(ci))
}

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
	locations, isPreviousMeasurementId, err := createLocations("New York")
	assert.Equal(t, []model.Locations{{Magic: "New York"}}, locations)
	assert.False(t, isPreviousMeasurementId)
	assert.Nil(t, err)
}

func testLocationsMultiple(t *testing.T) {
	locations, isPreviousMeasurementId, err := createLocations("New York,Los Angeles")
	assert.Equal(t, []model.Locations{{Magic: "New York"}, {Magic: "Los Angeles"}}, locations)
	assert.False(t, isPreviousMeasurementId)
	assert.Nil(t, err)
}

// Check if multiple locations with whitespace are parsed correctly
func testLocationsMultipleWhitespace(t *testing.T) {
	locations, isPreviousMeasurementId, err := createLocations("New York, Los Angeles ")
	assert.Equal(t, []model.Locations{{Magic: "New York"}, {Magic: "Los Angeles"}}, locations)
	assert.False(t, isPreviousMeasurementId)
	assert.Nil(t, err)
}
