package sysinfo

import (
	"strings"
)

// Options represents CLI options for the sysinfo command.
type Options struct {
	Full  bool // --full: show extended PATH and memory stats
	Plain bool // --plain
	JSON  bool // --json
}

// ParseFlags parses flags for nova sysinfo.
func ParseFlags(args []string) Options {
	var opts Options
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--plain" {
			opts.Plain = true
		} else if arg == "--json" {
			opts.JSON = true
		} else if arg == "--full" || arg == "-f" {
			opts.Full = true
		} else if strings.HasPrefix(arg, "--theme=") || strings.HasPrefix(arg, "--color=") || strings.HasPrefix(arg, "--config=") {
			// skip global flags
		} else if arg == "--theme" || arg == "--color" || arg == "--config" {
			i++
		}
	}
	return opts
}
