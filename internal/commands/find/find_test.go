package find

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

func setupFindHierarchy(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	// Structure:
	// tmpDir/
	//   app.go (10 bytes)
	//   README.MD (50 bytes)
	//   empty.txt (0 bytes)
	//   pkg/
	//     util.go (200 bytes)
	//     empty_dir/
	os.MkdirAll(filepath.Join(tmpDir, "pkg", "empty_dir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "app.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "README.MD"), []byte(strings.Repeat("a", 50)), 0644)
	os.WriteFile(filepath.Join(tmpDir, "empty.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "pkg", "util.go"), []byte(strings.Repeat("x", 200)), 0644)

	return tmpDir
}

func TestFindByName(t *testing.T) {
	dir := setupFindHierarchy(t)
	ctx, stdout, _ := newTestContext(output.ModePlain, terminal.ColorNone)

	err := Run(ctx, []string{dir, "-name", "*.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "app.go") || !strings.Contains(out, "util.go") {
		t.Errorf("expected to find app.go and util.go, got: %s", out)
	}
	if strings.Contains(out, "README.MD") {
		t.Errorf("README.MD should not match *.go, got: %s", out)
	}
}

func TestFindCaseInsensitive(t *testing.T) {
	dir := setupFindHierarchy(t)
	ctx, stdout, _ := newTestContext(output.ModePlain, terminal.ColorNone)

	err := Run(ctx, []string{dir, "-iname", "*.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "README.MD") {
		t.Errorf("expected README.MD with case-insensitive search, got: %s", out)
	}
}

func TestFindByType(t *testing.T) {
	dir := setupFindHierarchy(t)

	// Files only
	ctxF, stdoutF, _ := newTestContext(output.ModePlain, terminal.ColorNone)
	err := Run(ctxF, []string{dir, "-type", "f"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(stdoutF.String(), "pkg/empty_dir") {
		t.Errorf("expected files only, got dir: %s", stdoutF.String())
	}

	// Directories only
	ctxD, stdoutD, _ := newTestContext(output.ModePlain, terminal.ColorNone)
	err = Run(ctxD, []string{dir, "-type", "d"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdoutD.String(), "pkg") || !strings.Contains(stdoutD.String(), "empty_dir") {
		t.Errorf("expected directories in output, got: %s", stdoutD.String())
	}
	if strings.Contains(stdoutD.String(), "app.go") {
		t.Errorf("expected no files with -type d, got: %s", stdoutD.String())
	}
}

func TestFindBySize(t *testing.T) {
	dir := setupFindHierarchy(t)
	ctx, stdout, _ := newTestContext(output.ModePlain, terminal.ColorNone)

	// Files greater than 100 bytes (only util.go is 200 bytes)
	err := Run(ctx, []string{dir, "-size", "+100b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "util.go") {
		t.Errorf("expected util.go (>100b), got: %s", out)
	}
	if strings.Contains(out, "app.go") || strings.Contains(out, "README.MD") {
		t.Errorf("smaller files should not match +100b, got: %s", out)
	}
}

func TestFindEmpty(t *testing.T) {
	dir := setupFindHierarchy(t)
	ctx, stdout, _ := newTestContext(output.ModePlain, terminal.ColorNone)

	err := Run(ctx, []string{dir, "-empty"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "empty.txt") {
		t.Errorf("expected empty.txt with -empty, got: %s", out)
	}
	if strings.Contains(out, "app.go") {
		t.Errorf("non-empty file app.go matched -empty, got: %s", out)
	}
}

func TestFindJSON(t *testing.T) {
	dir := setupFindHierarchy(t)
	ctx, stdout, _ := newTestContext(output.ModeJSON, terminal.ColorNone)

	err := Run(ctx, []string{dir, "-name", "*.go", "--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, `"name": "app.go"`) || !strings.Contains(out, `"name": "util.go"`) {
		t.Errorf("expected JSON array of entries, got: %s", out)
	}
}
