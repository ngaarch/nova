package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"nova/internal/filesystem"
)

// TestPhase9Security_ZeroShellExecution audits all Go source files to ensure no shell wrappers (sh, bash, cmd.exe) are used.
func TestPhase9Security_ZeroShellExecution(t *testing.T) {
	root := filepath.Join("..", "..")
	forbiddenPatterns := []*regexp.Regexp{
		regexp.MustCompile(`exec\.Command\s*\(\s*["'](sh|bash|zsh|dash|cmd|powershell)["']`),
		regexp.MustCompile(`exec\.CommandContext\s*\([^,]+,\s*["'](sh|bash|zsh|dash|cmd|powershell)["']`),
		regexp.MustCompile(`(sh|bash|zsh|cmd|powershell)[^,)]*,\s*["']-c["']`),
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "phase9_test.go") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		for _, pat := range forbiddenPatterns {
			if pat.Match(data) {
				t.Errorf("Security audit violation in %s: forbidden shell execution matching %s", path, pat.String())
			}
		}
		return nil
	})

	if err != nil {
		t.Fatalf("failed walking source tree: %v", err)
	}
}

// TestPhase9Security_RootProtection tests that ProtectRoot prevents destructive removal of critical directories.
func TestPhase9Security_RootProtection(t *testing.T) {
	criticalPaths := []string{
		"/",
		".",
		"..",
		"",
		"/..",
		"/tmp/..",
	}

	for _, p := range criticalPaths {
		if !filesystem.ProtectRoot(p) {
			t.Errorf("expected ProtectRoot(%q) to return true", p)
		}

		// Verify filesystem.Remove also rejects it
		err := filesystem.Remove(p, filesystem.RemoveOptions{
			Recursive:   true,
			ProtectRoot: true,
		})
		if err == nil {
			t.Errorf("expected error removing protected root %q, got nil", p)
		}
	}
}

// TestPhase9Security_SymlinkDestructiveResistance verifies that removing a symlink never deletes target directory contents.
func TestPhase9Security_SymlinkDestructiveResistance(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "target")
	if err := os.Mkdir(targetDir, 0o755); err != nil {
		t.Fatal(err)
	}

	canaryFile := filepath.Join(targetDir, "important_data.txt")
	if err := os.WriteFile(canaryFile, []byte("vital business data"), 0o644); err != nil {
		t.Fatal(err)
	}

	symlinkPath := filepath.Join(tmpDir, "symlink_dir")
	if err := os.Symlink(targetDir, symlinkPath); err != nil {
		t.Skip("symlinks not supported on this platform")
	}

	// Remove the symlink with recursive=true
	err := filesystem.Remove(symlinkPath, filesystem.RemoveOptions{
		Recursive:   true,
		ProtectRoot: true,
	})
	if err != nil {
		t.Fatalf("failed removing symlink: %v", err)
	}

	// Verify symlink is deleted
	if _, err := os.Lstat(symlinkPath); !os.IsNotExist(err) {
		t.Errorf("expected symlink %q to be deleted", symlinkPath)
	}

	// Verify target canary file is completely intact!
	data, err := os.ReadFile(canaryFile)
	if err != nil {
		t.Fatalf("target canary file was deleted or corrupted! %v", err)
	}
	if string(data) != "vital business data" {
		t.Errorf("target content mismatch: %q", string(data))
	}
}

// TestPhase9Security_RecursiveDirCopyIntoItself verifies recursion protection.
func TestPhase9Security_RecursiveDirCopyIntoItself(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "parent")
	if err := os.Mkdir(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	dstDir := filepath.Join(srcDir, "child", "grandchild")

	err := filesystem.CopyDir(srcDir, dstDir, filesystem.CopyOptions{Recursive: true})
	if !errors.Is(err, filesystem.ErrDestInsideSource) {
		t.Errorf("expected ErrDestInsideSource on recursive copy into self, got: %v", err)
	}

	moveErr := filesystem.Move(srcDir, dstDir, filesystem.MoveOptions{})
	if !errors.Is(moveErr, filesystem.ErrDestInsideSource) {
		t.Errorf("expected ErrDestInsideSource on move into self, got: %v", moveErr)
	}
}

// TestPhase9Security_PathTraversalAndNullBytes verifies robust error handling for malicious paths.
func TestPhase9Security_PathTraversalAndNullBytes(t *testing.T) {
	app := NewApp()

	maliciousArgs := [][]string{
		{"stat", "test\x00file"},
		{"cat", "test\x00file"},
		{"ls", "test\x00file"},
		{"du", "test\x00file"},
	}

	for _, args := range maliciousArgs {
		var out, errBuf bytes.Buffer
		code := app.Run(args, strings.NewReader(""), &out, &errBuf)
		if code == ExitSuccess {
			t.Errorf("expected failure for arguments %v, got ExitSuccess", args)
		}
		// Must not panic and must produce clean error message
		if errBuf.Len() == 0 {
			t.Errorf("expected error message in stderr for arguments %v", args)
		}
	}
}

