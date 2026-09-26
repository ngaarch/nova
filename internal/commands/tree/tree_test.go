package tree

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

func setupTestHierarchy(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	// Hierarchy:
	// tmpDir/
	//   file1.txt
	//   docs/
	//     guide.md
	//     nested/
	//       info.txt
	if err := os.MkdirAll(filepath.Join(tmpDir, "docs", "nested"), 0755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("file 1 content"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "docs", "guide.md"), []byte("# Guide"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "docs", "nested", "info.txt"), []byte("info"), 0644)

	return tmpDir
}

func TestTreeBasic(t *testing.T) {
	dir := setupTestHierarchy(t)
	ctx, stdout, _ := newTestContext(output.ModeHuman, terminal.ColorTrueColor)

	err := Run(ctx, []string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "├── ") && !strings.Contains(out, "└── ") {
		t.Errorf("expected tree branches in output, got: %s", out)
	}
	if !strings.Contains(out, "docs/") || !strings.Contains(out, "guide.md") {
		t.Errorf("expected files in tree output, got: %s", out)
	}
	if !strings.Contains(out, "directories") || !strings.Contains(out, "files") {
		t.Errorf("expected summary footer, got: %s", out)
	}
}

func TestTreePlain(t *testing.T) {
	dir := setupTestHierarchy(t)
	ctx, stdout, _ := newTestContext(output.ModePlain, terminal.ColorNone)

	err := Run(ctx, []string{"--plain", dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if strings.Contains(out, "\x1b[") {
		t.Errorf("expected no ANSI codes in plain mode, got: %s", out)
	}
	if !strings.Contains(out, "|-- ") && !strings.Contains(out, "\\-- ") {
		t.Errorf("expected ASCII branch glyphs in plain mode, got: %s", out)
	}
}

func TestTreeDepthLimit(t *testing.T) {
	dir := setupTestHierarchy(t)
	ctx, stdout, _ := newTestContext(output.ModePlain, terminal.ColorNone)

	err := Run(ctx, []string{"-L", "1", dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "docs/") {
		t.Errorf("expected direct child docs/ at depth 1, got: %s", out)
	}
	if strings.Contains(out, "guide.md") || strings.Contains(out, "info.txt") {
		t.Errorf("expected depth 1 limit to exclude nested files, got: %s", out)
	}
}

func TestTreeDirsOnly(t *testing.T) {
	dir := setupTestHierarchy(t)
	ctx, stdout, _ := newTestContext(output.ModePlain, terminal.ColorNone)

	err := Run(ctx, []string{"-d", dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "docs/") {
		t.Errorf("expected docs/ dir, got: %s", out)
	}
	if strings.Contains(out, "file1.txt") || strings.Contains(out, "guide.md") {
		t.Errorf("expected no files with -d flag, got: %s", out)
	}
}

func TestTreeSizesAndPerms(t *testing.T) {
	dir := setupTestHierarchy(t)
	ctx, stdout, _ := newTestContext(output.ModePlain, terminal.ColorNone)

	err := Run(ctx, []string{"-s", "-p", dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "[-rw-r--r--]") {
		t.Errorf("expected permissions in output, got: %s", out)
	}
	if !strings.Contains(out, "total") {
		t.Errorf("expected total size in footer, got: %s", out)
	}
}

func TestTreeJSON(t *testing.T) {
	dir := setupTestHierarchy(t)
	ctx, stdout, _ := newTestContext(output.ModeJSON, terminal.ColorNone)

	err := Run(ctx, []string{"--json", dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, `"dir_count"`) || !strings.Contains(out, `"file_count"`) {
		t.Errorf("expected JSON tree structure, got: %s", out)
	}
}
