package storage

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/icza/backscanner"
	"github.com/jsdelivr/globalping-cli/utils"
)

var (
	ErrReadHistory = errors.New("failed to read history")
)

var (
	historyFileName = "history"

	invalidHistoryItemErr = "invalid history item: %s"
)

const (
	// <version>|<index>|<time>|<id>|<command>
	HistoryItemVersion1 string = "1"
)

func (s *LocalStorage) GetHistoryIndex() (int, error) {
	f, err := os.Open(s.historyPath())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return 1, nil
		}
		return 0, ErrReadHistory
	}
	defer f.Close()
	fStats, err := f.Stat()
	if err != nil {
		return 0, ErrReadHistory
	}
	if fStats.Size() == 0 {
		return 1, nil
	}
	scanner := backscanner.New(f, int(fStats.Size()-1)) // -1 to skip last newline
	for {
		b, _, err := scanner.LineBytes()
		if err != nil {
			if err == io.EOF {
				return 1, nil
			}
			return 0, ErrReadHistory
		}
		parts, err := getHistoryItem(string(b))
		if err != nil {
			return 0, err
		}
		if parts[1] == "-" {
			continue
		}
		index, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, err
		}
		return index + 1, nil
	}
}

func (s *LocalStorage) GetHistory(limit int) ([]string, error) {
	items := make([]string, 0)
	f, err := os.Open(s.historyPath())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return items, nil
		}
		return nil, ErrReadHistory
	}
	defer f.Close()
	// Read from the end of the file
	if limit < 0 {
		fStats, err := f.Stat()
		if err != nil {
			return nil, ErrReadHistory
		}
		if fStats.Size() == 0 {
			return items, nil
		}
		scanner := backscanner.New(f, int(fStats.Size()-1)) // -1 to skip last newline
		for {
			limit++
			b, _, err := scanner.LineBytes()
			if err != nil {
				if err == io.EOF {
					return items, nil
				}
				return nil, ErrReadHistory
			}
			item, err := parseHistoryItem(string(b))
			if err != nil {
				return nil, err
			}
			items = append(items, item)
			if limit == 0 {
				break
			}
		}
		return items, nil
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		limit--
		item, err := parseHistoryItem(scanner.Text())
		if err != nil {
			return nil, err
		}
		items = append(items, item)
		if limit == 0 {
			break
		}
	}
	return items, nil
}

func (s *LocalStorage) SaveCommandToHistory(
	index string,
	time int64,
	ids string,
	cmd string,
) error {
	f, err := os.OpenFile(s.historyPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("%s|%s|%d|%s|%s\n", HistoryItemVersion1, index, time, ids, cmd))
	if err != nil {
		return err
	}
	return nil
}

func (s *LocalStorage) historyPath() string {
	return s.joinSessionDir(historyFileName)
}

func parseHistoryItem(line string) (string, error) {
	parts, err := getHistoryItem(line)
	if err != nil {
		return "", err
	}
	if parts[0] == HistoryItemVersion1 {
		t, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return "", fmt.Errorf(invalidHistoryItemErr, line)
		}
		return fmt.Sprintf(
			"%s | %s | %s\n%s",
			parts[1],
			time.Unix(t, 0).Format("2006-01-02 15:04:05"),
			parts[4],
			"> "+utils.ShareURL+parts[3],
		), nil
	}
	return "", nil
}

func getHistoryItem(line string) ([]string, error) {
	parts := strings.Split(line, "|")
	if len(parts) < 1 {
		return nil, fmt.Errorf(invalidHistoryItemErr, line)
	}
	switch parts[0] {
	case HistoryItemVersion1:
		if len(parts) != 5 {
			return nil, fmt.Errorf(invalidHistoryItemErr, line)
		}
		return parts, nil
	default:
		return nil, fmt.Errorf(invalidHistoryItemErr, line)
	}
}
