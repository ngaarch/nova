package ls

import (
	"strings"
)

// Options holds configuration options parsed from command-line arguments for ls.
type Options struct {
	All           bool
	Long          bool
	HumanReadable bool
	Reverse       bool
	Recursive     bool
	Tree          bool
	Sort          string // "name", "size", "time", "ext"
	DirsFirst     bool
	Icons         string // "auto", "always", "never"
	Paths         []string
}

// DefaultOptions returns default listing options.
func DefaultOptions() Options {
	return Options{
		All:           false,
		Long:          false,
		HumanReadable: true,
		Reverse:       false,
		Recursive:     false,
		Tree:          false,
		Sort:          "name",
		DirsFirst:     true,
		Icons:         "auto",
		Paths:         nil,
	}
}

// ParseFlags parses flags specifically for the ls subcommand.
func ParseFlags(args []string) Options {
	opts := DefaultOptions()
	var paths []string

	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "-a" || arg == "-A" || arg == "--all" {
			opts.All = true
			i++
		} else if arg == "-l" || arg == "--long" {
			opts.Long = true
			i++
		} else if arg == "-h" || arg == "--human-readable" {
			opts.HumanReadable = true
			i++
		} else if arg == "-r" || arg == "--reverse" {
			opts.Reverse = true
			i++
		} else if arg == "-t" {
			opts.Sort = "time"
			i++
		} else if arg == "-S" {
			opts.Sort = "size"
			i++
		} else if arg == "-X" {
			opts.Sort = "ext"
			i++
		} else if arg == "-R" || arg == "--recursive" {
			opts.Recursive = true
			i++
		} else if arg == "--tree" {
			opts.Tree = true
			i++
		} else if arg == "--dirs-first" {
			opts.DirsFirst = true
			i++
		} else if arg == "--no-dirs-first" {
			opts.DirsFirst = false
			i++
		} else if strings.HasPrefix(arg, "--sort=") {
			opts.Sort = strings.ToLower(strings.TrimPrefix(arg, "--sort="))
			i++
		} else if arg == "--sort" && i+1 < len(args) {
			opts.Sort = strings.ToLower(args[i+1])
			i += 2
		} else if strings.HasPrefix(arg, "--icons=") {
			opts.Icons = strings.ToLower(strings.TrimPrefix(arg, "--icons="))
			i++
		} else if arg == "--icons" && i+1 < len(args) {
			opts.Icons = strings.ToLower(args[i+1])
			i += 2
		} else if arg == "--no-icons" {
			opts.Icons = "never"
			i++
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 && !strings.HasPrefix(arg, "--") {
			// Handle bundled flags e.g. -lah or -alrt
			for _, r := range arg[1:] {
				switch r {
				case 'a', 'A':
					opts.All = true
				case 'l':
					opts.Long = true
				case 'h':
					opts.HumanReadable = true
				case 'r':
					opts.Reverse = true
				case 't':
					opts.Sort = "time"
				case 'S':
					opts.Sort = "size"
				case 'X':
					opts.Sort = "ext"
				case 'R':
					opts.Recursive = true
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
