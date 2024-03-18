package cmd

import (
	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/spf13/cobra"
)

func (r *Root) initDNS() {
	dnsCmd := &cobra.Command{
		RunE:    r.RunDNS,
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
	}

	// dns specific flags
	flags := dnsCmd.Flags()
	flags.StringVar(&r.ctx.Protocol, "protocol", "", "Specifies the protocol to use for the DNS query (TCP or UDP) (default \"udp\")")
	flags.IntVar(&r.ctx.Port, "port", 0, "Send the query to a non-standard port on the server (default 53)")
	flags.StringVar(&r.ctx.Resolver, "resolver", "", "Resolver is the hostname or IP address of the name server to use (default empty)")
	flags.StringVar(&r.ctx.QueryType, "type", "", "Specifies the type of DNS query to perform (default \"A\")")
	flags.BoolVar(&r.ctx.Trace, "trace", false, "Toggle tracing of the delegation path from the root name servers (default false)")

	r.Cmd.AddCommand(dnsCmd)
}

func (r *Root) RunDNS(cmd *cobra.Command, args []string) error {
	err := r.updateContext(cmd.CalledAs(), args)
	if err != nil {
		return err
	}

	r.ctx.RecordToSession = true

	opts := &globalping.MeasurementCreate{
		Type:              "dns",
		Target:            r.ctx.Target,
		Limit:             r.ctx.Limit,
		InProgressUpdates: !r.ctx.CIMode,
		Options: &globalping.MeasurementOptions{
			Protocol: r.ctx.Protocol,
			Port:     r.ctx.Port,
			Resolver: r.ctx.Resolver,
			Query: &globalping.QueryOptions{
				Type: r.ctx.QueryType,
			},
			Trace: r.ctx.Trace,
		},
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
