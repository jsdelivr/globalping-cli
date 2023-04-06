package cmd

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type UrlData struct {
	Protocol string
	Path     string
	Query    string
	Host     string
	Port     int
}

// parse url data from user text input
func parseUrlData(input string) (*UrlData, error) {
	var urlData UrlData

	// add url scheme if missing
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		input = "http://" + input
	}

	// Parse input
	u, err := url.Parse(input)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse url input")
	}

	urlData.Protocol = u.Scheme
	urlData.Path = u.Path
	urlData.Query = u.RawQuery

	h, p, err := net.SplitHostPort(u.Host)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") {
			// u.Host is not in the format "host:port"
			h = u.Host
		} else {
			return nil, errors.Wrapf(err, "failed to parse url host/port")
		}
	}

	urlData.Host = h

	if p != "" {
		// parse port if present
		urlData.Port, err = strconv.Atoi(p)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse url port number: %s", p)
		}
	}

	return &urlData, nil
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
	Use:     "http [target] from [location]",
	GroupID: "Measurements",
	Short:   "Perform a HEAD or GET request to a host",
	Long: `The http command sends an HTTP request to a host and can perform HEAD or GET operations. GET is limited to 10KB responses, everything above will be cut by the API.

Examples:
  # HTTP HEAD request to jsdelivr.com from 2 probes in New York (protocol, port and path are inferred from the URL)
  http https://www.jsdelivr.com:443/package/npm/test?nav=stats from New York --limit 2

  # HTTP GET request to google.com from 2 probes from London or Belgium
  http google.com from London,Belgium --limit 2 --method get

  # HTTP HEAD request to jsdelivr.com from a probe that is from the AWS network and is located in Montreal using HTTP2
  http jsdelivr.com from aws+montreal --protocol http2

  # HTTP GET request google.com with ASN 12345 with json output
  http google.com from 12345 --json`,
	Args: checkCommandFormat(),
	RunE: httpCmdRun,
}

// httpCmdRun is the cobra run function for the http command
func httpCmdRun(cmd *cobra.Command, args []string) error {
	// Create context
	err := createContext(cmd.CalledAs(), args)
	if err != nil {
		return err
	}

	// build http measurement
	m, err := buildHttpMeasurementRequest()
	if err != nil {
		return err
	}

	opts = m
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
}

const PostMeasurementTypeHttp = "http"

// buildHttpMeasurementRequest builds the measurement request for the http type
func buildHttpMeasurementRequest() (model.PostMeasurement, error) {
	m := model.PostMeasurement{
		Type: PostMeasurementTypeHttp,
	}

	urlData, err := parseUrlData(ctx.Target)
	if err != nil {
		return m, err
	}

	m.Target = urlData.Host
	m.Locations = createLocations(ctx.From)
	m.Limit = ctx.Limit
	m.Options = &model.MeasurementOptions{
		Protocol: overrideOpt(urlData.Protocol, httpCmdOpts.Protocol),
		Port:     overrideOptInt(urlData.Port, httpCmdOpts.Port),
		Packets:  packets,
		Request: &model.RequestOptions{
			Path:  overrideOpt(urlData.Path, httpCmdOpts.Path),
			Query: overrideOpt(urlData.Query, httpCmdOpts.Query),
			Host:  overrideOpt(urlData.Host, httpCmdOpts.Host),
			// TODO: Headers: headers,
			Method: httpCmdOpts.Method,
		},
		Resolver: httpCmdOpts.Resolver,
	}

	return m, nil
}

// HttpCmdOpts represents the parsed http command line opts
type HttpCmdOpts struct {
	Path     string
	Query    string
	Host     string
	Method   string
	Protocol string
	Port     int
	Resolver string
}

func init() {
	rootCmd.AddCommand(httpCmd)

	httpCmdOpts = &HttpCmdOpts{}

	// http specific flags
	httpCmd.Flags().StringVar(&httpCmdOpts.Path, "path", "", "A URL pathname (default \"/\")")
	httpCmd.Flags().StringVar(&httpCmdOpts.Query, "query", "", "A query-string")
	httpCmd.Flags().StringVar(&httpCmdOpts.Host, "host", "", "Specifies the Host header, which is going to be added to the request (default host defined in target)")
	httpCmd.Flags().StringVar(&httpCmdOpts.Method, "method", "", "Specifies the HTTP method to use (HEAD or GET) (default \"HEAD\")")
	httpCmd.Flags().StringVar(&httpCmdOpts.Protocol, "protocol", "", "Specifies the query protocol (HTTP, HTTPS, HTTP2) (default \"HTTP\")")
	httpCmd.Flags().IntVar(&httpCmdOpts.Port, "port", 0, "Specifies the port to use (default 80 for HTTP, 443 for HTTPS and HTTP2)")
	httpCmd.Flags().StringVar(&httpCmdOpts.Resolver, "resolver", "", "Specifies the resolver server used for DNS lookup")

	// Extra flags
	httpCmd.Flags().BoolVar(&ctx.Latency, "latency", false, "Output only stats of a measurement (default false)")
}
