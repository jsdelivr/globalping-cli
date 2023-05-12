package cmd

import (
	"os"

	"github.com/jsdelivr/globalping-cli/lib"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	// cfgFile string

	// Additional flags
	packets   int
	protocol  string
	port      int
	resolver  string
	trace     bool
	queryType string

	httpCmdOpts *HttpCmdOpts

	opts = model.PostMeasurement{}
	ctx  = model.Context{}
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "globalping",
	Short: "A global network of probes to run network tests like ping, traceroute and DNS resolve.",
	Long: `Globalping is a platform that allows anyone to run networking commands such as ping, traceroute, dig and mtr on probes distributed all around the world.
The CLI tool allows you to interact with the API in a simple and human-friendly way to debug networking issues like anycast routing and script automated tests and benchmarks.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.AddGroup(&cobra.Group{ID: "Measurements", Title: "Measurement Commands:"})
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&ctx.From, "from", "F", "world", `Comma-separated list of location values to match against. For example the partial or full name of a continent, region (e.g eastern europe), country, US state, city or network (default "world").`)
	rootCmd.PersistentFlags().IntVarP(&ctx.Limit, "limit", "L", 1, "Limit the number of probes to use")
	rootCmd.PersistentFlags().BoolVarP(&ctx.JsonOutput, "json", "J", false, "Output results in JSON format (default false)")
	rootCmd.PersistentFlags().BoolVarP(&ctx.CI, "ci", "C", false, "Disable realtime terminal updates and color suitable for CI and scripting (default false)")
	rootCmd.PersistentFlags().BoolVar(&ctx.Latency, "latency", false, "Output only the stats of a measurement (default false). Only applies to the dns, http and ping commands")
	rootCmd.PersistentFlags().BoolVar(&ctx.Share, "share", false, "Prints a link at the end the results, allowing to vizualize the results online (default false)")
}

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
	stdoutFileInfo, err := os.Stdout.Stat()
	if err != nil {
		return errors.Wrapf(err, "stdout stat failed")
	}

	if (stdoutFileInfo.Mode() & os.ModeCharDevice) == 0 {
		// stdout is piped, run in ci mode
		ctx.CI = true
	}

	return nil
}
