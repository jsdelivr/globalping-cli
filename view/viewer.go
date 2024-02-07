package view

import "github.com/jsdelivr/globalping-cli/globalping"

type Viewer interface {
	Output(id string, m *globalping.MeasurementCreate) error
	OutputInfinite(id string) error
	OutputSummary()
}

type viewer struct {
	ctx     *Context
	printer *Printer
	gp      globalping.Client
}

func NewViewer(
	ctx *Context,
	printer *Printer,
	gp globalping.Client,
) Viewer {
	return &viewer{
		ctx:     ctx,
		printer: printer,
		gp:      gp,
	}
}
