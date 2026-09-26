package cp

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

func newTestContext(stdin string, mode output.Mode, profile terminal.ColorProfile) (*command.Context, *bytes.Buffer, *bytes.Buffer) {
	var in bytes.Buffer
	in.WriteString(stdin)
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

func TestCopySingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "source.txt")
	dst := filepath.Join(tmpDir, "dest.txt")
	os.WriteFile(src, []byte("hello copy"), 0644)

	ctx, _, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{src, dst})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("destination file not created: %v", err)
	}
	if string(data) != "hello copy" {
		t.Errorf("expected 'hello copy', got: %s", string(data))
	}
}

func TestCopyDirWithoutRecursiveFails(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src_dir")
	dst := filepath.Join(tmpDir, "dst_dir")
	os.Mkdir(src, 0755)

	ctx, _, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{src, dst})
	if err == nil {
		t.Fatalf("expected error copying directory without -r")
	}
	exitErr, ok := err.(*command.ExitError)
	if !ok || !strings.Contains(exitErr.Message, "-r not specified") {
		t.Errorf("expected '-r not specified' error, got: %v", err)
	}
}

func TestCopyDirRecursive(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src_dir")
	dst := filepath.Join(tmpDir, "dst_dir")
	os.MkdirAll(filepath.Join(src, "sub"), 0755)
	os.WriteFile(filepath.Join(src, "file.txt"), []byte("file1"), 0644)
	os.WriteFile(filepath.Join(src, "sub", "leaf.txt"), []byte("leaf"), 0644)

	ctx, _, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{"-r", src, dst})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	leafData, err := os.ReadFile(filepath.Join(dst, "sub", "leaf.txt"))
	if err != nil {
		t.Fatalf("leaf file not found in destination: %v", err)
	}
	if string(leafData) != "leaf" {
		t.Errorf("expected 'leaf', got %s", string(leafData))
	}
}

func TestCopyDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src.txt")
	dst := filepath.Join(tmpDir, "dst.txt")
	os.WriteFile(src, []byte("data"), 0644)

	ctx, stdout, _ := newTestContext("", output.ModeHuman, terminal.ColorTrueColor)
	err := Run(ctx, []string{"--dry-run", src, dst})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Errorf("destination should not exist after dry run")
	}
	if !strings.Contains(stdout.String(), "[dry-run]") {
		t.Errorf("expected dry-run notification in output, got: %s", stdout.String())
	}
}
