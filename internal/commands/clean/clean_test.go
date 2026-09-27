package clean

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

func TestCleanDryRun(t *testing.T) {
	tmpDir := t.TempDir()

	// Create disposable files
	junk1 := filepath.Join(tmpDir, ".DS_Store")
	junk2 := filepath.Join(tmpDir, "temp.swp")
	validFile := filepath.Join(tmpDir, "keep.go")

	_ = os.WriteFile(junk1, []byte("junk1"), 0644)
	_ = os.WriteFile(junk2, []byte("junk2"), 0644)
	_ = os.WriteFile(validFile, []byte("keep"), 0644)

	ctx, stdout, _ := newTestContext(output.ModeHuman)

	err := Run(ctx, []string{tmpDir})
	if err != nil {
		t.Fatalf("expected clean dry-run to succeed, got %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "DRY-RUN") {
		t.Errorf("expected DRY-RUN banner in output, got: %s", out)
	}
	if !strings.Contains(out, "os-junk") {
		t.Errorf("expected 'os-junk' in candidates, got: %s", out)
	}

	// Files must still exist in dry-run
	if _, err := os.Stat(junk1); os.IsNotExist(err) {
		t.Errorf("junk1 should NOT be deleted in dry-run")
	}
}

func TestCleanForceExecution(t *testing.T) {
	tmpDir := t.TempDir()

	junk1 := filepath.Join(tmpDir, ".DS_Store")
	validFile := filepath.Join(tmpDir, "keep.go")

	_ = os.WriteFile(junk1, []byte("junk1"), 0644)
	_ = os.WriteFile(validFile, []byte("keep"), 0644)

	ctx, stdout, _ := newTestContext(output.ModeHuman)

	err := Run(ctx, []string{"-f", tmpDir})
	if err != nil {
		t.Fatalf("expected clean force to succeed, got %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "Deleted") {
		t.Errorf("expected 'Deleted' in output, got: %s", out)
	}

	// junk1 should now be gone
	if _, err := os.Stat(junk1); !os.IsNotExist(err) {
		t.Errorf("junk1 should be deleted after -f")
	}

	// validFile must still exist
	if _, err := os.Stat(validFile); os.IsNotExist(err) {
		t.Errorf("validFile must NOT be deleted")
	}
}

func TestCleanPlainAndJSON(t *testing.T) {
	tmpDir := t.TempDir()
	junk := filepath.Join(tmpDir, "test.tmp")
	_ = os.WriteFile(junk, []byte("temp content"), 0644)

	// Plain mode
	ctxPlain, stdoutPlain, _ := newTestContext(output.ModePlain)
	err := Run(ctxPlain, []string{"--plain", tmpDir})
	if err != nil {
		t.Fatalf("plain clean failed: %v", err)
	}
	if !strings.Contains(stdoutPlain.String(), "temp-file") {
		t.Errorf("expected 'temp-file' in plain output, got: %s", stdoutPlain.String())
	}

	// JSON mode
	ctxJSON, stdoutJSON, _ := newTestContext(output.ModeJSON)
	err = Run(ctxJSON, []string{"--json", tmpDir})
	if err != nil {
		t.Fatalf("json clean failed: %v", err)
	}
	if !strings.Contains(stdoutJSON.String(), `"total_candidates": 1`) {
		t.Errorf("expected total_candidates: 1 in json, got: %s", stdoutJSON.String())
	}
}
