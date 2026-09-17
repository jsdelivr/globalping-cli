package view

import (
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/jsdelivr/globalping-go"
)

// InfinitePingRun is owned by the command loop. Rounds are registered in creation order;
// each update produces a snapshot that can be rendered without changing the run.
type InfinitePingRun struct {
	protocol         string
	packets          int
	startedAt        time.Time
	now              func() time.Time
	rounds           []*pingRound
	completed        []*MeasurementStats
	statuses         []pingProbeStatus
	hostname         string
	headerPrinted    bool
	creditCount      int
	creditsPerMinute int64
}

type pingRound struct {
	id           string
	startedAt    time.Time
	finished     bool
	stats        []*MeasurementStats
	statuses     []pingProbeStatus
	linesPrinted int
}

type pingProbeStatus struct {
	status globalping.TestStatus
	source globalping.FailureSource
	last   *float64
}

// InfinitePingOutput contains one prepared redraw or batch of newly received packet lines.
type InfinitePingOutput struct {
	Measurement      *globalping.Measurement
	probes           []infinitePingProbeStats
	streamRawOutput  bool
	initial          bool
	showHeader       bool
	packets          *ParsedPingOutput
	creditsPerMinute int64
	showCredits      bool
}

type infinitePingProbeStats struct {
	measurement   *globalping.ProbeMeasurement
	stats         *MeasurementStats
	statusMessage string
}

type PingSummary struct {
	Hostname string
	Stats    *MeasurementStats
}

func NewInfinitePingRun(protocol string, packets int, startedAt time.Time, now func() time.Time) *InfinitePingRun {
	return &InfinitePingRun{protocol: protocol, packets: packets, startedAt: startedAt, now: now}
}

func (r *InfinitePingRun) Add(id string, startedAt time.Time) {
	r.rounds = append(r.rounds, &pingRound{id: id, startedAt: startedAt})
}

func (r *InfinitePingRun) Update(m *globalping.Measurement, streamRawOutput bool, created int) (*InfinitePingOutput, error) {
	output := &InfinitePingOutput{Measurement: m, streamRawOutput: streamRawOutput}

	if !streamRawOutput {
		if err := validateFinishedPingStats(m); err != nil {
			return nil, err
		}
	}

	index := slices.IndexFunc(r.rounds, func(round *pingRound) bool { return round.id == m.ID })

	if index < 0 {
		return nil, fmt.Errorf("ping round %s has not been registered", m.ID)
	}

	round := r.rounds[index]

	if m.Status != globalping.MeasurementStatusInProgress && !isSomeTestFinished(m) {
		round.finished = true
		round.stats = nil
		r.compactRounds()

		return output, nil
	}

	output.initial = len(r.completed) == 0

	if output.initial {
		r.completed = make([]*MeasurementStats, len(m.Results))
		r.statuses = make([]pingProbeStatus, len(m.Results))

		for i := range r.completed {
			r.completed[i] = NewMeasurementStats()
		}
	}

	if streamRawOutput {
		r.updatePackets(round, m, output)
	} else {
		r.updateTable(round, m, output)
	}

	if m.Status != globalping.MeasurementStatusInProgress {
		round.finished = true

		for i, stats := range round.stats {
			r.completed[i] = mergeMeasurementStats(*r.completed[i], stats)
		}

		round.stats = nil
		r.compactRounds()
	}

	if !streamRawOutput && created >= 2 {
		if r.creditCount != created {
			r.creditCount = created
			minutes := r.now().Sub(r.startedAt).Minutes()
			r.creditsPerMinute = int64(math.Ceil(float64((created-1)*len(r.completed)) / minutes))
		}

		output.showCredits = true
		output.creditsPerMinute = r.creditsPerMinute
	}

	return output, nil
}

