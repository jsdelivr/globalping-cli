package cmd

import (
	"fmt"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/spf13/cobra"
)

func (r *Root) initTraceroute() {
	var tracerouteCmd = &cobra.Command{
		RunE:    r.RunTraceroute,
		Use:     "traceroute [target] from [location | measurement ID | @1 | first | @-1 | last | previous]",
		GroupID: "Measurements",
		Short:   "Run a traceroute test",
		Long: `The traceroute command traces the path packets take to reach a target, displaying each hop along the way, including its round-trip time. Use it to troubleshoot network connectivity issues and identify latency problems.

Examples:
  # Traceroute google.com from 2 probes in New York.
  traceroute google.com from New York --limit 2

  # Traceroute google.com using probes from a previous measurement by using its ID.
  traceroute google.com from rvasVvKnj48cxNjC

  # Traceroute google.com using the same probes from the first measurement in this session.
  traceroute google.com from @1

  # Traceroute google.com using the same probes from the last measurement in this session.
  traceroute google.com from last

  # Traceroute google.com using the same probes from the second-to-last measurement in this session.
  traceroute google.com from @-2

  # Traceroute 1.1.1.1 from 2 probes in the USA or Belgium and enable CI mode.
  traceroute 1.1.1.1 from USA,Belgium --limit 2 --ci

  # Traceroute jsdelivr.com from a probe on the AWS network located in Montreal using the UDP protocol.
  traceroute jsdelivr.com from aws+montreal --protocol udp

  # Traceroute jsdelivr.com from a probe in Paris using port 453.
  traceroute jsdelivr.com from Paris --port 453

  # Traceroute jsdelivr.com from a probe in ASN 123 and output the results in JSON format.
  traceroute jsdelivr.com from 123 --json
  
  # Traceroute jsdelivr.com from a non-data center probe in Europe and add a link to view the results online.
  traceroute jsdelivr.com from europe+eyeball --share`,
	}

	// traceroute specific flags
	flags := tracerouteCmd.Flags()
	flags.StringVar(&r.ctx.Protocol, "protocol", r.ctx.Protocol, "specify the protocol to use for tracerouting: ICMP, TCP, or UDP (default \"icmp\")")
	flags.IntVar(&r.ctx.Port, "port", r.ctx.Port, "specify the port to use for the traceroute; only applicable for the TCP protocol (default 80)")

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

	defer r.UpdateHistory()
	r.ctx.RecordToSession = true

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
