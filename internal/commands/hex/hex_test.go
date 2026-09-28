package hex

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

func newTestContext(stdin string, mode output.Mode) (*command.Context, *bytes.Buffer, *bytes.Buffer) {
	var stdout, stderr bytes.Buffer
	in := strings.NewReader(stdin)
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
	ctx := command.NewContext(in, &stdout, &stderr, cfg, caps, th, mode, logger)
	return ctx, &stdout, &stderr
}

func TestHex_FileDumpPlain(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "binary.dat")
	data := []byte{0x00, 0x41, 0x42, 0x43, 0x0a, 0xff} // null, A, B, C, newline, 255
	if err := os.WriteFile(f, data, 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	ctx, stdout, _ := newTestContext("", output.ModePlain)
	err := Run(ctx, []string{"--plain", f})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "00000000") {
		t.Errorf("expected offset 00000000, got: %s", out)
	}
	if !strings.Contains(out, "00 41 42 43 0a ff") {
		t.Errorf("expected hex bytes in output, got: %s", out)
	}
	if !strings.Contains(out, ".ABC.") {
		t.Errorf("expected ASCII representation, got: %s", out)
	}
}

func TestHex_SkipAndLength(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "numbers.bin")
	data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	_ = os.WriteFile(f, data, 0o644)

	ctx, stdout, _ := newTestContext("", output.ModePlain)
	err := Run(ctx, []string{"-s", "4", "-n", "6", f})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "00000004") {
		t.Errorf("expected offset 00000004, got: %s", out)
	}
	if !strings.Contains(out, "04 05 06 07 08 09") {
		t.Errorf("expected 6 bytes dumped starting at 4, got: %s", out)
	}
}

func TestHex_StdinDump(t *testing.T) {
	input := "Hello Nova!"
	ctx, stdout, _ := newTestContext(input, output.ModePlain)
	err := Run(ctx, []string{"-"})
	if err != nil {
		t.Fatalf("Run failed on stdin: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "Hello Nova!") {
		t.Errorf("expected ASCII string in stdin dump, got: %s", out)
	}
}

func TestHex_JSONOutput(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(f, []byte("json test"), 0o644)

	ctx, stdout, _ := newTestContext("", output.ModeJSON)
	err := Run(ctx, []string{"--json", f})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, `"rows":`) || !strings.Contains(out, `"offset":`) {
		t.Errorf("expected JSON output with rows, got: %s", out)
	}
}

func TestHex_DirectoryTarget(t *testing.T) {
	tmpDir := t.TempDir()
	ctx, stdout, _ := newTestContext("", output.ModeHuman)
	err := Run(ctx, []string{tmpDir})
	if err != nil {
		t.Fatalf("unexpected fatal error: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "is a directory") {
		t.Errorf("expected directory error in output, got: %s", out)
	}
}
