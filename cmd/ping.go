package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/spf13/cobra"
)

func (r *Root) initPing() {
	pingCmd := &cobra.Command{
		Use:     "ping [target] from [location | measurement ID | @1 | first | @-1 | last | previous]",
		GroupID: "Measurements",
		Short:   "Run a ping test",
		Long: `The ping command allows sending ping requests to a target. Often used to test the network latency and stability.

Examples:
  # Ping google.com from 2 probes in New York
  ping google.com from New York --limit 2

  # Ping google.com using probes from previous measurement
  ping google.com from rvasVvKnj48cxNjC

  # Ping google.com using probes from first measurement in session
  ping google.com from @1

  # Ping google.com using probes from last measurement in session
  ping google.com from last

  # Ping google.com using probes from second to last measurement in session
  ping google.com from @-2

  # Ping 1.1.1.1 from 2 probes from USA or Belgium with 10 packets in CI mode
  ping 1.1.1.1 from USA,Belgium --limit 2 --packets 10 --ci

  # Ping jsdelivr.com from a probe that is from the AWS network and is located in Montreal with latency output
  ping jsdelivr.com from aws+montreal --latency

  # Ping jsdelivr.com from a probe in ASN 123 with json output
  ping jsdelivr.com from 123 --json

  # Continuously ping google.com from New York
  ping google.com from New York --infinite`,
		RunE: r.RunPing,
	}

	// ping specific flags
	flags := pingCmd.Flags()
	flags.IntVar(&r.ctx.Packets, "packets", 0, "Specifies the desired amount of ECHO_REQUEST packets to be sent (default 3)")
	flags.BoolVar(&r.ctx.Infinite, "infinite", false, "Keep pinging the target continuously until stopped (default false)")

	r.Cmd.AddCommand(pingCmd)
}

func (r *Root) RunPing(cmd *cobra.Command, args []string) error {
	err := r.updateContext(cmd.CalledAs(), args)
	if err != nil {
		return err
	}
	if r.ctx.Infinite {
		return r.pingInfinite()
	}
	_, err = r.ping()
	return err
}

func (r *Root) ping() (string, error) {
	opts := &globalping.MeasurementCreate{
		Type:              "ping",
		Target:            r.ctx.Target,
		Limit:             r.ctx.Limit,
		InProgressUpdates: inProgressUpdates(r.ctx.CI),
		Options: &globalping.MeasurementOptions{
			Packets: r.ctx.Packets,
		},
	}
	var err error
	isPreviousMeasurementId := true
	if r.ctx.CallCount == 0 {
		opts.Locations, isPreviousMeasurementId, err = createLocations(r.ctx.From)
		if err != nil {
			r.Cmd.SilenceUsage = true
			return "", err
		}
	} else {
		opts.Locations = []globalping.Locations{{Magic: r.ctx.From}}
	}

	res, showHelp, err := r.gp.CreateMeasurement(opts)
	if err != nil {
		if !showHelp {
			r.Cmd.SilenceUsage = true
		}
		return "", err
	}

	r.ctx.CallCount++

	// Save measurement ID to history
	if !isPreviousMeasurementId {
		err := saveMeasurementID(res.ID)
		if err != nil {
			r.printer.Printf("Warning: %s\n", err)
		}
	}
	if r.ctx.Infinite {
		err = r.viewer.OutputInfinite(res.ID)
		r.Cmd.SilenceUsage = true
	} else {
		r.viewer.Output(res.ID, opts)
	}
	return res.ID, err
}

func (r *Root) pingInfinite() error {
	var err error
	if r.ctx.Limit > 5 {
		return fmt.Errorf("continous mode is currently limited to 5 probes")
	}
	r.ctx.Packets = 16 // Default to 16 packets

	// Trap sigterm or interupt to display info on exit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			r.ctx.From, err = r.ping()
			if err != nil {
				sig <- syscall.SIGINT
				return
			}
		}
	}()

	<-sig
	if err == nil {
		r.viewer.OutputSummary()
	}
	return err
}
