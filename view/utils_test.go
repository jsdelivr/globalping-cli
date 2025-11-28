package view

import (
	"encoding/json"
	"time"

	"github.com/jsdelivr/globalping-go"
)

var (
	measurementID1 = "1zGzfAGL7sZfUs3c"
	measurementID2 = "2aZfUs3cnzGzfAGL"
	// measurementID3 = "3sZfUs3cnzGz1I20"

	defaultCurrentTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
)

func createPingMeasurement(id string) *globalping.Measurement {
	return &globalping.Measurement{
		ID:          id,
		Type:        "ping",
		Status:      globalping.StatusFinished,
		CreatedAt:   "2024-01-18T14:09:41.250Z",
		UpdatedAt:   "2024-01-18T14:09:41.488Z",
		Target:      "cdn.jsdelivr.net",
		ProbesCount: 1,
		Results: []globalping.ProbeMeasurement{
			{
				Probe: globalping.ProbeDetails{
					Continent: "EU",
					Region:    "Western Europe",
					Country:   "DE",
					State:     "",
					City:      "Berlin",
					ASN:       3320,
					Network:   "Deutsche Telekom AG",
					Tags:      []string{"eyeball-network"},
				},
				Result: globalping.ProbeResult{
					Status: globalping.StatusFinished,
					RawOutput: `PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.
64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=60 time=17.6 ms

--- jsdelivr.map.fastly.net ping statistics ---
1 packets transmitted, 1 received, 0% packet loss, time 1000ms
rtt min/avg/max/mdev = 17.639/17.639/17.639/0.123 ms`,
					ResolvedAddress:  "151.101.1.229",
					ResolvedHostname: "151.101.1.229",
					StatsRaw:         json.RawMessage(`{"min":17.639,"avg":17.639,"max":17.639,"total":1,"rcv":1,"drop":0,"loss":0}`),
					TimingsRaw:       json.RawMessage(`[{"ttl":60,"rtt":17.639}]`),
				},
			},
		},
	}
}

func createPingMeasurement_MultipleProbes(id string) *globalping.Measurement {
	return &globalping.Measurement{
		ID:          id,
		Type:        "ping",
		Status:      globalping.StatusFinished,
		CreatedAt:   "2024-01-18T14:17:41.471Z",
		UpdatedAt:   "2024-01-18T14:17:41.571Z",
		Target:      "cdn.jsdelivr.net",
		ProbesCount: 3,
		Results: []globalping.ProbeMeasurement{
			{
				Probe: globalping.ProbeDetails{
					Continent: "EU",
					Region:    "Northern Europe",
					Country:   "GB",
					State:     "",
					City:      "London",
					Network:   "OVH SAS",
					Tags:      []string{"datacenter-network"},
				},
				Result: globalping.ProbeResult{
					Status: globalping.StatusFinished,
					RawOutput: `PING  (146.75.73.229) 56(84) bytes of data.
64 bytes from 146.75.73.229 (146.75.73.229): icmp_seq=1 ttl=52 time=0.770 ms

--  ping statistics ---
1 packets transmitted, 1 received, 0% packet loss, time 100ms
rtt min/avg/max/mdev = 0.770/0.770/0.770/0.001 ms`,
					ResolvedAddress:  "146.75.73.229",
					ResolvedHostname: "146.75.73.229",
					StatsRaw:         json.RawMessage(`{"min":0.77,"avg":0.77,"max":0.77,"total":1,"rcv":1,"drop":0,"loss":0}`),
					TimingsRaw:       json.RawMessage(`[{"ttl":52,"rtt":0.77}]`),
				},
			},
			{
				Probe: globalping.ProbeDetails{
					Continent: "EU",
					Region:    "Western Europe",
					Country:   "DE",
					State:     "",
					City:      "Falkenstein",
					Network:   "Hetzner Online GmbH",
					Tags:      []string{"datacenter-network"},
				},
				Result: globalping.ProbeResult{
					Status: globalping.StatusFinished,
					RawOutput: `PING  (104.16.85.20) 56(84) bytes of data.
64 bytes from 104.16.85.20 (104.16.85.20): icmp_seq=1 ttl=55 time=5.46 ms

---  ping statistics ---
1 packets transmitted, 1 received, 0% packet loss, time 200ms
rtt min/avg/max/mdev = 5.457/5.457/5.457/0.002 ms`,
					ResolvedAddress:  "104.16.85.20",
					ResolvedHostname: "104.16.85.20",
					StatsRaw:         json.RawMessage(`{"min":5.457,"avg":5.457,"max":5.457,"total":1,"rcv":1,"drop":0,"loss":0}`),
					TimingsRaw:       json.RawMessage(`[{"ttl":55,"rtt":5.46}]`),
				},
			},
			{
				Probe: globalping.ProbeDetails{
					Continent: "EU",
					Region:    "Western Europe",
					Country:   "DE",
					State:     "",
					City:      "Nuremberg",
					Network:   "Hetzner Online GmbH",
					Tags:      []string{"datacenter-network"},
				},
				Result: globalping.ProbeResult{
					Status: globalping.StatusFinished,
					RawOutput: `PING  (104.16.88.20) 56(84) bytes of data.
64 bytes from 104.16.88.20 (104.16.88.20): icmp_seq=1 ttl=58 time=4.07 ms

---  ping statistics ---
1 packets transmitted, 1 received, 0% packet loss, time 300ms
rtt min/avg/max/mdev = 4.069/4.069/4.069/0.003 ms`,
					ResolvedAddress:  "104.16.88.20",
					ResolvedHostname: "104.16.88.20",
					StatsRaw:         json.RawMessage(`{"min":4.069,"avg":4.069,"max":4.069,"total":1,"rcv":1,"drop":0,"loss":0}`),
					TimingsRaw:       json.RawMessage(`[{"ttl":58,"rtt":4.07}]`),
				},
			},
		},
	}
}

func createDefaultContext(cmd string) *Context {
	ctx := &Context{
		Cmd:                 cmd,
		MeasurementsCreated: 1,
		History:             NewHistoryBuffer(3),
		RunSessionStartedAt: defaultCurrentTime,
	}
	if cmd == "ping" {
		ctx.History.Push(&HistoryItem{
			Id:        measurementID1,
			StartedAt: defaultCurrentTime,
		})
	}
	return ctx
}