func (r *InfinitePingRun) updateTable(round *pingRound, m *globalping.Measurement, output *InfinitePingOutput) {
	round.stats = make([]*MeasurementStats, len(m.Results))
	round.statuses = make([]pingProbeStatus, len(m.Results))

	for i := range m.Results {
		result := &m.Results[i]
		var stats *MeasurementStats

		if result.Result.Status == globalping.TestStatusInProgress {
			stats = parsePingRawOutput(r.protocol, round.startedAt, r.now, result, -1).Stats
		} else {
			stats = decodePingMeasurementStats(&result.Result)
		}

		if stats == nil {
			stats = NewMeasurementStats()
		}

		if result.Result.Status == globalping.TestStatusFailed || result.Result.Status == globalping.TestStatusOffline {
			if missing := r.packets - stats.Sent; missing > 0 {
				stats.Sent += missing
				stats.Lost += missing
				stats.Loss = float64(stats.Lost) / float64(stats.Sent) * 100
			}
		}

		round.stats[i] = stats
		round.statuses[i] = pingProbeStatus{status: result.Result.Status, source: result.Result.FailureSource}

		if stats.Last >= 0 {
			round.statuses[i].last = &stats.Last
		}
	}

	statuses := slices.Clone(r.statuses)

	for _, active := range r.rounds {
		mergePingStatuses(statuses, active.statuses)
	}

	output.probes = make([]infinitePingProbeStats, len(m.Results))

	for i := range m.Results {
		stats := r.aggregatePending(r.completed[i], i)
		statusMessage := ""

		if statuses[i].status == globalping.TestStatusFailed || statuses[i].status == globalping.TestStatusOffline {
			statusMessage = resultStatusLabel(&globalping.ProbeResult{Status: statuses[i].status, FailureSource: statuses[i].source})
		}

		output.probes[i] = infinitePingProbeStats{
			measurement:   &m.Results[i],
			stats:         stats,
			statusMessage: statusMessage,
		}
	}
}

func (r *InfinitePingRun) updatePackets(round *pingRound, m *globalping.Measurement, output *InfinitePingOutput) {
	result := &m.Results[0]

	if result.Result.RawOutput == "" {
		round.stats = nil

		return
	}

	sent := r.completed[0].Sent

	for _, other := range r.rounds {
		if other != round && !other.finished && len(other.stats) > 0 {
			sent += other.stats[0].Sent
		}
	}

	parsed := parsePingRawOutput(r.protocol, round.startedAt, r.now, result, sent)
	round.stats = []*MeasurementStats{parsed.Stats}
	round.statuses = make([]pingProbeStatus, 1)

	if parsed.Stats.Last >= 0 {
		round.statuses[0].last = &parsed.Stats.Last
	}

	output.showHeader = !r.headerPrinted

	if output.showHeader {
		r.hostname = parsed.Hostname
		r.headerPrinted = true
	}

	lines := parsed.RawPacketLines
	parsed.RawPacketLines = lines[min(round.linesPrinted, len(lines)):]
	round.linesPrinted = max(round.linesPrinted, len(lines))
	output.packets = parsed
}

func (r *InfinitePingRun) aggregatePending(completed *MeasurementStats, probe int) *MeasurementStats {
	stats := *completed
	var last *float64

	if len(r.statuses) > 0 {
		last = r.statuses[probe].last
	}

	for _, round := range r.rounds {
		if !round.finished && len(round.stats) > 0 {
			stats = *mergeMeasurementStats(stats, round.stats[probe])
		}

		if len(round.statuses) > 0 && round.statuses[probe].last != nil {
			last = round.statuses[probe].last
		}
	}

	if last != nil {
		stats.Last = *last
	}

	return &stats
}

// Completed round statistics are already in the totals. Keep only the latest
// terminal statuses and available RTTs between pending rounds, so retained state stays bounded.
func (r *InfinitePingRun) compactRounds() {
	rounds := r.rounds[:0]

	for _, round := range r.rounds {
		switch {
		case round.finished && len(rounds) == 0:
			mergePingStatuses(r.statuses, round.statuses)
		case round.finished && rounds[len(rounds)-1].finished:
			mergePingStatuses(rounds[len(rounds)-1].statuses, round.statuses)
		default:
			rounds = append(rounds, round)
		}
	}

	clear(r.rounds[len(rounds):])
	r.rounds = rounds
}

func mergePingStatuses(statuses, newer []pingProbeStatus) {
	for i, status := range newer {
		switch status.status {
		case globalping.TestStatusFinished, globalping.TestStatusFailed, globalping.TestStatusOffline:
			statuses[i].status = status.status
			statuses[i].source = status.source
		}

		if status.last != nil {
			statuses[i].last = status.last
		}
	}
}

func (r *InfinitePingRun) Summary() *PingSummary {
	if len(r.completed) != 1 {
		return nil
	}

	return &PingSummary{Hostname: r.hostname, Stats: r.aggregatePending(r.completed[0], 0)}
}
