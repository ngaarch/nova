package runcmd

import (
	"strings"
)

// Options holds CLI configuration for nova run / task.
type Options struct {
	TaskName   string   // name of task to execute
	TaskArgs   []string // extra arguments to pass to the task
	List       bool     // force list mode
	Plain      bool     // --plain
	JSON       bool     // --json
	Dir        string   // workspace directory (default: ".")
}

// DefaultOptions returns standard run command options.
func DefaultOptions() Options {
	return Options{
		Dir: ".",
	}
}

// ParseFlags parses command line flags for nova run.
func ParseFlags(args []string) Options {
	opts := DefaultOptions()
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--plain" {
			opts.Plain = true
		} else if arg == "--json" {
			opts.JSON = true
		} else if arg == "-l" || arg == "--list" {
			opts.List = true
		} else if strings.HasPrefix(arg, "--dir=") {
			parts := strings.SplitN(arg, "=", 2)
			opts.Dir = strings.TrimSpace(parts[1])
		} else if arg == "--dir" {
			if i+1 < len(args) {
				opts.Dir = strings.TrimSpace(args[i+1])
				i++
			}
		} else if strings.HasPrefix(arg, "--theme=") || strings.HasPrefix(arg, "--color=") {
			// skip global flags
		} else if arg == "--theme" || arg == "--color" {
			i++
		} else if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
		}
	}

	if len(positional) > 0 {
		opts.TaskName = positional[0]
		opts.TaskArgs = positional[1:]
	}

	return opts
}
