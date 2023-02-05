package cmd

import (
	"fmt"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/spf13/cobra"
)

// mtrCmd represents the mtr command
var mtrCmd = &cobra.Command{
	Use:   "mtr [target]",
	Short: "Implementation of the native mtr command",
	Long: `mtr combines the functionality of the traceroute and ping programs in a single network diagnostic tool.
	
		Examples:
		# MTR google.com from a probe in the network
		globalping mtr google.com --from "New York" --limit 2`,
	Args: requireTarget(),
	Run: func(cmd *cobra.Command, args []string) {
		// Make post struct
		opts = model.PostMeasurement{
			Type:   "mtr",
			Target: args[0],
			Locations: model.Locations{
				{
					Magic: from,
				},
			},
			Limit: limit,
			Options: &model.MeasurementOptions{
				Protocol: protocol,
				Port:     port,
				Packets:  packets,
			},
		}

		res, err := client.PostAPI(opts)
		if err != nil {
			fmt.Println(err)
			return
		}

		client.OutputResults(res.ID, ctx)
	},
}

func init() {
	rootCmd.AddCommand(mtrCmd)

	// mtr specific flags
	mtrCmd.Flags().StringVar(&protocol, "protocol", "", "Specifies the protocol used for tracerouting (ICMP, TCP or UDP). (default \"icmp\")")
	mtrCmd.Flags().IntVar(&port, "port", 0, "Specifies the port to use for the traceroute. Only applicable for TCP protocol. (default 53)")
	mtrCmd.Flags().IntVar(&packets, "packets", 0, "Specifies the number of packets to send to each hop. (default 3)")
}
