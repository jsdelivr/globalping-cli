package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/spf13/cobra"
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
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
	RunE: func(cmd *cobra.Command, args []string) error {
		err := createContext(cmd.CalledAs(), args)
		if err != nil {
			return err
		}
		if ctx.Infinite {
			return infinitePing(cmd)
		}
		_, err = ping(cmd)
		return err
	},
}

func infinitePing(cmd *cobra.Command) error {
	var err error
	if ctx.Limit > 5 {
		return fmt.Errorf("continous mode is currently limited to 5 probes")
	}
	ctx.Packets = 16 // Default to 16 packets

	// Trap sigterm or interupt to display info on exit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			ctx.From, err = ping(cmd)
			if err != nil {
				sig <- syscall.SIGINT
				return
			}
		}
	}()

	<-sig
	if err == nil {
		view.OutputSummary(&ctx)
	}
	return err
}

func ping(cmd *cobra.Command) (string, error) {
	opts = model.PostMeasurement{
		Type:              "ping",
		Target:            ctx.Target,
		Limit:             ctx.Limit,
		InProgressUpdates: inProgressUpdates(ctx.CI),
		Options: &model.MeasurementOptions{
			Packets: ctx.Packets,
		},
	}
	var err error
	isPreviousMeasurementId := true
	if ctx.CallCount == 0 {
		opts.Locations, isPreviousMeasurementId, err = createLocations(ctx.From)
		if err != nil {
			cmd.SilenceUsage = true
			return "", err
		}
	} else {
		opts.Locations = []model.Locations{{Magic: ctx.From}}
	}

	res, showHelp, err := client.PostAPI(opts)
	if err != nil {
		if !showHelp {
			cmd.SilenceUsage = true
		}
		return "", err
	}

	ctx.CallCount++

	// Save measurement ID to history
	if !isPreviousMeasurementId {
		err := saveMeasurementID(res.ID)
		if err != nil {
			fmt.Printf("Warning: %s\n", err)
		}
	}

	if ctx.Infinite {
		err = view.OutputInfinite(res.ID, &ctx)
	} else {
		view.OutputResults(res.ID, ctx, opts)
	}
	return res.ID, err
}

func init() {
	rootCmd.AddCommand(pingCmd)

	// ping specific flags
	pingCmd.Flags().IntVar(&ctx.Packets, "packets", 0, "Specifies the desired amount of ECHO_REQUEST packets to be sent (default 3)")
	pingCmd.Flags().BoolVar(&ctx.Infinite, "infinite", false, "Keep pinging the target continuously until stopped (default false)")
}
