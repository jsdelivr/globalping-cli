package client

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/gosuri/uilive"
)

func OutputResults(id string) {
	// UI styles
	highlight := lipgloss.NewStyle().
		Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	// Create new writer
	writer := uilive.New()
	writer.Start()

	// Get results
	data, err := GetAPI(id)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Probe may not have started yet
	for len(data.Results) == 0 {
		fmt.Fprint(writer, highlight.Render("Pending..."))
		time.Sleep(100 * time.Millisecond)
		data, err = GetAPI(id)
		if err != nil {
			fmt.Fprintf(writer, "%s", err)
			writer.Stop()
			return
		}
	}

	// Poll API every 100 milliseconds until the measurement is complete
	for data.Status == "in-progress" {
		time.Sleep(100 * time.Millisecond)
		data, err = GetAPI(id)
		// Output every result in case of multiple probes
		var i int
		for _, result := range data.Results {
			i++
			fmt.Fprintf(writer, highlight.Render("Probe %d:\n"), i)
			fmt.Fprintf(writer, "%s\n", result.Result.RawOutput)
		}

		if err != nil {
			fmt.Fprintf(writer, "%s", err)
			writer.Stop()
			return
		}
	}

	writer.Stop()
}
