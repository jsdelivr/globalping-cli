package cmd

import (
	"fmt"
	"net/http"
	"syscall"
	"time"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/spf13/cobra"
)

func (r *Root) initPing() {
	pingCmd := &cobra.Command{
		RunE:    r.RunPing,
		Use:     "ping [target] from [location | measurement ID | @1 | first | @-1 | last | previous]",
		GroupID: "Measurements",
		Short:   "Perform a ping test",
		Long: `The ping command checks a target's reachability by sending small data packets. Use it to test network latency and stability, as well as obtain information about packet loss and round-trip times.

Examples:
  # Ping google.com from 2 probes in New York.
  ping google.com from New York --limit 2

  # Ping google.com using probes from a previous measurement by using its ID.
  ping google.com from rvasVvKnj48cxNjC

  # Ping google.com using the same probes from the first measurement in this session.
  ping google.com from @1

  # Ping google.com using the same probes from the last measurement in this session.
  ping google.com from last

  # Ping google.com using the same probes from the second-to-last measurement in this session.
  ping google.com from @-2

  # Ping 1.1.1.1 from 2 probes in the USA or Belgium. Send 10 packets and enable CI mode.
  ping 1.1.1.1 from USA,Belgium --limit 2 --packets 10 --ci

  # Ping jsdelivr.com from a probe on the AWS network located in Montreal and display only latency information.
  ping jsdelivr.com from aws+montreal --latency

  # Ping jsdelivr.com from a probe in ASN 123 and output the results in JSON format.
  ping jsdelivr.com from 123 --json

  # Ping jsdelivr.com from a non-data center probe in Europe and add a link to view the results online.
  ping jsdelivr.com from europe+eyeball --share

  # Start a continuous ping to google.com from a probe in New York.
  ping google.com from New York --infinite`,
	}

	// ping specific flags
	flags := pingCmd.Flags()
	flags.IntVar(&r.ctx.Packets, "packets", r.ctx.Packets, "specify the number of ECHO_REQUEST packets to send (default 3)")
	flags.BoolVar(&r.ctx.Infinite, "infinite", r.ctx.Infinite, "enable continuous pinging of the target until manually stopped (default false)")

	r.Cmd.AddCommand(pingCmd)
}

func (r *Root) RunPing(cmd *cobra.Command, args []string) error {
	err := r.updateContext(cmd.CalledAs(), args)
	if err != nil {
		return err
	}

	defer r.UpdateHistory()
	r.ctx.RecordToSession = true
	if r.ctx.Infinite {
		r.ctx.Packets = 16
	}

	opts := &globalping.MeasurementCreate{
		Type:              "ping",
		Target:            r.ctx.Target,
		Limit:             r.ctx.Limit,
		InProgressUpdates: !r.ctx.CIMode,
		Options: &globalping.MeasurementOptions{
			Packets: r.ctx.Packets,
		},
	}
	opts.Locations, err = r.getLocations()
	if err != nil {
		r.Cmd.SilenceUsage = true
		return err
	}

	if r.ctx.Ipv4 {
		opts.Options.IPVersion = globalping.IPVersion4
	} else if r.ctx.Ipv6 {
		opts.Options.IPVersion = globalping.IPVersion6
	}

	if r.ctx.Infinite {
		return r.pingInfinite(opts)
	}

	hm, err := r.createMeasurement(opts)
	if err != nil {
		return err
	}
	return r.viewer.Output(hm.Id, opts)
}

func (r *Root) pingInfinite(opts *globalping.MeasurementCreate) error {
	if r.ctx.Limit > 5 {
		return fmt.Errorf("continuous mode is currently limited to 5 probes")
	}

	var err error
	go func() {
		err = r.ping(opts)
		if err != nil {
			r.cancel <- syscall.SIGINT
			return
		}
	}()
	<-r.cancel

	r.viewer.OutputSummary()
	if err != nil && r.ctx.MeasurementsCreated > 0 {
		e, ok := err.(*globalping.MeasurementError)
		if ok && e.Code == http.StatusTooManyRequests {
			r.Cmd.SilenceErrors = true
			if r.ctx.CIMode {
				r.printer.Printf("> %s\n", e.Message)
			} else {
				r.printer.Printf(r.printer.Color("> "+e.Message, view.ColorLightYellow) + "\n")
			}
		}
	}
	r.viewer.OutputShare()
	return err
}

