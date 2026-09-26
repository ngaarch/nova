package output

import (
	"encoding/json"
	"fmt"
	"io"

	"nova/internal/terminal"
)

// Printer controls formatted writing to standard output and standard error.
type Printer struct {
	Out  io.Writer
	Err  io.Writer
	Mode Mode
	Caps terminal.Capabilities
}

// NewPrinter creates a new Printer configured with the specified output streams, mode, and terminal capabilities.
func NewPrinter(out, err io.Writer, mode Mode, caps terminal.Capabilities) *Printer {
	return &Printer{
		Out:  out,
		Err:  err,
		Mode: mode,
		Caps: caps,
	}
}

// Print writes arguments to Out.
func (p *Printer) Print(a ...any) {
	fmt.Fprint(p.Out, a...)
}

// Println writes arguments to Out followed by a newline.
func (p *Printer) Println(a ...any) {
	fmt.Fprintln(p.Out, a...)
}

// Printf writes formatted text to Out.
func (p *Printer) Printf(format string, a ...any) {
	fmt.Fprintf(p.Out, format, a...)
}

// Error writes arguments to Err followed by a newline.
func (p *Printer) Error(a ...any) {
	fmt.Fprintln(p.Err, a...)
}

// Errorf writes formatted text to Err.
func (p *Printer) Errorf(format string, a ...any) {
	fmt.Fprintf(p.Err, format, a...)
}

// PrintJSON serializes data as indented JSON to Out followed by a newline.
func (p *Printer) PrintJSON(v any) error {
	enc := json.NewEncoder(p.Out)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
