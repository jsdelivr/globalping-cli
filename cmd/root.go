package cmd

import (
	"errors"
	"os"

	"github.com/jsdelivr/globalping-cli/model"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	// cfgFile string
	from  string
	limit int
	// Additional flags
	packets   int
	protocol  string
	port      int
	resolver  string
	trace     bool
	queryType string
	path      string
	host      string
	query     string
	method    string
	// TODO: headers   map[string]string

	opts = model.PostMeasurement{}
	ctx  = model.ViewContext{}
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "globalping",
	Short: "A global network of probes to run network tests like ping, traceroute and DNS resolve.",
	Long: `Globalping is a platform that allows anyone to run networking commands such as ping, traceroute, dig and mtr on probes distributed all around the world. 
	Our goal is to provide a free and simple API for everyone out there to build interesting networking tools and services.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&from, "from", "F", "world", "A continent, region (e.g eastern europe), country, US state or city")
	rootCmd.PersistentFlags().IntVarP(&limit, "limit", "L", 1, "Limit the number of probes to use")
	rootCmd.PersistentFlags().BoolVarP(&ctx.JsonOutput, "json", "J", false, "Output results in JSON format")
}

// requireTarget returns an error if no target is specified.
func requireTarget() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("no target specified")
		}
		return nil
	}
}
