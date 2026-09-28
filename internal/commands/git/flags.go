package gitcmd

import (
	"strconv"
	"strings"
)

// Options holds CLI configuration for nova git.
type Options struct {
	Dir    string // repository root or subfolder (default ".")
	Action string // "status" (default), "log", "branch"
	Count  int    // -n, --count: number of commits for log (default 10)
	Plain  bool   // --plain
	JSON   bool   // --json
}

// DefaultOptions returns standard git command options.
func DefaultOptions() Options {
	return Options{
		Dir:    ".",
		Action: "status",
		Count:  10,
	}
}

// ParseFlags parses flags for nova git.
func ParseFlags(args []string) Options {
	opts := DefaultOptions()
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--plain" {
			opts.Plain = true
		} else if arg == "--json" {
			opts.JSON = true
		} else if strings.HasPrefix(arg, "-n=") || strings.HasPrefix(arg, "--count=") {
			parts := strings.SplitN(arg, "=", 2)
			if n, err := strconv.Atoi(parts[1]); err == nil && n > 0 {
				opts.Count = n
			}
		} else if arg == "-n" || arg == "--count" {
			if i+1 < len(args) {
				if n, err := strconv.Atoi(args[i+1]); err == nil && n > 0 {
					opts.Count = n
				}
				i++
			}
		} else if arg == "status" || arg == "log" || arg == "branch" || arg == "branches" {
			opts.Action = arg
			if opts.Action == "branches" {
				opts.Action = "branch"
			}
		} else if strings.HasPrefix(arg, "--theme=") || strings.HasPrefix(arg, "--color=") || strings.HasPrefix(arg, "--config=") {
			// skip global flags
		} else if arg == "--theme" || arg == "--color" || arg == "--config" {
			i++
		} else if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
		}
	}

	if len(positional) > 0 {
		opts.Dir = positional[0]
	}

	return opts
}
