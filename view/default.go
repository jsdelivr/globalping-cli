package view

import (
	"fmt"
	"os"
	"strings"

	"github.com/jsdelivr/globalping-cli/model"
)

// Outputs non-json non-latency results for a measurement
func OutputDefault(id string, data *model.GetMeasurement, ctx model.Context, m model.PostMeasurement) {
	for i := range data.Results {
		result := &data.Results[i]
		if i > 0 {
			// new line as separator if more than 1 result
			fmt.Println()
		}

		// Output slightly different format if state is available
		fmt.Fprintln(os.Stderr, generateHeader(result, !ctx.CI))

		if isBodyOnlyHttpGet(ctx, m) {
			fmt.Println(strings.TrimSpace(result.Result.RawBody))
		} else {
			fmt.Println(strings.TrimSpace(result.Result.RawOutput))
		}
	}

	if ctx.Share {
		fmt.Fprintln(os.Stderr, formatWithLeadingArrow(shareMessage(id), !ctx.CI))
	}
}
