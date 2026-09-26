package du

import (
	"strconv"
	"strings"
)

// Options holds command flags for du.
type Options struct {
	HumanReadable bool   // -h, --human-readable
	MaxDepth      int    // -d, --max-depth
	Summarize     bool   // -s, --summarize
	All           bool   // -a, --all
	Total         bool   // -c, --total
	SortBy        string // --sort=size|name
	Plain         bool   // --plain
	JSON          bool   // --json
	Paths         []string
}

// DefaultOptions returns standard du options.
func DefaultOptions() Options {
	return Options{
		HumanReadable: true,
		MaxDepth:      -1,
		Summarize:     false,
		All:           false,
		Total:         false,
		SortBy:        "size",
		Plain:         false,
		JSON:          false,
		Paths:         nil,
	}
}

// ParseFlags parses flags specifically for du.
func ParseFlags(args []string) Options {
	opts := DefaultOptions()
	var paths []string

	i := 0
	for i < len(args) {
		arg := args[i]

		// Global flags
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

		if arg == "-h" || arg == "--human-readable" {
			opts.HumanReadable = true
			i++
		} else if arg == "-s" || arg == "--summarize" {
			opts.Summarize = true
			opts.MaxDepth = 0
			i++
		} else if arg == "-a" || arg == "--all" {
			opts.All = true
			i++
		} else if arg == "-c" || arg == "--total" {
			opts.Total = true
			i++
		} else if strings.HasPrefix(arg, "-d") && len(arg) > 2 {
			if val, err := strconv.Atoi(arg[2:]); err == nil {
				opts.MaxDepth = val
			}
			i++
		} else if (arg == "-d" || arg == "--max-depth") && i+1 < len(args) {
			if val, err := strconv.Atoi(args[i+1]); err == nil {
				opts.MaxDepth = val
			}
			i += 2
		} else if strings.HasPrefix(arg, "--max-depth=") {
			if val, err := strconv.Atoi(strings.TrimPrefix(arg, "--max-depth=")); err == nil {
				opts.MaxDepth = val
			}
			i++
		} else if strings.HasPrefix(arg, "--sort=") {
			opts.SortBy = strings.TrimPrefix(arg, "--sort=")
			i++
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 && !strings.HasPrefix(arg, "--") {
			// Bundled flags: -hsc, etc.
			for _, r := range arg[1:] {
				switch r {
				case 'h':
					opts.HumanReadable = true
				case 's':
					opts.Summarize = true
					opts.MaxDepth = 0
				case 'a':
					opts.All = true
				case 'c':
					opts.Total = true
				}
			}
			i++
		} else {
			paths = append(paths, arg)
			i++
		}
	}

	if len(paths) == 0 {
		paths = []string{"."}
	}
	opts.Paths = paths
	return opts
}
