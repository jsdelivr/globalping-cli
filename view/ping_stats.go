package view

import (
	"bufio"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jsdelivr/globalping-go"
)

type MeasurementStats struct {
	Sent  int     // Number of packets sent
	Rcv   int     // Number of packets received
	Lost  int     // Number of packets lost
	Loss  float64 // Percentage of packets lost
	Last  float64 // Last RTT
	Min   float64 // Minimum RTT
	Avg   float64 // Average RTT
	Max   float64 // Maximum RTT
	Mdev  float64 // Mean deviation of RTT
	Time  float64 // Total time of measurement, in milliseconds
	Tsum  float64 // Total sum of RTT
	Tsum2 float64 // Total sum of RTT squared
}

func NewMeasurementStats() *MeasurementStats {
	return &MeasurementStats{Last: -1, Min: math.MaxFloat64, Avg: -1, Max: -1}
}

func decodePingMeasurementStats(result *globalping.ProbeResult) *MeasurementStats {
	stats, err := globalping.DecodePingStats(result.StatsRaw)
	timings, timingsErr := globalping.DecodePingTimings(result.TimingsRaw)
	decoded := NewMeasurementStats()

	if err != nil || stats.Total == 0 {
		if timingsErr != nil || len(timings) == 0 {
			return nil
		}

		// A command timeout can leave replies but no final summary.
		for _, timing := range timings {
			decoded.Min = math.Min(decoded.Min, timing.RTT)
			decoded.Max = math.Max(decoded.Max, timing.RTT)
			decoded.Tsum += timing.RTT
			decoded.Tsum2 += timing.RTT * timing.RTT
		}

		decoded.Sent = len(timings)
		decoded.Rcv = len(timings)
		decoded.Last = timings[len(timings)-1].RTT
		decoded.Avg = decoded.Tsum / float64(decoded.Rcv)
		decoded.Mdev = computeMdev(decoded.Tsum, decoded.Tsum2, decoded.Rcv, decoded.Avg)

		return decoded
	}

	decoded.Sent = stats.Total
	decoded.Rcv = stats.Rcv
	decoded.Lost = stats.Drop
	decoded.Loss = stats.Loss

	if stats.Rcv == 0 {
		return decoded
	}

	decoded.Mdev = stats.Mdev

	if stats.Min != nil {
		decoded.Min = *stats.Min
	}

	if stats.Max != nil {
		decoded.Max = *stats.Max
	}

	if stats.Avg != nil {
		decoded.Avg = *stats.Avg
		decoded.Tsum = *stats.Avg * float64(stats.Rcv)
		decoded.Tsum2 = float64(stats.Rcv) * (stats.Mdev*stats.Mdev + (*stats.Avg)*(*stats.Avg))
	}

	if timingsErr == nil && len(timings) > 0 {
		decoded.Last = timings[len(timings)-1].RTT
	}

	return decoded
}

func mergeMeasurementStats(stats MeasurementStats, newStats *MeasurementStats) *MeasurementStats {
	if newStats.Rcv > 0 {
		hadReceivedPackets := stats.Rcv > 0
		minUnavailable := newStats.Min == math.MaxFloat64 || (hadReceivedPackets && stats.Min == math.MaxFloat64)
		avgUnavailable := newStats.Avg == -1 || (hadReceivedPackets && stats.Avg == -1)
		maxUnavailable := newStats.Max == -1 || (hadReceivedPackets && stats.Max == -1)

		if newStats.Min < stats.Min {
			stats.Min = newStats.Min
		}

		if newStats.Max > stats.Max {
			stats.Max = newStats.Max
		}

		stats.Tsum += newStats.Tsum
		stats.Tsum2 += newStats.Tsum2
		stats.Rcv += newStats.Rcv
		stats.Avg = stats.Tsum / float64(stats.Rcv)
		stats.Mdev = computeMdev(stats.Tsum, stats.Tsum2, stats.Rcv, stats.Avg)
		stats.Last = newStats.Last

		if minUnavailable {
			stats.Min = math.MaxFloat64
		}

		if avgUnavailable {
			stats.Avg = -1
			stats.Mdev = 0
		}

		if maxUnavailable {
			stats.Max = -1
		}
	}

	stats.Sent += newStats.Sent
	stats.Lost += newStats.Lost
	stats.Time += newStats.Time

	if stats.Sent > 0 {
		stats.Loss = float64(stats.Lost) / float64(stats.Sent) * 100
	}

	return &stats
}

type ParsedPingOutput struct {
	Header         string
	Hostname       string
	Address        string
	BytesOfData    string
	RawPacketLines []string
	Timings        []globalping.PingTiming
	Stats          *MeasurementStats
}

