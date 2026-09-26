package tree

import (
	"strconv"
	"strings"
)

// Options represents parsed options for the tree command.
type Options struct {
	MaxDepth      int    // -L, --level, --max-depth
	All           bool   // -a, --all
	DirsOnly      bool   // -d, --dirs-only
	Sizes         bool   // -s, --sizes
	HumanReadable bool   // -h, --human-readable
	Permissions   bool   // -p, --permissions
	Icons         string // "auto", "always", "never"
	Plain         bool   // --plain
	JSON          bool   // --json
	Path          string
}

// DefaultOptions returns standard tree options.
func DefaultOptions() Options {
	return Options{
		MaxDepth:      -1,
		All:           false,
		DirsOnly:      false,
		Sizes:         false,
		HumanReadable: false,
		Permissions:   false,
		Icons:         "auto",
		Plain:         false,
		JSON:          false,
		Path:          ".",
	}
}

// ParseFlags parses command arguments for tree.
func ParseFlags(args []string) Options {
	opts := DefaultOptions()
	var targets []string

	i := 0
	for i < len(args) {
		arg := args[i]

		// Global flag passes
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

		if arg == "-a" || arg == "--all" {
			opts.All = true
			i++
		} else if arg == "-d" || arg == "--dirs-only" {
			opts.DirsOnly = true
			i++
		} else if arg == "-s" || arg == "--sizes" {
			opts.Sizes = true
			i++
		} else if arg == "-h" || arg == "--human-readable" {
			opts.HumanReadable = true
			opts.Sizes = true
			i++
		} else if arg == "-p" || arg == "--permissions" {
			opts.Permissions = true
			i++
		} else if strings.HasPrefix(arg, "-L") && len(arg) > 2 {
			if val, err := strconv.Atoi(arg[2:]); err == nil {
				opts.MaxDepth = val
			}
			i++
		} else if (arg == "-L" || arg == "--level" || arg == "--max-depth") && i+1 < len(args) {
			if val, err := strconv.Atoi(args[i+1]); err == nil {
				opts.MaxDepth = val
			}
			i += 2
		} else if strings.HasPrefix(arg, "--max-depth=") {
			if val, err := strconv.Atoi(strings.TrimPrefix(arg, "--max-depth=")); err == nil {
				opts.MaxDepth = val
			}
			i++
		} else if strings.HasPrefix(arg, "--level=") {
			if val, err := strconv.Atoi(strings.TrimPrefix(arg, "--level=")); err == nil {
				opts.MaxDepth = val
			}
			i++
		} else if strings.HasPrefix(arg, "--icons=") {
			opts.Icons = strings.TrimPrefix(arg, "--icons=")
			i++
		} else if arg == "--icons" && i+1 < len(args) {
			opts.Icons = args[i+1]
			i += 2
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 && !strings.HasPrefix(arg, "--") {
			// Bundled short flags: -adh, etc.
			for _, r := range arg[1:] {
				switch r {
				case 'a':
					opts.All = true
				case 'd':
					opts.DirsOnly = true
				case 's':
					opts.Sizes = true
				case 'h':
					opts.HumanReadable = true
					opts.Sizes = true
				case 'p':
					opts.Permissions = true
				}
			}
			i++
		} else {
			targets = append(targets, arg)
			i++
		}
	}

	if len(targets) > 0 {
		opts.Path = targets[0]
	}
	return opts
}
