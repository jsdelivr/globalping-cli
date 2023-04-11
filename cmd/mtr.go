package cmd

import (
	"fmt"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/spf13/cobra"
)

// mtrCmd represents the mtr command
var mtrCmd = &cobra.Command{
	Use:     "mtr [target] from [location]",
	GroupID: "Measurements",
	Short:   "Run an MTR test, similar to traceroute",
	Long: `mtr combines the functionality of the traceroute and ping programs in a single network diagnostic tool.

Examples:
  # MTR google.com from 2 probes in New York
  mtr google.com from New York --limit 2

  # MTR 1.1.1.1 from 2 probes from USA or Belgium with 10 packets in CI mode
  mtr 1.1.1.1 from USA,Belgium --limit 2 --packets 10 --ci

  # MTR jsdelivr.com from a probe that is from the AWS network and is located in Montreal using the TCP protocol and port 453
  mtr jsdelivr.com from aws+montreal --protocol tcp --port 453

  # MTR jsdelivr.com from a probe in ASN 123 with json output
  mtr jsdelivr.com from 123 --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create context
		err := createContext(cmd.CalledAs(), args)
		if err != nil {
			return err
		}

		if ctx.Latency {
			return fmt.Errorf("the latency flag is not supported by the mtr command")
		}

		// Make post struct
		opts = model.PostMeasurement{
			Type:              "mtr",
			Target:            ctx.Target,
			Locations:         createLocations(ctx.From),
			Limit:             ctx.Limit,
			InProgressUpdates: inProgressUpdates(ctx.CI),
			Options: &model.MeasurementOptions{
				Protocol: protocol,
				Port:     port,
				Packets:  packets,
			},
		}

		res, showHelp, err := client.PostAPI(opts)
		if err != nil {
			if showHelp {
				return err
			}
			fmt.Println(err)
			return nil
		}

		view.OutputResults(res.ID, ctx)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mtrCmd)

	// mtr specific flags
	mtrCmd.Flags().StringVar(&protocol, "protocol", "", "Specifies the protocol used (ICMP, TCP or UDP) (default \"icmp\")")
	mtrCmd.Flags().IntVar(&port, "port", 0, "Specifies the port to use. Only applicable for TCP protocol (default 53)")
	mtrCmd.Flags().IntVar(&packets, "packets", 0, "Specifies the number of packets to send to each hop (default 3)")
}
