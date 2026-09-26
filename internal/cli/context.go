package cli

import (
	"io"

	"nova/internal/config"
	"nova/internal/logging"
	"nova/internal/output"
	"nova/internal/terminal"
)

// Context encapsulates runtime dependencies, streams, configurations, and tools for command execution.
type Context struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Config  config.Config
	Caps    terminal.Capabilities
	Printer *output.Printer
	Logger  *logging.Logger
}

// NewContext constructs an initialized execution context.
func NewContext(stdin io.Reader, stdout, stderr io.Writer, cfg config.Config, caps terminal.Capabilities, mode output.Mode, logger *logging.Logger) *Context {
	printer := output.NewPrinter(stdout, stderr, mode, caps)
	return &Context{
		Stdin:   stdin,
		Stdout:  stdout,
		Stderr:  stderr,
		Config:  cfg,
		Caps:    caps,
		Printer: printer,
		Logger:  logger,
	}
}
