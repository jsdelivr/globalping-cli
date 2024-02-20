package cmd

import (
	"fmt"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/spf13/cobra"
)

func (r *Root) initMTR() {
	mtrCmd := &cobra.Command{
		RunE:    r.RunMTR,
		Use:     "mtr [target] from [location | measurement ID | @1 | first | @-1 | last | previous]",
		GroupID: "Measurements",
		Short:   "Run an MTR test, similar to traceroute",
		Long: `mtr combines the functionality of the traceroute and ping programs in a single network diagnostic tool.

Examples:
  # MTR google.com from 2 probes in New York
  mtr google.com from New York --limit 2

  # MTR google.com using probes from previous measurement
  mtr google.com from rvasVvKnj48cxNjC

  # MTR google.com using probes from first measurement in session
  mtr google.com from @1

  # MTR google.com using probes from last measurement in session
  mtr google.com from last

  # MTR google.com using probes from second to last measurement in session
  mtr google.com from @-2

  # MTR 1.1.1.1 from 2 probes from USA or Belgium with 10 packets in CI mode
  mtr 1.1.1.1 from USA,Belgium --limit 2 --packets 10 --ci

  # MTR jsdelivr.com from a probe that is from the AWS network and is located in Montreal using the TCP protocol and port 453
  mtr jsdelivr.com from aws+montreal --protocol tcp --port 453

  # MTR jsdelivr.com from a probe in ASN 123 with json output
  mtr jsdelivr.com from 123 --json`,
	}

	// mtr specific flags
	flags := mtrCmd.Flags()
	flags.StringVar(&r.ctx.Protocol, "protocol", "", "Specifies the protocol used (ICMP, TCP or UDP) (default \"icmp\")")
	flags.IntVar(&r.ctx.Port, "port", 0, "Specifies the port to use. Only applicable for TCP protocol (default 53)")
	flags.IntVar(&r.ctx.Packets, "packets", 0, "Specifies the number of packets to send to each hop (default 3)")

	r.Cmd.AddCommand(mtrCmd)
}

func (r *Root) RunMTR(cmd *cobra.Command, args []string) error {
	err := r.updateContext(cmd.CalledAs(), args)
	if err != nil {
		return err
	}

	if r.ctx.ToLatency {
		return fmt.Errorf("the latency flag is not supported by the mtr command")
	}

	r.ctx.RecordToSession = true

	opts := &globalping.MeasurementCreate{
		Type:              "mtr",
		Target:            r.ctx.Target,
		Limit:             r.ctx.Limit,
		InProgressUpdates: !r.ctx.CIMode,
		Options: &globalping.MeasurementOptions{
			Protocol: r.ctx.Protocol,
			Port:     r.ctx.Port,
			Packets:  r.ctx.Packets,
		},
	}
	opts.Locations, err = r.getLocations()
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

	r.ctx.MeasurementsCreated++

	if r.ctx.RecordToSession {
		r.ctx.RecordToSession = false
		err := saveIdToSession(res.ID)
		if err != nil {
			r.printer.Printf("Warning: %s\n", err)
		}
	}

	r.viewer.Output(res.ID, opts)
	return nil
}
