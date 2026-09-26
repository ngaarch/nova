package cli

import (
	"io"

	"nova/internal/config"
	"nova/internal/logging"
	"nova/internal/output"
	"nova/internal/terminal"
	"nova/internal/theme"
)

// Context encapsulates runtime dependencies, streams, configurations, and tools for command execution.
type Context struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Config  config.Config
	Caps    terminal.Capabilities
	Theme   *theme.Theme
	Printer *output.Printer
	Logger  *logging.Logger
}

// NewContext constructs an initialized execution context.
func NewContext(stdin io.Reader, stdout, stderr io.Writer, cfg config.Config, caps terminal.Capabilities, th *theme.Theme, mode output.Mode, logger *logging.Logger) *Context {
	printer := output.NewPrinter(stdout, stderr, mode, caps)
	return &Context{
		Stdin:   stdin,
		Stdout:  stdout,
		Stderr:  stderr,
		Config:  cfg,
		Caps:    caps,
		Theme:   th,
		Printer: printer,
		Logger:  logger,
	}
}
