package cmd

import (
	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/spf13/cobra"
)

func (r *Root) initDNS() {
	dnsCmd := &cobra.Command{
		RunE:    r.RunDNS,
		Use:     "dns [target] from [location | measurement ID | @1 | first | @-1 | last | previous]",
		GroupID: "Measurements",
		Short:   "Resolve DNS records, similar to the dig command",
		Long: `The dns command (similar to the "dig" command) performs DNS lookups and displays the responses from the queried name servers, helping you troubleshoot DNS-related issues.
Note that a probe's local settings or DHCP determine the default nameserver the command uses. To specify a DNS resolver, use the --resolver argument or @resolver format: 
- dns jsdelivr.com from Berlin --resolver 1.1.1.1
- dns jsdelivr.com @1.1.1.1 from Berlin

Examples:
  # Resolve google.com from 2 probes in New York.
  dns google.com from New York --limit 2

  # Resolve google.com using probes from a previous measurement by using its ID.
  dns google.com from rvasVvKnj48cxNjC

  # Resolve google.com using the same probes from the first measurement in this session.
  dns google.com from @1

  # Resolve google.com using the same probes from the last measurement in this session.
  dns google.com from last

  # Resolve google.com using the same probes from the second-to-last measurement in this session.
  dns google.com from @-2

  # Resolve google.com from 2 probes from London or Belgium with trace enabled.
  dns google.com from London,Belgium --limit 2 --trace

  # Resolve google.com from a probe in Paris using the TCP protocol.
  dns google.com from Paris --protocol tcp

  # Resolve the MX records for jsdelivr.com from a probe in Berlin with the resolver 1.1.1.1 and enable CI mode.
  dns jsdelivr.com from Berlin --type MX --resolver 1.1.1.1 --ci

  # Resolve jsdelivr.com from a probe on the AWS network located in Montreal and display only latency information.
  dns jsdelivr.com from aws+montreal --latency

  # Resolve jsdelivr.com from a probe in ASN 123 and output the results in JSON format.
  dns jsdelivr.com from 123 --json
  
  # Resolve jsdelivr.com from a non-data center probe in Europe and add a link to view the results online..
  dns jsdelivr.com from europe+eyeball --share`,
	}

	// dns specific flags
	flags := dnsCmd.Flags()
	flags.StringVar(&r.ctx.Protocol, "protocol", r.ctx.Protocol, "specify the protocol to use for the DNS query: TCP or UDP (default \"udp\")")
	flags.IntVar(&r.ctx.Port, "port", r.ctx.Port, "specify a non-standard port on the server to send the query to (default 53)")
	flags.StringVar(&r.ctx.Resolver, "resolver", r.ctx.Resolver, "specify the hostname or IP address of the name server to use as the resolver (default defined by the probe)")
	flags.StringVar(&r.ctx.QueryType, "type", r.ctx.QueryType, "specify the type of DNS query to perform (default \"A\")")
	flags.BoolVar(&r.ctx.Trace, "trace", r.ctx.Trace, "enable tracing of the delegation path from the root name servers (default false)")

	r.Cmd.AddCommand(dnsCmd)
}

func (r *Root) RunDNS(cmd *cobra.Command, args []string) error {
	err := r.updateContext(cmd.CalledAs(), args)
	if err != nil {
		return err
	}

	defer r.UpdateHistory()
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