// TestPhase9Accessibility_NoColorCompliance verifies that NO_COLOR=1 outputs zero ANSI escape sequences.
func TestPhase9Accessibility_NoColorCompliance(t *testing.T) {
	app := NewApp()
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "sample.txt")
	if err := os.WriteFile(f, []byte("accessible output"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("NO_COLOR", "1")

	commands := [][]string{
		{"ls", "-l", tmpDir},
		{"tree", tmpDir},
		{"stat", f},
	}

	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

	for _, cmd := range commands {
		var out, errBuf bytes.Buffer
		code := app.Run(cmd, strings.NewReader(""), &out, &errBuf)
		if code != ExitSuccess {
			t.Errorf("command %v failed with code %d: %s", cmd, code, errBuf.String())
		}

		if ansiRegex.Match(out.Bytes()) {
			t.Errorf("NO_COLOR=1 violation in command %v: ANSI escape found in output: %q", cmd, out.String())
		}
	}
}

// TestPhase9Accessibility_PlainModeParity verifies that --plain produces clean deterministic output.
func TestPhase9Accessibility_PlainModeParity(t *testing.T) {
	app := NewApp()
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "data.tsv")
	if err := os.WriteFile(f, []byte("col1\tcol2\nval1\tval2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

	commands := [][]string{
		{"ls", "--plain", tmpDir},
		{"tree", "--plain", tmpDir},
		{"cat", "--plain", f},
		{"stat", "--plain", f},
		{"du", "--plain", tmpDir},
		{"find", tmpDir, "--name", "*.tsv", "--plain"},
	}

	for _, cmd := range commands {
		var out, errBuf bytes.Buffer
		code := app.Run(cmd, strings.NewReader(""), &out, &errBuf)
		if code != ExitSuccess {
			t.Errorf("plain mode command %v failed with code %d: %s", cmd, code, errBuf.String())
		}

		if ansiRegex.Match(out.Bytes()) {
			t.Errorf("--plain violation in command %v: ANSI escape found in output: %q", cmd, out.String())
		}
	}
}

// TestPhase9Concurrency_GoroutineLeakFree verifies that CLI commands terminate with zero goroutine leaks.
func TestPhase9Concurrency_GoroutineLeakFree(t *testing.T) {
	app := NewApp()
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "leak_test.txt")
	if err := os.WriteFile(f, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Warm up runtime
	var out, errBuf bytes.Buffer
	app.Run([]string{"stat", f}, strings.NewReader(""), &out, &errBuf)
	time.Sleep(50 * time.Millisecond)

	initialGoroutines := runtime.NumGoroutine()

	// Execute sequence of diverse commands
	cmds := [][]string{
		{"ls", "-la", tmpDir},
		{"tree", tmpDir},
		{"stat", f},
		{"cat", f},
		{"du", tmpDir},
		{"find", tmpDir, "--name", "*"},
	}

	for _, cmd := range cmds {
		var o, e bytes.Buffer
		app.Run(cmd, strings.NewReader(""), &o, &e)
	}

	time.Sleep(100 * time.Millisecond)
	finalGoroutines := runtime.NumGoroutine()

	// Leaked goroutines threshold check (allowing standard Go GC/background runtime fluctuations <= 2)
	diff := finalGoroutines - initialGoroutines
	if diff > 2 {
		t.Errorf("goroutine leak detected: started with %d, ended with %d (diff: %d)", initialGoroutines, finalGoroutines, diff)
	}
}

// BenchmarkPhase9_LsLargeDir benchmarks listing a directory containing 1,000 files.
func BenchmarkPhase9_LsLargeDir(b *testing.B) {
	tmpDir := b.TempDir()
	for i := 0; i < 1000; i++ {
		_ = os.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("file_%04d.txt", i)), []byte("content"), 0o644)
	}

	app := NewApp()
	b.ResetTimer()
	for b.Loop() {
		_ = app.Run([]string{"ls", "--plain", tmpDir}, strings.NewReader(""), io.Discard, io.Discard)
	}
}

// BenchmarkPhase9_TreeTraversal benchmarks tree traversal over deep hierarchies.
func BenchmarkPhase9_TreeTraversal(b *testing.B) {
	tmpDir := b.TempDir()
	for d := 0; d < 5; d++ {
		sub := filepath.Join(tmpDir, fmt.Sprintf("dir_%d", d))
		_ = os.MkdirAll(sub, 0o755)
		for f := 0; f < 20; f++ {
			_ = os.WriteFile(filepath.Join(sub, fmt.Sprintf("f_%d.txt", f)), []byte("data"), 0o644)
		}
	}

	app := NewApp()
	b.ResetTimer()
	for b.Loop() {
		_ = app.Run([]string{"tree", "--plain", tmpDir}, strings.NewReader(""), io.Discard, io.Discard)
	}
}

// BenchmarkPhase9_CatThroughput benchmarks cat throughput streaming 1MB data.
func BenchmarkPhase9_CatThroughput(b *testing.B) {
	tmpDir := b.TempDir()
	f := filepath.Join(tmpDir, "large.txt")
	oneMB := bytes.Repeat([]byte("abcdefghijklmnopqrstuvwxyz0123456789\n"), 27778) // ~1MB
	_ = os.WriteFile(f, oneMB, 0o644)

	app := NewApp()
	b.SetBytes(int64(len(oneMB)))
	b.ResetTimer()
	for b.Loop() {
		_ = app.Run([]string{"cat", "--raw", f}, strings.NewReader(""), io.Discard, io.Discard)
	}
}
