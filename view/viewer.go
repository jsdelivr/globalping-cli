package view

import (
	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/utils"
)

type Viewer interface {
	Output(id string, m *globalping.MeasurementCreate) error
	OutputInfinite(id string) error
	OutputSummary()
}

type viewer struct {
	ctx     *Context
	printer *Printer
	time    utils.Time
	gp      globalping.Client
}

func NewViewer(
	ctx *Context,
	printer *Printer,
	time utils.Time,
	gp globalping.Client,
) Viewer {
	return &viewer{
		ctx:     ctx,
		printer: printer,
		time:    time,
		gp:      gp,
	}
}
