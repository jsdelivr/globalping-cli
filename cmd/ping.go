package cmd

import (
	"fmt"

	"github.com/jsdelivr/globalping-cli/api"
	"github.com/spf13/cobra"
)

// pingCmd represents the ping command
var (
	pingCmd = &cobra.Command{
		Use:   "ping [target]",
		Short: "Use ping command",
		Long: `Use ping command from a probe in the Globalping network.
	
	Examples:
	# Ping google.com from a probe in the network
	globalping ping google.com --from "New York" --limit 2`,
		Args: requireTarget(),
		Run: func(cmd *cobra.Command, args []string) {
			// Make post struct
			opts = api.PostMeasurement{
				Type:   "ping",
				Target: args[0],
				Locations: api.Locations{
					{
						Magic: from,
					},
				},
				Limit: limit,
				Options: &api.MeasurementOptions{
					Packets: packets,
				},
			}

			res, err := api.PostAPI(opts)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println(res)
		},
	}

	opts    = api.PostMeasurement{}
	packets int
)

func init() {
	rootCmd.AddCommand(pingCmd)

	// ping specific flags
	pingCmd.Flags().IntVar(&packets, "packets", 3, "Number of packets to send")
}
