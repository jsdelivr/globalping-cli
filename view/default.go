package view

import (
	"strings"
	"time"

	"github.com/jsdelivr/globalping-cli/globalping"
)

// Outputs non-json non-latency results for a measurement
func (v *viewer) outputDefault(id string, data *globalping.Measurement, m *globalping.MeasurementCreate) {
	for i := range data.Results {
		result := &data.Results[i]
		if i > 0 {
			// new line as separator if more than 1 result
			v.printer.Println()
		}

		v.printer.ErrPrintln(v.getProbeInfo(result))

		if v.ctx.Cmd == "http" {
			if v.ctx.Full {
				if result.Result.ResolvedAddress != "" {
					v.printer.ErrPrintf("Resolved address: %s\n\n", result.Result.ResolvedAddress)
				}

				tls := result.Result.TLS
				if tls != nil {
					colorize := func(s string) string {
						if tls.Authorized {
							return s
						}
						return v.printer.Color(s, FGRed)
					}

					v.printer.ErrPrintf(colorize("%s/%s\n"), tls.Protocol, tls.ChipherName)

					if tls.Authorized == false {
						v.printer.ErrPrintf(colorize("Error: %s\n"), tls.Error)
					}

					v.printer.ErrPrintf(colorize("Subject: %s; %s\n"), tls.Subject.CommonName, tls.Subject.AlternativeName)
					v.printer.ErrPrintf(colorize("Issuer: %s; %s; %s\n"), tls.Issuer.CommonName, tls.Issuer.Organization, tls.Issuer.Country)
					v.printer.ErrPrintf(colorize("Validity: %s; %s\n"), tls.CreatedAt.Format(time.RFC3339), tls.ExpiresAt.Format(time.RFC3339))
					v.printer.ErrPrintf(colorize("Serial number: %s\n"), tls.SerialNumber)
					v.printer.ErrPrintf(colorize("Fingerprint: %s\n"), tls.Fingerprint256)
					v.printer.ErrPrintf(colorize("Key type: %s%d\n"), tls.KeyType, tls.KeyBits)
					v.printer.ErrPrintln()
				}
				firstLineEnd := strings.Index(result.Result.RawOutput, "\n")
				if firstLineEnd > 0 {
					v.printer.ErrPrintln(result.Result.RawOutput[:firstLineEnd])
				}
				v.printer.ErrPrintln(result.Result.RawHeaders)
				if m.Options.Request.Method == "GET" {
					v.printer.ErrPrintln()
					v.printer.Println(strings.TrimSpace(result.Result.RawBody))
				}
			} else if m.Options.Request.Method == "GET" {
				v.printer.Println(strings.TrimSpace(result.Result.RawBody))
			} else {
				v.printer.Println(strings.TrimSpace(result.Result.RawOutput))
			}
		} else {
			v.printer.Println(strings.TrimSpace(result.Result.RawOutput))
		}
	}

	if v.ctx.Share {
		v.printer.ErrPrintln(v.getShareMessage(id))
	}
}
