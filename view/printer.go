package view

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"golang.org/x/term"
)

type Color string

const (
	ColorNone       Color = ""
	FGBlack         Color = "30"
	FGRed           Color = "31"
	FGGreen         Color = "32"
	FGYellow        Color = "33"
	FGBlue          Color = "34"
	FGMagenta       Color = "35"
	FGCyan          Color = "36"
	FGWhite         Color = "37"
	FGBrightBlack   Color = "90"
	FGBrightRed     Color = "91"
	FGBrightGreen   Color = "92"
	FGBrightYellow  Color = "93"
	FGBrightBlue    Color = "94"
	FGBrightMagenta Color = "95"
	FGBrightCyan    Color = "96"
	FGBrightWhite   Color = "97"
	BGBlack         Color = "40"
	BGRed           Color = "41"
	BGGreen         Color = "42"
	BGYellow        Color = "43"
	BGBlue          Color = "44"
	BGMagenta       Color = "45"
	BGCyan          Color = "46"
	BGWhite         Color = "47"
	BGBrightBlack   Color = "100"
	BGBrightRed     Color = "101"
	BGBrightGreen   Color = "102"
	BGBrightYellow  Color = "103"
	BGBrightBlue    Color = "104"
	BGBrightMagenta Color = "105"
	BGBrightCyan    Color = "106"
	BGBrightWhite   Color = "107"
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

func (p *Printer) ErrPrint(a ...any) {
	fmt.Fprint(p.ErrWriter, a...)
}

func (p *Printer) ErrPrintln(a ...any) {
	fmt.Fprintln(p.ErrWriter, a...)
}

func (p *Printer) ErrPrintf(format string, a ...any) {
	fmt.Fprintf(p.ErrWriter, format, a...)
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

func (p *Printer) ColorForeground(s string, color Color) string {
	if p.disableStyling {
		return s
	}
	return fmt.Sprintf("\033[38;5;%sm%s\033[0m", color, s)
}

func (p *Printer) ColorBackground(s string, color Color) string {
	if p.disableStyling {
		return s
	}
	return fmt.Sprintf("\033[48;5;%sm%s\033[0m", color, s)
}

func (p *Printer) Bold(s string) string {
	if p.disableStyling {
		return s
	}
	return fmt.Sprintf("\033[1m%s\033[0m", s)
}

func (p *Printer) BoldColor(s string, color Color) string {
	if p.disableStyling {
		return s
	}
	return fmt.Sprintf("\033[1;%sm%s\033[0m", color, s)
}

func (p *Printer) BoldForeground(s string, color Color) string {
	if p.disableStyling {
		return s
	}
	return fmt.Sprintf("\033[1;38;5;%sm%s\033[0m", color, s)
}

func (p *Printer) BoldBackground(s string, color Color) string {
	if p.disableStyling {
		return s
	}
	return fmt.Sprintf("\033[1;48;5;%sm%s\033[0m", color, s)
}

func (p *Printer) ReadPassword() (string, error) {
	if p.InReader == nil {
		return "", errors.New("no input reader")
	}
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	return string(bytePassword), nil
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
