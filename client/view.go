package client

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/gosuri/uilive"
)

func OutputResults(id string) {
	// UI styles
	highlight := lipgloss.NewStyle().
		Bold(true).Foreground(lipgloss.Color("#17D4A7"))

	arrow := lipgloss.NewStyle().SetString(">").Bold(true).Foreground(lipgloss.Color("#17D4A7")).PaddingRight(1).String()

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
		for _, result := range data.Results {
			// Output slightly different format if state is available
			if result.Probe.State != "" {
				fmt.Fprintf(writer, arrow+highlight.Render("%s, %s, (%s), %s, ASN:%d")+"\n%s\n\n", result.Probe.Continent, result.Probe.Country, result.Probe.State, result.Probe.City, result.Probe.ASN, strings.TrimSpace(result.Result.RawOutput))
			} else {
				fmt.Fprintf(writer, arrow+highlight.Render("%s, %s, %s, ASN:%d")+"\n%s\n\n", result.Probe.Continent, result.Probe.Country, result.Probe.City, result.Probe.ASN, strings.TrimSpace(result.Result.RawOutput))
			}
		}

		if err != nil {
			fmt.Fprintf(writer, "%s", err)
			writer.Stop()
			return
		}
	}

	writer.Stop()
}
