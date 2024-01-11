package cmd

import (
	"fmt"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/spf13/cobra"
)

// dnsCmd represents the dns command
var dnsCmd = &cobra.Command{
	Use:     "dns [target] from [location | measurement ID | @1 | first | @-1 | last | previous]",
	GroupID: "Measurements",
	Short:   "Resolve a DNS record similarly to dig",
	Long: `Performs DNS lookups and displays the answers that are returned from the name server(s) that were queried.
The default nameserver depends on the probe and is defined by the user's local settings or DHCP.
This command provides 2 different ways to provide the dns resolver:
Using the --resolver argument. For example:
  dns jsdelivr.com from Berlin --resolver 1.1.1.1
Using the dig format @resolver. For example:
  dns jsdelivr.com @1.1.1.1 from Berlin

  Examples:
  # Resolve google.com from 2 probes in New York
  dns google.com from New York --limit 2

  # Resolve google.com using probes from previous measurement
  dns google.com from rvasVvKnj48cxNjC

  # Resolve google.com using probes from first measurement in session
  dns google.com from @1

  # Resolve google.com using probes from last measurement in session
  dns google.com from last

  # Resolve google.com using probes from second to last measurement in session
  dns google.com from @-2

  # Resolve google.com from 2 probes from London or Belgium with trace enabled
  dns google.com from London,Belgium --limit 2 --trace

  # Resolve google.com from a probe in Paris using the TCP protocol
  dns google.com from Paris --protocol tcp

  # Resolve jsdelivr.com from a probe in Berlin using the type MX and the resolver 1.1.1.1 in CI mode
  dns jsdelivr.com from Berlin --type MX --resolver 1.1.1.1 --ci

  # Resolve jsdelivr.com from a probe that is from the AWS network and is located in Montreal with latency output
  dns jsdelivr.com from aws+montreal --latency

  # Resolve jsdelivr.com from a probe in ASN 123 with json output
  dns jsdelivr.com from 123 --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create context

		err := createContext(cmd.CalledAs(), args)
		if err != nil {
			return err
		}

		// Make post struct
		opts = model.PostMeasurement{
			Type:              "dns",
			Target:            ctx.Target,
			Limit:             ctx.Limit,
			InProgressUpdates: inProgressUpdates(ctx.CI),
			Options: &model.MeasurementOptions{
				Protocol: protocol,
				Port:     port,
				Resolver: overrideOpt(ctx.Resolver, resolver),
				Query: &model.QueryOptions{
					Type: queryType,
				},
				Trace: trace,
			},
		}
		isPreviousMeasurementId := false
		opts.Locations, isPreviousMeasurementId, err = createLocations(ctx.From)
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		res, showHelp, err := client.PostAPI(opts)
		if err != nil {
			if showHelp {
				return err
			}
			fmt.Println(err)
			return nil
		}

		// Save measurement ID to history
		if !isPreviousMeasurementId {
			err := saveMeasurementID(res.ID)
			if err != nil {
				fmt.Printf("warning: %s\n", err)
			}
		}

		view.OutputResults(res.ID, ctx, opts)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dnsCmd)

	// dns specific flags
	dnsCmd.Flags().StringVar(&protocol, "protocol", "", "Specifies the protocol to use for the DNS query (TCP or UDP) (default \"udp\")")
	dnsCmd.Flags().IntVar(&port, "port", 0, "Send the query to a non-standard port on the server (default 53)")
	dnsCmd.Flags().StringVar(&resolver, "resolver", "", "Resolver is the hostname or IP address of the name server to use (default empty)")
	dnsCmd.Flags().StringVar(&queryType, "type", "", "Specifies the type of DNS query to perform (default \"A\")")
	dnsCmd.Flags().BoolVar(&trace, "trace", false, "Toggle tracing of the delegation path from the root name servers (default false)")
}
