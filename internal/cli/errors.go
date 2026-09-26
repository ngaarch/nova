package cli

import (
	"nova/internal/command"
)

// Standard Unix exit codes adhering to POSIX conventions.
const (
	ExitSuccess = command.ExitSuccess
	ExitFailure = command.ExitFailure
	ExitUsage   = command.ExitUsage
)

// ExitError represents a structured, user-actionable command-line failure.
type ExitError = command.ExitError

// Error constructor aliases
var (
	NewUsageError = command.NewUsageError
	NewOpError    = command.NewOpError
)
