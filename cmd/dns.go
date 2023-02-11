package cmd

import (
	"fmt"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/spf13/cobra"
)

// dnsCmd represents the dns command
var dnsCmd = &cobra.Command{
	Use:   "dns [target] from [location]",
	Short: "Implementation of the native dig command",
	Long: `Performs DNS lookups and displays the answers that are returned from the name server(s) that were queried.
	
		Examples:
		# Resolve google.com from a probe in the network
		dns traceroute google.com --from "New York" --limit 2`,
	Args: checkCommandFormat(),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create context
		err := createContext(args)
		if err != nil {
			return err
		}

		// Make post struct
		opts = model.PostMeasurement{
			Type:      "dns",
			Target:    ctx.Target,
			Locations: createLocations(ctx.From),
			Limit:     ctx.Limit,
			Options: &model.MeasurementOptions{
				Protocol: protocol,
				Port:     port,
				Resolver: resolver,
				Query: &model.QueryOptions{
					Type: queryType,
				},
				Trace: trace,
			},
		}

		res, showHelp, err := client.PostAPI(opts)
		if err != nil {
			if showHelp {
				return err
			}
			fmt.Println(err)
			return nil
		}

		client.OutputResults(res.ID, ctx)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dnsCmd)

	// dns specific flags
	dnsCmd.Flags().StringVar(&protocol, "protocol", "", "Specifies the protocol to use for the DNS query (TCP or UDP) (default \"udp\")")
	dnsCmd.Flags().IntVar(&port, "port", 0, "Send the query to a non-standard port on the server (default 53)")
	dnsCmd.Flags().StringVar(&resolver, "resolver", "", "Resolver is the name or IP address of the name server to query (default empty)")
	dnsCmd.Flags().StringVar(&queryType, "type", "", "Specifies the type of DNS query to perform (default \"A\")")
	dnsCmd.Flags().BoolVar(&trace, "trace", false, "Toggle tracing of the delegation path from the root name servers (default false)")
}
