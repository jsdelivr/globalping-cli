package cmd

import (
	"fmt"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/spf13/cobra"
)

// mtrCmd represents the mtr command
var mtrCmd = &cobra.Command{
	Use:   "mtr [target] from [location]",
	Short: "Implementation of the native mtr command",
	Long: `mtr combines the functionality of the traceroute and ping programs in a single network diagnostic tool.
	
Examples:
# MTR google.com from 2 probes in New York
mtr google.com --from "New York" --limit 2`,
	Args: checkCommandFormat(),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create context
		err := createContext(args)
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
}
