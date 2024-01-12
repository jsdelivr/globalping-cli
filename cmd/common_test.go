package cmd

import (
	"errors"
	"io/fs"
	"os"
	"testing"

	"github.com/jsdelivr/globalping-cli/model"
	"github.com/stretchr/testify/assert"
)

var (
	measurementID1 = "WOOxHNyhdsBQYEjU"
	measurementID2 = "hhUicONd75250Z1b"
	measurementID3 = "YPDXL29YeGctf6iJ"
	measurementID4 = "hH3tBVPZEj5k6AcW"
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
		"valid_single":                       testLocationsSingle,
		"valid_multiple":                     testLocationsMultiple,
		"valid_multiple_whitespace":          testLocationsMultipleWhitespace,
		"valid_session_last_measurement":     testCreateLocationsSessionLastMeasurement,
		"valid_session_first_measurement":    testCreateLocationsSessionFirstMeasurement,
		"valid_session_measurement_at_index": testCreateLocationsSessionMeasurementAtIndex,
		"valid_session_no_session":           testCreateLocationsSessionNoSession,
		"invalid_session_index":              testCreateLocationsSessionInvalidIndex,
	} {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
			t.Cleanup(func() {
				sessionPath := getSessionPath()
				err := os.RemoveAll(sessionPath)
				if err != nil && !errors.Is(err, fs.ErrNotExist) {
					t.Fatalf("Failed to remove session path: %s", err)
				}
			})
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

func testCreateLocationsSessionLastMeasurement(t *testing.T) {
	_ = saveMeasurementID(measurementID1)
	locations, isPreviousMeasurementId, err := createLocations("@1")
	assert.Equal(t, []model.Locations{{Magic: measurementID1}}, locations)
	assert.True(t, isPreviousMeasurementId)
	assert.Nil(t, err)

	locations, isPreviousMeasurementId, err = createLocations("last")
	assert.Equal(t, []model.Locations{{Magic: measurementID1}}, locations)
	assert.True(t, isPreviousMeasurementId)
	assert.Nil(t, err)

	locations, isPreviousMeasurementId, err = createLocations("previous")
	assert.Equal(t, []model.Locations{{Magic: measurementID1}}, locations)
	assert.True(t, isPreviousMeasurementId)
	assert.Nil(t, err)
}

func testCreateLocationsSessionFirstMeasurement(t *testing.T) {
	_ = saveMeasurementID(measurementID1)
	_ = saveMeasurementID(measurementID2)
	locations, isPreviousMeasurementId, err := createLocations("@-1")
	assert.Equal(t, []model.Locations{{Magic: measurementID2}}, locations)
	assert.True(t, isPreviousMeasurementId)
	assert.Nil(t, err)

	locations, isPreviousMeasurementId, err = createLocations("last")
	assert.Equal(t, []model.Locations{{Magic: measurementID2}}, locations)
	assert.True(t, isPreviousMeasurementId)
	assert.Nil(t, err)
}

func testCreateLocationsSessionMeasurementAtIndex(t *testing.T) {
	_ = saveMeasurementID(measurementID1)
	_ = saveMeasurementID(measurementID2)
	_ = saveMeasurementID(measurementID3)
	_ = saveMeasurementID(measurementID4)
	locations, isPreviousMeasurementId, err := createLocations("@2")
	assert.Equal(t, []model.Locations{{Magic: measurementID2}}, locations)
	assert.True(t, isPreviousMeasurementId)
	assert.Nil(t, err)

	locations, isPreviousMeasurementId, err = createLocations("@-2")
	assert.Equal(t, []model.Locations{{Magic: measurementID3}}, locations)
	assert.True(t, isPreviousMeasurementId)
	assert.Nil(t, err)

	locations, isPreviousMeasurementId, err = createLocations("@-4")
	assert.Equal(t, []model.Locations{{Magic: measurementID1}}, locations)
	assert.True(t, isPreviousMeasurementId)
	assert.Nil(t, err)
}

func testCreateLocationsSessionNoSession(t *testing.T) {
	locations, isPreviousMeasurementId, err := createLocations("@1")
	assert.Nil(t, locations)
	assert.False(t, isPreviousMeasurementId)
	assert.Equal(t, ErrorNoPreviousMeasurements, err)
}

func testCreateLocationsSessionInvalidIndex(t *testing.T) {
	locations, isPreviousMeasurementId, err := createLocations("@0")
	assert.Nil(t, locations)
	assert.False(t, isPreviousMeasurementId)
	assert.Equal(t, ErrInvalidIndex, err)

	locations, isPreviousMeasurementId, err = createLocations("@")
	assert.Nil(t, locations)
	assert.False(t, isPreviousMeasurementId)
	assert.Equal(t, ErrInvalidIndex, err)

	locations, isPreviousMeasurementId, err = createLocations("@x")
	assert.Nil(t, locations)
	assert.False(t, isPreviousMeasurementId)
	assert.Equal(t, ErrInvalidIndex, err)

	_ = saveMeasurementID(measurementID1)
	locations, isPreviousMeasurementId, err = createLocations("@2")
	assert.Nil(t, locations)
	assert.False(t, isPreviousMeasurementId)
	assert.Equal(t, ErrIndexOutOfRange, err)

	locations, isPreviousMeasurementId, err = createLocations("@-2")
	assert.Nil(t, locations)
	assert.False(t, isPreviousMeasurementId)
	assert.Equal(t, ErrIndexOutOfRange, err)
}

func TestSaveMeasurementID(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T){
		"valid_new_session":      testSaveMeasurementIDNewSession,
		"valid_existing_session": testSaveMeasurementIDExistingSession,
	} {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
			t.Cleanup(func() {
				sessionPath := getSessionPath()
				err := os.RemoveAll(sessionPath)
				if err != nil && !os.IsNotExist(err) {
					t.Fatalf("Failed to remove session path: %s", err)
				}
			})
		})
	}
}

func testSaveMeasurementIDNewSession(t *testing.T) {
	_ = saveMeasurementID(measurementID1)
	assert.FileExists(t, getMeasurementsPath())
	b, err := os.ReadFile(getMeasurementsPath())
	assert.NoError(t, err)
	expected := []byte(measurementID1 + "\n")
	assert.Equal(t, expected, b)
}

func testSaveMeasurementIDExistingSession(t *testing.T) {
	err := os.Mkdir(getSessionPath(), 0755)
	if err != nil {
		t.Fatalf("Failed to create session path: %s", err)
	}
	err = os.WriteFile(getMeasurementsPath(), []byte(measurementID1+"\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create measurements file: %s", err)
	}
	_ = saveMeasurementID(measurementID2)
	b, err := os.ReadFile(getMeasurementsPath())
	assert.NoError(t, err)
	expected := []byte(measurementID1 + "\n" + measurementID2 + "\n")
	assert.Equal(t, expected, b)
}
