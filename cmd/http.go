package cmd

import (
	"fmt"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/spf13/cobra"
)

// httpCmd represents the http command
var httpCmd = &cobra.Command{
	Use:   "http [target] from [location]",
	Short: "Use http command",
	Long: `The http command sends an HTTP request to a host and can perform HEAD or GET operations.
	
		Examples:
		# HTTP google.com from a probe in the network
		globalping http google.com --from "New York" --limit 2`,
	Args: checkCommandFormat(),
	Run: func(cmd *cobra.Command, args []string) {
		// Create context
		err := createContext(args)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Make post struct
		opts = model.PostMeasurement{
			Type:      "http",
			Target:    ctx.Target,
			Locations: createLocations(ctx.From),
			Limit:     ctx.Limit,
			Options: &model.MeasurementOptions{
				Protocol: protocol,
				Port:     port,
				Packets:  packets,
				Request: &model.RequestOptions{
					Path:  path,
					Query: query,
					Host:  host,
					// TODO: Headers: headers,
					Method: method,
				},
				Resolver: resolver,
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
	rootCmd.AddCommand(httpCmd)

	// http specific flags
	httpCmd.Flags().StringVar(&path, "path", "", "A URL pathname (default \"/\")")
	httpCmd.Flags().StringVar(&query, "query", "", "A query-string")
	httpCmd.Flags().StringVar(&host, "host", "", "Specifies the Host header, which is going to be added to the request (default host defined in target)")
	httpCmd.Flags().StringVar(&method, "method", "", "Specifies the HTTP method to use (HEAD or GET).(default \"HEAD\")")
	httpCmd.Flags().StringVar(&protocol, "protocol", "", "Specifies the query protocol (HTTP, HTTPS, HTTP2) (default \"HTTP\")")
	httpCmd.Flags().IntVar(&port, "port", 0, "Specifies the port to use (default 80 for HTTP, 443 for HTTPS and HTTP2)")
	httpCmd.Flags().StringVar(&resolver, "resolver", "", "Specifies the resolver server used for DNS lookup")

}
