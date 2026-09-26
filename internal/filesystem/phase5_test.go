package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestPhase5ExitGate(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Deeply nested directories (10 levels)
	deepDir := tmpDir
	for i := 1; i <= 10; i++ {
		deepDir = filepath.Join(deepDir, fmt.Sprintf("level_%02d", i))
	}
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		t.Fatal(err)
	}
	deepFile := filepath.Join(deepDir, "deep_leaf.txt")
	os.WriteFile(deepFile, []byte("deep data"), 0644)

	// 2. Symlinks (valid file, valid dir, broken symlink)
	validFileLink := filepath.Join(tmpDir, "link_to_leaf")
	_ = os.Symlink(deepFile, validFileLink)

	brokenLink := filepath.Join(tmpDir, "broken_symlink")
	_ = os.Symlink(filepath.Join(tmpDir, "non_existent_file"), brokenLink)

	dirLink := filepath.Join(tmpDir, "link_to_dir")
	_ = os.Symlink(deepDir, dirLink)

	// 3. Symlink Cycle (loop -> ancestor)
	cycleDir := filepath.Join(tmpDir, "cycle_parent")
	os.MkdirAll(cycleDir, 0755)
	cycleLink := filepath.Join(cycleDir, "self_loop")
	_ = os.Symlink(cycleDir, cycleLink)

	// 4. Special Files: Named Pipe (FIFO)
	fifoPath := filepath.Join(tmpDir, "test_fifo")
	_ = syscall.Mknod(fifoPath, syscall.S_IFIFO|0666, 0)

	// 5. Test Walk with Cycle Detection on the complex tree
	opts := DefaultWalkOptions()
	opts.FollowSymlinks = true
	opts.DetectCycles = true

	seenCount := 0
	seenDeepLeaf := false
	seenBrokenLink := false
	seenFifo := false

	err := Walk(tmpDir, opts, func(path string, entry *Entry, depth int) error {
		seenCount++
		if entry.Name == "deep_leaf.txt" {
			seenDeepLeaf = true
		}
		if entry.Name == "broken_symlink" {
			seenBrokenLink = true
			if !entry.IsBroken {
				t.Errorf("expected broken_symlink to have IsBroken=true")
			}
		}
		if entry.Name == "test_fifo" {
			seenFifo = true
		}
		if seenCount > 500 {
			t.Fatal("cycle detection failed: excessive iterations in Walk")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected walk error: %v", err)
	}
	if !seenDeepLeaf {
		t.Errorf("failed to discover deeply nested leaf file")
	}
	if !seenBrokenLink {
		t.Errorf("failed to discover broken symlink")
	}

	// 6. Test BuildTree with cycles and symlinks
	treeOpts := DefaultWalkOptions()
	treeOpts.FollowSymlinks = true
	treeOpts.DetectCycles = true
	tree, err := BuildTree(tmpDir, treeOpts)
	if err != nil {
		t.Fatalf("unexpected BuildTree error: %v", err)
	}

	if tree.FileCount < 2 {
		t.Errorf("expected at least 2 files counted, got %d", tree.FileCount)
	}
	if tree.DirCount < 11 {
		t.Errorf("expected at least 11 directories counted, got %d", tree.DirCount)
	}

	// 7. Test Inaccessible / Permission Denied Directory
	noReadDir := filepath.Join(tmpDir, "no_read_dir")
	if err := os.Mkdir(noReadDir, 0755); err == nil {
		os.WriteFile(filepath.Join(noReadDir, "secret.txt"), []byte("secret"), 0644)
		// Restrict permissions
		if err := os.Chmod(noReadDir, 0000); err == nil {
			defer os.Chmod(noReadDir, 0755) // Clean up after test
			// Walk should not fail fatally or panic
			walkErr := Walk(noReadDir, DefaultWalkOptions(), func(path string, entry *Entry, depth int) error {
				return nil
			})
			_ = walkErr
		}
	}
	_ = seenFifo
}

func TestHugeTreeTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	// Generate 50 dirs with 20 files each = 1,000 files
	for d := 0; d < 50; d++ {
		sub := filepath.Join(tmpDir, fmt.Sprintf("sub_%02d", d))
		os.Mkdir(sub, 0755)
		for f := 0; f < 20; f++ {
			os.WriteFile(filepath.Join(sub, fmt.Sprintf("file_%02d.txt", f)), []byte("data"), 0644)
		}
	}

	tree, err := BuildTree(tmpDir, DefaultWalkOptions())
	if err != nil {
		t.Fatalf("unexpected BuildTree error on huge tree: %v", err)
	}

	if tree.FileCount != 1000 {
		t.Errorf("expected 1000 files in huge tree, got %d", tree.FileCount)
	}
	if tree.DirCount != 51 { // root + 50 subs
		t.Errorf("expected 51 directories in huge tree, got %d", tree.DirCount)
	}
}
