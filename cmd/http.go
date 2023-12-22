package cmd

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/jsdelivr/globalping-cli/client"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/jsdelivr/globalping-cli/view"
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
	Use:     "http [target] from [location | measurement ID | @1 | first | @-1 | last]",
	GroupID: "Measurements",
	Short:   "Perform a HEAD or GET request to a host",
	Long: `The http command sends an HTTP request to a host and can perform HEAD or GET operations. GET is limited to 10KB responses, everything above will be cut by the API. Detailed performance stats as available for every request.
The tool supports 2 formats for this command:
When the full url is supplied, the tool autoparses the scheme, host, port, domain, path and query. For example:
  http https://www.jsdelivr.com:443/package/npm/test?nav=stats
As an alternative that can be useful for scripting, the scheme, host, port, domain, path and query can be provided as separate command line flags. For example:
  http jsdelivr.com --host www.jsdelivr.com --protocol https --port 443 --path "/package/npm/test" --query "nav=stats"

This command also provides 2 different ways to provide the dns resolver:
Using the --resolver argument. For example:
 http jsdelivr.com from Berlin --resolver 1.1.1.1
Using the dig format @resolver. For example:
 http jsdelivr.com @1.1.1.1 from Berlin

Examples:
  # HTTP HEAD request to jsdelivr.com from 2 probes in New York (protocol, port and path are inferred from the URL)
  http https://www.jsdelivr.com:443/package/npm/test?nav=stats from New York --limit 2

  # HTTP GET request to google.com from 2 probes from London or Belgium in CI mode
  http google.com from London,Belgium --limit 2 --method get --ci

  # HTTP GET request google.com using probes from previous measurement
  http google.com from rvasVvKnj48cxNjC --method get

  # HTTP GET request google.com using probes from first measurement in session
  http google.com from @1 --method get

  # HTTP GET request google.com using probes from last measurement in session
  http google.com from last --method get

  # HTTP GET request to google.com from a probe in London. Returns the full output
  http google.com from London --method get --full

  # HTTP HEAD request to jsdelivr.com from a probe that is from the AWS network and is located in Montreal using HTTP2. 2 http headers are added to the request.
  http jsdelivr.com from aws+montreal --protocol http2 --header "Accept-Encoding: br,gzip" -H "Accept-Language: *"

  # HTTP HEAD request to jsdelivr.com from a probe that is located in Paris, using the /robots.txt path with "test=1" query string
  http jsdelivr.com from Paris --path /robots.txt --query "test=1"

  # HTTP HEAD request to example.com from a probe that is located in Berlin, specifying a different host example.org in the request headers
  http example.com from Berlin --host example.org

  # HTTP GET request google.com from a probe in ASN 123 with a dns resolver 1.1.1.1 and json output
  http google.com from 123 --resolver 1.1.1.1 --json`,
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
	opts, err := buildHttpMeasurementRequest()
	if err != nil {
		return err
	}
	isPreviousMeasurementId := false
	opts.Locations, isPreviousMeasurementId, err = createLocations(ctx.From)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	res, showHelp, err := client.PostAPI(*opts)
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

	view.OutputResults(res.ID, ctx, *opts)
	return nil
}

const PostMeasurementTypeHttp = "http"

// buildHttpMeasurementRequest builds the measurement request for the http type
func buildHttpMeasurementRequest() (*model.PostMeasurement, error) {
	opts := &model.PostMeasurement{
		Type:              PostMeasurementTypeHttp,
		Limit:             ctx.Limit,
		InProgressUpdates: inProgressUpdates(ctx.CI),
	}
	urlData, err := parseUrlData(ctx.Target)
	if err != nil {
		return nil, err
	}
	headers, err := parseHttpHeaders(httpCmdOpts.Headers)
	if err != nil {
		return nil, err
	}
	method := strings.ToUpper(httpCmdOpts.Method)
	if ctx.Full {
		// override method to GET
		method = "GET"
	}
	opts.Target = urlData.Host
	opts.Options = &model.MeasurementOptions{
		Protocol: overrideOpt(urlData.Protocol, httpCmdOpts.Protocol),
		Port:     overrideOptInt(urlData.Port, httpCmdOpts.Port),
		Packets:  packets,
		Request: &model.RequestOptions{
			Path:    overrideOpt(urlData.Path, httpCmdOpts.Path),
			Query:   overrideOpt(urlData.Query, httpCmdOpts.Query),
			Host:    overrideOpt(urlData.Host, httpCmdOpts.Host),
			Headers: headers,
			Method:  method,
		},
		Resolver: overrideOpt(ctx.Resolver, httpCmdOpts.Resolver),
	}
	return opts, nil
}

func parseHttpHeaders(headerStrings []string) (map[string]string, error) {
	h := map[string]string{}

	for _, r := range headerStrings {
		k, v, ok := strings.Cut(r, ": ")
		if !ok {
			return nil, fmt.Errorf("invalid header: %s", r)
		}

		h[k] = v
	}

	return h, nil
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
	Headers  []string
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
	httpCmd.Flags().StringVar(&httpCmdOpts.Resolver, "resolver", "", "Specifies the resolver server used for DNS lookup (default is defined by the probe's network)")
	httpCmd.Flags().StringArrayVarP(&httpCmdOpts.Headers, "header", "H", nil, "Specifies a HTTP header to be added to the request, in the format \"Key: Value\". Multiple headers can be added by adding multiple flags")
	httpCmd.Flags().BoolVar(&ctx.Full, "full", false, "Full output. Uses an HTTP GET request, and outputs the status, headers and body to the output")
}
