package clean

import (
	"strings"
)

// Options holds CLI configuration for nova clean.
type Options struct {
	Paths       []string
	Force       bool // -f, --force: actually delete files (otherwise dry-run)
	DryRun      bool // -n, --dry-run
	EmptyDirs   bool // --empty-dirs: also prune empty directories
	All         bool // -a, --all: aggressive cleaning (include caches)
	Plain       bool // --plain
	JSON        bool // --json
}

// DefaultOptions returns standard clean parameters.
func DefaultOptions() Options {
	return Options{
		Paths:  []string{"."},
		DryRun: true, // safe by default unless -f is specified
	}
}

// ParseFlags parses flags for nova clean.
func ParseFlags(args []string) Options {
	opts := DefaultOptions()
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--plain" {
			opts.Plain = true
		} else if arg == "--json" {
			opts.JSON = true
		} else if arg == "-f" || arg == "--force" {
			opts.Force = true
			opts.DryRun = false
		} else if arg == "-n" || arg == "--dry-run" {
			opts.DryRun = true
			opts.Force = false
		} else if arg == "--empty-dirs" {
			opts.EmptyDirs = true
		} else if arg == "-a" || arg == "--all" {
			opts.All = true
		} else if strings.HasPrefix(arg, "--theme=") || strings.HasPrefix(arg, "--color=") || strings.HasPrefix(arg, "--config=") {
			// skip global flags
		} else if arg == "--theme" || arg == "--color" || arg == "--config" {
			i++
		} else if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
		}
	}

	if len(positional) > 0 {
		opts.Paths = positional
	}

	return opts
}
