package historycmd

import (
	"strconv"
	"strings"
)

// Options holds CLI configuration for nova history.
type Options struct {
	Count  int    // -n, --count (default: 10)
	Filter string // -f, --filter
	File   string // --file
	Plain  bool   // --plain
	JSON   bool   // --json
}

// DefaultOptions returns standard history command options.
func DefaultOptions() Options {
	return Options{
		Count: 10,
	}
}

// ParseFlags parses flags for nova history.
func ParseFlags(args []string) Options {
	opts := DefaultOptions()
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--plain" {
			opts.Plain = true
		} else if arg == "--json" {
			opts.JSON = true
		} else if strings.HasPrefix(arg, "-n=") || strings.HasPrefix(arg, "--count=") || strings.HasPrefix(arg, "--limit=") {
			parts := strings.SplitN(arg, "=", 2)
			if n, err := strconv.Atoi(parts[1]); err == nil && n > 0 {
				opts.Count = n
			}
		} else if arg == "-n" || arg == "--count" || arg == "--limit" {
			if i+1 < len(args) {
				if n, err := strconv.Atoi(args[i+1]); err == nil && n > 0 {
					opts.Count = n
				}
				i++
			}
		} else if strings.HasPrefix(arg, "-f=") || strings.HasPrefix(arg, "--filter=") {
			parts := strings.SplitN(arg, "=", 2)
			opts.Filter = strings.TrimSpace(parts[1])
		} else if arg == "-f" || arg == "--filter" {
			if i+1 < len(args) {
				opts.Filter = strings.TrimSpace(args[i+1])
				i++
			}
		} else if strings.HasPrefix(arg, "--file=") {
			parts := strings.SplitN(arg, "=", 2)
			opts.File = strings.TrimSpace(parts[1])
		} else if arg == "--file" {
			if i+1 < len(args) {
				opts.File = strings.TrimSpace(args[i+1])
				i++
			}
		} else if strings.HasPrefix(arg, "--theme=") || strings.HasPrefix(arg, "--color=") {
			// skip global flags
		} else if arg == "--theme" || arg == "--color" {
			i++
		} else if !strings.HasPrefix(arg, "-") && opts.Filter == "" {
			positional = append(positional, arg)
		}
	}

	if len(positional) > 0 && opts.Filter == "" {
		opts.Filter = strings.Join(positional, " ")
	}

	if opts.Count <= 0 {
		opts.Count = 10
	}

	return opts
}
