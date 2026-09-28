package gitcmd

import (
	"bytes"
	"os"
	"os/exec"
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
	var stdout, stderr bytes.Buffer
	in := strings.NewReader("")
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

func setupGitRepo(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	runGit(t, tmpDir, "init")
	runGit(t, tmpDir, "config", "user.name", "TestUser")
	runGit(t, tmpDir, "config", "user.email", "test@nova.dev")

	_ = os.WriteFile(filepath.Join(tmpDir, "initial.txt"), []byte("init"), 0o644)
	runGit(t, tmpDir, "add", "initial.txt")
	runGit(t, tmpDir, "commit", "-m", "Initial commit")

	return tmpDir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s failed: %v: %s", strings.Join(args, " "), err, string(out))
	}
}

func TestGit_StatusClean(t *testing.T) {
	repo := setupGitRepo(t)

	ctx, stdout, _ := newTestContext(output.ModeHuman)
	err := Run(ctx, []string{"status", repo})
	if err != nil {
		t.Fatalf("Run status failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "Working tree clean") {
		t.Errorf("expected clean working tree message, got: %s", out)
	}
}

func TestGit_StatusDirty(t *testing.T) {
	repo := setupGitRepo(t)

	// Create unstaged and untracked files
	_ = os.WriteFile(filepath.Join(repo, "initial.txt"), []byte("modified"), 0o644)
	_ = os.WriteFile(filepath.Join(repo, "new.txt"), []byte("untracked"), 0o644)

	ctx, stdout, _ := newTestContext(output.ModePlain)
	err := Run(ctx, []string{"--plain", repo})
	if err != nil {
		t.Fatalf("Run plain status failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "unstaged:") || !strings.Contains(out, "untracked:") {
		t.Errorf("expected unstaged and untracked in output, got: %s", out)
	}
}

func TestGit_Log(t *testing.T) {
	repo := setupGitRepo(t)

	ctx, stdout, _ := newTestContext(output.ModeHuman)
	err := Run(ctx, []string{"log", "-n", "5", repo})
	if err != nil {
		t.Fatalf("Run log failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "Initial commit") {
		t.Errorf("expected 'Initial commit' in log output, got: %s", out)
	}
}

func TestGit_Branch(t *testing.T) {
	repo := setupGitRepo(t)

	ctx, stdout, _ := newTestContext(output.ModePlain)
	err := Run(ctx, []string{"branch", "--plain", repo})
	if err != nil {
		t.Fatalf("Run branch failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "* ") {
		t.Errorf("expected current branch marked with * in output, got: %s", out)
	}
}

func TestGit_NonRepo(t *testing.T) {
	tmpDir := t.TempDir()

	ctx, _, _ := newTestContext(output.ModePlain)
	err := Run(ctx, []string{tmpDir})
	if err == nil {
		t.Fatalf("expected error for non-git repository, got nil")
	}
}
