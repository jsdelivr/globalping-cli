package cmd

import (
	"fmt"
	"net"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/jsdelivr/globalping-cli/view"
	"github.com/jsdelivr/globalping-go"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func (r *Root) initHTTP(measurementFlags *pflag.FlagSet, localFlags *pflag.FlagSet) {
	httpCmd := &cobra.Command{
		RunE:    r.RunHTTP,
		Use:     "http [target] from [location | measurement ID | @1 | first | @-1 | last | previous]",
		GroupID: "Measurements",
		Short:   "Perform a HEAD, GET, or OPTIONS request to a host",
		Long: `The http command sends an HTTP request to a host and can perform a HEAD, GET, or OPTIONS operations, returning detailed performance statistics for each request. Use it to test and assess the performance and availability of your website, API, or other web services.
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
	localFlags.BoolP("help", "h", false, "help for http")
	localFlags.String("protocol", "HTTPS", "specify the protocol to use: HTTP, HTTPS, or HTTP2")
	localFlags.Uint16("port", 443, "specify the port to use")
	localFlags.StringVar(&r.ctx.Resolver, "resolver", r.ctx.Resolver, "specify the hostname or IP address of the name server to use for the DNS lookup (default defined by the probe)")
	localFlags.StringVar(&r.ctx.Host, "host", r.ctx.Host, "specify the Host header to add to the request (default host's defined in command target)")
	localFlags.StringVar(&r.ctx.Path, "path", r.ctx.Path, "specify the URL pathname (default \"/\")")
	localFlags.StringVar(&r.ctx.Query, "query", r.ctx.Query, "specify a query string to add")
	localFlags.StringVarP(&r.ctx.Method, "method", "X", r.ctx.Method, "specify the HTTP method to use: HEAD, GET, or OPTIONS (default \"HEAD\")")
	localFlags.StringArrayVarP(&r.ctx.Headers, "header", "H", r.ctx.Headers, "add HTTP headers to the request in the format \"Key: Value\"; to add multiple headers, define the flag for each one separately")
	localFlags.BoolVar(&r.ctx.Full, "full", r.ctx.Full, "enable full output to display TLS details, HTTP status, headers, and body (if available); changes the default HTTP method to GET")
	httpCmd.Flags().AddFlagSet(measurementFlags)
	httpCmd.Flags().AddFlagSet(localFlags)

	r.Cmd.AddCommand(httpCmd)
}

func (r *Root) RunHTTP(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	err := r.updateContext(cmd, args)
	if err != nil {
		return err
	}

	if !slices.Contains(globalping.HTTPProtocols, r.ctx.Protocol) {
		return fmt.Errorf("protocol %s is not supported", r.ctx.Protocol)
	}

	if r.ctx.Protocol == "HTTP" && !cmd.Flag("port").Changed {
		r.ctx.Port = 80
	}

	defer r.UpdateHistory()
	r.ctx.RecordToSession = true

	opts, err := r.buildHttpMeasurementRequest(cmd)
	if err != nil {
		return err
	}

	opts.Locations, err = r.getLocations()
	if err != nil {
		cmd.SilenceUsage = true
		return err
	}

	if r.ctx.Ipv4 {
		opts.Options.IPVersion = globalping.IPVersion4
	} else if r.ctx.Ipv6 {
		opts.Options.IPVersion = globalping.IPVersion6
	}

	res, err := r.client.CreateMeasurement(ctx, opts)
	if err != nil {
		cmd.SilenceUsage = silenceUsageOnCreateMeasurementError(err)
		r.evaluateError(err)
		return err
	}

	r.ctx.MeasurementsCreated++
	hm := &view.HistoryItem{
		Id:        res.ID,
		Status:    globalping.StatusInProgress,
		StartedAt: r.utils.Now(),
	}
	r.ctx.History.Push(hm)
	if r.ctx.RecordToSession {
		r.ctx.RecordToSession = false
		err := r.storage.SaveIdToSession(res.ID)
		if err != nil {
			r.printer.Printf("Warning: %s\n", err)
		}
	}

	r.handleMeasurement(ctx, res.ID, opts)
	return nil
}

const PostMeasurementTypeHttp = "http"

// buildHttpMeasurementRequest builds the measurement request for the http type
func (r *Root) buildHttpMeasurementRequest(cmd *cobra.Command) (*globalping.MeasurementCreate, error) {
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
	if r.ctx.Full && method == "" {
		// override method to GET unless it was specified by the user
		method = "GET"
	}
	opts.Target = urlData.Host
	opts.Options = &globalping.MeasurementOptions{
		Protocol: urlData.Protocol,
		Request: &globalping.RequestOptions{
			Path:    overrideOpt(urlData.Path, r.ctx.Path),
			Query:   overrideOpt(urlData.Query, r.ctx.Query),
			Host:    r.ctx.Host,
			Headers: headers,
			Method:  method,
		},
		Resolver: r.ctx.Resolver,
	}
	protocolFlag := cmd.Flag("protocol")
	if protocolFlag != nil && protocolFlag.Changed {
		opts.Options.Protocol = r.ctx.Protocol
	}
	if urlData.HasPort {
		portFlag := cmd.Flag("port")
		if portFlag != nil && portFlag.Changed {
			opts.Options.Port = r.ctx.Port
		} else {
			opts.Options.Port = urlData.Port
		}

	} else {
		opts.Options.Port = r.ctx.Port
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
	Port     uint16
	HasPort  bool
}

// parse url data from user text input
func parseUrlData(input string) (*UrlData, error) {
	var urlData UrlData

	// add url scheme if missing
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		input = "https://" + input
	}

	// Parse input
	u, err := url.Parse(input)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse url input")
	}

	urlData.Protocol = strings.ToUpper(u.Scheme)
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
		port, err := strconv.ParseUint(p, 10, 16)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse url port number: %s", p)
		}
		urlData.Port = uint16(port)
		urlData.HasPort = true
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
