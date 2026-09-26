package ls

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nova/internal/command"
	"nova/internal/config"
	"nova/internal/filesystem"
	"nova/internal/logging"
	"nova/internal/output"
	"nova/internal/terminal"
	"nova/internal/theme"
)

func TestParseFlags(t *testing.T) {
	opts := ParseFlags([]string{"-l", "-a", "-h", "-r", "--sort=size", "dir1", "dir2"})
	if !opts.Long || !opts.All || !opts.HumanReadable || !opts.Reverse {
		t.Errorf("flags parsed incorrectly: %+v", opts)
	}
	if opts.Sort != "size" {
		t.Errorf("expected sort=size, got %q", opts.Sort)
	}
	if len(opts.Paths) != 2 || opts.Paths[0] != "dir1" || opts.Paths[1] != "dir2" {
		t.Errorf("unexpected paths: %v", opts.Paths)
	}

	// Bundled flags
	bundled := ParseFlags([]string{"-alR"})
	if !bundled.All || !bundled.Long || !bundled.Recursive {
		t.Errorf("bundled flags failed: %+v", bundled)
	}
}

func TestFormatSize(t *testing.T) {
	cases := []struct {
		bytes    int64
		human    bool
		expected string
	}{
		{500, true, "500 B"},
		{1024, true, "1.0 KB"},
		{2048, true, "2.0 KB"},
		{1024 * 1024 * 5, true, "5.0 MB"},
		{1024 * 1024 * 1024 * 3, true, "3.0 GB"},
		{1500, false, "1500"},
	}

	for _, c := range cases {
		if got := FormatSize(c.bytes, c.human); got != c.expected {
			t.Errorf("FormatSize(%d, %v) = %q; want %q", c.bytes, c.human, got, c.expected)
		}
	}
}

func TestSortEntries(t *testing.T) {
	now := time.Now()
	entries := []filesystem.Entry{
		{Name: "b.txt", Size: 300, ModTime: now.Add(-time.Hour), IsDir: false},
		{Name: "a.txt", Size: 100, ModTime: now.Add(-2 * time.Hour), IsDir: false},
		{Name: "docs", Size: 4096, ModTime: now, IsDir: true},
	}

	// Dirs first, then by name
	SortEntries(entries, "name", false, true)
	if entries[0].Name != "docs" || entries[1].Name != "a.txt" || entries[2].Name != "b.txt" {
		t.Errorf("dirsFirst + name sort failed: %v, %v, %v", entries[0].Name, entries[1].Name, entries[2].Name)
	}

	// Sort by size ascending (without dirsFirst)
	SortEntries(entries, "size", false, false)
	if entries[0].Name != "a.txt" || entries[1].Name != "b.txt" || entries[2].Name != "docs" {
		t.Errorf("size sort failed: %v, %v, %v", entries[0].Name, entries[1].Name, entries[2].Name)
	}

	// Reverse sort by size
	SortEntries(entries, "size", true, false)
	if entries[0].Name != "docs" || entries[1].Name != "b.txt" || entries[2].Name != "a.txt" {
		t.Errorf("reverse size sort failed: %v, %v, %v", entries[0].Name, entries[1].Name, entries[2].Name)
	}
}

func setupTestContext(stdout, stderr *bytes.Buffer, mode output.Mode) *command.Context {
	caps := terminal.Capabilities{
		IsTTY:            mode == output.ModeHuman,
		Width:            80,
		Height:           24,
		ColorProfile:     terminal.ColorTrueColor,
		UnicodeSupported: true,
	}
	cfg := config.Default()
	th := theme.Get("default")
	logger := logging.New(stderr, false)
	return command.NewContext(strings.NewReader(""), stdout, stderr, cfg, caps, th, mode, logger)
}

func TestLSRun(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test hierarchy
	if err := os.WriteFile(filepath.Join(tmpDir, "fileA.txt"), []byte("aaa"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "fileB.go"), []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".hidden.cfg"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("nested"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("compact listing", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		ctx := setupTestContext(&stdout, &stderr, output.ModeHuman)

		if err := Run(ctx, []string{tmpDir}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out := stdout.String()
		if !strings.Contains(out, "fileA.txt") || !strings.Contains(out, "fileB.go") || !strings.Contains(out, "sub") {
			t.Errorf("missing entries in compact output: %s", out)
		}
		if strings.Contains(out, ".hidden.cfg") {
			t.Errorf("hidden file should not be shown without -a: %s", out)
		}
	})

	t.Run("all files (-a)", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		ctx := setupTestContext(&stdout, &stderr, output.ModeHuman)

		if err := Run(ctx, []string{"-a", tmpDir}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out := stdout.String()
		if !strings.Contains(out, ".hidden.cfg") {
			t.Errorf("expected .hidden.cfg with -a: %s", out)
		}
	})

	t.Run("long listing (-l)", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		ctx := setupTestContext(&stdout, &stderr, output.ModeHuman)

		if err := Run(ctx, []string{"-l", tmpDir}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out := stdout.String()
		if !strings.Contains(out, "Permissions") || !strings.Contains(out, "Size") {
			t.Errorf("expected long table headers in human mode: %s", out)
		}
		if !strings.Contains(out, "items (") || !strings.Contains(out, "total") {
			t.Errorf("expected summary line in long listing: %s", out)
		}
	})

	t.Run("plain mode (--plain)", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		ctx := setupTestContext(&stdout, &stderr, output.ModePlain)

		if err := Run(ctx, []string{"--plain", tmpDir}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out := stdout.String()
		lines := strings.Split(strings.TrimSpace(out), "\n")
		if len(lines) != 3 { // fileA.txt, fileB.go, sub
			t.Errorf("expected 3 plain lines, got %d: %q", len(lines), out)
		}
	})

	t.Run("JSON mode (--json)", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		ctx := setupTestContext(&stdout, &stderr, output.ModeJSON)

		if err := Run(ctx, []string{"--json", tmpDir}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var entries []filesystem.Entry
		if err := json.Unmarshal(stdout.Bytes(), &entries); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v; raw: %s", err, stdout.String())
		}
		if len(entries) != 3 {
			t.Errorf("expected 3 entries in JSON, got %d", len(entries))
		}
	})

	t.Run("recursive listing (-R)", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		ctx := setupTestContext(&stdout, &stderr, output.ModeHuman)

		if err := Run(ctx, []string{"-R", tmpDir}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out := stdout.String()
		if !strings.Contains(out, "nested.txt") {
			t.Errorf("expected recursive listing to find nested.txt: %s", out)
		}
	})

	t.Run("non-existent path error", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		ctx := setupTestContext(&stdout, &stderr, output.ModeHuman)

		err := Run(ctx, []string{filepath.Join(tmpDir, "missing")})
		if err == nil {
			t.Fatalf("expected error on missing path, got nil")
		}
	})
}
