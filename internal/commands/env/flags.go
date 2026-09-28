package env

import (
	"strings"
)

// Options holds CLI configuration for nova env.
type Options struct {
	Filter      string // -f, --filter: filter key by pattern
	ShowSecrets bool   // -s, --show-secrets: reveal masked secret values
	Group       bool   // -g, --group: group by category (runtime, system, cloud, etc)
	Export      string // --export=sh, --export=fish
	Plain       bool   // --plain
	JSON        bool   // --json
}

// DefaultOptions returns standard environment options.
func DefaultOptions() Options {
	return Options{
		Group: true,
	}
}

// ParseFlags parses flags for nova env.
func ParseFlags(args []string) Options {
	opts := DefaultOptions()

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--plain" {
			opts.Plain = true
		} else if arg == "--json" {
			opts.JSON = true
		} else if arg == "-s" || arg == "--show-secrets" {
			opts.ShowSecrets = true
		} else if arg == "--no-group" {
			opts.Group = false
		} else if arg == "-g" || arg == "--group" {
			opts.Group = true
		} else if strings.HasPrefix(arg, "-f=") || strings.HasPrefix(arg, "--filter=") {
			parts := strings.SplitN(arg, "=", 2)
			opts.Filter = parts[1]
		} else if arg == "-f" || arg == "--filter" {
			if i+1 < len(args) {
				opts.Filter = args[i+1]
				i++
			}
		} else if strings.HasPrefix(arg, "--export=") {
			parts := strings.SplitN(arg, "=", 2)
			opts.Export = strings.ToLower(parts[1])
		} else if arg == "--export" {
			if i+1 < len(args) {
				opts.Export = strings.ToLower(args[i+1])
				i++
			} else {
				opts.Export = "sh"
			}
		} else if strings.HasPrefix(arg, "--theme=") || strings.HasPrefix(arg, "--color=") || strings.HasPrefix(arg, "--config=") {
			// skip global flags
		} else if arg == "--theme" || arg == "--color" || arg == "--config" {
			i++
		} else if !strings.HasPrefix(arg, "-") && opts.Filter == "" {
			opts.Filter = arg
		}
	}

	return opts
}
