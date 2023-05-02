package cmd

import (
	"fmt"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/spf13/cobra"
)

// tracerouteCmd represents the traceroute command
var tracerouteCmd = &cobra.Command{
	Use:     "traceroute [target] from [location]",
	GroupID: "Measurements",
	Short:   "Run a traceroute test",
	Long: `traceroute tracks the route packets take from an IP network on their way to a given host.

Examples:
  # Traceroute google.com from 2 probes in New York
  traceroute google.com from New York --limit 2

  # Traceroute 1.1.1.1 from 2 probes from USA or Belgium in CI mode
  traceroute 1.1.1.1 from USA,Belgium --limit 2 --ci

  # Traceroute jsdelivr.com from a probe that is from the AWS network and is located in Montreal using the UDP protocol
  traceroute jsdelivr.com from aws+montreal --protocol udp

  # Traceroute jsdelivr.com from a probe that is located in Paris to port 453
  traceroute jsdelivr.com from Paris --port 453

  # Traceroute jsdelivr.com from a probe in ASN 123 with json output
  traceroute jsdelivr.com from 123 --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create context
		err := createContext(cmd.CalledAs(), args)
		if err != nil {
			return err
		}

		if ctx.Latency {
			return fmt.Errorf("the latency flag is not supported by the traceroute command")
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

		view.OutputResults(res.ID, ctx, opts)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tracerouteCmd)

	// traceroute specific flags
	tracerouteCmd.Flags().StringVar(&protocol, "protocol", "", "Specifies the protocol used for tracerouting (ICMP, TCP or UDP) (default \"icmp\")")
	tracerouteCmd.Flags().IntVar(&port, "port", 0, "Specifies the port to use for the traceroute. Only applicable for TCP protocol (default 80)")
}
