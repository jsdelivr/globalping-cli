package view

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"golang.org/x/term"
)

type Color string

const (
	ColorNone        Color = ""
	ColorLightYellow Color = "93"
	ColorLightCyan   Color = "96"
	ColorHighlight   Color = "38;2;23;212;167"
)

type Printer struct {
	InReader  io.Reader
	OutWriter io.Writer
	ErrWriter io.Writer

	areaHeight     int
	disableStyling bool
}

func NewPrinter(
	inReader io.Reader,
	outWriter io.Writer,
	errWriter io.Writer,
) *Printer {
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

func (p *Printer) FillLeft(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return strings.Repeat(" ", w-len(s)) + s
}

func (p *Printer) FillRight(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}

func (p *Printer) FillLeftAndColor(s string, w int, color Color) string {
	if len(s) < w {
		s = strings.Repeat(" ", w-len(s)) + s
	}
	if p.disableStyling || color == ColorNone {
		return s
	}
	return p.Color(s, color)
}

func (p *Printer) FillRightAndColor(s string, w int, color Color) string {
	if len(s) < w {
		s += strings.Repeat(" ", w-len(s))
	}
	if p.disableStyling || color == ColorNone {
		return s
	}
	return p.Color(s, color)
}

func (p *Printer) Color(s string, color Color) string {
	if p.disableStyling {
		return s
	}
	return fmt.Sprintf("\033[%sm%s\033[0m", color, s)
}

func (p *Printer) Bold(s string) string {
	if p.disableStyling {
		return s
	}
	return fmt.Sprintf("\033[1m%s\033[0m", s)
}

func (p *Printer) BoldWithColor(s string, color Color) string {
	if p.disableStyling {
		return s
	}
	return fmt.Sprintf("\033[1;%sm%s\033[0m", color, s)
}

func (p *Printer) GetSize() (width, height int) {
	f, ok := p.OutWriter.(*os.File)
	if !ok {
		return math.MaxInt, math.MaxInt
	}
	w, h, _ := term.GetSize(int(f.Fd()))
	if w <= 0 {
		w = math.MaxInt
	}
	if h <= 0 {
		h = math.MaxInt
	}
	return w, h
}

func (p *Printer) AreaUpdate(content *string) {
	p.AreaClear()
	p.areaHeight = strings.Count(*content, "\n")
	fmt.Fprint(p.OutWriter, *content)
}

func (p *Printer) AreaClear() {
	if p.areaHeight == 0 {
		return
	}
	fmt.Fprintf(p.OutWriter, "\033[%dA\033[0J", p.areaHeight)
	p.areaHeight = 0
}

func (p *Printer) DisableStyling() {
	p.disableStyling = true
}
