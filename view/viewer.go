package view

import (
	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/utils"
)

type Viewer interface {
	Output(id string, m *globalping.MeasurementCreate) error
	OutputInfinite(m *globalping.Measurement) error
	OutputSummary()
	OutputShare()
}

type viewer struct {
	ctx        *Context
	printer    *Printer
	time       utils.Time
	globalping globalping.Client
}

func NewViewer(
	ctx *Context,
	printer *Printer,
	time utils.Time,
	globalpingClient globalping.Client,
) Viewer {
	return &viewer{
		ctx:        ctx,
		printer:    printer,
		time:       time,
		globalping: globalpingClient,
	}
}
