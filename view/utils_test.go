package view

import (
	"encoding/json"

	"github.com/jsdelivr/globalping-cli/model"
)

var (
	MeasurementID1 = "nzGzfAGL7sZfUs3c"
	MeasurementID2 = "A2ZfUs3cnzGzfAGL"
	MeasurementID3 = "7sZfUs3cnzGz1I20"
)

func getPingGetMeasurement(id string) *model.GetMeasurement {
	return &model.GetMeasurement{
		ID:          id,
		Type:        "ping",
		Status:      "finished",
		CreatedAt:   "2024-01-18T14:09:41.250Z",
		UpdatedAt:   "2024-01-18T14:09:41.488Z",
		Target:      "cdn.jsdelivr.net",
		ProbesCount: 1,
		Results: []model.MeasurementResponse{
			{
				Probe: model.ProbeData{
					Continent: "EU",
					Region:    "Western Europe",
					Country:   "DE",
					State:     "",
					City:      "Berlin",
					ASN:       3320,
					Network:   "Deutsche Telekom AG",
					Tags:      []string{"eyeball-network"},
				},
				Result: model.ResultData{
					Status:           "finished",
					RawOutput:        "PING jsdelivr.map.fastly.net (151.101.1.229) 56(84) bytes of data.\n64 bytes from 151.101.1.229 (151.101.1.229): icmp_seq=1 ttl=60 time=17.6 ms\n\n--- jsdelivr.map.fastly.net ping statistics ---\n1 packets transmitted, 1 received, 0% packet loss, time 0ms\nrtt min/avg/max/mdev = 17.639/17.639/17.639/0.000 ms",
					ResolvedAddress:  "151.101.1.229",
					ResolvedHostname: "151.101.1.229",
					StatsRaw:         json.RawMessage(`{"min":17.639,"avg":17.639,"max":17.639,"total":1,"rcv":1,"drop":0,"loss":0}`),
					TimingsRaw:       json.RawMessage(`[{"ttl":60,"rtt":17.639}]`),
				},
			},
		},
	}
}

func getPingGetMeasurementMultipleLocations(id string) *model.GetMeasurement {
	return &model.GetMeasurement{
		ID:          id,
		Type:        "ping",
		Status:      "finished",
		CreatedAt:   "2024-01-18T14:17:41.471Z",
		UpdatedAt:   "2024-01-18T14:17:41.571Z",
		Target:      "cdn.jsdelivr.net",
		ProbesCount: 3,
		Results: []model.MeasurementResponse{
			{
				Probe: model.ProbeData{
					Continent: "EU",
					Region:    "Northern Europe",
					Country:   "GB",
					State:     "",
					City:      "London",
					Network:   "OVH SAS",
					Tags:      []string{"datacenter-network"},
				},
				Result: model.ResultData{
					Status:           "finished",
					RawOutput:        "PING  (146.75.73.229) 56(84) bytes of data.\n64 bytes from 146.75.73.229 (146.75.73.229): icmp_seq=1 ttl=52 time=0.770 ms\n\n---  ping statistics ---\n1 packets transmitted, 1 received, 0% packet loss, time 0ms\nrtt min/avg/max/mdev = 0.770/0.770/0.770/0.000 ms",
					ResolvedAddress:  "146.75.73.229",
					ResolvedHostname: "146.75.73.229",
					StatsRaw:         json.RawMessage(`{"min":0.77,"avg":0.77,"max":0.77,"total":1,"rcv":1,"drop":0,"loss":0}`),
					TimingsRaw:       json.RawMessage(`[{"ttl":52,"rtt":0.77}]`),
				},
			},
			{
				Probe: model.ProbeData{
					Continent: "EU",
					Region:    "Western Europe",
					Country:   "DE",
					State:     "",
					City:      "Falkenstein",
					Network:   "Hetzner Online GmbH",
					Tags:      []string{"datacenter-network"},
				},
				Result: model.ResultData{
					Status:           "finished",
					RawOutput:        "PING  (104.16.85.20) 56(84) bytes of data.\n64 bytes from 104.16.85.20 (104.16.85.20): icmp_seq=1 ttl=55 time=5.46 ms\n\n---  ping statistics ---\n1 packets transmitted, 1 received, 0% packet loss, time 0ms\nrtt min/avg/max/mdev = 5.457/5.457/5.457/0.000 ms",
					ResolvedAddress:  "104.16.85.20",
					ResolvedHostname: "104.16.85.20",
					StatsRaw:         json.RawMessage(`{"min":5.457,"avg":5.457,"max":5.457,"total":1,"rcv":1,"drop":0,"loss":0}`),
					TimingsRaw:       json.RawMessage(`[{"ttl":55,"rtt":5.46}]`),
				},
			},
			{
				Probe: model.ProbeData{
					Continent: "EU",
					Region:    "Western Europe",
					Country:   "DE",
					State:     "",
					City:      "Nuremberg",
					Network:   "Hetzner Online GmbH",
					Tags:      []string{"datacenter-network"},
				},
				Result: model.ResultData{
					Status:           "finished",
					RawOutput:        "PING  (104.16.88.20) 56(84) bytes of data.\n64 bytes from 104.16.88.20 (104.16.88.20): icmp_seq=1 ttl=58 time=4.07 ms\n\n---  ping statistics ---\n1 packets transmitted, 1 received, 0% packet loss, time 0ms\nrtt min/avg/max/mdev = 4.069/4.069/4.069/0.000 ms",
					ResolvedAddress:  "104.16.88.20",
					ResolvedHostname: "104.16.88.20",
					StatsRaw:         json.RawMessage(`{"min":4.069,"avg":4.069,"max":4.069,"total":1,"rcv":1,"drop":0,"loss":0}`),
					TimingsRaw:       json.RawMessage(`[{"ttl":58,"rtt":4.07}]`),
				},
			},
		},
	}
}
