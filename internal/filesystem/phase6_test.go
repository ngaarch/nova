package filesystem

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPhase6ExitGate_CopyOps(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. CopyFile with metadata preservation
	srcFile := filepath.Join(tmpDir, "source.txt")
	dstFile := filepath.Join(tmpDir, "dest.txt")
	content := []byte("Phase 6 copy metadata test data")
	if err := os.WriteFile(srcFile, content, 0755); err != nil {
		t.Fatal(err)
	}

	// Set a known past mtime
	pastTime := time.Now().Add(-24 * time.Hour).Truncate(time.Second)
	_ = os.Chtimes(srcFile, pastTime, pastTime)

	err := CopyFile(srcFile, dstFile, CopyOptions{PreserveMetadata: true})
	if err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	dstFi, err := os.Stat(dstFile)
	if err != nil {
		t.Fatalf("Stat dstFile failed: %v", err)
	}
	if dstFi.ModTime().Truncate(time.Second).Unix() != pastTime.Unix() {
		t.Errorf("expected mtime %v, got %v", pastTime, dstFi.ModTime())
	}
	if dstFi.Mode().Perm()&0700 != 0700 {
		t.Errorf("expected preserved user permissions, got %o", dstFi.Mode().Perm())
	}

	// 2. CopyDir recursive
	srcDir := filepath.Join(tmpDir, "source_dir")
	subDir := filepath.Join(srcDir, "sub")
	_ = os.MkdirAll(subDir, 0755)
	_ = os.WriteFile(filepath.Join(subDir, "leaf.txt"), []byte("leaf data"), 0644)

	dstDir := filepath.Join(tmpDir, "dest_dir")
	err = CopyDir(srcDir, dstDir, CopyOptions{Recursive: true})
	if err != nil {
		t.Fatalf("CopyDir failed: %v", err)
	}

	leafContent, err := os.ReadFile(filepath.Join(dstDir, "sub", "leaf.txt"))
	if err != nil || string(leafContent) != "leaf data" {
		t.Fatalf("leaf file in copied dir missing or invalid: %v", err)
	}

	// 3. Subdirectory recursion guard: cannot copy directory into itself
	nestedDst := filepath.Join(srcDir, "nested_child")
	err = CopyDir(srcDir, nestedDst, CopyOptions{Recursive: true})
	if !errors.Is(err, ErrDestInsideSource) {
		t.Errorf("expected ErrDestInsideSource when copying into child, got: %v", err)
	}

	// 4. NoClobber (silently skips overwrite like cp -n)
	err = CopyFile(srcFile, dstFile, CopyOptions{NoClobber: true})
	if err != nil {
		t.Fatalf("unexpected error under NoClobber: %v", err)
	}
	existingBytes, _ := os.ReadFile(dstFile)
	if string(existingBytes) != string(content) {
		t.Errorf("destination content changed under NoClobber")
	}

	// 5. DryRun
	dryDst := filepath.Join(tmpDir, "dry_copy.txt")
	err = CopyFile(srcFile, dryDst, CopyOptions{DryRun: true})
	if err != nil {
		t.Fatalf("CopyFile DryRun failed: %v", err)
	}
	if _, err := os.Stat(dryDst); !os.IsNotExist(err) {
		t.Fatalf("dryDst should not exist under DryRun")
	}

	// 6. Interactive rejection on existing destination
	confirmDst := filepath.Join(tmpDir, "confirm_copy.txt")
	_ = os.WriteFile(confirmDst, []byte("original unoverwritten"), 0644)
	err = CopyFile(srcFile, confirmDst, CopyOptions{
		Confirm: func(target string) (bool, error) {
			return false, nil
		},
	})
	if err != nil {
		t.Fatalf("expected nil when declined, got %v", err)
	}
	cBytes, _ := os.ReadFile(confirmDst)
	if string(cBytes) != "original unoverwritten" {
		t.Fatalf("declined destination was overwritten")
	}
}