// Parse ping's raw output. Adapted from iputils ping: https://github.com/iputils/iputils/tree/1c08152/ping
//
// - If startSequence is -1, RawPacketLines will be empty
func parsePingRawOutput(
	protocol string,
	startedAt time.Time,
	now func() time.Time,
	m *globalping.ProbeMeasurement,
	startSequence int,
) *ParsedPingOutput {
	res := &ParsedPingOutput{
		Timings: make([]globalping.PingTiming, 0),
		Stats:   NewMeasurementStats(),
	}

	if m.Result.RawOutput == "" {
		return res
	}

	scanner := bufio.NewScanner(strings.NewReader(m.Result.RawOutput))
	scanner.Scan()
	res.Header = scanner.Text()
	words := strings.Split(res.Header, " ")

	if len(words) > 2 {
		res.Hostname = words[1]

		if len(words[2]) > 1 && words[2][0] == '(' {
			res.Address = words[2][1 : len(words[2])-1]
		} else {
			res.Address = words[2]
		}

		if protocol == "ICMP" {
			res.BytesOfData = words[3]
		}
	}

	sentMap := make([]bool, 0)

	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 {
			break
		}

		if protocol == "TCP" {
			line, sentMap = parseTCPLine(line, sentMap, startSequence, res)
		} else {
			line, sentMap = parseICMPLine(line, sentMap, startSequence, res)
		}

		if startSequence != -1 {
			res.RawPacketLines = append(res.RawPacketLines, line)
		}
	}

	hasSummary := scanner.Scan()

	if hasSummary {
		// Parse summary
		scanner.Scan() // skip ---  ping statistics ---
		line := scanner.Text()
		words = strings.Split(line, " ")

		if len(words) > 9 && words[1] == "packets" && words[2] == "transmitted," {
			res.Stats.Sent, _ = strconv.Atoi(words[0])
			res.Stats.Rcv, _ = strconv.Atoi(words[3])

			if protocol == "TCP" {
				res.Stats.Time, _ = strconv.ParseFloat(words[9], 64)
			} else {
				res.Stats.Time, _ = strconv.ParseFloat(words[9][:len(words[9])-2], 64)
			}
		}
	} else {
		res.Stats.Time = float64(now().Sub(startedAt).Milliseconds())
	}

	if res.Stats.Sent > 0 {
		res.Stats.Lost = res.Stats.Sent - res.Stats.Rcv
		res.Stats.Loss = float64(res.Stats.Lost) / float64(res.Stats.Sent) * 100

		if res.Stats.Rcv > 0 {
			res.Stats.Avg = res.Stats.Tsum / float64(res.Stats.Rcv)
			res.Stats.Mdev = computeMdev(res.Stats.Tsum, res.Stats.Tsum2, res.Stats.Rcv, res.Stats.Avg)
			res.Stats.Last = res.Timings[len(res.Timings)-1].RTT
		}
	}

	return res
}

func parseICMPLine(line string, sentMap []bool, startSequence int, res *ParsedPingOutput) (string, []bool) {
	seq := -1
	seqIndex := 0
	words := strings.Split(line, " ")

	for seqIndex < len(words) {
		if strings.HasPrefix(words[seqIndex], "icmp_seq=") {
			n, err := strconv.Atoi(words[seqIndex][9:])

			if err == nil {
				seq = n - 1 // seq starts at 1
			}

			break
		}

		seqIndex++
	}

	if seq >= len(sentMap) {
		sentMap = append(sentMap, false)
	}

	// Get timing
	if seq != -1 {
		if words[1] == "bytes" && words[2] == "from" {
			if !sentMap[seq] {
				res.Stats.Sent++
			}

			res.Stats.Rcv++
			ttl, _ := strconv.Atoi(words[seqIndex+1][4:])
			rtt, _ := strconv.ParseFloat(words[seqIndex+2][5:], 64)
			res.Stats.Min = math.Min(res.Stats.Min, rtt)
			res.Stats.Max = math.Max(res.Stats.Max, rtt)
			res.Stats.Tsum += rtt
			res.Stats.Tsum2 += rtt * rtt
			res.Timings = append(res.Timings, globalping.PingTiming{
				TTL: &ttl,
				RTT: rtt,
			})
		} else {
			if !sentMap[seq] {
				res.Stats.Sent++
			}

			sentMap[seq] = true
		}

		// replace sequence number
		if startSequence != -1 {
			words[seqIndex] = "icmp_seq=" + strconv.Itoa(startSequence+seq+1)
			line = strings.Join(words, " ")
		}
	}

	return line, sentMap
}

func parseTCPLine(line string, sentMap []bool, startSequence int, res *ParsedPingOutput) (string, []bool) {
	seq := -1
	seqIndex := 0
	words := strings.Split(line, " ")

	for seqIndex < len(words) {
		if strings.HasPrefix(words[seqIndex], "tcp_conn=") {
			n, err := strconv.Atoi(words[seqIndex][9:])

			if err == nil {
				seq = n - 1 // seq starts at 1
			}

			break
		}

		seqIndex++
	}

	if seq >= len(sentMap) {
		sentMap = append(sentMap, false)
	}

	// Get timing
	if seq != -1 {
		if words[0] == "Reply" && words[1] == "from" {
			if !sentMap[seq] {
				res.Stats.Sent++
			}

			res.Stats.Rcv++
			rtt, _ := strconv.ParseFloat(words[seqIndex+1][5:], 64)
			res.Stats.Min = math.Min(res.Stats.Min, rtt)
			res.Stats.Max = math.Max(res.Stats.Max, rtt)
			res.Stats.Tsum += rtt
			res.Stats.Tsum2 += rtt * rtt
			res.Timings = append(res.Timings, globalping.PingTiming{
				RTT: rtt,
			})
		} else {
			if !sentMap[seq] {
				res.Stats.Sent++
			}

			sentMap[seq] = true
		}

		// replace sequence number
		if startSequence != -1 {
			words[seqIndex] = "tcp_conn=" + strconv.Itoa(startSequence+seq+1)
			line = strings.Join(words, " ")
		}
	}

	return line, sentMap
}

// https://github.com/iputils/iputils/tree/1c08152/ping/ping_common.c#L917
func computeMdev(tsum float64, tsum2 float64, rcv int, avg float64) float64 {
	if tsum < math.MaxInt32 {
		return math.Sqrt((tsum2 - ((tsum * tsum) / float64(rcv))) / float64(rcv))
	}

	return math.Sqrt(tsum2/float64(rcv) - avg*avg)
}
