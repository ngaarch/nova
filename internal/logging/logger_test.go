package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoggerDisabled(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, false)

	logger.Debug("should not print")
	logger.Debugf("format %s", "disabled")

	if buf.Len() > 0 {
		t.Errorf("expected no output when logger is disabled, got %q", buf.String())
	}
	if logger.IsEnabled() {
		t.Errorf("expected IsEnabled() = false")
	}
}

func TestLoggerEnabled(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, true)

	logger.Debug("hello debug")
	logger.Debugf("count: %d", 42)

	out := buf.String()
	if !strings.Contains(out, "[DEBUG") || !strings.Contains(out, "hello debug") {
		t.Errorf("expected debug message, got %q", out)
	}
	if !strings.Contains(out, "count: 42") {
		t.Errorf("expected formatted message, got %q", out)
	}
	if !logger.IsEnabled() {
		t.Errorf("expected IsEnabled() = true")
	}

	logger.SetEnabled(false)
	buf.Reset()
	logger.Debug("after disable")
	if buf.Len() > 0 {
		t.Errorf("expected no output after SetEnabled(false)")
	}
}
