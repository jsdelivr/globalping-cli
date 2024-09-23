package cmd

import (
	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/utils"
	"github.com/spf13/cobra"
)

func (r *Root) initLimits() {
	limitsCmd := &cobra.Command{
		Use:   "limits",
		Short: "Show the current rate limits",
		Long:  `Show the current rate limits.`,
		RunE:  r.RunLimits,
	}

	r.Cmd.AddCommand(limitsCmd)
}

func (r *Root) RunLimits(cmd *cobra.Command, args []string) error {
	introspection, _ := r.client.TokenIntrospection("")
	username := ""
	if introspection != nil {
		username = introspection.Username
	}
	limits, err := r.client.Limits()
	if err != nil {
		return err
	}
	createLimit := utils.Pluralize(limits.RateLimits.Measurements.Create.Limit, "test")
	createConsumed := limits.RateLimits.Measurements.Create.Limit - limits.RateLimits.Measurements.Create.Remaining
	createRemaining := limits.RateLimits.Measurements.Create.Remaining
	createResets := utils.FormatSeconds(limits.RateLimits.Measurements.Create.Reset)
	t := limits.RateLimits.Measurements.Create.Type
	if t == globalping.CreateLimitTypeUser {
		r.printer.Printf("Authentication: token (%s)\n\n", username)
	} else {
		r.printer.Printf("Authentication: IP address\n\n")
	}
	r.printer.Printf(`Creating measurements: 
 - %s per hour
 - %d consumed, %d remaining
 - resets in %s
`,
		createLimit,
		createConsumed,
		createRemaining,
		createResets,
	)
	if t == globalping.CreateLimitTypeUser {
		credits := utils.Pluralize(limits.Credits.Remaining, "credit")
		r.printer.Printf(`
Credits:
 - %s remaining (may be used to create measurements above the hourly limits)
`, credits)
	}
	return nil
}
