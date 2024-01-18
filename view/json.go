package view

import (
	"fmt"
	"os"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
)

// Outputs the raw JSON for a measurement
func OutputJson(id string, fetcher client.MeasurementsFetcher, ctx model.Context) error {
	output, err := fetcher.GetRawMeasurement(id)
	if err != nil {
		return err
	}
	fmt.Println(string(output))

	if ctx.Share {
		fmt.Fprintln(os.Stderr, formatWithLeadingArrow(shareMessage(id), !ctx.CI))
	}
	fmt.Println()

	return nil
}
