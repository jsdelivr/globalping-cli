package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/globalping/probe"
	"github.com/jsdelivr/globalping-cli/utils"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/spf13/cobra"
)

var fromShortDesc = `Comma-separated list of location values to match against or a measurement ID
  For example, the partial or full name of a continent, region (e.g eastern europe), country, US state, city or network
  Or use [@1 | first, @2 ... @-2, @-1 | last | previous] to run with the probes from previous measurements.`
var limitShortDesc = `Limit the number of probes to use`
var jsonShortDesc = `Output results in JSON format (default false)`
var ciModeShortDesc = `Disable realtime terminal updates and color suitable for CI and scripting (default false)`
var latencyShortDesc = `Output only the stats of a measurement (default false)`
var shareShortDesc = `Prints a link at the end the results, allowing to vizualize the results online (default false)`

type Root struct {
	printer *view.Printer
	ctx     *view.Context
	viewer  view.Viewer
	client  globalping.Client
	probe   probe.Probe
	time    utils.Time
	Cmd     *cobra.Command
	cancel  chan os.Signal
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	utime := utils.NewTime()
	printer := view.NewPrinter(os.Stdin, os.Stdout, os.Stderr)
	ctx := &view.Context{
		APIMinInterval: globalping.API_MIN_INTERVAL,
		History:        view.NewHistoryBuffer(10),
		From:           "world",
		Limit:          1,
	}
	globalpingClient := globalping.NewClient(globalping.API_URL)
	globalpingProbe := probe.NewProbe()
	viewer := view.NewViewer(ctx, printer, utime, globalpingClient)
	root := NewRoot(printer, ctx, viewer, utime, globalpingClient, globalpingProbe)

	err := root.Cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
	root.UpdateHistory()
}

func NewRoot(
	printer *view.Printer,
	ctx *view.Context,
	viewer view.Viewer,
	time utils.Time,
	globalpingClient globalping.Client,
	globalpingProbe probe.Probe,
) *Root {
	root := &Root{
		printer: printer,
		ctx:     ctx,
		viewer:  viewer,
		time:    time,
		client:  globalpingClient,
		probe:   globalpingProbe,
		cancel:  make(chan os.Signal, 1),
	}

	signal.Notify(root.cancel, syscall.SIGINT, syscall.SIGTERM)

	// rootCmd represents the base command when called without any subcommands
	root.Cmd = &cobra.Command{
		Use:   "globalping",
		Short: "A global network of probes to run network tests like ping, traceroute and DNS resolve.",
		Long: `Globalping is a platform that allows anyone to run networking commands such as ping, traceroute, dig and mtr on probes distributed all around the world.
The CLI tool allows you to interact with the API in a simple and human-friendly way to debug networking issues like anycast routing and script automated tests and benchmarks.`,
	}

	root.Cmd.SetOut(printer.OutWriter)
	root.Cmd.SetErr(printer.ErrWriter)

	root.Cmd.AddGroup(&cobra.Group{ID: "Measurements", Title: "Measurement Commands:"})

	root.initDNS()
	root.initHTTP()
	root.initMTR()
	root.initPing()
	root.initTraceroute()
	root.initInstallProbe()
	root.initVersion()
	root.initHistory()

	return root
}
