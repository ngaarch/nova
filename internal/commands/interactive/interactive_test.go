package interactive

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

func newTestContext(isTTY bool) (*command.Context, *bytes.Buffer, *bytes.Buffer) {
	var in bytes.Buffer
	var stdout, stderr bytes.Buffer

	cfg := config.Default()
	caps := terminal.Capabilities{
		IsTTY:            isTTY,
		Width:            80,
		Height:           24,
		ColorProfile:     terminal.ColorNone,
		UnicodeSupported: true,
	}
	th := theme.Get("default")
	logger := logging.New(&stderr, false)

	ctx := command.NewContext(&in, &stdout, &stderr, cfg, caps, th, output.ModePlain, logger)
	return ctx, &stdout, &stderr
}

func TestInteractiveFlags(t *testing.T) {
	opts, err := ParseFlags([]string{"-a", "/tmp/my_dir"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.ShowHidden {
		t.Errorf("expected ShowHidden to be true")
	}
	if opts.TargetDir != "/tmp/my_dir" {
		t.Errorf("expected TargetDir /tmp/my_dir, got %q", opts.TargetDir)
	}

	// Help flag
	opts, err = ParseFlags([]string{"--help"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.Help {
		t.Errorf("expected Help to be true")
	}

	// Unknown flag
	_, err = ParseFlags([]string{"--unknown-flag"})
	if err == nil {
		t.Fatalf("expected error for unknown flag")
	}
}

func TestInteractiveNonTTYFailsGracefully(t *testing.T) {
	ctx, _, _ := newTestContext(false) // non-TTY

	err := Run(ctx, []string{})
	if err == nil {
		t.Fatalf("expected error running interactive mode without TTY")
	}
	if !strings.Contains(err.Error(), "requires a controlling terminal TTY") {
		t.Errorf("expected TTY requirement error, got: %v", err)
	}
}

func TestInteractiveHelpOutputsUsage(t *testing.T) {
	ctx, stdout, _ := newTestContext(false)

	err := Run(ctx, []string{"--help"})
	if err != nil {
		t.Fatalf("unexpected error for --help: %v", err)
	}
	if !strings.Contains(stdout.String(), "nova interactive") {
		t.Errorf("expected usage in stdout, got: %s", stdout.String())
	}
}
