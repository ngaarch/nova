package rm

import "strings"

// Options represents parsed options for rm.
type Options struct {
	Recursive   bool // -r, -R, --recursive
	Force       bool // -f, --force
	Interactive bool // -i, --interactive
	Verbose     bool // -v, --verbose
	DryRun      bool // --dry-run
	Targets     []string
}

// ParseFlags parses flags for rm.
func ParseFlags(args []string) Options {
	var opts Options
	var targets []string

	i := 0
	for i < len(args) {
		arg := args[i]

		// Global flags pass-through
		if strings.HasPrefix(arg, "--theme=") || strings.HasPrefix(arg, "--color=") || strings.HasPrefix(arg, "--config=") {
			i++
			continue
		}
		if arg == "--theme" || arg == "--color" || arg == "--config" {
			i += 2
			continue
		}
		if arg == "--plain" || arg == "--json" {
			i++
			continue
		}

		if arg == "-r" || arg == "-R" || arg == "--recursive" {
			opts.Recursive = true
			i++
		} else if arg == "-f" || arg == "--force" {
			opts.Force = true
			opts.Interactive = false
			i++
		} else if arg == "-i" || arg == "--interactive" {
			opts.Interactive = true
			opts.Force = false
			i++
		} else if arg == "-v" || arg == "--verbose" {
			opts.Verbose = true
			i++
		} else if arg == "--dry-run" {
			opts.DryRun = true
			i++
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 && !strings.HasPrefix(arg, "--") {
			for _, r := range arg[1:] {
				switch r {
				case 'r', 'R':
					opts.Recursive = true
				case 'f':
					opts.Force = true
					opts.Interactive = false
				case 'i':
					opts.Interactive = true
					opts.Force = false
				case 'v':
					opts.Verbose = true
				}
			}
			i++
		} else {
			targets = append(targets, arg)
			i++
		}
	}

	opts.Targets = targets
	return opts
}
