package touch

import (
	"bytes"
	"os"
	"path/filepath"
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

func TestTouchCreateFile(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "newfile.txt")

	ctx, stdout, _ := newTestContext(output.ModeHuman)
	err := Run(ctx, []string{target})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(target); os.IsNotExist(err) {
		t.Errorf("file was not created: %s", target)
	}

	out := stdout.String()
	if !strings.Contains(out, "Created:") || !strings.Contains(out, "newfile.txt") {
		t.Errorf("expected Created notice in output, got: %s", out)
	}
}

func TestTouchUpdateFile(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "existing.txt")
	past := time.Now().Add(-24 * time.Hour)
	_ = os.WriteFile(target, []byte("content"), 0644)
	_ = os.Chtimes(target, past, past)

	ctx, stdout, _ := newTestContext(output.ModeHuman)
	err := Run(ctx, []string{target})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fi, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if fi.ModTime().Before(time.Now().Add(-10 * time.Second)) {
		t.Errorf("timestamp was not updated to current time")
	}

	out := stdout.String()
	if !strings.Contains(out, "Updated:") {
		t.Errorf("expected Updated notice, got: %s", out)
	}
}

func TestTouchParents(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "deep", "nested", "dir", "file.txt")

	ctx, _, _ := newTestContext(output.ModeHuman)
	err := Run(ctx, []string{"-p", target})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(target); os.IsNotExist(err) {
		t.Errorf("deep nested file was not created with -p")
	}
}

func TestTouchNoCreate(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "should_not_exist.txt")

	ctx, _, _ := newTestContext(output.ModeHuman)
	err := Run(ctx, []string{"-c", target})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf("file should not have been created with -c")
	}
}

func TestTouchCustomDate(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "dated.txt")

	ctx, _, _ := newTestContext(output.ModeHuman)
	err := Run(ctx, []string{"--date=2024-01-01 12:00:00", target})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fi, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if fi.ModTime().Year() != 2024 {
		t.Errorf("expected year 2024, got: %v", fi.ModTime().Year())
	}
}

func TestTouchJSON(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "json_touch.txt")

	ctx, stdout, _ := newTestContext(output.ModeJSON)
	err := Run(ctx, []string{"--json", target})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, `"created": true`) || !strings.Contains(out, `"path":`) {
		t.Errorf("expected JSON output, got: %s", out)
	}
}

func TestTouchMissingOperand(t *testing.T) {
	ctx, _, _ := newTestContext(output.ModeHuman)
	err := Run(ctx, []string{})
	if err == nil {
		t.Fatalf("expected error on missing operand")
	}
}
