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
	readHistoryErr        = "failed to read history: %s"
	invalidHistoryItemErr = "invalid history item: %s"
	saveToHistoryErr      = "failed to save to history: %s"
)

const (
	// <version>|<time>|<id>|<command>
	HistoryItemVersion1 int = 1
)

func (r *Root) initHistory() {
	historyCmd := &cobra.Command{
		Run:   r.RunHistory,
		Use:   "history",
		Short: "Show the history of your measurements",
		Long: `Show the history of your measurements
Examples:
  # Show the last measurements
  history

  # Show the last 10 measurements
  history --last 10
  
  # Show the first 5 measurements
  history --first 5`,
	}

	flags := historyCmd.Flags()
	flags.UintVar(&r.ctx.First, "first", r.ctx.First, "Number of first measurements to show")
	flags.UintVar(&r.ctx.Last, "last", r.ctx.Last, "Number of last measurements to show")

	r.Cmd.AddCommand(historyCmd)
}

func (r *Root) RunHistory(cmd *cobra.Command, args []string) {
	var limit int = 5 // default to last 5
	if r.ctx.First > 0 {
		limit = -int(r.ctx.First)
	} else if r.ctx.Last > 0 {
		limit = int(r.ctx.Last)
	}
	items, err := r.GetHistory(limit)
	if err != nil {
		r.printer.Println(err)
		return
	}
	if len(items) == 0 {
		r.printer.Println("No history found")
		return
	}
	for _, item := range items {
		r.printer.Println(item)
	}
}

func (r *Root) UpdateHistory() error {
	ids := r.ctx.History.ToString("+")
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
	f, err := os.OpenFile(getHistoryPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf(saveToHistoryErr, err)
	}
	defer f.Close()
	time := r.time.Now().Unix()
	cmd := strings.Join(os.Args[1:], " ")
	_, err = f.WriteString(fmt.Sprintf("%d|%d|%s|%s\n", HistoryItemVersion1, time, ids, cmd))
	if err != nil {
		return fmt.Errorf(saveToHistoryErr, err)
	}
	return nil
}

func (r *Root) GetHistory(limit int) ([]string, error) {
	f, err := os.Open(getHistoryPath())
	if err != nil {
		return nil, fmt.Errorf(readHistoryErr, err)
	}
	defer f.Close()
	items := make([]string, 0)
	// Read from the start of the file
	if limit < 0 {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			limit++
			item, err := parseHistoryItem(scanner.Text(), len(items))
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
	fStats, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf(readHistoryErr, err)
	}
	if fStats.Size() == 0 {
		return items, nil
	}
	scanner := backscanner.New(f, int(fStats.Size()-1)) // -1 to skip last newline
	for {
		limit--
		b, _, err := scanner.LineBytes()
		if err != nil {
			if err == io.EOF {
				return items, nil
			}
			return nil, fmt.Errorf(readHistoryErr, err)
		}
		item, err := parseHistoryItem(string(b), len(items))
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

func parseHistoryItem(line string, index int) (string, error) {
	parts := strings.Split(line, "|")
	if len(parts) < 1 {
		return "", fmt.Errorf(invalidHistoryItemErr, line)
	}
	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", fmt.Errorf(invalidHistoryItemErr, line)
	}
	switch version {
	case HistoryItemVersion1:
		if len(parts) != 4 {
			return "", fmt.Errorf(invalidHistoryItemErr, line)
		}
		t, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return "", fmt.Errorf(invalidHistoryItemErr, line)
		}
		return fmt.Sprintf(
			"%d | %s | %s\n%s",
			index+1,
			time.Unix(t, 0).Format("2006-01-02 15:04:05"),
			parts[3],
			"> "+view.ShareURL+parts[2],
		), nil
	default:
		return "", fmt.Errorf("invalid history item version: %d", version)
	}
}
