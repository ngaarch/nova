package output

import (
	"bytes"
	"strings"
	"testing"

	"nova/internal/terminal"
)

func TestResolveMode(t *testing.T) {
	tests := []struct {
		name        string
		flagPlain   bool
		flagJSON    bool
		isStdoutTTY bool
		expected    Mode
	}{
		{
			name:        "JSON flag overrides plain and TTY",
			flagPlain:   true,
			flagJSON:    true,
			isStdoutTTY: true,
			expected:    ModeJSON,
		},
		{
			name:        "JSON flag on pipe returns JSON",
			flagPlain:   false,
			flagJSON:    true,
			isStdoutTTY: false,
			expected:    ModeJSON,
		},
		{
			name:        "Plain flag on TTY returns Plain",
			flagPlain:   true,
			flagJSON:    false,
			isStdoutTTY: true,
			expected:    ModePlain,
		},
		{
			name:        "No flags on TTY returns Human",
			flagPlain:   false,
			flagJSON:    false,
			isStdoutTTY: true,
			expected:    ModeHuman,
		},
		{
			name:        "No flags on pipe returns Plain",
			flagPlain:   false,
			flagJSON:    false,
			isStdoutTTY: false,
			expected:    ModePlain,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveMode(tt.flagPlain, tt.flagJSON, tt.isStdoutTTY)
			if got != tt.expected {
				t.Errorf("ResolveMode(%v, %v, %v) = %v; want %v",
					tt.flagPlain, tt.flagJSON, tt.isStdoutTTY, got, tt.expected)
			}
		})
	}
}

func TestPrinterOutputs(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	caps := terminal.Capabilities{Width: 80, Height: 24}
	p := NewPrinter(&outBuf, &errBuf, ModeHuman, caps)

	p.Print("hello ")
	p.Println("world")
	p.Printf("number: %d\n", 42)

	p.Error("an error occurred")
	p.Errorf("failed code: %d\n", 1)

	expectedOut := "hello world\nnumber: 42\n"
	if outBuf.String() != expectedOut {
		t.Errorf("outBuf = %q; want %q", outBuf.String(), expectedOut)
	}

	expectedErr := "an error occurred\nfailed code: 1\n"
	if errBuf.String() != expectedErr {
		t.Errorf("errBuf = %q; want %q", errBuf.String(), expectedErr)
	}
}

func TestPrinterJSON(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	p := NewPrinter(&outBuf, &errBuf, ModeJSON, terminal.Capabilities{})

	data := map[string]any{
		"name":    "nova",
		"version": "0.1.0",
		"tools":   []string{"ls", "cat"},
	}

	if err := p.PrintJSON(data); err != nil {
		t.Fatalf("PrintJSON error: %v", err)
	}

	outputStr := outBuf.String()
	if !strings.Contains(outputStr, `"name": "nova"`) {
		t.Errorf("JSON output missing name: %s", outputStr)
	}
	if !strings.Contains(outputStr, `"tools": [`) {
		t.Errorf("JSON output missing tools array: %s", outputStr)
	}
}
