package stat

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nova/internal/command"
	"nova/internal/config"
	"nova/internal/logging"
	"nova/internal/output"
	"nova/internal/terminal"
	"nova/internal/theme"
)

func newTestContext(mode output.Mode, profile terminal.ColorProfile) (*command.Context, *bytes.Buffer, *bytes.Buffer) {
	var in bytes.Buffer
	var stdout, stderr bytes.Buffer

	cfg := config.Default()
	caps := terminal.Capabilities{
		IsTTY:            mode == output.ModeHuman,
		Width:            80,
		Height:           24,
		ColorProfile:     profile,
		UnicodeSupported: true,
	}
	th := theme.Get("default")
	logger := logging.New(&stderr, false)

	ctx := command.NewContext(&in, &stdout, &stderr, cfg, caps, th, mode, logger)
	return ctx, &stdout, &stderr
}

func TestStatFile(t *testing.T) {
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "hello.txt")
	os.WriteFile(fPath, []byte("hello stat!"), 0644)

	ctx, stdout, _ := newTestContext(output.ModeHuman, terminal.ColorTrueColor)
	err := Run(ctx, []string{fPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "File:") || !strings.Contains(out, "hello.txt") {
		t.Errorf("expected File: hello.txt, got: %s", out)
	}
	if !strings.Contains(out, "Size:") || !strings.Contains(out, "11") {
		t.Errorf("expected Size 11 in stat output, got: %s", out)
	}
	if !strings.Contains(out, "Access:") || !strings.Contains(out, "Modify:") {
		t.Errorf("expected timestamps in output, got: %s", out)
	}
}

func TestStatPlain(t *testing.T) {
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "plain.txt")
	os.WriteFile(fPath, []byte("data"), 0644)

	ctx, stdout, _ := newTestContext(output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{"--plain", fPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "File:\t") || !strings.Contains(out, "Size:\t4") {
		t.Errorf("expected tab-separated plain stat output, got: %s", out)
	}
}

func TestStatJSON(t *testing.T) {
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "meta.json")
	os.WriteFile(fPath, []byte("{}"), 0644)

	ctx, stdout, _ := newTestContext(output.ModeJSON, terminal.ColorNone)
	err := Run(ctx, []string{"--json", fPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, `"name": "meta.json"`) || !strings.Contains(out, `"mode_octal":`) {
		t.Errorf("expected JSON metadata, got: %s", out)
	}
}

func TestStatSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	origPath := filepath.Join(tmpDir, "orig.txt")
	linkPath := filepath.Join(tmpDir, "sym.txt")
	os.WriteFile(origPath, []byte("orig"), 0644)
	if err := os.Symlink(origPath, linkPath); err != nil {
		t.Skip("symlinks not supported")
	}

	ctx, stdout, _ := newTestContext(output.ModeHuman, terminal.ColorTrueColor)
	err := Run(ctx, []string{linkPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "Target:") || !strings.Contains(out, "orig.txt") {
		t.Errorf("expected symlink target in output, got: %s", out)
	}
}

func TestStatMissingOperand(t *testing.T) {
	ctx, _, _ := newTestContext(output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{})
	if err == nil {
		t.Fatalf("expected usage error when no operands provided")
	}
	exitErr, ok := err.(*command.ExitError)
	if !ok || exitErr.Code != command.ExitUsage {
		t.Errorf("expected ExitUsage code, got: %v", err)
	}
}
