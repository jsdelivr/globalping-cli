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
	"time"

	"github.com/icza/backscanner"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/spf13/cobra"
)

var (
	ErrReadHistory = errors.New("failed to read history")
)

var (
	invalidHistoryItemErr = "invalid history item: %s"
	saveToHistoryErr      = "failed to save to history: %s"
)

const (
	// <version>|<index>|<time>|<id>|<command>
	HistoryItemVersion1 string = "1"
)

func (r *Root) initHistory() {
	historyCmd := &cobra.Command{
		Run:   r.RunHistory,
		Use:   "history",
		Short: "Display the measurement history of your current session",
		Long: `Display the measurement history of your current session.

Examples:
  # Display all measurements of the current session.
  history

  # Display the first 5 measurements of the current session.
  history --head 5

  # Display the last 10 measurements of the current session.
  history --tail 10`,
	}

	flags := historyCmd.Flags()
	flags.UintVar(&r.ctx.Head, "head", r.ctx.Head, "specify the number of measurements to display from the beginning of the history")
	flags.UintVar(&r.ctx.Tail, "tail", r.ctx.Tail, "specify the number of measurements to display from the end of the history")

	r.Cmd.AddCommand(historyCmd)
}

func (r *Root) RunHistory(cmd *cobra.Command, args []string) {
	var limit int = 0
	if r.ctx.Head > 0 {
		limit = int(r.ctx.Head)
	} else if r.ctx.Tail > 0 {
		limit = -int(r.ctx.Tail)
	}
	items, err := r.GetHistory(limit)
	if err != nil {
		r.printer.Println(err)
		return
	}
	if len(items) == 0 {
		r.printer.Println("No history items found")
		return
	}
	for _, item := range items {
		r.printer.Println(item)
	}
}

func (r *Root) UpdateHistory() error {
	ids := r.ctx.History.ToString(".")
	if ids == "" {
		return nil
	}
	_, err := os.Stat(getSessionPath())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err := os.Mkdir(getSessionPath(), 0755)
			if err != nil {
				return fmt.Errorf(saveToHistoryErr, err)
			}
		} else {
			return fmt.Errorf(saveToHistoryErr, err)
		}
	}
	index := "-"
	if !r.ctx.IsLocationFromSession {
		i, err := getIndex()
		if err != nil {
			return err
		}
		index = fmt.Sprintf("%d", i)
	}
	time := r.time.Now().Unix()
	cmd := strings.Join(os.Args[1:], " ")
	f, err := os.OpenFile(getHistoryPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf(saveToHistoryErr, err)
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("%s|%s|%d|%s|%s\n", HistoryItemVersion1, index, time, ids, cmd))
	if err != nil {
		return fmt.Errorf(saveToHistoryErr, err)
	}
	return nil
}

func (r *Root) GetHistory(limit int) ([]string, error) {
	items := make([]string, 0)
	f, err := os.Open(getHistoryPath())
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
			"> "+view.ShareURL+parts[3],
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

func getIndex() (int, error) {
	f, err := os.Open(getHistoryPath())
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
