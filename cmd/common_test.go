package cmd

import (
	"errors"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/stretchr/testify/assert"
)

var (
	measurementID1 = "WOOxHNyhdsBQYEjU"
	measurementID2 = "hhUicONd75250Z1b"
	measurementID3 = "YPDXL29YeGctf6iJ"
	measurementID4 = "hH3tBVPZEj5k6AcW"

	defaultCurrentTime = time.Unix(0, 0)
)

func Test_UpdateContext(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T){
		"no_arg":             test_updateContext_NoArg,
		"country":            test_updateContext_Country,
		"country_whitespace": test_updateContext_CountryWhitespace,
		"no_target":          test_updateContext_NoTarget,
		"ci_env":             test_uodateContext_CIEnv,
	} {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
		})
	}
}

func test_updateContext_NoArg(t *testing.T) {
	ctx := &view.Context{}
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil)

	err := root.updateContext("test", []string{"1.1.1.1"})
	assert.Equal(t, "test", ctx.Cmd)
	assert.Equal(t, "1.1.1.1", ctx.Target)
	assert.Equal(t, "world", ctx.From)
	assert.NoError(t, err)
}

func test_updateContext_Country(t *testing.T) {
	ctx := &view.Context{}
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil)

	err := root.updateContext("test", []string{"1.1.1.1", "from", "Germany"})
	assert.Equal(t, "test", ctx.Cmd)
	assert.Equal(t, "1.1.1.1", ctx.Target)
	assert.Equal(t, "Germany", ctx.From)
	assert.NoError(t, err)
}

// Check if country with whitespace is parsed correctly
func test_updateContext_CountryWhitespace(t *testing.T) {
	ctx := &view.Context{}
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil)

	err := root.updateContext("test", []string{"1.1.1.1", "from", " Germany, France"})
	assert.Equal(t, "test", ctx.Cmd)
	assert.Equal(t, "1.1.1.1", ctx.Target)
	assert.Equal(t, "Germany, France", ctx.From)
	assert.NoError(t, err)
}

func test_updateContext_NoTarget(t *testing.T) {
	ctx := &view.Context{}
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil)

	err := root.updateContext("test", []string{})
	assert.Error(t, err)
}

func test_uodateContext_CIEnv(t *testing.T) {
	oldCI := os.Getenv("CI")
	t.Setenv("CI", "true")
	defer t.Setenv("CI", oldCI)

	ctx := &view.Context{}
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil)

	err := root.updateContext("test", []string{"1.1.1.1"})
	assert.Equal(t, "test", ctx.Cmd)
	assert.Equal(t, "1.1.1.1", ctx.Target)
	assert.Equal(t, "world", ctx.From)
	assert.True(t, ctx.CIMode)
	assert.NoError(t, err)
}

func Test_ParseTargetQuery_Simple(t *testing.T) {
	cmd := "ping"
	args := []string{"example.com"}

	q, err := parseTargetQuery(cmd, args)
	assert.NoError(t, err)

	assert.Equal(t, TargetQuery{Target: "example.com", From: ""}, *q)
}

func Test_ParseTargetQuery_SimpleWithResolver(t *testing.T) {
	cmd := "dns"
	args := []string{"example.com", "@1.1.1.1"}

	q, err := parseTargetQuery(cmd, args)
	assert.NoError(t, err)

	assert.Equal(t, TargetQuery{Target: "example.com", From: "", Resolver: "1.1.1.1"}, *q)
}

func Test_ParseTargetQuery_ResolverNotAllowed(t *testing.T) {
	cmd := "ping"
	args := []string{"example.com", "@1.1.1.1"}

	_, err := parseTargetQuery(cmd, args)
	assert.ErrorContains(t, err, "does not accept a resolver argument")
}

func Test_ParseTargetQuery_TargetFromX(t *testing.T) {
	cmd := "ping"
	args := []string{"example.com", "from", "London"}

	q, err := parseTargetQuery(cmd, args)
	assert.NoError(t, err)

	assert.Equal(t, TargetQuery{Target: "example.com", From: "London"}, *q)
}

func Test_ParseTargetQuery_TargetFromXWithResolver(t *testing.T) {
	cmd := "http"
	args := []string{"example.com", "from", "London", "@1.1.1.1"}

	q, err := parseTargetQuery(cmd, args)
	assert.NoError(t, err)

	assert.Equal(t, TargetQuery{Target: "example.com", From: "London", Resolver: "1.1.1.1"}, *q)
}

func Test_FindAndRemoveResolver_SimpleNoResolver(t *testing.T) {
	args := []string{"example.com"}

	resolver, argsWithoutResolver := findAndRemoveResolver(args)

	assert.Equal(t, "", resolver)
	assert.Equal(t, args, argsWithoutResolver)
}

func Test_FindAndRemoveResolver_NoResolver(t *testing.T) {
	args := []string{"example.com", "from", "London"}

	resolver, argsWithoutResolver := findAndRemoveResolver(args)

	assert.Equal(t, "", resolver)
	assert.Equal(t, args, argsWithoutResolver)
}

func Test_FindAndRemoveResolver_ResolverAndFrom(t *testing.T) {
	args := []string{"example.com", "@1.1.1.1", "from", "London"}

	resolver, argsWithoutResolver := findAndRemoveResolver(args)

	assert.Equal(t, "1.1.1.1", resolver)
	assert.Equal(t, []string{"example.com", "from", "London"}, argsWithoutResolver)
}

func Test_FindAndRemoveResolver_ResolverOnly(t *testing.T) {
	args := []string{"example.com", "@1.1.1.1"}

	resolver, argsWithoutResolver := findAndRemoveResolver(args)

	assert.Equal(t, "1.1.1.1", resolver)
	assert.Equal(t, []string{"example.com"}, argsWithoutResolver)
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

func test_CreateLocations_Multiple_Whitespace(t *testing.T) {
	locations, isPreviousMeasurementId, err := createLocations("New York, Los Angeles ")
	assert.Equal(t, []globalping.Locations{{Magic: "New York"}, {Magic: "Los Angeles"}}, locations)
	assert.False(t, isPreviousMeasurementId)
	assert.Nil(t, err)
}

func test_CreateLocations_Session_Last_Measurement(t *testing.T) {
	_ = saveIdToHistory(measurementID1)
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
	_ = saveIdToHistory(measurementID1)
	_ = saveIdToHistory(measurementID2)
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
	_ = saveIdToHistory(measurementID1)
	_ = saveIdToHistory(measurementID2)
	_ = saveIdToHistory(measurementID3)
	_ = saveIdToHistory(measurementID4)
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

	_ = saveIdToHistory(measurementID1)
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
	_ = saveIdToHistory(measurementID1)
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
	_ = saveIdToHistory(measurementID2)
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

func getMeasurementCreateResponse(id string) *globalping.MeasurementCreateResponse {
	return &globalping.MeasurementCreateResponse{
		ID:          id,
		ProbesCount: 1,
	}
}
