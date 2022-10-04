package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "globalping-cli",
		Short: "globalping-cli is a command line interface for the GlobalPing API",
		Long:  "globalping-cli is a command line interface for the GlobalPing API",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(cmd.UsageString())

			return nil
		},
	}

	cmd.AddCommand(newVersionCmd(version)) // version subcommand
	// cmd.AddCommand(newExampleCmd())        // example subcommand

	return cmd
}

// Execute invokes the command.
func Execute(version string) error {
	if err := newRootCmd(version).Execute(); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}

	return nil
}
