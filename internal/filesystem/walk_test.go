package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWalkBasic(t *testing.T) {
	tmpDir := t.TempDir()

	// Structure:
	// tmpDir/
	//   a.txt
	//   sub/
	//     b.txt
	//     nested/
	//       c.txt
	if err := os.MkdirAll(filepath.Join(tmpDir, "sub", "nested"), 0755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("aaa"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "sub", "b.txt"), []byte("bbbb"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "sub", "nested", "c.txt"), []byte("ccccc"), 0644)

	var visitedPaths []string
	opts := DefaultWalkOptions()
	err := Walk(tmpDir, opts, func(path string, entry *Entry, depth int) error {
		rel, _ := filepath.Rel(tmpDir, path)
		visitedPaths = append(visitedPaths, rel)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected walk error: %v", err)
	}

	if len(visitedPaths) != 6 { // ".", "a.txt", "sub", "sub/b.txt", "sub/nested", "sub/nested/c.txt"
		t.Errorf("expected 6 visited paths, got %d: %v", len(visitedPaths), visitedPaths)
	}
}

func TestWalkDepthLimit(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "d1", "d2", "d3"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "d1", "d2", "d3", "file.txt"), []byte("hi"), 0644)

	opts := DefaultWalkOptions()
	opts.MaxDepth = 1 // only root + direct children

	var maxDepthSeen int
	err := Walk(tmpDir, opts, func(path string, entry *Entry, depth int) error {
		if depth > maxDepthSeen {
			maxDepthSeen = depth
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected walk error: %v", err)
	}

	if maxDepthSeen > 1 {
		t.Errorf("expected max depth 1, got %d", maxDepthSeen)
	}
}

func TestWalkCycleDetection(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(subDir, "f.txt"), []byte("ok"), 0644)

	// Create a symlink cycle: sub/loop -> sub
	loopLink := filepath.Join(subDir, "loop")
	if err := os.Symlink(subDir, loopLink); err != nil {
		t.Skip("symlinks not supported on this filesystem")
	}

	opts := DefaultWalkOptions()
	opts.FollowSymlinks = true
	opts.DetectCycles = true

	count := 0
	err := Walk(tmpDir, opts, func(path string, entry *Entry, depth int) error {
		count++
		if count > 50 {
			t.Fatal("infinite loop detected in Walk with symlink cycle!")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected walk error: %v", err)
	}
}

func TestBuildTreeAggregations(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "sub"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("12345"), 0644)        // 5 bytes
	os.WriteFile(filepath.Join(tmpDir, "sub", "b.txt"), []byte("12345678"), 0644) // 8 bytes

	opts := DefaultWalkOptions()
	tree, err := BuildTree(tmpDir, opts)
	if err != nil {
		t.Fatalf("unexpected BuildTree error: %v", err)
	}

	if tree.FileCount != 2 {
		t.Errorf("expected 2 files, got %d", tree.FileCount)
	}
	if tree.DirCount != 2 { // root + sub
		t.Errorf("expected 2 dirs, got %d", tree.DirCount)
	}
	if tree.TotalSize < 13 {
		t.Errorf("expected total size >= 13, got %d", tree.TotalSize)
	}
}

func TestGetDetailedStat(t *testing.T) {
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "sample.txt")
	os.WriteFile(fPath, []byte("stat test data\n"), 0644)

	st, err := GetDetailedStat(fPath)
	if err != nil {
		t.Fatalf("unexpected GetDetailedStat error: %v", err)
	}

	if st.Name != "sample.txt" {
		t.Errorf("expected name sample.txt, got %s", st.Name)
	}
	if st.Size != 15 {
		t.Errorf("expected size 15, got %d", st.Size)
	}
	if st.ModeString != "-rw-r--r--" {
		t.Errorf("expected mode string -rw-r--r--, got %s", st.ModeString)
	}
	if st.Inode == 0 {
		t.Errorf("expected non-zero inode")
	}
}
