package archive

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

func TestArchivePackAndUnpackZip(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	_ = os.WriteFile(file1, []byte("contents of file 1"), 0644)
	_ = os.WriteFile(file2, []byte("contents of file 2 with more text"), 0644)

	zipOut := filepath.Join(tmpDir, "bundle.zip")

	// 1. Pack
	ctx, stdout, _ := newTestContext(output.ModeHuman)
	err := Run(ctx, []string{"pack", "-o", zipOut, file1, file2})
	if err != nil {
		t.Fatalf("pack failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "Archive Created") {
		t.Errorf("expected 'Archive Created' in output, got: %s", stdout.String())
	}

	// 2. List
	stdout.Reset()
	err = Run(ctx, []string{"list", zipOut})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "file1.txt") {
		t.Errorf("expected file1.txt in archive list, got: %s", stdout.String())
	}

	// 3. Unpack
	unpackDir := filepath.Join(tmpDir, "unpacked")
	stdout.Reset()
	err = Run(ctx, []string{"unpack", zipOut, "-C", unpackDir})
	if err != nil {
		t.Fatalf("unpack failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "Archive Extracted") {
		t.Errorf("expected 'Archive Extracted' in output, got: %s", stdout.String())
	}

	// Verify extracted contents
	unpackedFile1 := filepath.Join(unpackDir, "file1.txt")
	data, err := os.ReadFile(unpackedFile1)
	if err != nil || string(data) != "contents of file 1" {
		t.Errorf("extracted content mismatch: %v, got %s", err, string(data))
	}
}

func TestArchiveTarGz(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "tar_sample.txt")
	_ = os.WriteFile(file1, []byte("tar sample test"), 0644)

	tarOut := filepath.Join(tmpDir, "bundle.tar.gz")

	ctx, stdout, _ := newTestContext(output.ModePlain)
	err := Run(ctx, []string{"pack", "-o", tarOut, file1})
	if err != nil {
		t.Fatalf("tar pack failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "bundle.tar.gz") {
		t.Errorf("expected bundle.tar.gz in plain output, got: %s", stdout.String())
	}
}
