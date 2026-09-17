package view

import "github.com/jsdelivr/globalping-go"

type Viewer interface {
	OutputDefault(id string, measurement *globalping.Measurement, opts *globalping.MeasurementCreate)
	OutputJSON(id string, measurement []byte)
	OutputLatency(id string, measurement *globalping.Measurement) error
	OutputInfinite(output *InfinitePingOutput) (string, error)
	OutputTable(measurement *globalping.Measurement) (string, error)
	OutputComparisonTable(first, second *globalping.Measurement) (string, error)
	OutputLive(measurement *globalping.Measurement, opts *globalping.MeasurementCreate, w, h int)
	OutputPingSummary(infiniteTableOutput string, summary *PingSummary)
	OutputShare()
}

type viewer struct {
	ctx     *Context
	printer *Printer
}

func NewViewer(
	ctx *Context,
	printer *Printer,
) Viewer {
	return &viewer{
		ctx:     ctx,
		printer: printer,
	}
}
