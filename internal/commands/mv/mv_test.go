package mv

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

func TestRenameFile(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "old.txt")
	dst := filepath.Join(tmpDir, "new.txt")
	os.WriteFile(src, []byte("moving"), 0644)

	ctx, _, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{src, dst})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source file should not exist after move")
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("destination file not found: %v", err)
	}
	if string(data) != "moving" {
		t.Errorf("expected 'moving', got: %s", string(data))
	}
}

func TestMoveIntoDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "file.txt")
	dstDir := filepath.Join(tmpDir, "target_dir")
	os.WriteFile(src, []byte("in dir"), 0644)
	os.Mkdir(dstDir, 0755)

	ctx, _, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{src, dstDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dstFile := filepath.Join(dstDir, "file.txt")
	if _, err := os.Stat(dstFile); err != nil {
		t.Fatalf("file not found in target dir: %v", err)
	}
}

func TestMoveDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "keep.txt")
	dst := filepath.Join(tmpDir, "wont_exist.txt")
	os.WriteFile(src, []byte("data"), 0644)

	ctx, stdout, _ := newTestContext("", output.ModeHuman, terminal.ColorTrueColor)
	err := Run(ctx, []string{"--dry-run", src, dst})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(src); err != nil {
		t.Errorf("source should still exist after dry-run")
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Errorf("destination should not exist after dry-run")
	}
	if !strings.Contains(stdout.String(), "[dry-run]") {
		t.Errorf("expected dry-run notice in output, got: %s", stdout.String())
	}
}
