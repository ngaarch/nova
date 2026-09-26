package cp

import "strings"

// Options represents parsed options for cp.
type Options struct {
	Recursive   bool // -r, -R, --recursive
	Preserve    bool // -p, --preserve
	Interactive bool // -i, --interactive
	Force       bool // -f, --force
	NoClobber   bool // -n, --no-clobber
	Verbose     bool // -v, --verbose
	DryRun      bool // --dry-run
	Dereference bool // -L, --dereference
	Sources     []string
	Destination string
}

// ParseFlags parses flags for cp.
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
		} else if arg == "-p" || arg == "--preserve" {
			opts.Preserve = true
			i++
		} else if arg == "-i" || arg == "--interactive" {
			opts.Interactive = true
			opts.Force = false
			i++
		} else if arg == "-f" || arg == "--force" {
			opts.Force = true
			opts.Interactive = false
			i++
		} else if arg == "-n" || arg == "--no-clobber" {
			opts.NoClobber = true
			i++
		} else if arg == "-v" || arg == "--verbose" {
			opts.Verbose = true
			i++
		} else if arg == "--dry-run" {
			opts.DryRun = true
			i++
		} else if arg == "-L" || arg == "--dereference" {
			opts.Dereference = true
			i++
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 && !strings.HasPrefix(arg, "--") {
			// Bundled short flags: -rpv, etc.
			for _, r := range arg[1:] {
				switch r {
				case 'r', 'R':
					opts.Recursive = true
				case 'p':
					opts.Preserve = true
				case 'i':
					opts.Interactive = true
					opts.Force = false
				case 'f':
					opts.Force = true
					opts.Interactive = false
				case 'n':
					opts.NoClobber = true
				case 'v':
					opts.Verbose = true
				case 'L':
					opts.Dereference = true
				}
			}
			i++
		} else {
			targets = append(targets, arg)
			i++
		}
	}

	if len(targets) > 1 {
		opts.Sources = targets[:len(targets)-1]
		opts.Destination = targets[len(targets)-1]
	} else if len(targets) == 1 {
		opts.Sources = targets
	}

	return opts
}
