package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"nova/internal/theme"
)

func TestReadDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	f1 := filepath.Join(tmpDir, "file1.txt")
	if err := os.WriteFile(f1, []byte("hello"), 0o644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}

	dotFile := filepath.Join(tmpDir, ".hidden")
	if err := os.WriteFile(dotFile, []byte("secret"), 0o600); err != nil {
		t.Fatalf("failed to create .hidden: %v", err)
	}

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// ReadDir without hidden
	entries, err := ReadDir(tmpDir, false, true)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries (excluding hidden), got %d", len(entries))
	}

	// ReadDir with hidden
	entriesWithHidden, err := ReadDir(tmpDir, true, true)
	if err != nil {
		t.Fatalf("ReadDir with hidden failed: %v", err)
	}
	if len(entriesWithHidden) != 3 {
		t.Errorf("expected 3 entries (including hidden), got %d", len(entriesWithHidden))
	}
}

func TestReadDirSingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "single.go")
	if err := os.WriteFile(filePath, []byte("package main"), 0o644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	entries, err := ReadDir(filePath, false, true)
	if err != nil {
		t.Fatalf("ReadDir single file failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Name != "single.go" || entries[0].IsDir {
		t.Errorf("unexpected entry metadata: %+v", entries[0])
	}
}

func TestSymlinkHandling(t *testing.T) {
	tmpDir := t.TempDir()

	targetPath := filepath.Join(tmpDir, "target.txt")
	if err := os.WriteFile(targetPath, []byte("target"), 0o644); err != nil {
		t.Fatalf("failed to write target: %v", err)
	}

	// Valid symlink
	validLink := filepath.Join(tmpDir, "link.txt")
	if err := os.Symlink(targetPath, validLink); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// Broken symlink
	brokenLink := filepath.Join(tmpDir, "broken.txt")
	if err := os.Symlink(filepath.Join(tmpDir, "nonexistent.txt"), brokenLink); err != nil {
		t.Fatalf("failed to create broken symlink: %v", err)
	}

	entries, err := ReadDir(tmpDir, false, true)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	var foundValid, foundBroken bool
	for _, e := range entries {
		if e.Name == "link.txt" {
			foundValid = true
			if !e.IsSymlink || e.IsBroken || e.LinkTarget != targetPath {
				t.Errorf("valid symlink mismatch: %+v", e)
			}
			if e.EntityType != theme.TypeSymlink {
				t.Errorf("expected TypeSymlink, got %v", e.EntityType)
			}
		}
		if e.Name == "broken.txt" {
			foundBroken = true
			if !e.IsSymlink || !e.IsBroken {
				t.Errorf("broken symlink mismatch: %+v", e)
			}
			if e.EntityType != theme.TypeBrokenSymlink {
				t.Errorf("expected TypeBrokenSymlink, got %v", e.EntityType)
			}
		}
	}

	if !foundValid || !foundBroken {
		t.Errorf("failed to locate valid or broken symlink: valid=%v, broken=%v", foundValid, foundBroken)
	}
}

func TestClassifyEntityType(t *testing.T) {
	cases := []struct {
		mode      os.FileMode
		isSymlink bool
		isBroken  bool
		expected  theme.EntityType
	}{
		{os.ModeDir | 0o755, false, false, theme.TypeDirectory},
		{0o755, false, false, theme.TypeExecutable},
		{0o644, false, false, theme.TypeRegular},
		{os.ModeSymlink | 0o777, true, false, theme.TypeSymlink},
		{os.ModeSymlink | 0o777, true, true, theme.TypeBrokenSymlink},
		{os.ModeNamedPipe | 0o600, false, false, theme.TypePipe},
		{os.ModeSocket | 0o600, false, false, theme.TypeSocket},
		{os.ModeDevice | 0o600, false, false, theme.TypeDevice},
	}

	for _, c := range cases {
		got := ClassifyEntityType(c.mode, c.isSymlink, c.isBroken)
		if got != c.expected {
			t.Errorf("ClassifyEntityType(%v, %v, %v) = %v; want %v",
				c.mode, c.isSymlink, c.isBroken, got, c.expected)
		}
	}
}
