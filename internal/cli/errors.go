package cli

import (
	"fmt"
	"strings"
)

// Standard Unix exit codes adhering to POSIX conventions.
const (
	ExitSuccess = 0 // Successful execution
	ExitFailure = 1 // Operational error (filesystem failure, uncompleted task, etc.)
	ExitUsage   = 2 // Command line usage / syntax error
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
