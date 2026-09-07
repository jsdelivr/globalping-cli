package cmd

import (
	"errors"

	"github.com/jsdelivr/globalping-cli/api"
	"github.com/jsdelivr/globalping-cli/utils"
	"github.com/jsdelivr/globalping-go"
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

func (r *Root) RunLimits(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	introspection, err := r.client.TokenIntrospection(ctx, "")
	username := ""

	if err != nil {
		var authorizeErr *api.AuthorizeError

		if !errors.As(err, &authorizeErr) || authorizeErr.ErrorType != api.ErrTypeNotAuthorized {
			r.printer.ErrPrintf("Warning: failed to retrieve authentication details: %s\n", err)
		}
	}

	if introspection != nil {
		username = introspection.Username
	}

	limits, err := r.client.Limits(ctx)

	if err != nil {
		r.Cmd.SilenceUsage = true

		return err
	}

	createLimit := utils.Pluralize(limits.RateLimits.Measurements.Create.Limit, "test")
	createConsumed := limits.RateLimits.Measurements.Create.Limit - limits.RateLimits.Measurements.Create.Remaining
	createRemaining := limits.RateLimits.Measurements.Create.Remaining
	t := limits.RateLimits.Measurements.Create.Type

	if t == globalping.CreateLimitTypeUser {
		if username == "" {
			r.printer.Printf("Authentication: token\n\n")
		} else {
			r.printer.Printf("Authentication: token (%s)\n\n", username)
		}
	} else {
		r.printer.Printf("Authentication: IP address\n\n")
	}

	r.printer.Printf(`Creating measurements:
 - %s per hour
 - %d consumed, %d remaining
`,
		createLimit,
		createConsumed,
		createRemaining,
	)

	if limits.RateLimits.Measurements.Create.Reset > 0 {
		createResets := utils.FormatSeconds(limits.RateLimits.Measurements.Create.Reset)
		r.printer.Printf(" - resets in %s\n", createResets)
	}

	if t == globalping.CreateLimitTypeUser {
		credits := utils.Pluralize(limits.Credits.Remaining, "credit")
		r.printer.Printf(`
Credits:
 - %s remaining (may be used to create measurements above the hourly limits)
`, credits)
	}

	return nil
}