func (r *Root) ping(opts *globalping.MeasurementCreate) error {
	var runErr error
	mbuf := NewMeasurementsBuffer(10) // 10 is the maximum number of measurements that can be in progress at the same time
	r.ctx.RunSessionStartedAt = r.time.Now()
	for {
		mbuf.Restart()
		elapsedTime := time.Duration(0)
		el := mbuf.Next()
		for el != nil {
			m, err := r.client.GetMeasurement(el.Id)
			if err != nil {
				r.Cmd.SilenceUsage = true
				return err
			}
			el.Status = m.Status
			if len(m.Results) == 0 {
				el = mbuf.Next()
				continue
			}
			err = r.viewer.OutputInfinite(m)
			if err != nil {
				r.Cmd.SilenceUsage = true
				return err
			}
			if m.Status != globalping.StatusInProgress {
				mbuf.Remove(el)
			} else {
				el.ProbeStatus = make([]globalping.MeasurementStatus, len(m.Results))
				for i := range m.Results {
					el.ProbeStatus[i] = m.Results[i].Result.Status
				}
			}
			if runErr == nil && mbuf.CanAppend() {
				opts.Locations = []globalping.Locations{{Magic: r.ctx.History.Last().Id}}
				start := r.time.Now()
				hm, err := r.createMeasurement(opts)
				if err != nil {
					runErr = err // Return the error after all measurements have finished
				} else {
					mbuf.Append(hm)
				}
				elapsedTime += r.time.Now().Sub(start)
			}
			el = mbuf.Next()
		}
		if mbuf.Len() > 0 {
			time.Sleep(r.ctx.APIMinInterval - elapsedTime)
			continue
		}
		if runErr != nil {
			return runErr
		}
		last := r.ctx.History.Last()
		if last != nil {
			opts.Locations = []globalping.Locations{{Magic: r.ctx.History.Last().Id}}
		}
		hm, err := r.createMeasurement(opts)
		if err != nil {
			return err
		}
		mbuf.Append(hm)
	}
}

func (r *Root) createMeasurement(opts *globalping.MeasurementCreate) (*view.HistoryItem, error) {
	res, err := r.client.CreateMeasurement(opts)
	if err != nil {
		r.Cmd.SilenceUsage = silenceUsageOnCreateMeasurementError(err)
		return nil, err
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
	return hm, nil
}

type MeasurementsBuffer struct {
	capacity int
	items    []*view.HistoryItem
	pos      int
}

func NewMeasurementsBuffer(capacity int) *MeasurementsBuffer {
	return &MeasurementsBuffer{
		capacity: capacity,
		items:    make([]*view.HistoryItem, 0, capacity),
	}
}

func (b *MeasurementsBuffer) Len() int {
	return len(b.items)
}

func (b *MeasurementsBuffer) Next() *view.HistoryItem {
	if b.pos >= len(b.items) {
		return nil
	}
	b.pos++
	return b.items[b.pos-1]
}

func (b *MeasurementsBuffer) Restart() {
	b.pos = 0
}

func (b *MeasurementsBuffer) Append(hm *view.HistoryItem) {
	b.items = append(b.items, hm)
}

func (b *MeasurementsBuffer) Remove(el *view.HistoryItem) {
	if len(b.items) == 0 {
		return
	}
	newb := make([]*view.HistoryItem, 0, b.capacity)
	for i, item := range b.items {
		if item != el {
			newb = append(newb, item)
		} else if i < b.pos {
			b.pos--
		}
	}
	b.items = newb
}

func (b *MeasurementsBuffer) CanAppend() bool {
	if len(b.items) >= b.capacity {
		return false
	}
	if len(b.items) == 0 {
		return true
	}
	// If there is at least one probe that has finished in all measurements then we can append
	inProgressMat := make([]bool, len(b.items[0].ProbeStatus))
	for i := range b.items {
		if len(b.items[i].ProbeStatus) == 0 {
			return false
		}
		for j := range b.items[i].ProbeStatus {
			inProgressMat[j] = inProgressMat[j] || b.items[i].ProbeStatus[j] != globalping.StatusFinished
		}
	}
	for _, inProgress := range inProgressMat {
		if !inProgress {
			return true
		}
	}
	return false
}
