package calccmd

import (
	"strconv"
	"strings"
)

// Options holds CLI configuration for nova calc.
type Options struct {
	Expression string
	Precision  int  // decimal precision (default: 6)
	Plain      bool // --plain
	JSON       bool // --json
	Base       int  // forced output base (10, 16, 2, 8)
}

// DefaultOptions returns standard calc command options.
func DefaultOptions() Options {
	return Options{
		Precision: 6,
		Base:      10,
	}
}

// ParseFlags parses flags for nova calc.
func ParseFlags(args []string) Options {
	opts := DefaultOptions()
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--plain" {
			opts.Plain = true
		} else if arg == "--json" {
			opts.JSON = true
		} else if strings.HasPrefix(arg, "-p=") || strings.HasPrefix(arg, "--precision=") {
			parts := strings.SplitN(arg, "=", 2)
			if p, err := strconv.Atoi(parts[1]); err == nil && p >= 0 {
				opts.Precision = p
			}
		} else if arg == "-p" || arg == "--precision" {
			if i+1 < len(args) {
				if p, err := strconv.Atoi(args[i+1]); err == nil && p >= 0 {
					opts.Precision = p
				}
				i++
			}
		} else if arg == "--hex" || arg == "-x" {
			opts.Base = 16
		} else if arg == "--bin" || arg == "-b" {
			opts.Base = 2
		} else if arg == "--oct" || arg == "-o" {
			opts.Base = 8
		} else if strings.HasPrefix(arg, "--theme=") || strings.HasPrefix(arg, "--color=") {
			// skip global flags
		} else if arg == "--theme" || arg == "--color" {
			i++
		} else if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
		}
	}

	if len(positional) > 0 {
		opts.Expression = strings.Join(positional, " ")
	}

	return opts
}
