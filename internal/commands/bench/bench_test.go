package bench

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

func TestBenchHuman(t *testing.T) {
	ctx, stdout, _ := newTestContext(output.ModeHuman)

	// Small size and iterations for fast test
	err := Run(ctx, []string{"--size", "1", "--iterations", "100"})
	if err != nil {
		t.Fatalf("expected bench to succeed, got %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "Filesystem Throughput") {
		t.Errorf("expected 'Filesystem Throughput' in output, got: %s", out)
	}
	if !strings.Contains(out, "Stat Latency Profile") {
		t.Errorf("expected 'Stat Latency Profile' in output, got: %s", out)
	}
}

func TestBenchPlain(t *testing.T) {
	ctx, stdout, _ := newTestContext(output.ModePlain)

	err := Run(ctx, []string{"--plain", "--size", "1", "--iterations", "50"})
	if err != nil {
		t.Fatalf("expected plain bench to succeed, got %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "seq_write_mb_s=") {
		t.Errorf("expected 'seq_write_mb_s=' in plain output, got: %s", out)
	}
	if !strings.Contains(out, "seq_read_mb_s=") {
		t.Errorf("expected 'seq_read_mb_s=' in plain output, got: %s", out)
	}
	if !strings.Contains(out, "stat_ops_per_sec=") {
		t.Errorf("expected 'stat_ops_per_sec=' in plain output, got: %s", out)
	}
}

func TestBenchJSON(t *testing.T) {
	ctx, stdout, _ := newTestContext(output.ModeJSON)

	err := Run(ctx, []string{"--json", "--size", "1", "--iterations", "50"})
	if err != nil {
		t.Fatalf("expected json bench to succeed, got %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, `"seq_write_mb_s":`) {
		t.Errorf("expected '\"seq_write_mb_s\":' in json output, got: %s", out)
	}
	if !strings.Contains(out, `"stat_latency_us":`) {
		t.Errorf("expected '\"stat_latency_us\":' in json output, got: %s", out)
	}
}
