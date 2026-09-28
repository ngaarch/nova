package calccmd

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"nova/internal/command"
	"nova/internal/config"
	"nova/internal/output"
	"nova/internal/terminal"
	"nova/internal/theme"
)

func TestEvaluateArithmetic(t *testing.T) {
	tests := []struct {
		expr     string
		expected float64
	}{
		{"2 + 3 * 4", 14},
		{"(2 + 3) * 4", 20},
		{"100 / 4 - 5", 20},
		{"2 ^ 8", 256},
		{"2 ** 10", 1024},
		{"15 % 4", 3},
		{"-5 + 10", 5},
	}

	for _, tt := range tests {
		val, err := Evaluate(tt.expr)
		if err != nil {
			t.Fatalf("Evaluate(%q) failed: %v", tt.expr, err)
		}
		if math.Abs(val-tt.expected) > 1e-6 {
			t.Errorf("Evaluate(%q) = %v; want %v", tt.expr, val, tt.expected)
		}
	}
}

func TestEvaluateFunctionsAndConstants(t *testing.T) {
	val, err := Evaluate("sqrt(16) + abs(-10)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 14 {
		t.Errorf("expected 14, got %v", val)
	}

	piVal, err := Evaluate("pi")
	if err != nil || math.Abs(piVal-math.Pi) > 1e-6 {
		t.Errorf("expected pi, got %v (err: %v)", piVal, err)
	}
}

func TestEvaluateMultiBases(t *testing.T) {
	valHex, err := Evaluate("0x10 + 0x20")
	if err != nil || valHex != 48 {
		t.Errorf("expected 48 from 0x10 + 0x20, got %v (err: %v)", valHex, err)
	}

	valBin, err := Evaluate("0b1010 + 0b0101")
	if err != nil || valBin != 15 {
		t.Errorf("expected 15 from 0b1010 + 0b0101, got %v (err: %v)", valBin, err)
	}

	valOct, err := Evaluate("0o77")
	if err != nil || valOct != 63 {
		t.Errorf("expected 63 from 0o77, got %v (err: %v)", valOct, err)
	}
}

func TestEvaluateByteUnits(t *testing.T) {
	val, err := Evaluate("4GB / 2GB")
	if err != nil || val != 2 {
		t.Errorf("expected 2 from 4GB / 2GB, got %v (err: %v)", val, err)
	}

	valKB, err := Evaluate("1KB")
	if err != nil || valKB != 1024 {
		t.Errorf("expected 1024 from 1KB, got %v (err: %v)", valKB, err)
	}
}

func TestRunCommand(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	caps := terminal.Capabilities{
		ColorProfile: terminal.ColorNone,
		Width:        80,
	}
	ctx := command.NewContext(nil, stdout, stderr, config.Config{}, caps, theme.Get("default"), output.ModePlain, nil)

	cmd := Command()
	if cmd.Name != "calc" {
		t.Errorf("expected name calc, got %s", cmd.Name)
	}

	// Plain execution
	err := cmd.Run(ctx, []string{"2^16", "--plain"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !strings.Contains(stdout.String(), "65536") {
		t.Errorf("expected 65536 in plain output, got: %s", stdout.String())
	}

	// JSON execution
	stdout.Reset()
	ctx = command.NewContext(nil, stdout, stderr, config.Config{}, caps, theme.Get("default"), output.ModeJSON, nil)
	err = cmd.Run(ctx, []string{"0xFF + 1", "--json"})
	if err != nil {
		t.Fatalf("expected nil error for JSON, got %v", err)
	}

	var parsed CalcResult
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v, raw: %s", err, stdout.String())
	}
	if parsed.Value != 256 {
		t.Errorf("expected Value 256, got %v", parsed.Value)
	}
	if parsed.Hex != "0x100" {
		t.Errorf("expected Hex 0x100, got %s", parsed.Hex)
	}
}
