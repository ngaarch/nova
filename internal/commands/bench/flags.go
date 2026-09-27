package bench

import (
	"strconv"
	"strings"
)

// Options holds CLI configuration for nova bench.
type Options struct {
	SizeMB     int  // --size (default 16 MB)
	Iterations int  // -n, --iterations (default 1000)
	Plain      bool // --plain
	JSON       bool // --json
}

// DefaultOptions returns standard benchmarking parameters.
func DefaultOptions() Options {
	return Options{
		SizeMB:     16,
		Iterations: 1000,
	}
}

// ParseFlags extracts bench options from command arguments.
func ParseFlags(args []string) Options {
	opts := DefaultOptions()
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--plain" {
			opts.Plain = true
		} else if arg == "--json" {
			opts.JSON = true
		} else if (arg == "-s" || arg == "--size") && i+1 < len(args) {
			if n, err := strconv.Atoi(args[i+1]); err == nil && n > 0 {
				opts.SizeMB = n
			}
			i++
		} else if (arg == "-n" || arg == "--iterations") && i+1 < len(args) {
			if n, err := strconv.Atoi(args[i+1]); err == nil && n > 0 {
				opts.Iterations = n
			}
			i++
		} else if strings.HasPrefix(arg, "--theme=") || strings.HasPrefix(arg, "--color=") || strings.HasPrefix(arg, "--config=") {
			// skip
		} else if arg == "--theme" || arg == "--color" || arg == "--config" {
			i++
		}
	}
	return opts
}
