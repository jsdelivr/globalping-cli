package cmd

import (
	"fmt"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/spf13/cobra"
)

// mtrCmd represents the mtr command
var mtrCmd = &cobra.Command{
	Use:     "mtr [target] from [location]",
	GroupID: "Measurements",
	Short:   "Use the native mtr command",
	Long: `mtr combines the functionality of the traceroute and ping programs in a single network diagnostic tool.
	
Examples:
  # MTR google.com from 2 probes in New York
  mtr google.com from New York --limit 2

  # MTR 1.1.1.1 from 2 probes from North America or Belgium with 10 packets
  mtr 1.1.1.1 from North America,Belgium --limit 2 --packets 10

  # MTR jsdelivr.com from a probe that is from the AWS network and is located in Montreal using the TCP protocol
  mtr jsdelivr.com from aws+montreal --protocol tcp

  # MTR jsdelivr.com with ASN 12345 with json output
  mtr jsdelivr.com from 12345 --json`,
	Args: checkCommandFormat(),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create context
		err := createContext(cmd.CalledAs(), args)
		if err != nil {
			return err
		}

		// Make post struct
		opts = model.PostMeasurement{
			Type:      "mtr",
			Target:    ctx.Target,
			Locations: createLocations(ctx.From),
			Limit:     ctx.Limit,
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

		client.OutputResults(res.ID, ctx)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mtrCmd)

	// mtr specific flags
	mtrCmd.Flags().StringVar(&protocol, "protocol", "", "Specifies the protocol used for tracerouting (ICMP, TCP or UDP) (default \"icmp\")")
	mtrCmd.Flags().IntVar(&port, "port", 0, "Specifies the port to use for the traceroute. Only applicable for TCP protocol (default 53)")
	mtrCmd.Flags().IntVar(&packets, "packets", 0, "Specifies the number of packets to send to each hop (default 3)")

	// Extra flags
	// mtrCmd.Flags().BoolVar(&ctx.Latency, "latency", false, "Output only stats of a measurement (default false)")
}
