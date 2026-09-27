package watch

import (
	"strconv"
	"strings"
	"time"
)

// Options holds configuration for nova watch.
type Options struct {
	Paths       []string
	ExecCommand string        // -e, --exec
	Interval    time.Duration // -i, --interval (default 400ms)
	Debounce    time.Duration // -d, --debounce (default 150ms)
	ClearScreen bool          // -c, --clear
	Iterations  int           // -n, --iterations (0 = infinite)
	Extensions  []string      // --ext (e.g. "go,md")
	Plain       bool          // --plain
	JSON        bool          // --json
}

// DefaultOptions returns standard watch settings.
func DefaultOptions() Options {
	return Options{
		Paths:    []string{"."},
		Interval: 400 * time.Millisecond,
		Debounce: 150 * time.Millisecond,
	}
}

// ParseFlags parses flags for nova watch.
func ParseFlags(args []string) Options {
	opts := DefaultOptions()
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--plain" {
			opts.Plain = true
		} else if arg == "--json" {
			opts.JSON = true
		} else if (arg == "-e" || arg == "--exec") && i+1 < len(args) {
			opts.ExecCommand = args[i+1]
			i++
		} else if (arg == "-i" || arg == "--interval") && i+1 < len(args) {
			if d, err := time.ParseDuration(args[i+1]); err == nil && d > 0 {
				opts.Interval = d
			}
			i++
		} else if (arg == "-d" || arg == "--debounce") && i+1 < len(args) {
			if d, err := time.ParseDuration(args[i+1]); err == nil && d > 0 {
				opts.Debounce = d
			}
			i++
		} else if arg == "-c" || arg == "--clear" {
			opts.ClearScreen = true
		} else if (arg == "-n" || arg == "--iterations") && i+1 < len(args) {
			if n, err := strconv.Atoi(args[i+1]); err == nil && n > 0 {
				opts.Iterations = n
			}
			i++
		} else if (arg == "--ext") && i+1 < len(args) {
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
			i++
		} else if strings.HasPrefix(arg, "--theme=") || strings.HasPrefix(arg, "--color=") || strings.HasPrefix(arg, "--config=") {
			// skip
		} else if arg == "--theme" || arg == "--color" || arg == "--config" {
			i++
		} else if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
		}
	}

	if len(positional) > 0 {
		opts.Paths = positional
	}

	return opts
}
