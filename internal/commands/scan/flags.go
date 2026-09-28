package scancmd

import (
	"strconv"
	"strings"
)

// Options holds configuration for secret scanning.
type Options struct {
	Target       string   `json:"target"`
	Rule         string   `json:"rule,omitempty"`
	MinEntropy   float64  `json:"min_entropy"`
	IgnorePaths  []string `json:"ignore_paths,omitempty"`
	Plain        bool     `json:"plain"`
	JSON         bool     `json:"json"`
}

// ParseFlags parses command line arguments for the scan command.
func ParseFlags(args []string) Options {
	opts := Options{
		Target:      ".",
		MinEntropy:  4.5,
		IgnorePaths: []string{".git", "node_modules", "vendor", "dist", "build", ".next", ".cache"},
	}

	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "--plain" {
			opts.Plain = true
			i++
		} else if arg == "--json" {
			opts.JSON = true
			i++
		} else if strings.HasPrefix(arg, "--pattern=") {
			opts.Rule = strings.TrimPrefix(arg, "--pattern=")
			i++
		} else if (arg == "--pattern" || arg == "-p") && i+1 < len(args) {
			opts.Rule = args[i+1]
			i += 2
		} else if strings.HasPrefix(arg, "--entropy=") {
			if f, err := strconv.ParseFloat(strings.TrimPrefix(arg, "--entropy="), 64); err == nil && f > 0 {
				opts.MinEntropy = f
			}
			i++
		} else if (arg == "--entropy" || arg == "-e") && i+1 < len(args) {
			if f, err := strconv.ParseFloat(args[i+1], 64); err == nil && f > 0 {
				opts.MinEntropy = f
			}
			i += 2
		} else if strings.HasPrefix(arg, "--ignore=") {
			opts.IgnorePaths = append(opts.IgnorePaths, strings.Split(strings.TrimPrefix(arg, "--ignore="), ",")...)
			i++
		} else if (arg == "--ignore" || arg == "-i") && i+1 < len(args) {
			opts.IgnorePaths = append(opts.IgnorePaths, strings.Split(args[i+1], ",")...)
			i += 2
		} else if !strings.HasPrefix(arg, "-") && (opts.Target == "." || opts.Target == "") {
			opts.Target = arg
			i++
		} else {
			i++
		}
	}

	return opts
}
