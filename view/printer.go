package view

import (
	"fmt"
	"io"

	"github.com/pterm/pterm"
)

type Printer struct {
	InReader  io.Reader
	OutWriter io.Writer
	ErrWriter io.Writer
}

func NewPrinter(
	inReader io.Reader,
	outWriter io.Writer,
	errWriter io.Writer,
) *Printer {
	pterm.SetDefaultOutput(outWriter) // TODO: Set writer for AreaPrinter
	return &Printer{
		InReader:  inReader,
		OutWriter: outWriter,
		ErrWriter: errWriter,
	}
}

func (p *Printer) Print(a ...any) {
	fmt.Fprint(p.OutWriter, a...)
}

func (p *Printer) Println(a ...any) {
	fmt.Fprintln(p.OutWriter, a...)
}

func (p *Printer) Printf(format string, a ...any) {
	fmt.Fprintf(p.OutWriter, format, a...)
}
