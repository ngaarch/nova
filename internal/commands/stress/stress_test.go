package stresscmd

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"nova/internal/command"
	"nova/internal/config"
	"nova/internal/logging"
	"nova/internal/output"
	"nova/internal/terminal"
	"nova/internal/theme"
)

func newTestContext(stdin string, stdout, stderr *bytes.Buffer, mode output.Mode) *command.Context {
	cfg := config.Config{
		Theme:     "default",
		ColorMode: "never",
	}
	caps := terminal.Capabilities{
		Width:        80,
		Height:       24,
		ColorProfile: terminal.ColorNone,
		IsTTY:        false,
	}
	th := theme.Get("default")
	logger := logging.New(stderr, false)
	return command.NewContext(
		strings.NewReader(stdin),
		stdout,
		stderr,
		cfg,
		caps,
		th,
		mode,
		logger,
	)
}

func TestParseFlags(t *testing.T) {
	opts := ParseFlags([]string{"--duration=50ms", "--threads=4", "--algo=sha256", "--plain", "--json"})
	if opts.Duration != 50*time.Millisecond {
		t.Fatalf("expected 50ms, got %v", opts.Duration)
	}
	if opts.Threads != 4 {
		t.Fatalf("expected 4 threads, got %d", opts.Threads)
	}
	if opts.Algo != "sha256" {
		t.Fatalf("expected sha256, got %s", opts.Algo)
	}
	if !opts.Plain {
		t.Fatalf("expected Plain=true")
	}
	if !opts.JSON {
		t.Fatalf("expected JSON=true")
	}

	// Positional algo and shorthand flags
	opts2 := ParseFlags([]string{"-d", "100ms", "-t", "2", "-a", "math"})
	if opts2.Duration != 100*time.Millisecond {
		t.Fatalf("expected 100ms, got %v", opts2.Duration)
	}
	if opts2.Threads != 2 {
		t.Fatalf("expected 2 threads, got %d", opts2.Threads)
	}
	if opts2.Algo != "math" {
		t.Fatalf("expected math, got %s", opts2.Algo)
	}
}

func TestStressWorkloads(t *testing.T) {
	dur := 50 * time.Millisecond
	threads := 2

	shaRes := runSHA256Stress(threads, dur)
	if shaRes.TotalOps <= 0 {
		t.Errorf("expected positive sha ops, got %d", shaRes.TotalOps)
	}
	if shaRes.Throughput == "" {
		t.Errorf("expected non-empty sha throughput")
	}

	mathRes := runMathStress(threads, dur)
	if mathRes.TotalOps <= 0 {
		t.Errorf("expected positive math ops, got %d", mathRes.TotalOps)
	}
	if mathRes.Throughput == "" {
		t.Errorf("expected non-empty math throughput")
	}

	memRes := runMemStress(threads, dur)
	if memRes.TotalOps <= 0 {
		t.Errorf("expected positive mem ops, got %d", memRes.TotalOps)
	}
	if memRes.Throughput == "" {
		t.Errorf("expected non-empty mem throughput")
	}
}

func TestStressRunJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ctx := newTestContext("", &stdout, &stderr, output.ModeJSON)

	err := Run(ctx, []string{"--duration=50ms", "--threads=2", "--algo=math", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, `"threads": 2`) || !strings.Contains(out, `"Floating Math"`) {
		t.Fatalf("expected json with Floating Math, got: %s", out)
	}
}

func TestStressRunPlain(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ctx := newTestContext("", &stdout, &stderr, output.ModePlain)

	err := Run(ctx, []string{"--duration=50ms", "--threads=1", "--algo=mem", "--plain"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "Memory Bandwidth") || !strings.Contains(out, "WORKLOAD") {
		t.Fatalf("expected plain output with Memory Bandwidth, got: %s", out)
	}
}

func TestStressRunDashboard(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ctx := newTestContext("", &stdout, &stderr, output.ModeHuman)

	err := Run(ctx, []string{"--duration=50ms", "--threads=1", "--algo=sha256"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "Multi-Core Hardware Stress") || !strings.Contains(out, "SHA-256") {
		t.Fatalf("expected dashboard output, got: %s", out)
	}
}

func TestStressCommandMetadata(t *testing.T) {
	cmd := Command()
	if cmd.Name != "stress" {
		t.Errorf("expected cmd name stress, got %s", cmd.Name)
	}
	if cmd.Phase != 31 {
		t.Errorf("expected phase 31, got %d", cmd.Phase)
	}
	if len(cmd.Aliases) == 0 {
		t.Errorf("expected aliases for stress")
	}
}
