package rm

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

func TestRemoveSingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "del.txt")
	os.WriteFile(path, []byte("goodbye"), 0644)

	ctx, _, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("file should not exist after remove")
	}
}

func TestRemoveDirWithoutRecursiveFails(t *testing.T) {
	tmpDir := t.TempDir()
	dir := filepath.Join(tmpDir, "folder")
	os.Mkdir(dir, 0755)

	ctx, _, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{dir})
	if err == nil {
		t.Fatalf("expected error removing directory without -r")
	}
	exitErr, ok := err.(*command.ExitError)
	if !ok || !strings.Contains(exitErr.Cause, "Is a directory") {
		t.Errorf("expected 'Is a directory' error, got: %v", err)
	}
}

func TestRemoveDirRecursive(t *testing.T) {
	tmpDir := t.TempDir()
	dir := filepath.Join(tmpDir, "nested_folder")
	os.MkdirAll(filepath.Join(dir, "sub"), 0755)
	os.WriteFile(filepath.Join(dir, "sub", "item.txt"), []byte("item"), 0644)

	ctx, _, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{"-r", dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("directory should be deleted after recursive remove")
	}
}

func TestRemoveRootProtection(t *testing.T) {
	ctx, _, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{"-rf", "/"})
	if err == nil {
		t.Fatalf("expected error removing root directory")
	}
	exitErr, ok := err.(*command.ExitError)
	if !ok || !strings.Contains(exitErr.Cause, "refusing to remove root directory") {
		t.Errorf("expected root protection error, got: %v", err)
	}

	// Current directory protection
	err2 := Run(ctx, []string{"-rf", "."})
	if err2 == nil {
		t.Fatalf("expected error removing current directory")
	}
}

func TestRemoveSymlinkDoesNotDeleteTarget(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "target_dir")
	targetFile := filepath.Join(targetDir, "important.txt")
	os.Mkdir(targetDir, 0755)
	os.WriteFile(targetFile, []byte("important"), 0644)

	symlinkPath := filepath.Join(tmpDir, "link_to_target")
	if err := os.Symlink(targetDir, symlinkPath); err != nil {
		t.Skip("symlinks not supported")
	}

	// Remove the symlink
	ctx, _, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{symlinkPath})
	if err != nil {
		t.Fatalf("unexpected error removing symlink: %v", err)
	}

	// Symlink should be removed
	if _, err := os.Lstat(symlinkPath); !os.IsNotExist(err) {
		t.Errorf("symlink should be removed")
	}

	// Target directory and its contents MUST remain intact!
	if _, err := os.Stat(targetFile); err != nil {
		t.Fatalf("CRITICAL BUG: removing symlink deleted the target file!")
	}
}

func TestRemoveDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "keep.txt")
	os.WriteFile(path, []byte("keep"), 0644)

	ctx, stdout, _ := newTestContext("", output.ModeHuman, terminal.ColorTrueColor)
	err := Run(ctx, []string{"--dry-run", path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("file should not be deleted during dry-run")
	}
	if !strings.Contains(stdout.String(), "[dry-run]") {
		t.Errorf("expected dry-run notice in output, got: %s", stdout.String())
	}
}
