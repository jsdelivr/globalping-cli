package cmd

import (
	"github.com/spf13/pflag"
	"golang.org/x/term"
	"os"
	"os/signal"
	"syscall"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/globalping/probe"
	"github.com/jsdelivr/globalping-cli/utils"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/spf13/cobra"
)

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

	cobra.AddTemplateFunc("wrappedFlagUsages", wrappedFlagUsages)
	root.Cmd.SetUsageTemplate(usageTemplate)
	root.Cmd.SetHelpTemplate(helpTemplate)

	// Global flags
	flags := root.Cmd.PersistentFlags()
	flags.StringVarP(&ctx.From, "from", "F", ctx.From, `Comma-separated list of location values to match against or a measurement ID
	For example, the partial or full name of a continent, region (e.g eastern europe), country, US state, city or network
	Or use [@1 | first, @2 ... @-2, @-1 | last | previous] to run with the probes from previous measurements.`)
	flags.IntVarP(&ctx.Limit, "limit", "L", ctx.Limit, "Limit the number of probes to use")
	flags.BoolVarP(&ctx.ToJSON, "json", "J", ctx.ToJSON, "Output results in JSON format (default false)")
	flags.BoolVarP(&ctx.CIMode, "ci", "C", ctx.CIMode, "Disable realtime terminal updates and color suitable for CI and scripting (default false)")
	flags.BoolVar(&ctx.ToLatency, "latency", ctx.ToLatency, "Output only the stats of a measurement (default false). Only applies to the dns, http and ping commands")
	flags.BoolVar(&ctx.Share, "share", ctx.Share, "Prints a link at the end the results, allowing to vizualize the results online (default false)")

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

// Uses the users terminal size or width of 80 if cannot determine users width
// Based on https://github.com/spf13/cobra/issues/1805#issuecomment-1246192724
func wrappedFlagUsages(cmd *pflag.FlagSet) string {
	fd := int(os.Stdout.Fd())
	width := 80

	// Get the terminal width and dynamically set
	termWidth, _, err := term.GetSize(fd)
	if err == nil {
		width = termWidth
	}

	return cmd.FlagUsagesWrapped(width - 1)
}

// Identical to the default cobra usage template,
// but utilizes wrappedFlagUsages to ensure flag usages don't wrap around
var usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{wrappedFlagUsages .LocalFlags | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{wrappedFlagUsages .InheritedFlags | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

var helpTemplate = `
{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`
