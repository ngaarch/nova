package git

import (
	"os"
	"testing"
	"time"

	"nova/internal/terminal"
	"nova/internal/theme"
)

func TestFindRepoRoot(t *testing.T) {
	// Inside real repo
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root, err := FindRepoRoot(cwd)
	if err != nil {
		t.Fatalf("unexpected error finding repo root: %v", err)
	}
	if root == "" {
		t.Errorf("expected to find repo root from cwd %s", cwd)
	}

	// Outside repo in temp dir
	tmpDir := t.TempDir()
	outRoot, err := FindRepoRoot(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error on temp dir: %v", err)
	}
	if outRoot != "" {
		t.Errorf("expected empty root for temp dir, got %s", outRoot)
	}
}

func TestGetBranch(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root, err := FindRepoRoot(cwd)
	if err != nil || root == "" {
		t.Skip("skipping, not in git repo")
	}

	branch, _, err := GetBranch(root)
	if err != nil {
		t.Fatalf("GetBranch failed: %v", err)
	}
	if branch == "" {
		t.Errorf("expected non-empty branch name")
	}
}

func TestParsePorcelainZ(t *testing.T) {
	rs := &RepoStatus{
		Root:     "/mock/repo",
		Statuses: make(map[string]FileStatus),
	}

	// Simulating git status -z stream:
	// M  mod.go\0?? untracked.txt\0A  added.rs\0D  del.py\0R  renamed.c\0orig.c\0
	mockOutput := []byte(" M mod.go\x00?? untracked.txt\x00A  added.rs\x00 D del.py\x00R  renamed.c\x00orig.c\x00")
	rs.parsePorcelainZ(mockOutput)

	if rs.Statuses["mod.go"] != StatusModified {
		t.Errorf("expected mod.go to be Modified, got %v", rs.Statuses["mod.go"])
	}
	if rs.Statuses["untracked.txt"] != StatusUntracked {
		t.Errorf("expected untracked.txt to be Untracked, got %v", rs.Statuses["untracked.txt"])
	}
	if rs.Statuses["added.rs"] != StatusAdded {
		t.Errorf("expected added.rs to be Added, got %v", rs.Statuses["added.rs"])
	}
	if rs.Statuses["del.py"] != StatusDeleted {
		t.Errorf("expected del.py to be Deleted, got %v", rs.Statuses["del.py"])
	}
	if rs.Statuses["renamed.c"] != StatusRenamed {
		t.Errorf("expected renamed.c to be Renamed, got %v", rs.Statuses["renamed.c"])
	}
}

func TestGetStatus_DirectoryAggregation(t *testing.T) {
	rs := &RepoStatus{
		Root: "/mock/repo",
		Statuses: map[string]FileStatus{
			"pkg/sub/modified.go":  StatusModified,
			"pkg/other/clean.go":   StatusClean,
			"docs/new_page.md":     StatusUntracked,
		},
	}

	// Direct file query
	if rs.GetStatus("/mock/repo/pkg/sub/modified.go") != StatusModified {
		t.Errorf("expected StatusModified for direct file")
	}

	// Directory query (should inherit StatusModified from child)
	if rs.GetStatus("/mock/repo/pkg/sub") != StatusModified {
		t.Errorf("expected StatusModified for directory containing modified file")
	}

	// Directory query with untracked
	if rs.GetStatus("/mock/repo/docs") != StatusUntracked {
		t.Errorf("expected StatusUntracked for docs directory")
	}

	// Clean directory
	if rs.GetStatus("/mock/repo/src") != StatusClean {
		t.Errorf("expected StatusClean for untouched dir")
	}
}

func TestFormatStatusBadge(t *testing.T) {
	th := theme.Get("default")
	profile := terminal.ColorNone

	mBadge := FormatStatusBadge(StatusModified, th, profile)
	if mBadge != "M" {
		t.Errorf("expected 'M', got %q", mBadge)
	}

	uBadge := FormatStatusBadge(StatusUntracked, th, profile)
	if uBadge != "?" {
		t.Errorf("expected '?', got %q", uBadge)
	}

	cleanBadge := FormatStatusBadge(StatusClean, th, profile)
	if cleanBadge != " " {
		t.Errorf("expected ' ', got %q", cleanBadge)
	}
}

func TestFormatBranch(t *testing.T) {
	th := theme.Get("default")
	profile := terminal.ColorNone

	branchStr := FormatBranch("main", false, true, th, profile)
	if branchStr != "⎇ main" {
		t.Errorf("expected '⎇ main', got %q", branchStr)
	}

	asciiBranch := FormatBranch("feature", false, false, th, profile)
	if asciiBranch != "git:(feature)" {
		t.Errorf("expected 'git:(feature)', got %q", asciiBranch)
	}

	detached := FormatBranch("a1b2c3d", true, true, th, profile)
	if detached != "⎇ detached:a1b2c3d" {
		t.Errorf("expected detached format, got %q", detached)
	}
}

func TestTimeoutGracefulDegradation(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root, err := FindRepoRoot(cwd)
	if err != nil || root == "" {
		t.Skip("skipping, not in git repo")
	}

	// Impossibly short timeout (1 nanosecond)
	rs, err := GetRepoStatus(root, 1*time.Nanosecond)
	if err != nil {
		t.Fatalf("expected nil error on timeout fallback, got: %v", err)
	}
	if rs == nil {
		t.Fatalf("expected non-nil RepoStatus fallback")
	}
	if rs.Branch == "" {
		t.Errorf("expected Branch to still be read from .git/HEAD on timeout")
	}
}
