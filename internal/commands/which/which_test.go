package which

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

func setupMockBinDir(t *testing.T) (string, string) {
	t.Helper()
	tmpDir := t.TempDir()

	binDir1 := filepath.Join(tmpDir, "bin1")
	binDir2 := filepath.Join(tmpDir, "bin2")
	_ = os.Mkdir(binDir1, 0755)
	_ = os.Mkdir(binDir2, 0755)

	toolA1 := filepath.Join(binDir1, "mytool")
	toolA2 := filepath.Join(binDir2, "mytool")
	_ = os.WriteFile(toolA1, []byte("#!/bin/sh\n"), 0755)
	_ = os.WriteFile(toolA2, []byte("#!/bin/sh\n"), 0755)

	symlinkTool := filepath.Join(binDir1, "mytool_link")
	_ = os.Symlink(toolA1, symlinkTool)

	pathVal := binDir1 + string(filepath.ListSeparator) + binDir2
	t.Setenv("PATH", pathVal)

	return binDir1, binDir2
}

func TestWhichFoundHuman(t *testing.T) {
	binDir1, _ := setupMockBinDir(t)
	ctx, stdout, _ := newTestContext(output.ModeHuman)

	err := Run(ctx, []string{"mytool"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, filepath.Join(binDir1, "mytool")) {
		t.Errorf("expected path %s in output, got: %s", binDir1, out)
	}
}

func TestWhichAll(t *testing.T) {
	binDir1, binDir2 := setupMockBinDir(t)
	ctx, stdout, _ := newTestContext(output.ModePlain)

	err := Run(ctx, []string{"-a", "mytool"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines with -a, got %d: %q", len(lines), out)
	}
	if !strings.Contains(out, binDir1) || !strings.Contains(out, binDir2) {
		t.Errorf("expected both bin directories in output")
	}
}

func TestWhichSymlink(t *testing.T) {
	setupMockBinDir(t)
	ctx, stdout, _ := newTestContext(output.ModeHuman)

	err := Run(ctx, []string{"mytool_link"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "->") || !strings.Contains(out, "mytool") {
		t.Errorf("expected symlink target indicator, got: %s", out)
	}
}

func TestWhichNotFound(t *testing.T) {
	setupMockBinDir(t)
	ctx, stdout, _ := newTestContext(output.ModeHuman)

	err := Run(ctx, []string{"nonexistent_binary_xyz"})
	if err == nil {
		t.Fatalf("expected error on non-existent binary")
	}

	out := stdout.String()
	if !strings.Contains(out, "not found") {
		t.Errorf("expected 'not found' message, got: %s", out)
	}
}

func TestWhichJSON(t *testing.T) {
	setupMockBinDir(t)
	ctx, stdout, _ := newTestContext(output.ModeJSON)

	err := Run(ctx, []string{"--json", "mytool"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, `"found": true`) || !strings.Contains(out, `"mytool"`) {
		t.Errorf("expected structured JSON output, got: %s", out)
	}
}

func TestWhichMissingOperand(t *testing.T) {
	ctx, _, _ := newTestContext(output.ModeHuman)
	err := Run(ctx, []string{})
	if err == nil {
		t.Fatalf("expected error on missing operand")
	}
}
