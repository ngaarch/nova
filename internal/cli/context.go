package cli

import (
	"nova/internal/command"
)

// Context encapsulates runtime dependencies, streams, configurations, and tools for command execution.
type Context = command.Context

// NewContext constructs an initialized execution context.
var NewContext = command.NewContext
