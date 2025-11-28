package view

import (
	"github.com/jsdelivr/globalping-cli/utils"
	"github.com/jsdelivr/globalping-go"
)

type Viewer interface {
	OutputDefault(id string, measurement *globalping.Measurement, opts *globalping.MeasurementCreate)
	OutputJSON(id string, measurement []byte)
	OutputLatency(id string, measurement *globalping.Measurement) error
	OutputInfinite(measurement *globalping.Measurement) error
	OutputLive(measurement *globalping.Measurement, opts *globalping.MeasurementCreate, w, h int)
	OutputSummary()
	OutputShare()
}

type viewer struct {
	ctx     *Context
	printer *Printer
	utils   utils.Utils
}

func NewViewer(
	ctx *Context,
	printer *Printer,
	utils utils.Utils,
) Viewer {
	return &viewer{
		ctx:     ctx,
		printer: printer,
		utils:   utils,
	}
}
