package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"github.com/jsdelivr/globalping-cli/model"
)

func inProgressUpdates(ci bool) bool {
	return !(ci)
}

func createLocations(from string) ([]model.Locations, bool, error) {
	fromArr := strings.Split(from, ",")
	if len(fromArr) == 1 {
		mId, err := mapToMeasurementID(fromArr[0])
		if err != nil {
			return nil, false, err
		}
		isPreviousMeasurementId := false
		if mId == "" {
			mId = strings.TrimSpace(fromArr[0])
		} else {
			isPreviousMeasurementId = true
		}
		return []model.Locations{
			{
				Magic: mId,
			},
		}, isPreviousMeasurementId, nil
	}
	locations := make([]model.Locations, len(fromArr))
	for i, v := range fromArr {
		locations[i] = model.Locations{
			Magic: strings.TrimSpace(v),
		}
	}
	return locations, false, nil
}

// Maps a location like @1, ~1, last or previous to a measurement ID
func mapToMeasurementID(location string) (string, error) {
	if location == "" {
		return "", nil
	}
	if location[0] == '@' {
		index, err := strconv.Atoi(location[1:])
		if err != nil {
			return "", fmt.Errorf("%s: invalid index", location[1:])
		}
		return getMeasurementID(index)
	}
	if location[0] == '~' {
		index, err := strconv.Atoi(location[1:])
		if err != nil {
			return "", fmt.Errorf("%s: invalid index", location[1:])
		}
		return getMeasurementID(-index)
	}
	if location == "last" || location == "previous" {
		return getMeasurementID(-1)
	}
	return "", nil
}

// Returns the measurement ID at the given index from the session history
func getMeasurementID(index int) (string, error) {
	if index == 0 {
		return "", fmt.Errorf("%d: invalid index", index)
	}
	f, err := os.Open(getMeasurementsPath())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", errors.New("no previous measurements found")
		}
		return "", fmt.Errorf("failed to open previous measurements file: %s", err)
	}
	defer f.Close()
	// TODO: Read the file in reverse
	if index < 0 {
		return "", errors.New("not supported")
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		index--
		if index == 0 {
			return scanner.Text(), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read previous measurements: %s", err)
	}
	return "", fmt.Errorf("%d: index out of range", index)
}

// Saves the measurement ID to the session history
func saveMeasurementID(id string) error {
	_, err := os.Stat(getSessionPath())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err := os.Mkdir(getSessionPath(), 0755)
			if err != nil {
				return fmt.Errorf("failed to save measurement ID: %s", err)
			}
		} else {
			return fmt.Errorf("failed to save measurement ID: %s", err)
		}
	}
	f, err := os.OpenFile(getMeasurementsPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to save measurement ID: %s", err)
	}
	defer f.Close()
	_, err = f.WriteString(id + "\n")
	if err != nil {
		return fmt.Errorf("failed to save measurement ID: %s", err)
	}
	return nil
}

func getSessionPath() string {
	return fmt.Sprintf("%s/globalping_%d", os.TempDir(), os.Getppid())
}

func getMeasurementsPath() string {
	return getSessionPath() + "/measurements"
}
