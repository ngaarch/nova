package mkdir

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

func TestMakeDirSingle(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "new_folder")

	ctx, _, _ := newTestContext("", output.ModePlain, terminal.ColorNone)

	err := Run(ctx, []string{target})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fi, err := os.Stat(target)
	if err != nil {
		t.Fatalf("expected directory to exist: %v", err)
	}
	if !fi.IsDir() {
		t.Fatalf("expected target to be directory")
	}
}

func TestMakeDirParents(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "nested", "sub", "dir")

	ctx, _, _ := newTestContext("", output.ModePlain, terminal.ColorNone)

	err := Run(ctx, []string{"-p", target})
	if err != nil {
		t.Fatalf("unexpected error with -p: %v", err)
	}

	fi, err := os.Stat(target)
	if err != nil {
		t.Fatalf("expected nested dir to exist: %v", err)
	}
	if !fi.IsDir() {
		t.Fatalf("expected target to be directory")
	}

	// Running again with -p should succeed without error
	err = Run(ctx, []string{"-p", target})
	if err != nil {
		t.Fatalf("re-running with -p should not fail: %v", err)
	}
}

func TestMakeDirWithoutParentsFails(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "missing", "sub", "dir")

	ctx, _, _ := newTestContext("", output.ModePlain, terminal.ColorNone)

	err := Run(ctx, []string{target})
	if err == nil {
		t.Fatalf("expected error without -p for non-existent parents")
	}
}

func TestMakeDirMode(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "mode_dir")

	ctx, _, _ := newTestContext("", output.ModePlain, terminal.ColorNone)

	err := Run(ctx, []string{"-m", "0700", target})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fi, err := os.Stat(target)
	if err != nil {
		t.Fatalf("expected dir to exist: %v", err)
	}
	// On Unix, filemode is subject to umask, but user permissions should be rwx (0700)
	perm := fi.Mode().Perm()
	if perm&0700 != 0700 {
		t.Errorf("expected user permissions 0700, got: %o", perm)
	}
}

func TestMakeDirDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "dry_dir")

	ctx, stdout, _ := newTestContext("", output.ModePlain, terminal.ColorNone)

	err := Run(ctx, []string{"--dry-run", target})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("expected directory to not exist under --dry-run")
	}

	if !strings.Contains(stdout.String(), "[dry-run]") {
		t.Errorf("expected dry-run output, got: %q", stdout.String())
	}
}

func TestMakeDirVerbose(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "verbose_dir")

	ctx, stdout, _ := newTestContext("", output.ModePlain, terminal.ColorNone)

	err := Run(ctx, []string{"-v", target})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "created directory") {
		t.Errorf("expected verbose message, got: %q", stdout.String())
	}
}

func TestMakeDirInvalidMode(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "bad_mode")

	ctx, _, _ := newTestContext("", output.ModePlain, terminal.ColorNone)

	err := Run(ctx, []string{"--mode=invalid", target})
	if err == nil {
		t.Fatalf("expected error for invalid mode")
	}
}

func TestMakeDirMissingOperand(t *testing.T) {
	ctx, _, _ := newTestContext("", output.ModePlain, terminal.ColorNone)

	err := Run(ctx, []string{})
	if err == nil {
		t.Fatalf("expected error when no operands provided")
	}
}
