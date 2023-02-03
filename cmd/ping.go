package cmd

import (
	"fmt"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
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
			opts = model.PostMeasurement{
				Type:   "ping",
				Target: args[0],
				Locations: model.Locations{
					{
						Magic: from,
					},
				},
				Limit: limit,
				Options: &model.MeasurementOptions{
					Packets: packets,
				},
			}

			res, err := client.PostAPI(opts)
			if err != nil {
				fmt.Println(err)
				return
			}

			client.OutputResults(res.ID)
		},
	}

	opts    = model.PostMeasurement{}
	packets int
)

func init() {
	rootCmd.AddCommand(pingCmd)

	// ping specific flags
	pingCmd.Flags().IntVar(&packets, "packets", 3, "Number of packets to send")
}
