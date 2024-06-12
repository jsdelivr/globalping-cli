package cmd

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *Root) initHTTP() {
	httpCmd := &cobra.Command{
		RunE:    r.RunHTTP,
		Use:     "http [target] from [location | measurement ID | @1 | first | @-1 | last | previous]",
		GroupID: "Measurements",
		Short:   "Perform a HEAD or GET request to a host.",
		Long: `The http command sends an HTTP request to a host and can perform either HEAD or GET operations, returning detailed performance statistics for each request. Use it to test and assess the performance and availability of your website, API, or other web services. 
Note that GET responses are limited to 10KB, with anything beyond this cut by the API. 

The CLI tool supports two formats: 
1. Full URL: The tool automatically parses the scheme, host, port, domain, path, and query. For example:
	http https://www.jsdelivr.com:443/package/npm/test?nav=stats
2. Separate flags: Specify the scheme, host, port, domain, path, and query as separate command line flags, useful for scripting. For example:
	http jsdelivr.com --host www.jsdelivr.com --protocol https --port 443 --path "/package/npm/test" --query "nav=stats"

Note that a probe's local settings or DHCP determine the default nameserver the command uses. To specify a DNS resolver, use the --resolver argument or @resolver format:  
- http jsdelivr.com from Berlin --resolver 1.1.1.1
- http jsdelivr.com @1.1.1.1 from Berlin

Examples:
  # Perform an HTTP HEAD request to jsdelivr.com from 2 probes in New York. Protocol, port, and path are derived from the URL.
  http https://www.jsdelivr.com:443/package/npm/test?nav=stats from New York --limit 2

  # Perform an HTTP GET request to google.com from 2 probes from London or Belgium and enable CI mode.
  http google.com from London,Belgium --limit 2 --method get --ci

  # Perform an HTTP GET request to google.com using probes from a previous measurement by using its ID.
  http google.com from rvasVvKnj48cxNjC --method get

  # Perform an HTTP GET request to google.com using the same probes from the first measurement in this session.
  http google.com from @1 --method get

  # Perform an HTTP GET request to google.com using the same probes from the last measurement in this session.
  http google.com from last --method get

  # Perform an HTTP GET request to google.com using the same probes from the second-to-last measurement in this session.
  http google.com from @-2 --method get

  # Perform an HTTP GET request to google.com from a probe in London and return the full output.
  http google.com from London --method get --full

  # Perform an HTTP HEAD request to jsdelivr.com from a probe on the AWS network located in Montreal using HTTP2. Include the http headers "Accept-Encoding" and "Accept-Language" to the request.
  http jsdelivr.com from aws+montreal --protocol http2 --header "Accept-Encoding: br,gzip" -H "Accept-Language: *"

  # Perform an HTTP HEAD request to jsdelivr.com from a non-data center probe in Europe and add the path /robots.txt and query string "test=1" to the request.
  http jsdelivr.com from europe+eyeball --path /robots.txt --query "test=1"

  # Perform an HTTP HEAD request to example.com from a probe in Berlin. Override the "Host" header by specifying a different host (example.org) from the target (example.com).
  http example.com from Berlin --host example.org

  # Perform an HTTP GET request to google.com from a probe in ASN 123 using 1.1.1.1 as the DNS resolver and output the results in JSON format.
  http google.com from 123 --resolver 1.1.1.1 --json`,
	}

	// http specific flags
	flags := httpCmd.Flags()
	flags.StringVar(&r.ctx.Protocol, "protocol", r.ctx.Protocol, "Specify the protocol to use: HTTP, HTTPS, or HTTP2.  (default \"HTTP\")")
	flags.IntVar(&r.ctx.Port, "port", r.ctx.Port, "Specify the port to use. (default 80 for HTTP, 443 for HTTPS and HTTP2)")
	flags.StringVar(&r.ctx.Resolver, "resolver", r.ctx.Resolver, "Specify the hostname or IP address of the name server to use for the DNS lookup. (default defined by the probe)")
	flags.StringVar(&r.ctx.Host, "host", r.ctx.Host, "Specify the Host header to add to the request. (default host's defined in command target)")
	flags.StringVar(&r.ctx.Path, "path", r.ctx.Path, "Specify the URL pathname. (default \"/\")")
	flags.StringVar(&r.ctx.Query, "query", r.ctx.Query, "Specify a query string to add.")
	flags.StringVar(&r.ctx.Method, "method", r.ctx.Method, "Specify the HTTP method to use: HEAD or GET. (default \"HEAD\")")
	flags.StringArrayVarP(&r.ctx.Headers, "header", "H", r.ctx.Headers, "Specify an HTTP header to add to the request in the format \"Key: Value\". You can add multiple headers by providing the flag for each one separately.")
	flags.BoolVar(&r.ctx.Full, "full", r.ctx.Full, "Enable full output when performing an HTTP GET request to display the status, headers, and body.")

	r.Cmd.AddCommand(httpCmd)
}

func (r *Root) RunHTTP(cmd *cobra.Command, args []string) error {
	err := r.updateContext(cmd.CalledAs(), args)
	if err != nil {
		return err
	}

	defer r.UpdateHistory()
	r.ctx.RecordToSession = true

	opts, err := r.buildHttpMeasurementRequest()
	if err != nil {
		return err
	}

	opts.Locations, err = r.getLocations()
	if err != nil {
		cmd.SilenceUsage = true
		return err
	}

	res, showHelp, err := r.client.CreateMeasurement(opts)
	if err != nil {
		if !showHelp {
			cmd.SilenceUsage = true
		}
		return err
	}

	r.ctx.MeasurementsCreated++
	hm := &view.HistoryItem{
		Id:        res.ID,
		Status:    globalping.StatusInProgress,
		StartedAt: r.time.Now(),
	}
	r.ctx.History.Push(hm)
	if r.ctx.RecordToSession {
		r.ctx.RecordToSession = false
		err := saveIdToSession(res.ID)
		if err != nil {
			r.printer.Printf("Warning: %s\n", err)
		}
	}

	r.viewer.Output(res.ID, opts)
	return nil
}

const PostMeasurementTypeHttp = "http"

// buildHttpMeasurementRequest builds the measurement request for the http type
func (r *Root) buildHttpMeasurementRequest() (*globalping.MeasurementCreate, error) {
	opts := &globalping.MeasurementCreate{
		Type:              PostMeasurementTypeHttp,
		Limit:             r.ctx.Limit,
		InProgressUpdates: !r.ctx.CIMode,
	}
	urlData, err := parseUrlData(r.ctx.Target)
	if err != nil {
		return nil, err
	}
	headers, err := parseHttpHeaders(r.ctx.Headers)
	if err != nil {
		return nil, err
	}
	method := strings.ToUpper(r.ctx.Method)
	if r.ctx.Full {
		// override method to GET
		method = "GET"
	}
	opts.Target = urlData.Host
	opts.Options = &globalping.MeasurementOptions{
		Protocol: overrideOpt(urlData.Protocol, r.ctx.Protocol),
		Port:     overrideOptInt(urlData.Port, r.ctx.Port),
		Request: &globalping.RequestOptions{
			Path:    overrideOpt(urlData.Path, r.ctx.Path),
			Query:   overrideOpt(urlData.Query, r.ctx.Query),
			Host:    overrideOpt(urlData.Host, r.ctx.Host),
			Headers: headers,
			Method:  method,
		},
		Resolver: r.ctx.Resolver,
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
