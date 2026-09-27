package diff

import (
	"strconv"
	"strings"
)

// Options represents parsed command-line flags for the diff command.
type Options struct {
	ContextLines   int    `json:"context_lines"`
	Brief          bool   `json:"brief"`
	IgnoreCase     bool   `json:"ignore_case"`
	IgnoreAllSpace bool   `json:"ignore_all_space"`
	Plain          bool   `json:"plain"`
	JSON           bool   `json:"json"`
	File1          string `json:"file1"`
	File2          string `json:"file2"`
}

// ParseFlags parses command-line arguments for the diff command.
func ParseFlags(args []string) Options {
	opts := Options{
		ContextLines: 3,
	}

	var positional []string
	i := 0
	for i < len(args) {
		arg := args[i]

		if arg == "-q" || arg == "--brief" {
			opts.Brief = true
		} else if arg == "-i" || arg == "--ignore-case" {
			opts.IgnoreCase = true
		} else if arg == "-w" || arg == "--ignore-all-space" {
			opts.IgnoreAllSpace = true
		} else if arg == "--plain" {
			opts.Plain = true
		} else if arg == "--json" {
			opts.JSON = true
		} else if arg == "-u" || arg == "--unified" {
			if i+1 < len(args) {
				if n, err := strconv.Atoi(args[i+1]); err == nil && n >= 0 {
					opts.ContextLines = n
					i++
				}
			}
		} else if strings.HasPrefix(arg, "-u") && len(arg) > 2 {
			if n, err := strconv.Atoi(arg[2:]); err == nil && n >= 0 {
				opts.ContextLines = n
			}
		} else if strings.HasPrefix(arg, "--unified=") {
			val := strings.TrimPrefix(arg, "--unified=")
			if n, err := strconv.Atoi(val); err == nil && n >= 0 {
				opts.ContextLines = n
			}
		} else if !strings.HasPrefix(arg, "-") && arg != "" {
			positional = append(positional, arg)
		}
		i++
	}

	if len(positional) >= 1 {
		opts.File1 = positional[0]
	}
	if len(positional) >= 2 {
		opts.File2 = positional[1]
	}

	return opts
}
