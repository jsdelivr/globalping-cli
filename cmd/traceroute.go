package cmd

import (
	"fmt"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/spf13/cobra"
)

// tracerouteCmd represents the traceroute command
var tracerouteCmd = &cobra.Command{
	Use:     "traceroute [target] from [location]",
	GroupID: "Measurements",
	Short:   "Use the native traceroute command",
	Long: `traceroute tracks the route packets taken from an IP network on their way to a given host. It utilizes the IP protocol's time to live (TTL) field and attempts to elicit an ICMP TIME_EXCEEDED response from each gateway along the path to the host.

Examples:
  # Traceroute google.com from 2 probes in New York
  traceroute google.com from New York --limit 2

  # Traceroute 1.1.1.1 from 2 probes from North America or Belgium
  traceroute 1.1.1.1 from North America,Belgium --limit 2

  # Traceroute jsdelivr.com from a probe that is from the AWS network and is located in Montreal using the UDP protocol
  traceroute jsdelivr.com from aws+montreal --protocol udp

  # Traceroute jsdelivr.com with ASN 12345 with json output
  traceroute jsdelivr.com from 12345 --json`,
	Args: checkCommandFormat(),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create context
		err := createContext(cmd.CalledAs(), args)
		if err != nil {
			return err
		}

		// Make post struct
		opts = model.PostMeasurement{
			Type:              "traceroute",
			Target:            ctx.Target,
			Locations:         createLocations(ctx.From),
			Limit:             ctx.Limit,
			InProgressUpdates: inProgressUpdates(ctx.CI),
			Options: &model.MeasurementOptions{
				Protocol: protocol,
				Port:     port,
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
	rootCmd.AddCommand(tracerouteCmd)

	// traceroute specific flags
	tracerouteCmd.Flags().StringVar(&protocol, "protocol", "", "Specifies the protocol used for tracerouting (ICMP, TCP or UDP) (default \"icmp\")")
	tracerouteCmd.Flags().IntVar(&port, "port", 0, "Specifies the port to use for the traceroute. Only applicable for TCP protocol (default 80)")
}
