package hash

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

func TestHash_ComputeAlgorithms(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "sample.txt")
	content := "hello nova hash\n"
	if err := os.WriteFile(testFile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	algorithms := []string{"sha256", "sha512", "sha1", "md5", "crc32"}
	for _, algo := range algorithms {
		t.Run(algo, func(t *testing.T) {
			ctx, stdout, _ := newTestContext("", output.ModePlain)
			err := Run(ctx, []string{"--algo=" + algo, testFile})
			if err != nil {
				t.Fatalf("Run failed for algo %s: %v", algo, err)
			}
			out := stdout.String()
			if !strings.Contains(out, testFile) {
				t.Errorf("expected output to contain file path %s, got: %s", testFile, out)
			}
			parts := strings.Fields(out)
			if len(parts) < 2 {
				t.Fatalf("expected at least hash and filename, got %v", parts)
			}
			digest := parts[0]
			if len(digest) == 0 {
				t.Errorf("empty digest for algo %s", algo)
			}
		})
	}
}

func TestHash_JSONOutput(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "data.bin")
	_ = os.WriteFile(testFile, []byte("structured test"), 0o644)

	ctx, stdout, _ := newTestContext("", output.ModeJSON)
	err := Run(ctx, []string{"--json", testFile})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, `"hash":`) || !strings.Contains(out, `"file":`) {
		t.Errorf("expected valid JSON array, got: %s", out)
	}
}

func TestHash_CheckMode(t *testing.T) {
	tmpDir := t.TempDir()
	f1 := filepath.Join(tmpDir, "file1.txt")
	f2 := filepath.Join(tmpDir, "file2.txt")
	_ = os.WriteFile(f1, []byte("content 1\n"), 0o644)
	_ = os.WriteFile(f2, []byte("content 2\n"), 0o644)

	h1, _, _ := HashFile(f1, "sha256")
	h2, _, _ := HashFile(f2, "sha256")

	// Generate checkfile
	checkFile := filepath.Join(tmpDir, "checksums.txt")
	checkContent := h1 + "  file1.txt\n" + h2 + "  file2.txt\n"
	_ = os.WriteFile(checkFile, []byte(checkContent), 0o644)

	// Run check
	ctx, stdout, _ := newTestContext("", output.ModePlain)
	err := Run(ctx, []string{"-c", checkFile})
	if err != nil {
		t.Fatalf("expected check to pass, got err: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "file1.txt: PASS") || !strings.Contains(out, "file2.txt: PASS") {
		t.Errorf("expected PASS results in plain mode, got: %s", out)
	}

	// Mismatched file check
	corruptContent := "0000000000000000000000000000000000000000000000000000000000000000  file1.txt\n"
	corruptCheck := filepath.Join(tmpDir, "corrupt.txt")
	_ = os.WriteFile(corruptCheck, []byte(corruptContent), 0o644)

	ctxFail, _, _ := newTestContext("", output.ModePlain)
	errFail := Run(ctxFail, []string{"-c", corruptCheck})
	if errFail == nil {
		t.Fatalf("expected failure for corrupt checksum, got nil")
	}
}

func TestHash_Recursive(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	_ = os.MkdirAll(subDir, 0o755)
	_ = os.WriteFile(filepath.Join(subDir, "a.txt"), []byte("a"), 0o644)
	_ = os.WriteFile(filepath.Join(subDir, "b.txt"), []byte("b"), 0o644)

	ctx, stdout, _ := newTestContext("", output.ModePlain)
	err := Run(ctx, []string{"-r", tmpDir})
	if err != nil {
		t.Fatalf("recursive hash failed: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "a.txt") || !strings.Contains(out, "b.txt") {
		t.Errorf("expected both recursive files in output, got: %s", out)
	}
}

func TestHash_InvalidAlgo(t *testing.T) {
	ctx, _, _ := newTestContext("", output.ModePlain)
	err := Run(ctx, []string{"--algo=nonexistent", "dummy.txt"})
	if err == nil {
		t.Fatalf("expected error on unsupported algorithm, got nil")
	}
}
