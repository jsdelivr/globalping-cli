package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
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
	items, err := r.storage.GetHistory(limit)
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
	index := "-"
	if !r.ctx.IsLocationFromSession {
		i, err := r.storage.GetHistoryIndex()
		if err != nil {
			return err
		}
		index = fmt.Sprintf("%d", i)
	}
	err := r.storage.SaveCommandToHistory(
		index,
		r.utils.Now().Unix(),
		ids,
		strings.Join(os.Args[1:], " "),
	)
	if err != nil {
		return fmt.Errorf("failed to save command to history: %s", err)
	}
	err = r.storage.Cleanup()
	if err != nil {
		return fmt.Errorf("failed to cleanup history: %s", err)
	}
	return nil
}
