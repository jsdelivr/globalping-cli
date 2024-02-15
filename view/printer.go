package view

import (
	"fmt"
	"io"

	"github.com/pterm/pterm"
)

type Printer struct {
	w io.Writer
}

func NewPrinter(writer io.Writer) *Printer {
	pterm.SetDefaultOutput(writer) // TODO: Set writer for AreaPrinter
	return &Printer{
		w: writer,
	}
}

func (p *Printer) Print(a ...any) {
	fmt.Fprint(p.w, a...)
}

func (p *Printer) Println(a ...any) {
	fmt.Fprintln(p.w, a...)
}

func (p *Printer) Printf(format string, a ...any) {
	fmt.Fprintf(p.w, format, a...)
}
