package storage

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/icza/backscanner"
)

var (
	ErrNoPreviousMeasurements = errors.New("no previous measurements found")
	ErrInvalidIndex           = errors.New("invalid index")
	ErrIndexOutOfRange        = errors.New("index out of range")
)

var (
	measurementsFileName = "measurements"

	saveIdToSessionErr = "failed to save measurement ID: %s"
	readMeasuremetsErr = "failed to read previous measurements: %s"
)

// Returns the measurement ID at the given index from the session history
func (s *LocalStorage) GetIdFromSession(index int) (string, error) {
	if index == 0 {
		return "", ErrInvalidIndex
	}
	f, err := os.Open(s.measurementsPath())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", ErrNoPreviousMeasurements
		}
		return "", fmt.Errorf(readMeasuremetsErr, err)
	}
	defer f.Close()
	// Read ids from the end of the file
	if index < 0 {
		fStats, err := f.Stat()
		if err != nil {
			return "", fmt.Errorf(readMeasuremetsErr, err)
		}
		if fStats.Size() == 0 {
			return "", ErrNoPreviousMeasurements
		}
		scanner := backscanner.New(f, int(fStats.Size()-1)) // -1 to skip last newline
		for {
			index++
			b, _, err := scanner.LineBytes()
			if err != nil {
				if err == io.EOF {
					return "", ErrIndexOutOfRange
				}
				return "", fmt.Errorf(readMeasuremetsErr, err)
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

func (s *LocalStorage) SaveIdToSession(id string) error {
	f, err := os.OpenFile(s.measurementsPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf(saveIdToSessionErr, err)
	}
	defer f.Close()
	_, err = f.WriteString(id + "\n")
	if err != nil {
		return fmt.Errorf(saveIdToSessionErr, err)
	}
	return nil
}

func (s *LocalStorage) GetMeasurements() ([]byte, error) {
	return os.ReadFile(s.measurementsPath())
}

func (s *LocalStorage) measurementsPath() string {
	return s.joinSessionDir(measurementsFileName)
}
