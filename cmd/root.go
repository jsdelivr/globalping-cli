package cmd

import (
	"io"
	"os"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/lib"
	"github.com/jsdelivr/globalping-cli/utils"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// TODO: Remove global variables

var (
	// Global flags
	// cfgFile string

	// Additional flags
	protocol  string
	port      int
	resolver  string
	trace     bool
	queryType string

	httpCmdOpts *HttpCmdOpts

	opts = globalping.MeasurementCreate{}
	ctx  = &view.Context{
		APIMinInterval: globalping.API_MIN_INTERVAL,
		MaxHistory:     10,
	}
	utime   = utils.NewTime()
	gp      = globalping.NewClient(globalping.API_URL)
	outW    = os.Stdout
	errW    = os.Stderr
	printer = view.NewPrinter(outW)
	viewer  = view.NewViewer(ctx, printer, utime, gp)
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "globalping",
	Short: "A global network of probes to run network tests like ping, traceroute and DNS resolve.",
	Long: `Globalping is a platform that allows anyone to run networking commands such as ping, traceroute, dig and mtr on probes distributed all around the world.
The CLI tool allows you to interact with the API in a simple and human-friendly way to debug networking issues like anycast routing and script automated tests and benchmarks.`,
}

var root = NewRoot(outW, errW, printer, ctx, viewer, utime, gp, rootCmd)

type Root struct {
	outW    io.Writer
	printer *view.Printer
	ctx     *view.Context
	viewer  view.Viewer
	gp      globalping.Client
	time    utils.Time
	Cmd     *cobra.Command
}

func NewRoot(
	outW io.Writer,
	errW io.Writer,
	printer *view.Printer,
	ctx *view.Context,
	viewer view.Viewer,
	time utils.Time,
	gp globalping.Client,
	cmd *cobra.Command,
) *Root {
	root := &Root{
		Cmd:     cmd,
		outW:    outW,
		printer: printer,
		ctx:     ctx,
		viewer:  viewer,
		time:    time,
		gp:      gp,
	}

	root.Cmd.SetOut(outW)
	root.Cmd.SetErr(errW)

	// Global flags
	flags := root.Cmd.PersistentFlags()
	flags.StringVarP(&ctx.From, "from", "F", "world", `Comma-separated list of location values to match against or a measurement ID
	For example, the partial or full name of a continent, region (e.g eastern europe), country, US state, city or network
	Or use [@1 | first, @2 ... @-2, @-1 | last | previous] to run with the probes from previous measurements.`)
	flags.IntVarP(&ctx.Limit, "limit", "L", 1, "Limit the number of probes to use")
	flags.BoolVarP(&ctx.ToJSON, "json", "J", false, "Output results in JSON format (default false)")
	flags.BoolVarP(&ctx.CI, "ci", "C", false, "Disable realtime terminal updates and color suitable for CI and scripting (default false)")
	flags.BoolVar(&ctx.ToLatency, "latency", false, "Output only the stats of a measurement (default false). Only applies to the dns, http and ping commands")
	flags.BoolVar(&ctx.Share, "share", false, "Prints a link at the end the results, allowing to vizualize the results online (default false)")

	root.Cmd.AddGroup(&cobra.Group{ID: "Measurements", Title: "Measurement Commands:"})

	root.initPing()

	return root
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := root.Cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func (c *Root) updateContext(cmd string, args []string) error {
	c.ctx.Cmd = cmd // Get the command name

	// parse target query
	targetQuery, err := lib.ParseTargetQuery(cmd, args)
	if err != nil {
		return err
	}

	c.ctx.Target = targetQuery.Target

	if targetQuery.From != "" {
		c.ctx.From = targetQuery.From
	}

	if targetQuery.Resolver != "" {
		c.ctx.Resolver = targetQuery.Resolver
	}

	// Check env for CI
	if os.Getenv("CI") != "" {
		c.ctx.CI = true
	}

	// Check if it is a terminal or being piped/redirected
	// We want to disable realtime updates if that is the case
	f, ok := c.outW.(*os.File)
	if ok {
		stdoutFileInfo, err := f.Stat()
		if err != nil {
			return errors.Wrapf(err, "stdout stat failed")
		}
		if (stdoutFileInfo.Mode() & os.ModeCharDevice) == 0 {
			// stdout is piped, run in ci mode
			c.ctx.CI = true
		}
	} else {
		c.ctx.CI = true
	}

	return nil
}

// Todo: Remove this function
func createContext(cmd string, args []string) error {
	ctx.Cmd = cmd // Get the command name

	// parse target query
	targetQuery, err := lib.ParseTargetQuery(cmd, args)
	if err != nil {
		return err
	}

	ctx.Target = targetQuery.Target

	if targetQuery.From != "" {
		ctx.From = targetQuery.From
	}

	if targetQuery.Resolver != "" {
		ctx.Resolver = targetQuery.Resolver
	}

	// Check env for CI
	if os.Getenv("CI") != "" {
		ctx.CI = true
	}

	// Check if it is a terminal or being piped/redirected
	// We want to disable realtime updates if that is the case
	stdoutFileInfo, err := outW.Stat()
	if err != nil {
		return errors.Wrapf(err, "stdout stat failed")
	}

	if (stdoutFileInfo.Mode() & os.ModeCharDevice) == 0 {
		// stdout is piped, run in ci mode
		ctx.CI = true
	}

	return nil
}
