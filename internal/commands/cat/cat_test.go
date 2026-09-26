package cat

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

	cfg := config.DefaultConfig()
	caps := terminal.Capabilities{
		IsTTY:        mode == output.ModeHuman,
		Width:        80,
		Height:       24,
		ColorProfile: profile,
	}
	th := theme.ThemeDefault
	logger := logging.NewLogger(&stderr, false)

	ctx := command.NewContext(&in, &stdout, &stderr, cfg, caps, th, mode, logger)
	return ctx, &stdout, &stderr
}

func TestCatTinyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tiny.txt")
	os.WriteFile(path, []byte("Hello, Nova!\n"), 0644)

	ctx, stdout, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stdout.String() != "Hello, Nova!\n" {
		t.Errorf("got %q, want %q", stdout.String(), "Hello, Nova!\n")
	}
}

func TestCatEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")
	os.WriteFile(path, []byte(""), 0644)

	ctx, stdout, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stdout.Len() != 0 {
		t.Errorf("expected 0 bytes output for empty file, got: %q", stdout.String())
	}
}

func TestCatStdin(t *testing.T) {
	ctx, stdout, _ := newTestContext("piped input line 1\npiped input line 2\n", output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{"-"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "piped input line 1\npiped input line 2\n"
	if stdout.String() != expected {
		t.Errorf("got %q, want %q", stdout.String(), expected)
	}
}

func TestCatLineNumbers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lines.txt")
	os.WriteFile(path, []byte("first\n\nthird\n"), 0644)

	// Test -n (number all)
	ctx, stdout, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{"-n", path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "   1 │ first") {
		t.Errorf("expected line 1 numbered, got: %s", out)
	}
	if !strings.Contains(out, "   2 │ ") {
		t.Errorf("expected line 2 numbered, got: %s", out)
	}
	if !strings.Contains(out, "   3 │ third") {
		t.Errorf("expected line 3 numbered, got: %s", out)
	}

	// Test -b (number non-blank)
	ctx2, stdout2, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err2 := Run(ctx2, []string{"-b", path})
	if err2 != nil {
		t.Fatalf("unexpected error: %v", err2)
	}
	out2 := stdout2.String()
	if !strings.Contains(out2, "   1 │ first") {
		t.Errorf("expected line 1 numbered, got: %s", out2)
	}
	if strings.Contains(out2, "   2 │ \n") {
		t.Errorf("expected line 2 to not be numbered with -b, got: %s", out2)
	}
	if !strings.Contains(out2, "   2 │ third") {
		t.Errorf("expected third line to be numbered 2 with -b, got: %s", out2)
	}
}

func TestCatSqueezeBlank(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "squeeze.txt")
	os.WriteFile(path, []byte("line 1\n\n\n\nline 2\n"), 0644)

	ctx, stdout, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{"-s", path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "line 1\n\nline 2\n"
	if stdout.String() != expected {
		t.Errorf("got %q, want %q", stdout.String(), expected)
	}
}

func TestCatShowEndsAndTabs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tabs.txt")
	os.WriteFile(path, []byte("col1\tcol2\n"), 0644)

	ctx, stdout, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{"-E", "-T", path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "col1^Icol2$\n"
	if stdout.String() != expected {
		t.Errorf("got %q, want %q", stdout.String(), expected)
	}
}

func TestCatBinaryFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.bin")
	binData := []byte("ELF\x02\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00")
	os.WriteFile(path, binData, 0644)

	// Human mode: displays warning banner and hex preview
	ctxHuman, stdoutHuman, _ := newTestContext("", output.ModeHuman, terminal.ColorNone)
	err := Run(ctxHuman, []string{path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	outHuman := stdoutHuman.String()
	if !strings.Contains(outHuman, "binary file") {
		t.Errorf("expected binary file warning banner in human mode, got: %s", outHuman)
	}

	// Plain mode: passes raw bytes for pipeline safety
	ctxPlain, stdoutPlain, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err = Run(ctxPlain, []string{path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(stdoutPlain.Bytes(), binData) {
		t.Errorf("expected exact raw binary bytes in plain mode, got %q", stdoutPlain.Bytes())
	}

	// Hex dump mode (--hex)
	ctxHex, stdoutHex, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err = Run(ctxHex, []string{"--hex", path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdoutHex.String(), "00000000") || !strings.Contains(stdoutHex.String(), "|ELF") {
		t.Errorf("expected hex dump output with --hex, got: %s", stdoutHex.String())
	}
}

func TestCatUTF8(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "utf8.txt")
	content := "こんにちは世界！ 🚀 ✨ 🌟 — Modern CLI suite\n"
	os.WriteFile(path, []byte(content), 0644)

	ctx, stdout, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout.String() != content {
		t.Errorf("got %q, want %q", stdout.String(), content)
	}
}

func TestCatInvalidUTF8(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid_utf8.txt")
	// Valid ASCII followed by invalid UTF-8 sequence \xff\xfe
	data := []byte("Valid text with invalid: \xff\xfe end\n")
	os.WriteFile(path, data, 0644)

	ctx, stdout, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "\\xff\\xfe") {
		t.Errorf("expected sanitized hex representation of invalid UTF-8 bytes, got: %s", out)
	}
}

func TestCatLongLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "long.txt")
	// 100,000 characters line (would fail with standard bufio.Scanner default buffer)
	longLine := strings.Repeat("A", 100000) + "\n"
	os.WriteFile(path, []byte(longLine), 0644)

	ctx, stdout, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout.Len() != len(longLine) {
		t.Errorf("expected %d bytes, got %d", len(longLine), stdout.Len())
	}
}

func TestCatLargeFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "large.txt")
	var b strings.Builder
	for i := 0; i < 5000; i++ {
		b.WriteString("line item in large file\n")
	}
	os.WriteFile(path, []byte(b.String()), 0644)

	ctx, stdout, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lineCount := strings.Count(stdout.String(), "\n")
	if lineCount != 5000 {
		t.Errorf("expected 5000 lines, got %d", lineCount)
	}
}

func TestCatDirectoryError(t *testing.T) {
	dir := t.TempDir()
	ctx, _, _ := newTestContext("", output.ModePlain, terminal.ColorNone)
	err := Run(ctx, []string{dir})
	if err == nil {
		t.Fatalf("expected error when trying to cat a directory")
	}
	exitErr, ok := err.(*command.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got %T: %v", err, err)
	}
	if !strings.Contains(exitErr.Message, "is a directory") {
		t.Errorf("expected 'is a directory' message, got: %s", exitErr.Message)
	}
}

func TestCatJSONFormatting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	compactJSON := "{\"name\":\"nova\",\"active\":true,\"count\":10}"
	os.WriteFile(path, []byte(compactJSON), 0644)

	ctx, stdout, _ := newTestContext("", output.ModeHuman, terminal.ColorNone)
	err := Run(ctx, []string{"--json", path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "{\n  \"active\": true") || !strings.Contains(out, "  \"count\": 10") {
		t.Errorf("expected pretty-printed JSON, got: %s", out)
	}
}
