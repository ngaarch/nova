package du

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

func setupDUHierarchy(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	// Structure:
	// tmpDir/
	//   file1.bin (1000 bytes)
	//   sub/
	//     file2.bin (2000 bytes)
	//     nested/
	//       file3.bin (5000 bytes)
	if err := os.MkdirAll(filepath.Join(tmpDir, "sub", "nested"), 0755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(tmpDir, "file1.bin"), make([]byte, 1000), 0644)
	os.WriteFile(filepath.Join(tmpDir, "sub", "file2.bin"), make([]byte, 2000), 0644)
	os.WriteFile(filepath.Join(tmpDir, "sub", "nested", "file3.bin"), make([]byte, 5000), 0644)

	return tmpDir
}

func TestDUHumanMode(t *testing.T) {
	dir := setupDUHierarchy(t)
	ctx, stdout, _ := newTestContext(output.ModeHuman, terminal.ColorTrueColor)

	err := Run(ctx, []string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	// Should contain visual bars or percentages
	if !strings.Contains(out, "%") || !strings.Contains(out, "sub") {
		t.Errorf("expected visual bars and percentage in human du, got: %s", out)
	}
}

func TestDUSummarize(t *testing.T) {
	dir := setupDUHierarchy(t)
	ctx, stdout, _ := newTestContext(output.ModePlain, terminal.ColorNone)

	err := Run(ctx, []string{"-s", dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 line for -s (summarize), got %d lines: %v", len(lines), lines)
	}
}

func TestDUMaxDepth(t *testing.T) {
	dir := setupDUHierarchy(t)
	ctx, stdout, _ := newTestContext(output.ModePlain, terminal.ColorNone)

	err := Run(ctx, []string{"-d", "1", dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "sub") {
		t.Errorf("expected direct child sub at depth 1, got: %s", out)
	}
	if strings.Contains(out, "nested") {
		t.Errorf("nested at depth 2 should not be reported with -d 1, got: %s", out)
	}
}

func TestDUAllFiles(t *testing.T) {
	dir := setupDUHierarchy(t)
	ctx, stdout, _ := newTestContext(output.ModePlain, terminal.ColorNone)

	err := Run(ctx, []string{"-a", dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "file1.bin") || !strings.Contains(out, "file2.bin") {
		t.Errorf("expected individual files to be listed with -a, got: %s", out)
	}
}

func TestDUTotal(t *testing.T) {
	dir := setupDUHierarchy(t)
	ctx, stdout, _ := newTestContext(output.ModePlain, terminal.ColorNone)

	err := Run(ctx, []string{"-c", dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "total") {
		t.Errorf("expected grand total with -c, got: %s", out)
	}
}

func TestDUJSON(t *testing.T) {
	dir := setupDUHierarchy(t)
	ctx, stdout, _ := newTestContext(output.ModeJSON, terminal.ColorNone)

	err := Run(ctx, []string{"--json", dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, `"bytes":`) || !strings.Contains(out, `"human_size":`) {
		t.Errorf("expected structured JSON output, got: %s", out)
	}
}
