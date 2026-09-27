package sysinfo

import (
	"bytes"
	"strings"
	"testing"

	"nova/internal/command"
	"nova/internal/config"
	"nova/internal/logging"
	"nova/internal/output"
	"nova/internal/terminal"
	"nova/internal/theme"
)

func newTestContext(mode output.Mode) (*command.Context, *bytes.Buffer, *bytes.Buffer) {
	var in, stdout, stderr bytes.Buffer
	cfg := config.Default()
	caps := terminal.Capabilities{
		IsTTY:            mode == output.ModeHuman,
		Width:            80,
		Height:           24,
		ColorProfile:     terminal.ColorTrueColor,
		UnicodeSupported: true,
	}
	th := theme.Get("default")
	logger := logging.New(&stderr, false)
	ctx := command.NewContext(&in, &stdout, &stderr, cfg, caps, th, mode, logger)
	return ctx, &stdout, &stderr
}

func TestSysinfoHuman(t *testing.T) {
	ctx, stdout, _ := newTestContext(output.ModeHuman)

	err := Run(ctx, []string{})
	if err != nil {
		t.Fatalf("expected sysinfo to succeed, got %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "System & Hardware") {
		t.Errorf("expected 'System & Hardware' in output, got: %s", out)
	}
	if !strings.Contains(out, "Memory & Storage") {
		t.Errorf("expected 'Memory & Storage' in output, got: %s", out)
	}
	if !strings.Contains(out, "Environment & Nova Context") {
		t.Errorf("expected 'Environment & Nova Context' in output, got: %s", out)
	}
}

func TestSysinfoPlain(t *testing.T) {
	ctx, stdout, _ := newTestContext(output.ModePlain)

	err := Run(ctx, []string{"--plain"})
	if err != nil {
		t.Fatalf("expected plain sysinfo to succeed, got %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "hostname=") {
		t.Errorf("expected 'hostname=' in plain output, got: %s", out)
	}
	if !strings.Contains(out, "os=") {
		t.Errorf("expected 'os=' in plain output, got: %s", out)
	}
	if !strings.Contains(out, "cpus=") {
		t.Errorf("expected 'cpus=' in plain output, got: %s", out)
	}
}

func TestSysinfoJSON(t *testing.T) {
	ctx, stdout, _ := newTestContext(output.ModeJSON)

	err := Run(ctx, []string{"--json"})
	if err != nil {
		t.Fatalf("expected json sysinfo to succeed, got %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, `"hostname":`) {
		t.Errorf("expected '\"hostname\":' in json output, got: %s", out)
	}
	if !strings.Contains(out, `"go_version":`) {
		t.Errorf("expected '\"go_version\":' in json output, got: %s", out)
	}
}

func TestSysinfoFull(t *testing.T) {
	ctx, stdout, _ := newTestContext(output.ModeHuman)

	err := Run(ctx, []string{"--full"})
	if err != nil {
		t.Fatalf("expected full sysinfo to succeed, got %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "PATH Directories:") {
		t.Errorf("expected 'PATH Directories:' in full output, got: %s", out)
	}
}
