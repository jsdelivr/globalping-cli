package cmd

import (
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/spf13/cobra"
)

type urlFlags struct {
	Path     string
	Query    string
	Host     string
	Protocol string
	Port     int
}

// This allows the user to specify a URL as the target without additional flags
func parseURL(input string) (urlFlags, error) {
	var flags urlFlags

	// Parse URL
	u, err := url.Parse(input)
	if err != nil {
		return flags, err
	}

	// Set flags
	flags.Protocol = u.Scheme
	flags.Path = u.Path
	flags.Query = u.RawQuery
	host, port, _ := net.SplitHostPort(u.Host)
	flags.Host = host
	flags.Port, _ = strconv.Atoi(port)

	return flags, nil

}

// Helper functions to override flags in command
func overrideOpt(orig, new string) string {
	if new != "" {
		return new
	}
	return orig
}

func overrideOptInt(orig, new int) int {
	if new != 0 {
		return new
	}
	return orig
}

// httpCmd represents the http command
var httpCmd = &cobra.Command{
	Use:   "http [target] from [location]",
	Short: "Use http command",
	Long: `The http command sends an HTTP request to a host and can perform HEAD or GET operations. GET is limited to 10KB responses, everything above will be cut by the API.
	
Examples:
# HTTP HEAD request to jsdelivr.com from 2 probes in New York
http https://www.jsdelivr.com/package/npm/test?nav=stats --from "New York" --limit 2`,
	Args: checkCommandFormat(),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create context
		err := createContext(cmd.CalledAs(), args)
		if err != nil {
			return err
		}

		flags, err := parseURL(ctx.Target)
		if err != nil {
			return err
		}

		// Make post struct
		opts = model.PostMeasurement{
			Type:      "http",
			Target:    overrideOpt(ctx.Target, flags.Host),
			Locations: createLocations(ctx.From),
			Limit:     ctx.Limit,
			Options: &model.MeasurementOptions{
				Protocol: overrideOpt(protocol, flags.Protocol),
				Port:     overrideOptInt(port, flags.Port),
				Packets:  packets,
				Request: &model.RequestOptions{
					Path:  overrideOpt(path, flags.Path),
					Query: overrideOpt(query, flags.Query),
					Host:  overrideOpt(host, flags.Host),
					// TODO: Headers: headers,
					Method: method,
				},
				Resolver: resolver,
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
	rootCmd.AddCommand(httpCmd)

	// http specific flags
	httpCmd.Flags().StringVar(&path, "path", "", "A URL pathname (default \"/\")")
	httpCmd.Flags().StringVar(&query, "query", "", "A query-string")
	httpCmd.Flags().StringVar(&host, "host", "", "Specifies the Host header, which is going to be added to the request (default host defined in target)")
	httpCmd.Flags().StringVar(&method, "method", "", "Specifies the HTTP method to use (HEAD or GET).(default \"HEAD\")")
	httpCmd.Flags().StringVar(&protocol, "protocol", "", "Specifies the query protocol (HTTP, HTTPS, HTTP2) (default \"HTTP\")")
	httpCmd.Flags().IntVar(&port, "port", 0, "Specifies the port to use (default 80 for HTTP, 443 for HTTPS and HTTP2)")
	httpCmd.Flags().StringVar(&resolver, "resolver", "", "Specifies the resolver server used for DNS lookup")

	// Extra flags
	httpCmd.Flags().BoolVar(&ctx.Latency, "latency", false, "Output only stats of a measurement (default false)")
}
