package cmd

import (
	"fmt"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/spf13/cobra"
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping [target] from [location]",
	Short: "Use ping command",
	Long: `The ping command sends an ICMP ECHO_REQUEST to obtain an ICMP ECHO_RESPONSE from a host or gateway.
	
	Examples:
	# Ping google.com from a probe in the network
	globalping ping google.com --from "New York" --limit 2`,
	Args: checkCommandFormat(),
	Run: func(cmd *cobra.Command, args []string) {
		// Create context
		createContext(args)

		// Make post struct
		opts = model.PostMeasurement{
			Type:      "ping",
			Target:    ctx.Target,
			Locations: createLocations(ctx.From),
			Limit:     ctx.Limit,
			Options: &model.MeasurementOptions{
				Packets: packets,
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
	rootCmd.AddCommand(pingCmd)

	// ping specific flags
	pingCmd.Flags().IntVar(&packets, "packets", 0, "Specifies the desired amount of ECHO_REQUEST packets to be sent (default 3)")
}
