package grep

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

func TestGrepBasicSearch(t *testing.T) {
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "sample.txt")
	content := "hello world\nalpha beta\nhello gophers\n"
	if err := os.WriteFile(file1, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	ctx, stdout, _ := newTestContext(output.ModeHuman)

	err := Run(ctx, []string{"hello", tempDir})
	if err != nil {
		t.Fatalf("expected search to succeed, got %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "sample.txt") {
		t.Errorf("expected sample.txt in output, got: %s", out)
	}
	if !strings.Contains(out, "2 matches") {
		t.Errorf("expected 2 matches in output, got: %s", out)
	}
}

func TestGrepPlainAndJSON(t *testing.T) {
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "code.go")
	content := "package main\n\nfunc main() {\n\tprintln(\"test\")\n}\n"
	if err := os.WriteFile(file1, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Test Plain
	ctxPlain, stdoutPlain, _ := newTestContext(output.ModePlain)

	err := Run(ctxPlain, []string{"--plain", "func main", tempDir})
	if err != nil {
		t.Fatalf("plain search failed: %v", err)
	}
	outPlain := stdoutPlain.String()
	if !strings.Contains(outPlain, "code.go:3:func main() {") {
		t.Errorf("expected code.go:3:func main() {, got: %s", outPlain)
	}

	// Test JSON
	ctxJSON, stdoutJSON, _ := newTestContext(output.ModeJSON)

	err = Run(ctxJSON, []string{"--json", "func main", tempDir})
	if err != nil {
		t.Fatalf("json search failed: %v", err)
	}
	outJSON := stdoutJSON.String()
	if !strings.Contains(outJSON, `"total_matches": 1`) {
		t.Errorf("expected json total_matches: 1, got: %s", outJSON)
	}
}

func TestGrepIgnoreCaseAndFilesWithMatches(t *testing.T) {
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "f1.txt")
	file2 := filepath.Join(tempDir, "f2.txt")
	os.WriteFile(file1, []byte("CASE_INSENSITIVE_TEST\n"), 0644)
	os.WriteFile(file2, []byte("nothing here\n"), 0644)

	ctx, stdout, _ := newTestContext(output.ModeHuman)

	err := Run(ctx, []string{"-i", "-l", "case_insensitive", tempDir})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "f1.txt") {
		t.Errorf("expected f1.txt in output, got: %s", out)
	}
	if strings.Contains(out, "f2.txt") {
		t.Errorf("f2.txt should not be present, got: %s", out)
	}
}

func TestGrepSkipBinary(t *testing.T) {
	tempDir := t.TempDir()
	binFile := filepath.Join(tempDir, "binary.dat")
	binContent := []byte{0x00, 0x01, 'f', 'o', 'o', 'b', 'a', 'r', 0x00}
	os.WriteFile(binFile, binContent, 0644)

	ctx, _, _ := newTestContext(output.ModeHuman)

	err := Run(ctx, []string{"foobar", tempDir})
	if err == nil {
		t.Fatalf("expected binary file to be skipped resulting in no matches error")
	}
}
