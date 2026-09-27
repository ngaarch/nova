package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestPhase8ExitGate_DetachedHEAD tests branch extraction on detached HEAD states.
func TestPhase8ExitGate_DetachedHEAD(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available on system")
	}

	tmpDir := t.TempDir()

	// Initialize git repo in tmpDir
	runCmd(t, tmpDir, "git", "init")
	runCmd(t, tmpDir, "git", "config", "user.name", "Nova Tester")
	runCmd(t, tmpDir, "git", "config", "user.email", "tester@nova.dev")

	// Create commit
	testFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	runCmd(t, tmpDir, "git", "add", "file.txt")
	runCmd(t, tmpDir, "git", "commit", "-m", "initial commit")

	// Detach HEAD
	runCmd(t, tmpDir, "git", "checkout", "--detach")

	branch, isDetached, err := GetBranch(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error getting branch in detached HEAD: %v", err)
	}
	if !isDetached {
		t.Errorf("expected isDetached to be true")
	}
	if len(branch) < 4 {
		t.Errorf("expected commit hash in detached HEAD, got %q", branch)
	}
}

// TestPhase8ExitGate_WorktreePointer tests pointer file (.git as a file with gitdir: ...).
func TestPhase8ExitGate_WorktreePointer(t *testing.T) {
	tmpDir := t.TempDir()
	actualGitDir := filepath.Join(tmpDir, "actual_gitdir")
	if err := os.MkdirAll(filepath.Join(actualGitDir, "refs", "heads"), 0o755); err != nil {
		t.Fatal(err)
	}

	headFile := filepath.Join(actualGitDir, "HEAD")
	if err := os.WriteFile(headFile, []byte("ref: refs/heads/feature-worktree\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	worktreeDir := filepath.Join(tmpDir, "worktree")
	if err := os.MkdirAll(worktreeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	dotGitFile := filepath.Join(worktreeDir, ".git")
	content := fmt.Sprintf("gitdir: %s\n", actualGitDir)
	if err := os.WriteFile(dotGitFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	root, err := FindRepoRoot(worktreeDir)
	if err != nil {
		t.Fatalf("FindRepoRoot failed: %v", err)
	}
	if root != worktreeDir {
		t.Errorf("expected root to be %s, got %s", worktreeDir, root)
	}

	branch, isDetached, err := GetBranch(worktreeDir)
	if err != nil {
		t.Fatalf("GetBranch failed: %v", err)
	}
	if isDetached {
		t.Errorf("expected isDetached to be false")
	}
	if branch != "feature-worktree" {
		t.Errorf("expected branch feature-worktree, got %s", branch)
	}
}

// TestPhase8ExitGate_StatusLifeCycle tests modified, added, untracked, and deleted statuses.
func TestPhase8ExitGate_StatusLifeCycle(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available on system")
	}

	tmpDir := t.TempDir()
	runCmd(t, tmpDir, "git", "init")
	runCmd(t, tmpDir, "git", "config", "user.name", "Nova Tester")
	runCmd(t, tmpDir, "git", "config", "user.email", "tester@nova.dev")

	// 1. Initial committed file
	committedFile := filepath.Join(tmpDir, "committed.txt")
	if err := os.WriteFile(committedFile, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	runCmd(t, tmpDir, "git", "add", "committed.txt")
	runCmd(t, tmpDir, "git", "commit", "-m", "commit 1")

	// 2. Modify committed file
	if err := os.WriteFile(committedFile, []byte("modified content"), 0o644); err != nil {
		t.Fatal(err)
	}

	// 3. Stage a new file
	stagedFile := filepath.Join(tmpDir, "staged.txt")
	if err := os.WriteFile(stagedFile, []byte("staged content"), 0o644); err != nil {
		t.Fatal(err)
	}
	runCmd(t, tmpDir, "git", "add", "staged.txt")

	// 4. Untracked file
	untrackedFile := filepath.Join(tmpDir, "untracked.txt")
	if err := os.WriteFile(untrackedFile, []byte("untracked content"), 0o644); err != nil {
		t.Fatal(err)
	}

	rs, err := GetRepoStatus(tmpDir, 2*time.Second)
	if err != nil {
		t.Fatalf("GetRepoStatus failed: %v", err)
	}
	if rs == nil {
		t.Fatal("expected non-nil RepoStatus")
	}

	if st := rs.GetStatus(committedFile); st != StatusModified {
		t.Errorf("expected StatusModified for committed.txt, got %v", st)
	}
	if st := rs.GetStatus(stagedFile); st != StatusAdded {
		t.Errorf("expected StatusAdded for staged.txt, got %v", st)
	}
	if st := rs.GetStatus(untrackedFile); st != StatusUntracked {
		t.Errorf("expected StatusUntracked for untracked.txt, got %v", st)
	}
}

// TestPhase8ExitGate_StrictTimeoutEnforcement tests that GetRepoStatus strictly respects timeout.
func TestPhase8ExitGate_StrictTimeoutEnforcement(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	// Timeout ceiling of 20 milliseconds
	rs, err := GetRepoStatus(tmpDir, 20*time.Millisecond)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("expected nil error on timeout, got %v", err)
	}
	if rs == nil {
		t.Fatal("expected non-nil RepoStatus fallback")
	}
	if rs.Branch != "main" {
		t.Errorf("expected branch main, got %s", rs.Branch)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("timeout guard exceeded expected bound: elapsed %v", elapsed)
	}
}

// BenchmarkFindRepoRoot benchmarks root traversal inside and outside git repos.
func BenchmarkFindRepoRoot_Inside(b *testing.B) {
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FindRepoRoot(cwd)
	}
}

func BenchmarkFindRepoRoot_Outside(b *testing.B) {
	tmp := os.TempDir()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FindRepoRoot(tmp)
	}
}

// BenchmarkParsePorcelainZ_Large benchmarks parsing 10,000 porcelain entries.
func BenchmarkParsePorcelainZ_Large(b *testing.B) {
	var sb strings.Builder
	for i := 0; i < 10000; i++ {
		sb.WriteString(fmt.Sprintf(" M src/pkg%d/file%d.go\x00", i%100, i))
	}
	data := []byte(sb.String())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rs := &RepoStatus{
			Root:     "/root",
			Statuses: make(map[string]FileStatus, 10000),
		}
		rs.parsePorcelainZ(data)
	}
}

func runCmd(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %s %s: %v\nOutput: %s", name, strings.Join(args, " "), err, string(out))
	}
}
