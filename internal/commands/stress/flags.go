package stresscmd

import (
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Options holds configuration for stress tests.
type Options struct {
	Duration time.Duration `json:"duration"`
	Threads  int           `json:"threads"`
	Algo     string        `json:"algo"` // "all", "sha256", "math", "mem"
	Plain    bool          `json:"plain"`
	JSON     bool          `json:"json"`
}

// ParseFlags parses arguments for the stress command.
func ParseFlags(args []string) Options {
	opts := Options{
		Duration: 3 * time.Second,
		Threads:  runtime.NumCPU(),
		Algo:     "all",
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
		} else if strings.HasPrefix(arg, "--duration=") {
			if d, err := time.ParseDuration(strings.TrimPrefix(arg, "--duration=")); err == nil && d > 0 {
				opts.Duration = d
			}
			i++
		} else if (arg == "--duration" || arg == "-d") && i+1 < len(args) {
			if d, err := time.ParseDuration(args[i+1]); err == nil && d > 0 {
				opts.Duration = d
			}
			i += 2
		} else if strings.HasPrefix(arg, "--threads=") {
			if t, err := strconv.Atoi(strings.TrimPrefix(arg, "--threads=")); err == nil && t > 0 {
				opts.Threads = t
			}
			i++
		} else if (arg == "--threads" || arg == "-t") && i+1 < len(args) {
			if t, err := strconv.Atoi(args[i+1]); err == nil && t > 0 {
				opts.Threads = t
			}
			i += 2
		} else if strings.HasPrefix(arg, "--algo=") {
			opts.Algo = strings.ToLower(strings.TrimPrefix(arg, "--algo="))
			i++
		} else if (arg == "--algo" || arg == "-a") && i+1 < len(args) {
			opts.Algo = strings.ToLower(args[i+1])
			i += 2
		} else if !strings.HasPrefix(arg, "-") {
			opts.Algo = strings.ToLower(arg)
			i++
		} else {
			i++
		}
	}

	if opts.Threads < 1 {
		opts.Threads = 1
	}
	if opts.Threads > 64 {
		opts.Threads = 64
	}

	return opts
}
