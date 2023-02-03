package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	from    string
	limit   int
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
	cobra.OnInitialize(initConfig)
	// Load persistent flags from config file if present
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.globalping-cli.yaml)")

	// Other global flags
	rootCmd.PersistentFlags().StringVarP(&from, "from", "F", "world", "A continent, region (e.g eastern europe), country, US state or city")
	rootCmd.PersistentFlags().IntVarP(&limit, "limit", "L", 1, "Limit the number of probes to use")

}

// initConfig reads in config file and ENV variables if set.
// This is to store or read API keys from a config file or env
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".globalping-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".globalping-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
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
