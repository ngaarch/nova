package mv

import "strings"

// Options represents parsed options for mv.
type Options struct {
	Interactive bool // -i, --interactive
	Force       bool // -f, --force
	NoClobber   bool // -n, --no-clobber
	Verbose     bool // -v, --verbose
	DryRun      bool // --dry-run
	Sources     []string
	Destination string
}

// ParseFlags parses flags for mv.
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

		if arg == "-i" || arg == "--interactive" {
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
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 && !strings.HasPrefix(arg, "--") {
			for _, r := range arg[1:] {
				switch r {
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