func TestPhase6ExitGate_MoveOps(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Single file move
	srcFile := filepath.Join(tmpDir, "mv_src.txt")
	dstFile := filepath.Join(tmpDir, "mv_dst.txt")
	_ = os.WriteFile(srcFile, []byte("moving"), 0644)

	err := Move(srcFile, dstFile, MoveOptions{})
	if err != nil {
		t.Fatalf("Move failed: %v", err)
	}
	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Fatalf("srcFile should not exist after move")
	}
	if _, err := os.Stat(dstFile); err != nil {
		t.Fatalf("dstFile should exist after move: %v", err)
	}

	// 2. Directory move
	srcDir := filepath.Join(tmpDir, "mv_dir_src")
	dstDir := filepath.Join(tmpDir, "mv_dir_dst")
	_ = os.MkdirAll(filepath.Join(srcDir, "sub"), 0755)
	_ = os.WriteFile(filepath.Join(srcDir, "sub", "item.txt"), []byte("item"), 0644)

	err = Move(srcDir, dstDir, MoveOptions{})
	if err != nil {
		t.Fatalf("Move directory failed: %v", err)
	}
	if _, err := os.Stat(srcDir); !os.IsNotExist(err) {
		t.Fatalf("srcDir should not exist after move")
	}
	if _, err := os.Stat(filepath.Join(dstDir, "sub", "item.txt")); err != nil {
		t.Fatalf("item.txt should exist in dstDir: %v", err)
	}

	// 3. Move NoClobber (silently skips overwrite like mv -n)
	anotherFile := filepath.Join(tmpDir, "another.txt")
	_ = os.WriteFile(anotherFile, []byte("data"), 0644)
	err = Move(anotherFile, dstFile, MoveOptions{NoClobber: true})
	if err != nil {
		t.Fatalf("unexpected error under Move NoClobber: %v", err)
	}
	if _, err := os.Stat(anotherFile); err != nil {
		t.Errorf("anotherFile should not be moved under NoClobber")
	}
	dstContent, _ := os.ReadFile(dstFile)
	if string(dstContent) != "moving" {
		t.Errorf("dstFile should retain original content under NoClobber")
	}

	// 4. Move into self/sub child
	childDir := filepath.Join(dstDir, "sub", "recursive_child")
	err = Move(dstDir, childDir, MoveOptions{})
	if !errors.Is(err, ErrDestInsideSource) {
		t.Errorf("expected ErrDestInsideSource when moving into child, got: %v", err)
	}

	// 5. Move DryRun
	drySrc := filepath.Join(tmpDir, "dry_src.txt")
	dryDst := filepath.Join(tmpDir, "dry_dst.txt")
	_ = os.WriteFile(drySrc, []byte("dry"), 0644)
	err = Move(drySrc, dryDst, MoveOptions{DryRun: true})
	if err != nil {
		t.Fatalf("Move DryRun failed: %v", err)
	}
	if _, err := os.Stat(drySrc); err != nil {
		t.Fatalf("drySrc should still exist under DryRun")
	}
	if _, err := os.Stat(dryDst); !os.IsNotExist(err) {
		t.Fatalf("dryDst should not exist under DryRun")
	}
}

