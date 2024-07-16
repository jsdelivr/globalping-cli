package cmd

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"
	"golang.org/x/term"

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
	config := utils.NewConfig()
	config.Load()
	ctx := &view.Context{
		APIMinInterval: config.GlobalpingAPIInterval,
		History:        view.NewHistoryBuffer(10),
		From:           "world",
		Limit:          1,
	}
	t := time.NewTicker(10 * time.Second)
	globalpingClient := globalping.NewClientWithCacheCleanup(globalping.Config{
		APIURL:   config.GlobalpingAPIURL,
		APIToken: config.GlobalpingToken,
	}, t, 30)
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
		Short: "A global network of probes to perform network tests such as ping, traceroute, and DNS resolution",
		Long: `The Globalping platform allows anyone to run networking commands such as ping, traceroute, dig, and mtr on probes distributed around the globe.
For more information about the platform, tips, and best practices, visit our GitHub repository at https://github.com/jsdelivr/globalping.`,
	}

	root.Cmd.SetOut(printer.OutWriter)
	root.Cmd.SetErr(printer.ErrWriter)

	cobra.AddTemplateFunc("wrappedFlagUsages", wrappedFlagUsages)
	root.Cmd.SetUsageTemplate(usageTemplate)
	root.Cmd.SetHelpTemplate(helpTemplate)

	// Global flags
	flags := root.Cmd.PersistentFlags()
	flags.StringVarP(&ctx.From, "from", "F", ctx.From, `specify the probe locations as a comma-separated list; you may use:
 - names of continents, regions, countries, US states, cities, or networks
 - [@1 | first, @2 ... @-2, @-1 | last | previous] to run with the probes from previous measurements in this session
 - an ID of a previous measurement to run with its probes
`)
	flags.IntVarP(&ctx.Limit, "limit", "L", ctx.Limit, "define the number of probes to use")
	flags.BoolVarP(&ctx.ToJSON, "json", "J", ctx.ToJSON, "output results in JSON format (default false)")
	flags.BoolVarP(&ctx.CIMode, "ci", "C", ctx.CIMode, "disable real-time terminal updates and colors, suitable for CI and scripting (default false)")
	flags.BoolVar(&ctx.ToLatency, "latency", ctx.ToLatency, "output only the latency stats; applicable only to dns, http, and ping commands (default false)")
	flags.BoolVar(&ctx.Share, "share", ctx.Share, "print a link at the end of the results to visualize them online (default false)")
	flags.BoolVarP(&ctx.Ipv4, "ipv4", "4", ctx.Ipv4, "resolve names to IPv4 addresses")
	flags.BoolVarP(&ctx.Ipv6, "ipv6", "6", ctx.Ipv6, "resolve names to IPv6 addresses")

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
var usageTemplate = `
Use '{{.CommandPath}} --help' for more information about the command.`

var helpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{wrappedFlagUsages .LocalFlags | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{wrappedFlagUsages .InheritedFlags | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}{{end}}
`
