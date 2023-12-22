package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"github.com/icza/backscanner"
	"github.com/jsdelivr/globalping-cli/model"
)

var (
	ErrorNoPreviousMeasurements = errors.New("no previous measurements found")
	ErrInvalidIndex             = errors.New("invalid index")
	ErrIndexOutOfRange          = errors.New("index out of range")
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

// Maps a location to a measurement ID if possible
func mapToMeasurementID(location string) (string, error) {
	if location == "" {
		return "", nil
	}
	if location[0] == '@' {
		index, err := strconv.Atoi(location[1:])
		if err != nil {
			return "", ErrInvalidIndex
		}
		return getMeasurementID(index)
	}
	if location == "first" {
		return getMeasurementID(1)
	}
	if location == "last" {
		return getMeasurementID(-1)
	}
	return "", nil
}

// Returns the measurement ID at the given index from the session history
func getMeasurementID(index int) (string, error) {
	if index == 0 {
		return "", ErrInvalidIndex
	}
	f, err := os.Open(getMeasurementsPath())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", ErrorNoPreviousMeasurements
		}
		return "", fmt.Errorf("failed to open previous measurements file: %s", err)
	}
	defer f.Close()
	// Read ids from the end of the file
	if index < 0 {
		fStats, err := f.Stat()
		if err != nil {
			return "", fmt.Errorf("failed to read previous measurements: %s", err)
		}
		if fStats.Size() == 0 {
			return "", ErrorNoPreviousMeasurements
		}
		scanner := backscanner.New(f, int(fStats.Size()-1)) // -1 to skip last newline
		for {
			index++
			b, _, err := scanner.LineBytes()
			if err != nil {
				if err == io.EOF {
					return "", ErrIndexOutOfRange
				}
				return "", fmt.Errorf("failed to read previous measurements: %s", err)
			}
			if index == 0 {
				return string(b), nil
			}
		}
	}
	// Read ids from the beginning of the file
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
	return "", ErrIndexOutOfRange
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
