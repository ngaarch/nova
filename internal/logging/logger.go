package logging

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// Logger provides structured debug logging for diagnostic operations.
type Logger struct {
	mu      sync.Mutex
	out     io.Writer
	enabled bool
}

// New creates a new Logger writing to out with the given enabled state.
func New(out io.Writer, enabled bool) *Logger {
	return &Logger{
		out:     out,
		enabled: enabled,
	}
}

// SetEnabled dynamically toggles debug logging output.
func (l *Logger) SetEnabled(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = enabled
}

// IsEnabled returns true if debug logging is enabled.
func (l *Logger) IsEnabled() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.enabled
}

// Debugf formats and prints a debug message to the configured writer if enabled.
func (l *Logger) Debugf(format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.enabled || l.out == nil {
		return
	}

	ts := time.Now().Format("15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(l.out, "[DEBUG %s] %s\n", ts, msg)
}

// Debug prints arguments to the configured writer if enabled.
func (l *Logger) Debug(args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.enabled || l.out == nil {
		return
	}

	ts := time.Now().Format("15:04:05.000")
	msg := fmt.Sprint(args...)
	fmt.Fprintf(l.out, "[DEBUG %s] %s\n", ts, msg)
}
