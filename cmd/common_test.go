package cmd

import (
	"errors"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/stretchr/testify/assert"
)

var (
	measurementID1 = "WOOxHNyhdsBQYEjU"
	measurementID2 = "hhUicONd75250Z1b"
	measurementID3 = "YPDXL29YeGctf6iJ"
	measurementID4 = "hH3tBVPZEj5k6AcW"

	defaultCurrentTime = time.Unix(0, 0)
)

func Test_InProgressUpdates_CI(t *testing.T) {
	ci := true
	assert.Equal(t, false, inProgressUpdates(ci))
}

func Test_InProgressUpdates_NotCI(t *testing.T) {
	ci := false
	assert.Equal(t, true, inProgressUpdates(ci))
}

func Test_CreateLocations(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T){
		"valid_single":                       test_CreateLocations_Single,
		"valid_multiple":                     test_CreateLocations_Multiple,
		"valid_multiple_whitespace":          test_CreateLocations_Multiple_Whitespace,
		"valid_session_last_measurement":     test_CreateLocations_Session_Last_Measurement,
		"valid_session_first_measurement":    test_CreateLocations_Session_First_Measurement,
		"valid_session_measurement_at_index": test_CreateLocations_Session_Measurement_At_Index,
		"valid_session_no_session":           test_CreateLocations_Session_No_Session,
		"invalid_session_index":              test_CreateLocations_Session_Invalid_Index,
	} {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
			t.Cleanup(sessionCleanup)
		})
	}
}

func test_CreateLocations_Single(t *testing.T) {
	locations, isPreviousMeasurementId, err := createLocations("New York")
	assert.Equal(t, []globalping.Locations{{Magic: "New York"}}, locations)
	assert.False(t, isPreviousMeasurementId)
	assert.Nil(t, err)
}

func test_CreateLocations_Multiple(t *testing.T) {
	locations, isPreviousMeasurementId, err := createLocations("New York,Los Angeles")
	assert.Equal(t, []globalping.Locations{{Magic: "New York"}, {Magic: "Los Angeles"}}, locations)
	assert.False(t, isPreviousMeasurementId)
	assert.Nil(t, err)
}

// Check if multiple locations with whitespace are parsed correctly
func test_CreateLocations_Multiple_Whitespace(t *testing.T) {
	locations, isPreviousMeasurementId, err := createLocations("New York, Los Angeles ")
	assert.Equal(t, []globalping.Locations{{Magic: "New York"}, {Magic: "Los Angeles"}}, locations)
	assert.False(t, isPreviousMeasurementId)
	assert.Nil(t, err)
}

func test_CreateLocations_Session_Last_Measurement(t *testing.T) {
	_ = saveMeasurementID(measurementID1)
	locations, isPreviousMeasurementId, err := createLocations("@1")
	assert.Equal(t, []globalping.Locations{{Magic: measurementID1}}, locations)
	assert.True(t, isPreviousMeasurementId)
	assert.Nil(t, err)

	locations, isPreviousMeasurementId, err = createLocations("last")
	assert.Equal(t, []globalping.Locations{{Magic: measurementID1}}, locations)
	assert.True(t, isPreviousMeasurementId)
	assert.Nil(t, err)

	locations, isPreviousMeasurementId, err = createLocations("previous")
	assert.Equal(t, []globalping.Locations{{Magic: measurementID1}}, locations)
	assert.True(t, isPreviousMeasurementId)
	assert.Nil(t, err)
}

func test_CreateLocations_Session_First_Measurement(t *testing.T) {
	_ = saveMeasurementID(measurementID1)
	_ = saveMeasurementID(measurementID2)
	locations, isPreviousMeasurementId, err := createLocations("@-1")
	assert.Equal(t, []globalping.Locations{{Magic: measurementID2}}, locations)
	assert.True(t, isPreviousMeasurementId)
	assert.Nil(t, err)

	locations, isPreviousMeasurementId, err = createLocations("last")
	assert.Equal(t, []globalping.Locations{{Magic: measurementID2}}, locations)
	assert.True(t, isPreviousMeasurementId)
	assert.Nil(t, err)
}

func test_CreateLocations_Session_Measurement_At_Index(t *testing.T) {
	_ = saveMeasurementID(measurementID1)
	_ = saveMeasurementID(measurementID2)
	_ = saveMeasurementID(measurementID3)
	_ = saveMeasurementID(measurementID4)
	locations, isPreviousMeasurementId, err := createLocations("@2")
	assert.Equal(t, []globalping.Locations{{Magic: measurementID2}}, locations)
	assert.True(t, isPreviousMeasurementId)
	assert.Nil(t, err)

	locations, isPreviousMeasurementId, err = createLocations("@-2")
	assert.Equal(t, []globalping.Locations{{Magic: measurementID3}}, locations)
	assert.True(t, isPreviousMeasurementId)
	assert.Nil(t, err)

	locations, isPreviousMeasurementId, err = createLocations("@-4")
	assert.Equal(t, []globalping.Locations{{Magic: measurementID1}}, locations)
	assert.True(t, isPreviousMeasurementId)
	assert.Nil(t, err)
}

func test_CreateLocations_Session_No_Session(t *testing.T) {
	locations, isPreviousMeasurementId, err := createLocations("@1")
	assert.Nil(t, locations)
	assert.False(t, isPreviousMeasurementId)
	assert.Equal(t, ErrorNoPreviousMeasurements, err)
}

func test_CreateLocations_Session_Invalid_Index(t *testing.T) {
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

func Test_SaveMeasurementID(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T){
		"valid_new_session":      test_SaveMeasurementID_New_Session,
		"valid_existing_session": test_SaveMeasurementID_Existing_Session,
	} {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
			t.Cleanup(sessionCleanup)
		})
	}
}

func test_SaveMeasurementID_New_Session(t *testing.T) {
	_ = saveMeasurementID(measurementID1)
	assert.FileExists(t, getMeasurementsPath())
	b, err := os.ReadFile(getMeasurementsPath())
	assert.NoError(t, err)
	expected := []byte(measurementID1 + "\n")
	assert.Equal(t, expected, b)
}

func test_SaveMeasurementID_Existing_Session(t *testing.T) {
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

func sessionCleanup() {
	sessionPath := getSessionPath()
	err := os.RemoveAll(sessionPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		panic("Failed to remove session path: " + err.Error())
	}
}
