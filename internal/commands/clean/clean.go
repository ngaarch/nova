package clean

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"nova/internal/command"
	"nova/internal/output"
)

// Item represents a single detected disposable candidate.
type Item struct {
	Path        string `json:"path"`
	Category    string `json:"category"`
	Size        int64  `json:"size"`
	IsDir       bool   `json:"is_dir"`
	IsBroken    bool   `json:"is_broken,omitempty"`
	Deleted     bool   `json:"deleted"`
	DeleteError string `json:"delete_error,omitempty"`
}

// Summary holds aggregate cleanup metrics.
type Summary struct {
	TotalCandidates int    `json:"total_candidates"`
	TotalDeleted    int    `json:"total_deleted"`
	ReclaimedBytes  int64  `json:"reclaimed_bytes"`
	IsDryRun        bool   `json:"is_dry_run"`
	ScannedDirs     int    `json:"scanned_dirs"`
}

// Results holds candidates and summary.
type Results struct {
	Items   []Item  `json:"items"`
	Summary Summary `json:"summary"`
}

// Command returns the registered Command instance for the clean subcommand.
func Command() *command.Command {
	return &command.Command{
		Name:        "clean",
		Aliases:     []string{"tidy", "sweep"},
		Summary:     "Clean temporary files, build artifacts, broken symlinks, and empty directories",
		Usage:       "nova clean [flags] [paths...]",
		Description: "Identify and safely delete disposable workspace caches, build outputs, and junk files.",
		Phase:       15,
		Run:         Run,
	}
}

// Run executes the clean command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)
	if opts.Plain {
		ctx.Printer.Mode = output.ModePlain
	} else if opts.JSON {
		ctx.Printer.Mode = output.ModeJSON
	}

	results, err := executeClean(opts)
	if err != nil {
		return err
	}

	if ctx.Printer.Mode == output.ModeJSON {
		return RenderJSON(ctx, results)
	}

	if ctx.Printer.Mode == output.ModePlain {
		RenderPlain(ctx, results)
	} else {
		RenderHuman(ctx, results)
	}

	return nil
}

func executeClean(opts Options) (Results, error) {
	var items []Item
	var totalReclaimed int64
	var scannedDirs int

	for _, root := range opts.Paths {
		absRoot, err := filepath.Abs(root)
		if err != nil {
			continue
		}

		// Prevent cleaning system roots
		if absRoot == "/" || absRoot == filepath.Clean(os.Getenv("HOME")) {
			return Results{}, fmt.Errorf("refusing to clean critical system root %q without explicit subfolder", absRoot)
		}

		emptyDirs := make([]string, 0)

		err = filepath.Walk(absRoot, func(path string, fi os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return nil
			}

			if fi.IsDir() {
				scannedDirs++
				base := filepath.Base(path)
				// Never enter .git or version control trees
				if base == ".git" || base == ".hg" || base == ".svn" {
					return filepath.SkipDir
				}

				if opts.EmptyDirs && path != absRoot {
					if isEmptyDir(path) {
						emptyDirs = append(emptyDirs, path)
					}
				}
				return nil
			}

			// Check symlinks
			if fi.Mode()&os.ModeSymlink != 0 {
				if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
					item := Item{
						Path:     path,
						Category: "broken-symlink",
						Size:     0,
						IsBroken: true,
					}
					items = append(items, item)
					return nil
				}
			}

			// Check junk patterns
			cat, isJunk := classifyJunk(filepath.Base(path), fi.Size(), opts.All)
			if isJunk {
				item := Item{
					Path:     path,
					Category: cat,
					Size:     fi.Size(),
				}
				items = append(items, item)
			}

			return nil
		})
		if err != nil {
			return Results{}, err
		}

		// Process empty dirs bottom-up
		for i := len(emptyDirs) - 1; i >= 0; i-- {
			d := emptyDirs[i]
			items = append(items, Item{
				Path:     d,
				Category: "empty-dir",
				Size:     0,
				IsDir:    true,
			})
		}
	}

	deletedCount := 0
	for i := range items {
		if !opts.DryRun && opts.Force {
			var delErr error
			if items[i].IsDir {
				delErr = os.Remove(items[i].Path)
			} else {
				delErr = os.Remove(items[i].Path)
			}
			if delErr == nil {
				items[i].Deleted = true
				deletedCount++
				totalReclaimed += items[i].Size
			} else {
				items[i].DeleteError = delErr.Error()
			}
		} else {
			totalReclaimed += items[i].Size
		}
	}

	summary := Summary{
		TotalCandidates: len(items),
		TotalDeleted:    deletedCount,
		ReclaimedBytes:  totalReclaimed,
		IsDryRun:        opts.DryRun,
		ScannedDirs:     scannedDirs,
	}

	return Results{
		Items:   items,
		Summary: summary,
	}, nil
}

func classifyJunk(name string, size int64, aggressive bool) (string, bool) {
	lower := strings.ToLower(name)

	// OS Artifacts
	if lower == ".ds_store" || lower == "thumbs.db" || lower == "desktop.ini" {
		return "os-junk", true
	}

	// Editor & Temporary Files
	if strings.HasSuffix(lower, "~") || strings.HasSuffix(lower, ".swp") || strings.HasSuffix(lower, ".swo") || strings.HasSuffix(lower, ".tmp") {
		return "temp-file", true
	}

	// Go build artifacts
	if strings.HasSuffix(lower, ".test") {
		return "test-binary", true
	}

	// Aggressive / cache mode
	if aggressive {
		if strings.HasSuffix(lower, ".log") && size > 10*1024*1024 {
			return "large-log", true
		}
		if strings.HasSuffix(lower, ".bak") || strings.HasSuffix(lower, ".old") {
			return "backup-file", true
		}
	}

	return "", false
}

func isEmptyDir(dirPath string) bool {
	f, err := os.Open(dirPath)
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	return err == io.EOF
}
