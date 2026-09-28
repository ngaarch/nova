package gitcmd

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"nova/internal/command"
	"nova/internal/git"
	"nova/internal/output"
)

// FileItem represents a tracked or untracked repository file.
type FileItem struct {
	Path   string `json:"path"`
	Status string `json:"status"` // STAGED, MODIFIED, UNTRACKED, DELETED, CONFLICT
	Staged bool   `json:"staged"`
}

// StatusData holds the repository status summary.
type StatusData struct {
	Root       string     `json:"root"`
	Branch     string     `json:"branch"`
	Ahead      int        `json:"ahead"`
	Behind     int        `json:"behind"`
	Clean      bool       `json:"clean"`
	Staged     []FileItem `json:"staged"`
	Unstaged   []FileItem `json:"unstaged"`
	Untracked  []FileItem `json:"untracked"`
	TotalFiles int        `json:"total_files"`
}

// CommitItem represents one commit entry in the log.
type CommitItem struct {
	Hash    string `json:"hash"`
	Author  string `json:"author"`
	RelTime string `json:"relative_time"`
	Message string `json:"message"`
	Ref     string `json:"ref,omitempty"`
}

// BranchItem represents a git branch.
type BranchItem struct {
	Name      string `json:"name"`
	IsCurrent bool   `json:"is_current"`
	IsRemote  bool   `json:"is_remote"`
}

// Command returns the registered Command instance for git.
func Command() *command.Command {
	return &command.Command{
		Name:        "git",
		Aliases:     []string{"gstatus", "glog", "repo"},
		Summary:     "Display Git repository status dashboard, visual branch overview, and commit logs",
		Usage:       "nova git [status|log|branch] [flags]",
		Description: "Inspect git repository health, staged/unstaged changes, ahead/behind counters, and commit graph.",
		Phase:       20,
		Run:         Run,
	}
}

// Run executes the git command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)

	root, err := git.FindRepoRoot(opts.Dir)
	if err != nil || root == "" {
		return fmt.Errorf("fatal: not a git repository (or any of the parent directories): %s", opts.Dir)
	}

	switch opts.Action {
	case "log":
		return runLog(ctx, root, opts)
	case "branch":
		return runBranch(ctx, root, opts)
	default:
		return runStatus(ctx, root, opts)
	}
}

func runStatus(ctx *command.Context, root string, opts Options) error {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	branch, _, _ := git.GetBranch(root)
	ahead, behind := getAheadBehind(timeoutCtx, root)

	cmd := exec.CommandContext(timeoutCtx, "git", "-C", root, "status", "--porcelain=v1")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git status: %w", err)
	}

	var staged, unstaged, untracked []FileItem
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if len(line) < 4 {
			continue
		}
		x := line[0]
		y := line[1]
		filePath := strings.TrimSpace(line[3:])

		if x == '?' && y == '?' {
			untracked = append(untracked, FileItem{Path: filePath, Status: "UNTRACKED"})
			continue
		}

		if x != ' ' && x != '?' {
			statusStr := "STAGED"
			if x == 'D' {
				statusStr = "DELETED"
			} else if x == 'R' {
				statusStr = "RENAMED"
			}
			staged = append(staged, FileItem{Path: filePath, Status: statusStr, Staged: true})
		}

		if y != ' ' && y != '?' {
			statusStr := "MODIFIED"
			if y == 'D' {
				statusStr = "DELETED"
			}
			unstaged = append(unstaged, FileItem{Path: filePath, Status: statusStr, Staged: false})
		}
	}

	data := StatusData{
		Root:       root,
		Branch:     branch,
		Ahead:      ahead,
		Behind:     behind,
		Clean:      len(staged) == 0 && len(unstaged) == 0 && len(untracked) == 0,
		Staged:     staged,
		Unstaged:   unstaged,
		Untracked:  untracked,
		TotalFiles: len(staged) + len(unstaged) + len(untracked),
	}

	return RenderStatus(ctx, data, opts)
}

func getAheadBehind(ctx context.Context, root string) (int, int) {
	cmd := exec.CommandContext(ctx, "git", "-C", root, "rev-list", "--left-right", "--count", "HEAD...@{u}")
	out, err := cmd.Output()
	if err != nil {
		return 0, 0
	}
	parts := strings.Fields(string(out))
	if len(parts) == 2 {
		ahead, _ := strconv.Atoi(parts[0])
		behind, _ := strconv.Atoi(parts[1])
		return ahead, behind
	}
	return 0, 0
}

func runLog(ctx *command.Context, root string, opts Options) error {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	countStr := fmt.Sprintf("-n%d", opts.Count)
	cmd := exec.CommandContext(timeoutCtx, "git", "-C", root, "log", countStr, "--pretty=format:%h|%an|%cr|%s|%d")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git log: %w", err)
	}

	var commits []CommitItem
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 4 {
			continue
		}
		item := CommitItem{
			Hash:    parts[0],
			Author:  parts[1],
			RelTime: parts[2],
			Message: parts[3],
		}
		if len(parts) >= 5 {
			item.Ref = strings.TrimSpace(parts[4])
		}
		commits = append(commits, item)
	}

	return RenderLog(ctx, commits, opts)
}

func runBranch(ctx *command.Context, root string, opts Options) error {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "git", "-C", root, "branch", "--all")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git branch: %w", err)
	}

	var branches []BranchItem
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		isCurrent := strings.HasPrefix(line, "*")
		name := strings.TrimPrefix(trimmed, "* ")
		isRemote := strings.HasPrefix(name, "remotes/")

		branches = append(branches, BranchItem{
			Name:      name,
			IsCurrent: isCurrent,
			IsRemote:  isRemote,
		})
	}

	return RenderBranch(ctx, branches, opts)
}
