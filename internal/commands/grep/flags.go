package grep

import (
	"strconv"
	"strings"
)

// Options holds command-line configuration for nova grep.
type Options struct {
	Query         string   // Search pattern
	Paths         []string // Target files or directories
	IgnoreCase    bool     // -i, --ignore-case
	LineNumbers   bool     // -n, --line-number (default true in human)
	CountOnly     bool     // -c, --count
	FilesWithMatch bool    // -l, --files-with-matches
	ContextBefore int      // -B, --before-context
	ContextAfter  int      // -A, --after-context
	Extensions    []string // --ext (e.g. "go,md")
	Hidden        bool     // --hidden
	MaxCount      int      // -m, --max-count per file (0 = unlimited)
	Plain         bool     // --plain
	JSON          bool     // --json
}

// DefaultOptions returns standard defaults for grep.
func DefaultOptions() Options {
	return Options{
		LineNumbers: true,
		Paths:       []string{"."},
	}
}

// ParseFlags parses grep command-line arguments.
func ParseFlags(args []string) Options {
	opts := DefaultOptions()
	var positional []string

	i := 0
	for i < len(args) {
		arg := args[i]

		// Skip global flags
		if arg == "--plain" {
			opts.Plain = true
			i++
			continue
		}
		if arg == "--json" {
			opts.JSON = true
			i++
			continue
		}
		if strings.HasPrefix(arg, "--theme=") || strings.HasPrefix(arg, "--color=") || strings.HasPrefix(arg, "--config=") {
			i++
			continue
		}
		if arg == "--theme" || arg == "--color" || arg == "--config" {
			i += 2
			continue
		}

		if arg == "-i" || arg == "--ignore-case" {
			opts.IgnoreCase = true
			i++
		} else if arg == "-n" || arg == "--line-number" {
			opts.LineNumbers = true
			i++
		} else if arg == "-N" || arg == "--no-line-number" {
			opts.LineNumbers = false
			i++
		} else if arg == "-c" || arg == "--count" {
			opts.CountOnly = true
			i++
		} else if arg == "-l" || arg == "--files-with-matches" {
			opts.FilesWithMatch = true
			i++
		} else if arg == "--hidden" {
			opts.Hidden = true
			i++
		} else if (arg == "-C" || arg == "--context") && i+1 < len(args) {
			if n, err := strconv.Atoi(args[i+1]); err == nil && n >= 0 {
				opts.ContextBefore = n
				opts.ContextAfter = n
			}
			i += 2
		} else if (arg == "-B" || arg == "--before-context") && i+1 < len(args) {
			if n, err := strconv.Atoi(args[i+1]); err == nil && n >= 0 {
				opts.ContextBefore = n
			}
			i += 2
		} else if (arg == "-A" || arg == "--after-context") && i+1 < len(args) {
			if n, err := strconv.Atoi(args[i+1]); err == nil && n >= 0 {
				opts.ContextAfter = n
			}
			i += 2
		} else if (arg == "-m" || arg == "--max-count") && i+1 < len(args) {
			if n, err := strconv.Atoi(args[i+1]); err == nil && n > 0 {
				opts.MaxCount = n
			}
			i += 2
		} else if (arg == "--ext" || arg == "-e") && i+1 < len(args) {
			parts := strings.Split(args[i+1], ",")
			for _, p := range parts {
				cleaned := strings.TrimSpace(p)
				if cleaned != "" {
					if !strings.HasPrefix(cleaned, ".") {
						cleaned = "." + cleaned
					}
					opts.Extensions = append(opts.Extensions, strings.ToLower(cleaned))
				}
			}
			i += 2
		} else if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
			i++
		} else {
			i++
		}
	}

	if len(positional) > 0 {
		opts.Query = positional[0]
		if len(positional) > 1 {
			opts.Paths = positional[1:]
		}
	}

	return opts
}
