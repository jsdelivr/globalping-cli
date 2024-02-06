package cmd

import (
	"fmt"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/spf13/cobra"
)

// tracerouteCmd represents the traceroute command
var tracerouteCmd = &cobra.Command{
	Use:     "traceroute [target] from [location | measurement ID | @1 | first | @-1 | last | previous]",
	GroupID: "Measurements",
	Short:   "Run a traceroute test",
	Long: `traceroute tracks the route packets take from an IP network on their way to a given host.

Examples:
  # Traceroute google.com from 2 probes in New York
  traceroute google.com from New York --limit 2

  # Traceroute google.com using probes from previous measurement
  traceroute google.com from rvasVvKnj48cxNjC

  # Traceroute google.com using probes from first measurement in session
  traceroute google.com from @1

  # Traceroute google.com using probes from last measurement in session
  traceroute google.com from last

  # Traceroute google.com using probes from second to last measurement in session
  traceroute google.com from @-2

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

		if ctx.ToLatency {
			return fmt.Errorf("the latency flag is not supported by the traceroute command")
		}

		// Make post struct
		opts = globalping.MeasurementCreate{
			Type:              "traceroute",
			Target:            ctx.Target,
			Limit:             ctx.Limit,
			InProgressUpdates: inProgressUpdates(ctx.CI),
			Options: &globalping.MeasurementOptions{
				Protocol: protocol,
				Port:     port,
			},
		}
		isPreviousMeasurementId := false
		opts.Locations, isPreviousMeasurementId, err = createLocations(ctx.From)
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		res, showHelp, err := gp.CreateMeasurement(&opts)
		if err != nil {
			if !showHelp {
				cmd.SilenceUsage = true
			}
			return err
		}

		// Save measurement ID to history
		if !isPreviousMeasurementId {
			err := saveMeasurementID(res.ID)
			if err != nil {
				fmt.Printf("Warning: %s\n", err)
			}
		}

		viewer.Output(res.ID, &opts)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tracerouteCmd)

	// traceroute specific flags
	tracerouteCmd.Flags().StringVar(&protocol, "protocol", "", "Specifies the protocol used for tracerouting (ICMP, TCP or UDP) (default \"icmp\")")
	tracerouteCmd.Flags().IntVar(&port, "port", 0, "Specifies the port to use for the traceroute. Only applicable for TCP protocol (default 80)")
}
