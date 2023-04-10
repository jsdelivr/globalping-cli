package cmd

import (
	"errors"
	"os"
	"strings"

	"github.com/jsdelivr/globalping-cli/model"
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

	// TODO: headers   map[string]string

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
	rootCmd.PersistentFlags().StringVarP(&ctx.From, "from", "F", "", `Comma-separated list of location values to match against. For example the partial or full name of a continent, region (e.g eastern europe), country, US state, city or network (default "world").`)
	rootCmd.PersistentFlags().IntVarP(&ctx.Limit, "limit", "L", 1, "Limit the number of probes to use")
	rootCmd.PersistentFlags().BoolVarP(&ctx.JsonOutput, "json", "J", false, "Output results in JSON format (default false)")
	rootCmd.PersistentFlags().BoolVarP(&ctx.CI, "ci", "C", false, "Disable realtime terminal updates and color suitable for CI and scripting (default false)")
	rootCmd.PersistentFlags().BoolVar(&ctx.Latency, "latency", false, "Output only the stats of a measurement (default false). Only applies to the dns, http and ping commands")
}

// checkCommandFormat checks if the command is in the correct format if using the from arg
func checkCommandFormat() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 && args[1] != "from" {
			return errors.New("invalid command format")
		}
		return nil
	}
}

func createContext(cmd string, args []string) error {
	ctx.Cmd = cmd // Get the command name

	// Target
	if len(args) == 0 {
		return errors.New("provided target is empty")
	}
	ctx.Target = args[0]

	// If no from arg is provided, use the default value
	if len(args) == 1 && ctx.From == "" {
		ctx.From = "world"
	}

	// If from args are provided, use it
	if len(args) > 1 && args[1] == "from" {
		ctx.From = strings.TrimSpace(strings.Join(args[2:], " "))
	}

	// Check env for CI
	if os.Getenv("CI") != "" {
		ctx.CI = true
	}

	// Check if it is a terminal or being piped/redirected
	// We want to disable realtime updates if that is the case
	o, _ := os.Stdout.Stat()
	if (o.Mode() & os.ModeCharDevice) != os.ModeCharDevice {
		ctx.CI = true
	}

	return nil
}
