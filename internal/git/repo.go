package git

import (
	"os"
	"path/filepath"
	"strings"
)

// FindRepoRoot traverses upward from startDir to find the root of a Git repository.
// It detects standard .git directories and .git worktree/submodule pointer files.
// Returns an empty string without error if startDir is not within a Git repository.
func FindRepoRoot(startDir string) (string, error) {
	current, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	for {
		gitPath := filepath.Join(current, ".git")
		fi, err := os.Stat(gitPath)
		if err == nil {
			if fi.IsDir() {
				return current, nil
			}
			// .git might be a file (submodule or worktree pointing to actual gitdir)
			data, err := os.ReadFile(gitPath)
			if err == nil && strings.HasPrefix(strings.TrimSpace(string(data)), "gitdir:") {
				return current, nil
			}
		}

		parent := filepath.Dir(current)
		if parent == current {
			// Reached filesystem root
			break
		}
		current = parent
	}

	return "", nil
}

// GetBranch resolves the active branch name or short detached HEAD commit hash.
// It reads directly from .git/HEAD on disk without launching subprocesses.
func GetBranch(repoRoot string) (branch string, isDetached bool, err error) {
	gitPath := filepath.Join(repoRoot, ".git")
	headPath := filepath.Join(gitPath, "HEAD")

	fi, err := os.Stat(gitPath)
	if err != nil {
		return "", false, err
	}

	// Handle .git file (worktree or submodule)
	if !fi.IsDir() {
		data, err := os.ReadFile(gitPath)
		if err != nil {
			return "", false, err
		}
		content := strings.TrimSpace(string(data))
		if strings.HasPrefix(content, "gitdir:") {
			gitDir := strings.TrimSpace(strings.TrimPrefix(content, "gitdir:"))
			if !filepath.IsAbs(gitDir) {
				gitDir = filepath.Join(repoRoot, gitDir)
			}
			headPath = filepath.Join(gitDir, "HEAD")
		}
	}

	headData, err := os.ReadFile(headPath)
	if err != nil {
		return "", false, err
	}

	line := strings.TrimSpace(string(headData))
	if strings.HasPrefix(line, "ref: refs/heads/") {
		branchName := strings.TrimPrefix(line, "ref: refs/heads/")
		return branchName, false, nil
	}

	// Detached HEAD (commit hash)
	commitHash := line
	if len(commitHash) > 7 {
		commitHash = commitHash[:7]
	}
	return commitHash, true, nil
}
