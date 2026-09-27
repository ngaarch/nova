package which

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"nova/internal/command"
	"nova/internal/output"
	"nova/internal/renderer"
)

// Result represents the resolved metadata of a single executable query.
type Result struct {
	Query       string `json:"query"`
	Path        string `json:"path"`
	IsSymlink   bool   `json:"is_symlink"`
	Target      string `json:"target,omitempty"`
	Size        int64  `json:"size"`
	HumanSize   string `json:"human_size"`
	Permissions string `json:"permissions"`
	Found       bool   `json:"found"`
}

// Command returns the registered Command instance for the which subcommand.
func Command() *command.Command {
	return &command.Command{
		Name:        "which",
		Aliases:     []string{"where"},
		Summary:     "Locate executables in PATH with symlink resolution, size, and permissions",
		Usage:       "nova which [flags] <program...>",
		Description: "Locate and inspect binary executables in $PATH, resolving symlinks and metadata.",
		Phase:       11,
		Run:         Run,
	}
}

// Run executes the which command.
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

	if len(opts.ExecNames) == 0 {
		return command.NewUsageError("missing executable name", "Specify one or more program names: nova which <program...>")
	}

	pathEnv := os.Getenv("PATH")
	pathDirs := filepath.SplitList(pathEnv)

	var allResults []Result
	var anyNotFound bool

	for _, name := range opts.ExecNames {
		results := locateExecutable(name, pathDirs, opts.All)
		if len(results) == 0 {
			anyNotFound = true
			allResults = append(allResults, Result{
				Query: name,
				Found: false,
			})
		} else {
			allResults = append(allResults, results...)
		}
	}

	if opts.Silent {
		if anyNotFound {
			return errors.New("one or more executables not found")
		}
		return nil
	}

	if ctx.Printer.Mode == output.ModeJSON {
		return RenderJSON(ctx, allResults)
	}

	if ctx.Printer.Mode == output.ModePlain {
		RenderPlain(ctx, allResults)
	} else {
		RenderHuman(ctx, allResults)
	}

	if anyNotFound {
		return errors.New("one or more executables not found")
	}

	return nil
}

func locateExecutable(query string, pathDirs []string, findAll bool) []Result {
	var results []Result

	// If query has path separator, check path directly
	if strings.ContainsRune(query, filepath.Separator) {
		res, ok := inspectCandidate(query, query)
		if ok {
			results = append(results, res)
		}
		return results
	}

	seenPaths := make(map[string]bool)

	for _, dir := range pathDirs {
		candidate := filepath.Join(dir, query)
		if seenPaths[candidate] {
			continue
		}
		seenPaths[candidate] = true

		res, ok := inspectCandidate(query, candidate)
		if ok {
			results = append(results, res)
			if !findAll {
				break
			}
		}
	}

	return results
}

func inspectCandidate(query, candidatePath string) (Result, bool) {
	fi, err := os.Lstat(candidatePath)
	if err != nil {
		return Result{}, false
	}

	// Must be regular file or symlink
	isSymlink := fi.Mode()&os.ModeSymlink != 0
	if !isSymlink && fi.IsDir() {
		return Result{}, false
	}

	var target string
	var size int64 = fi.Size()
	var perm string = fi.Mode().String()
	isExec := fi.Mode()&0111 != 0

	if isSymlink {
		t, readErr := os.Readlink(candidatePath)
		if readErr == nil {
			target = t
		}
		resolvedFi, statErr := os.Stat(candidatePath)
		if statErr == nil {
			size = resolvedFi.Size()
			perm = resolvedFi.Mode().String()
			isExec = resolvedFi.Mode()&0111 != 0
		}
	}

	if !isExec {
		return Result{}, false
	}

	return Result{
		Query:       query,
		Path:        candidatePath,
		IsSymlink:   isSymlink,
		Target:      target,
		Size:        size,
		HumanSize:   renderer.FormatSize(size, true),
		Permissions: perm,
		Found:       true,
	}, true
}
