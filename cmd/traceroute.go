package cmd

import (
	"fmt"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/spf13/cobra"
)

func (r *Root) initTraceroute() {
	var tracerouteCmd = &cobra.Command{
		RunE:    r.RunTraceroute,
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
	}

	// traceroute specific flags
	flags := tracerouteCmd.Flags()
	flags.StringVar(&r.ctx.Protocol, "protocol", "", "Specifies the protocol used for tracerouting (ICMP, TCP or UDP) (default \"icmp\")")
	flags.IntVar(&r.ctx.Port, "port", 0, "Specifies the port to use for the traceroute. Only applicable for TCP protocol (default 80)")

	r.Cmd.AddCommand(tracerouteCmd)
}

func (r *Root) RunTraceroute(cmd *cobra.Command, args []string) error {
	err := r.updateContext(cmd.CalledAs(), args)
	if err != nil {
		return err
	}

	if r.ctx.ToLatency {
		return fmt.Errorf("the latency flag is not supported by the traceroute command")
	}

	opts := &globalping.MeasurementCreate{
		Type:              "traceroute",
		Target:            r.ctx.Target,
		Limit:             r.ctx.Limit,
		InProgressUpdates: !r.ctx.CIMode,
		Options: &globalping.MeasurementOptions{
			Protocol: r.ctx.Protocol,
			Port:     r.ctx.Port,
		},
	}
	isPreviousMeasurementId := false
	opts.Locations, isPreviousMeasurementId, err = createLocations(r.ctx.From)
	if err != nil {
		cmd.SilenceUsage = true
		return err
	}

	res, showHelp, err := r.client.CreateMeasurement(opts)
	if err != nil {
		if !showHelp {
			cmd.SilenceUsage = true
		}
		return err
	}

	// Save measurement ID to history
	if !isPreviousMeasurementId {
		err := saveIdToHistory(res.ID)
		if err != nil {
			r.printer.Printf("Warning: %s\n", err)
		}
	}

	r.viewer.Output(res.ID, opts)
	return nil
}
