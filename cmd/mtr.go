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
	Args: checkCommandFormat(),
	Run: func(cmd *cobra.Command, args []string) {
		// Create context
		createContext(args)

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
