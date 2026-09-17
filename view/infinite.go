package view

import (
	"fmt"

	"github.com/jsdelivr/globalping-cli/utils"
	"github.com/jsdelivr/globalping-go"
)

const apiCreditInfo = "Consuming 1 API credit for every 16 packets until stopped.\n"

func (v *viewer) OutputInfinite(output *InfinitePingOutput) (string, error) {
	m := output.Measurement

	if !output.streamRawOutput && !v.ctx.ToLatency {
		v.ctx.Table = true
	}

	if m.Status != globalping.MeasurementStatusInProgress && !isSomeTestFinished(m) {
		if !output.streamRawOutput {
			v.clearInfiniteTableOutput()
		}

		return "", v.outputFailSummary(m)
	}

	if output.streamRawOutput {
		if output.initial {
			v.printer.ErrPrint(v.printer.Color(apiCreditInfo, FGBrightYellow))
		}

		if output.showHeader {
			v.printer.ErrPrintln(v.getProbeInfo(&m.Results[0]))

			if v.ctx.Protocol == "ICMP" {
				v.printer.Printf("PING %s (%s) %s bytes of data.\n",
					output.packets.Hostname, output.packets.Address, output.packets.BytesOfData)
			} else {
				v.printer.Println(output.packets.Header)
			}
		}

		if output.packets != nil {
			for _, line := range output.packets.RawPacketLines {
				v.printer.Println(line)
			}
		}

		return "", nil
	}

	if v.ctx.ToLatency {
		return v.outputInfinitePingLatencyTable(m, output)
	}

	return v.outputInfinitePingTableView(output)
}

func (v *viewer) getAPICreditConsumptionInfo(output *InfinitePingOutput, width int) string {
	if !output.showCredits {
		return ""
	}

	info := fmt.Sprintf("Consuming ~%s/minute.\n", utils.Pluralize(output.creditsPerMinute, "API credit"))

	if len(info) > width-4 {
		info = info[:max(width-5, 0)] + "..."
	}

	return v.printer.Color(info, FGBrightYellow)
}
