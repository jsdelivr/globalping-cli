package cmd

import (
	"fmt"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/spf13/cobra"
)

// dnsCmd represents the dns command
var dnsCmd = &cobra.Command{
	Use:   "dns [target]",
	Short: "Implementation of the native dig command",
	Long: `Performs DNS lookups and displays the answers that are returned from the name server(s) that were queried.
	
		Examples:
		# Resolve google.com from a probe in the network
		dns traceroute google.com --from "New York" --limit 2`,
	Args: requireTarget(),
	Run: func(cmd *cobra.Command, args []string) {
		// Make post struct
		opts = model.PostMeasurement{
			Type:   "dns",
			Target: args[0],
			Locations: model.Locations{
				{
					Magic: from,
				},
			},
			Limit: limit,
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

		res, err := client.PostAPI(opts)
		if err != nil {
			fmt.Println(err)
			return
		}

		client.OutputResults(res.ID, ctx)
	},
}

func init() {
	rootCmd.AddCommand(dnsCmd)

	// dns specific flags
	dnsCmd.Flags().StringVar(&protocol, "protocol", "", "Specifies the protocol to use for the DNS query (TCP or UDP). (default \"udp\")")
	dnsCmd.Flags().IntVar(&port, "port", 0, "Send the query to a non-standard port on the server, instead of the default port 53.")
	dnsCmd.Flags().StringVar(&resolver, "resolver", "", "Resolver is the name or IP address of the name server to query.")
	dnsCmd.Flags().StringVar(&queryType, "type", "", "Specifies the type of DNS query to perform. (default \"A\")")
	dnsCmd.Flags().BoolVar(&trace, "trace", false, "Toggle tracing of the delegation path from the root name servers for the name being looked up.")
}
