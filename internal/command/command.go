package command

import (
	"fmt"
	"io"
	"strings"

	"nova/internal/config"
	"nova/internal/logging"
	"nova/internal/output"
	"nova/internal/terminal"
	"nova/internal/theme"
)

// Standard Unix exit codes adhering to POSIX conventions.
const (
	ExitSuccess = 0 // Successful execution
	ExitFailure = 1 // Operational error
	ExitUsage   = 2 // Usage or syntax error
)

// ExitError represents a structured, user-actionable command-line failure.
type ExitError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Path       string `json:"path,omitempty"`
	Cause      string `json:"cause,omitempty"`
	Resolution string `json:"resolution,omitempty"`
}

// Error formats the error into an actionable message for the user.
func (e *ExitError) Error() string {
	var b strings.Builder
	b.WriteString("nova: error: ")
	b.WriteString(e.Message)
	if e.Path != "" {
		b.WriteString(fmt.Sprintf("\n  path: %s", e.Path))
	}
	if e.Cause != "" {
		b.WriteString(fmt.Sprintf("\n  reason: %s", e.Cause))
	}
	if e.Resolution != "" {
		b.WriteString(fmt.Sprintf("\n  resolution: %s", e.Resolution))
	}
	return b.String()
}

// NewUsageError creates an ExitError with exit code ExitUsage.
func NewUsageError(msg string, resolution string) *ExitError {
	return &ExitError{
		Code:       ExitUsage,
		Message:    msg,
		Resolution: resolution,
	}
}

// NewOpError creates an operational ExitError with exit code ExitFailure.
func NewOpError(msg string, path string, cause error, resolution string) *ExitError {
	var causeStr string
	if cause != nil {
		causeStr = cause.Error()
	}
	return &ExitError{
		Code:       ExitFailure,
		Message:    msg,
		Path:       path,
		Cause:      causeStr,
		Resolution: resolution,
	}
}

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

// Command represents a runnable CLI subcommand.
type Command struct {
	Name        string
	Aliases     []string
	Summary     string
	Usage       string
	Description string
	Phase       int
	Run         func(ctx *Context, args []string) error
}

// Matches returns true if the given name matches the command name or any alias.
func (c *Command) Matches(name string) bool {
	if c.Name == name {
		return true
	}
	for _, alias := range c.Aliases {
		if alias == name {
			return true
		}
	}
	return false
}
