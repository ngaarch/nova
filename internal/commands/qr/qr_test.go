package qrcmd

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
	var stdout, stderr bytes.Buffer
	in := strings.NewReader("")
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

func TestQR_EncodeText(t *testing.T) {
	code, err := Encode("https://github.com/ngaarch/nova")
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if code.Size <= 0 {
		t.Errorf("expected positive matrix size, got %d", code.Size)
	}
	if len(code.Matrix) != code.Size {
		t.Errorf("matrix dimension mismatch: %d != %d", len(code.Matrix), code.Size)
	}
}

func TestQR_RunHuman(t *testing.T) {
	ctx, stdout, _ := newTestContext(output.ModeHuman)
	err := Run(ctx, []string{"Hello Nova!"})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "Terminal QR Code") {
		t.Errorf("expected header in human output, got: %s", out)
	}
	if !strings.Contains(out, "Hello Nova!") {
		t.Errorf("expected text in summary, got: %s", out)
	}
}

func TestQR_PlainAndJSON(t *testing.T) {
	ctxPlain, stdoutPlain, _ := newTestContext(output.ModePlain)
	errPlain := Run(ctxPlain, []string{"--plain", "test"})
	if errPlain != nil {
		t.Fatalf("plain Run failed: %v", errPlain)
	}
	if !strings.Contains(stdoutPlain.String(), "##") {
		t.Errorf("expected ## in plain ASCII matrix output")
	}

	ctxJSON, stdoutJSON, _ := newTestContext(output.ModeJSON)
	errJSON := Run(ctxJSON, []string{"--json", "test"})
	if errJSON != nil {
		t.Fatalf("json Run failed: %v", errJSON)
	}
	if !strings.Contains(stdoutJSON.String(), `"matrix":`) || !strings.Contains(stdoutJSON.String(), `"version":`) {
		t.Errorf("expected valid JSON structure")
	}
}

func TestQR_FileInput(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "link.txt")
	_ = os.WriteFile(f, []byte("https://nova.dev"), 0o644)

	ctx, stdout, _ := newTestContext(output.ModePlain)
	err := Run(ctx, []string{"--plain", "-f", f})
	if err != nil {
		t.Fatalf("file input failed: %v", err)
	}
	if len(stdout.String()) == 0 {
		t.Errorf("expected non-empty output")
	}
}

func TestQR_EmptyInput(t *testing.T) {
	ctx, _, _ := newTestContext(output.ModePlain)
	err := Run(ctx, []string{})
	if err == nil {
		t.Fatalf("expected error for empty text, got nil")
	}
}
