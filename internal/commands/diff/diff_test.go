package diff

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

func setupDiffFiles(t *testing.T) (string, string) {
	t.Helper()
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	content1 := "apple\nbanana\ncherry\ndate\n"
	content2 := "apple\nblueberry\ncherry\ndate\nelderberry\n"

	_ = os.WriteFile(file1, []byte(content1), 0644)
	_ = os.WriteFile(file2, []byte(content2), 0644)

	return file1, file2
}

func TestDiffHuman(t *testing.T) {
	f1, f2 := setupDiffFiles(t)
	ctx, stdout, _ := newTestContext(output.ModeHuman)

	err := Run(ctx, []string{f1, f2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "banana") || !strings.Contains(out, "blueberry") {
		t.Errorf("expected diff to show changed lines, got: %s", out)
	}
	if !strings.Contains(out, "additions") || !strings.Contains(out, "deletions") {
		t.Errorf("expected summary in human output, got: %s", out)
	}
}

func TestDiffIdentical(t *testing.T) {
	tmpDir := t.TempDir()
	f1 := filepath.Join(tmpDir, "same1.txt")
	f2 := filepath.Join(tmpDir, "same2.txt")
	_ = os.WriteFile(f1, []byte("exact content\n"), 0644)
	_ = os.WriteFile(f2, []byte("exact content\n"), 0644)

	ctx, stdout, _ := newTestContext(output.ModeHuman)
	err := Run(ctx, []string{f1, f2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "identical") {
		t.Errorf("expected identical notice, got: %s", out)
	}
}

func TestDiffBrief(t *testing.T) {
	f1, f2 := setupDiffFiles(t)
	ctx, stdout, _ := newTestContext(output.ModePlain)

	err := Run(ctx, []string{"-q", f1, f2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "differ") {
		t.Errorf("expected 'differ' in brief output, got: %s", out)
	}
}

func TestDiffPlain(t *testing.T) {
	f1, f2 := setupDiffFiles(t)
	ctx, stdout, _ := newTestContext(output.ModePlain)

	err := Run(ctx, []string{"--plain", f1, f2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "--- ") || !strings.Contains(out, "+++ ") || !strings.Contains(out, "@@") {
		t.Errorf("expected unified diff format, got: %s", out)
	}
	if !strings.Contains(out, "-banana") || !strings.Contains(out, "+blueberry") {
		t.Errorf("expected -banana and +blueberry in plain diff, got: %s", out)
	}
}

func TestDiffJSON(t *testing.T) {
	f1, f2 := setupDiffFiles(t)
	ctx, stdout, _ := newTestContext(output.ModeJSON)

	err := Run(ctx, []string{"--json", f1, f2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, `"identical": false`) || !strings.Contains(out, `"additions": 2`) {
		t.Errorf("expected structured JSON output, got: %s", out)
	}
}

func TestDiffBinary(t *testing.T) {
	tmpDir := t.TempDir()
	bin1 := filepath.Join(tmpDir, "b1.bin")
	bin2 := filepath.Join(tmpDir, "b2.bin")
	_ = os.WriteFile(bin1, []byte{0x00, 0x01, 0x02}, 0644)
	_ = os.WriteFile(bin2, []byte{0x00, 0x01, 0x03}, 0644)

	ctx, stdout, _ := newTestContext(output.ModeHuman)
	err := Run(ctx, []string{bin1, bin2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "Binary") || !strings.Contains(out, "differ") {
		t.Errorf("expected binary notice, got: %s", out)
	}
}

func TestDiffMissingOperand(t *testing.T) {
	ctx, _, _ := newTestContext(output.ModeHuman)
	err := Run(ctx, []string{"only_one_file"})
	if err == nil {
		t.Fatalf("expected error on missing operand")
	}
}