func TestPhase6ExitGate_RemoveOps(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. ProtectRoot
	rootPaths := []string{"/", "//", ".", "..", ""}
	for _, p := range rootPaths {
		if !ProtectRoot(p) {
			t.Errorf("expected ProtectRoot to return true for %q", p)
		}
		err := Remove(p, RemoveOptions{ProtectRoot: true, Recursive: true, Force: true})
		if !errors.Is(err, ErrRootProtected) {
			t.Errorf("expected ErrRootProtected for %q, got: %v", p, err)
		}
	}

	// 2. Directory without Recursive fails
	testDir := filepath.Join(tmpDir, "rm_dir")
	_ = os.Mkdir(testDir, 0755)
	err := Remove(testDir, RemoveOptions{Recursive: false})
	if err == nil {
		t.Fatalf("expected error when removing directory without Recursive: true")
	}

	// 3. Directory with Recursive succeeds
	err = Remove(testDir, RemoveOptions{Recursive: true})
	if err != nil {
		t.Fatalf("Remove directory with Recursive failed: %v", err)
	}
	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Fatalf("testDir should be deleted")
	}

	// 4. Symlink removal deletes symlink only, NOT target
	targetFile := filepath.Join(tmpDir, "target_file.txt")
	symlinkPath := filepath.Join(tmpDir, "file_symlink")
	_ = os.WriteFile(targetFile, []byte("protected target"), 0644)
	if err := os.Symlink(targetFile, symlinkPath); err != nil {
		t.Fatal(err)
	}

	err = Remove(symlinkPath, RemoveOptions{})
	if err != nil {
		t.Fatalf("Remove symlink failed: %v", err)
	}
	if _, err := os.Lstat(symlinkPath); !os.IsNotExist(err) {
		t.Fatalf("symlink should be removed")
	}
	if _, err := os.Stat(targetFile); err != nil {
		t.Fatalf("target file must NOT be removed when symlink is deleted: %v", err)
	}

	// Symlink to directory
	targetDir := filepath.Join(tmpDir, "target_directory")
	_ = os.Mkdir(targetDir, 0755)
	dirSymlink := filepath.Join(tmpDir, "dir_symlink")
	if err := os.Symlink(targetDir, dirSymlink); err != nil {
		t.Fatal(err)
	}

	err = Remove(dirSymlink, RemoveOptions{Recursive: false})
	if err != nil {
		t.Fatalf("Remove dir symlink without recursive should succeed: %v", err)
	}
	if _, err := os.Stat(targetDir); err != nil {
		t.Fatalf("target dir must NOT be removed when symlink is deleted: %v", err)
	}

	// 5. DryRun
	remFile := filepath.Join(tmpDir, "rem_file.txt")
	_ = os.WriteFile(remFile, []byte("keep"), 0644)
	err = Remove(remFile, RemoveOptions{DryRun: true})
	if err != nil {
		t.Fatalf("Remove DryRun failed: %v", err)
	}
	if _, err := os.Stat(remFile); err != nil {
		t.Fatalf("file should still exist under DryRun")
	}
}

func TestPhase6ExitGate_MkdirOps(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Single dir
	single := filepath.Join(tmpDir, "single")
	err := MakeDir(single, MkdirOptions{})
	if err != nil {
		t.Fatalf("MakeDir single failed: %v", err)
	}
	if fi, err := os.Stat(single); err != nil || !fi.IsDir() {
		t.Fatalf("expected single dir to exist")
	}

	// 2. Parents hierarchy
	nested := filepath.Join(tmpDir, "nest_a", "nest_b", "nest_c")
	err = MakeDir(nested, MkdirOptions{Parents: true})
	if err != nil {
		t.Fatalf("MakeDir parents failed: %v", err)
	}
	if fi, err := os.Stat(nested); err != nil || !fi.IsDir() {
		t.Fatalf("expected nested dir to exist")
	}

	// 3. Custom permissions
	permDir := filepath.Join(tmpDir, "perm_dir")
	err = MakeDir(permDir, MkdirOptions{Mode: 0700})
	if err != nil {
		t.Fatalf("MakeDir with mode failed: %v", err)
	}
	fi, err := os.Stat(permDir)
	if err != nil {
		t.Fatalf("Stat permDir failed: %v", err)
	}
	if fi.Mode().Perm()&0700 != 0700 {
		t.Errorf("expected 0700 user permissions, got: %o", fi.Mode().Perm())
	}

	// 4. DryRun
	dryDir := filepath.Join(tmpDir, "dry_mkdir")
	err = MakeDir(dryDir, MkdirOptions{DryRun: true})
	if err != nil {
		t.Fatalf("MakeDir DryRun failed: %v", err)
	}
	if _, err := os.Stat(dryDir); !os.IsNotExist(err) {
		t.Fatalf("dryDir should not exist under DryRun")
	}
}
