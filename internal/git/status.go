package git

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// FileStatus represents the Git tracking state of a file or directory.
type FileStatus string

const (
	StatusClean     FileStatus = ""
	StatusModified  FileStatus = "M"
	StatusAdded     FileStatus = "A"
	StatusUntracked FileStatus = "?"
	StatusDeleted   FileStatus = "D"
	StatusRenamed   FileStatus = "R"
	StatusIgnored   FileStatus = "!"
	StatusConflict  FileStatus = "U"
)

// RepoStatus holds Git metadata and individual file tracking statuses for a repository.
type RepoStatus struct {
	Root       string
	Branch     string
	IsDetached bool
	Statuses   map[string]FileStatus
}

// GetRepoStatus inspects the Git repository containing dir within a strict timeout.
// If dir is not in a repository, it returns nil, nil with zero overhead.
// If the Git command times out or fails, it gracefully falls back to branch-only metadata.
func GetRepoStatus(dir string, timeout time.Duration) (*RepoStatus, error) {
	root, err := FindRepoRoot(dir)
	if err != nil || root == "" {
		return nil, err
	}

	branch, isDetached, _ := GetBranch(root)
	rs := &RepoStatus{
		Root:       root,
		Branch:     branch,
		IsDetached: isDetached,
		Statuses:   make(map[string]FileStatus),
	}

	if timeout <= 0 {
		timeout = 50 * time.Millisecond
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "-C", root, "status", "--porcelain=v1", "-z", "--untracked-files=all")
	out, err := cmd.Output()
	if err != nil {
		// Degrade gracefully if git binary is unavailable or command timed out
		return rs, nil
	}

	rs.parsePorcelainZ(out)
	return rs, nil
}

func (rs *RepoStatus) parsePorcelainZ(data []byte) {
	chunks := bytes.Split(data, []byte{0})
	i := 0
	for i < len(chunks) {
		chunk := chunks[i]
		if len(chunk) < 3 {
			i++
			continue
		}

		x := chunk[0]
		y := chunk[1]
		filePath := string(chunk[3:])

		status := resolveStatus(x, y)
		rs.Statuses[filepath.Clean(filePath)] = status

		// In rename mode (R), git porcelain -z emits:
		// "R  <new_path>\0<old_path>\0"
		if x == 'R' || y == 'R' {
			i++ // Skip original path
		}
		i++
	}
}

func resolveStatus(x, y byte) FileStatus {
	switch {
	case x == '?' && y == '?':
		return StatusUntracked
	case x == '!' && y == '!':
		return StatusIgnored
	case x == 'U' || y == 'U' || (x == 'A' && y == 'A') || (x == 'D' && y == 'D'):
		return StatusConflict
	case x == 'M' || y == 'M':
		return StatusModified
	case x == 'A' || y == 'A':
		return StatusAdded
	case x == 'D' || y == 'D':
		return StatusDeleted
	case x == 'R' || y == 'R':
		return StatusRenamed
	default:
		return StatusModified
	}
}

// GetStatus returns the FileStatus for a given file or directory path.
// If the target is a directory, it returns the aggregated status of any child items.
func (rs *RepoStatus) GetStatus(path string) FileStatus {
	if rs == nil || len(rs.Statuses) == 0 {
		return StatusClean
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return StatusClean
	}

	rel, err := filepath.Rel(rs.Root, absPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return StatusClean
	}
	cleanRel := filepath.Clean(rel)

	if st, ok := rs.Statuses[cleanRel]; ok {
		return st
	}

	// Check if directory contains modified or untracked children
	prefix := cleanRel + string(filepath.Separator)
	if cleanRel == "." {
		prefix = ""
	}

	hasModified := false
	hasUntracked := false
	hasAdded := false

	for p, st := range rs.Statuses {
		if prefix == "" || strings.HasPrefix(p, prefix) {
			switch st {
			case StatusModified, StatusConflict:
				hasModified = true
			case StatusUntracked:
				hasUntracked = true
			case StatusAdded, StatusRenamed:
				hasAdded = true
			}
		}
	}

	if hasModified {
		return StatusModified
	}
	if hasAdded {
		return StatusAdded
	}
	if hasUntracked {
		return StatusUntracked
	}

	return StatusClean
}
