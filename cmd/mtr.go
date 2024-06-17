package cmd

import (
	"fmt"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/spf13/cobra"
)

func (r *Root) initMTR() {
	mtrCmd := &cobra.Command{
		RunE:    r.RunMTR,
		Use:     "mtr [target] from [location | measurement ID | @1 | first | @-1 | last | previous]",
		GroupID: "Measurements",
		Short:   "Run a MTR test, which combines traceroute and ping",
		Long: `The MTR command combines the functionalities of traceroute and ping, providing real-time insights into the sent packets' routes. Use it to diagnose network issues such as packet loss, latency, and route instability.

Examples:
  # MTR google.com from 2 probes in New York.
  mtr google.com from New York --limit 2

  # MTR google.com using probes from a previous measurement by using its ID.
  mtr google.com from rvasVvKnj48cxNjC

  # MTR google.com using the same probes from the first measurement in this session.
  mtr google.com from @1

  # MTR google.com using the same probes from the last measurement in this session.
  mtr google.com from last

  # MTR google.com using the same probes from the second-to-last measurement in this session.
  mtr google.com from @-2

  # MTR 1.1.1.1 from 2 probes in the USA or Belgium. Send 10 packets and enable CI mode.
  mtr 1.1.1.1 from USA,Belgium --limit 2 --packets 10 --ci

  # MTR jsdelivr.com from a probe on the AWS network located in Montreal using the TCP protocol and port 453.
  mtr jsdelivr.com from aws+montreal --protocol tcp --port 453

  # MTR jsdelivr.com from a probe in ASN 123 and output the results in JSON format.
  mtr jsdelivr.com from 123 --json
  
  # MTR jsdelivr.com from a non-data center probe in Europe and add a link to view the results online.
  mtr jsdelivr.com from europe+eyeball --share `,
	}

	// mtr specific flags
	flags := mtrCmd.Flags()
	flags.StringVar(&r.ctx.Protocol, "protocol", r.ctx.Protocol, "specify the protocol to use for MTR: ICMP, TCP, or UDP (default \"icmp\")")
	flags.IntVar(&r.ctx.Port, "port", r.ctx.Port, "specify the port to use for MTR; only applicable for the TCP protocol (default 53)")
	flags.IntVar(&r.ctx.Packets, "packets", r.ctx.Packets, "specify the number of packets to send to each hop (default 3)")

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

	defer r.UpdateHistory()
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
	hm := &view.HistoryItem{
		Id:        res.ID,
		Status:    globalping.StatusInProgress,
		StartedAt: r.time.Now(),
	}
	r.ctx.History.Push(hm)
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
