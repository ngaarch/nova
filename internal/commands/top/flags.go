package topcmd

import (
	"strconv"
	"strings"
	"time"
)

// Options defines runtime flags for the top / proc command.
type Options struct {
	SortBy     string        // "cpu" (default), "mem", "pid", "name"
	Filter     string        // name or command substring filter
	Limit      int           // max processes to display (default: 15)
	Plain      bool          // --plain
	JSON       bool          // --json
	KillPID    int           // -k, --kill
	Signal     int           // --signal (default: 15 / SIGTERM)
	Live       bool          // -w, --live
	Interval   time.Duration // --interval (default: 1s)
	Iterations int           // --iter (default: 1)
}

// DefaultOptions returns standard top command options.
func DefaultOptions() Options {
	return Options{
		SortBy:     "cpu",
		Limit:      15,
		Signal:     15,
		Interval:   1 * time.Second,
		Iterations: 1,
	}
}

// ParseFlags parses flags for nova top.
func ParseFlags(args []string) Options {
	opts := DefaultOptions()

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--plain" {
			opts.Plain = true
		} else if arg == "--json" {
			opts.JSON = true
		} else if arg == "-w" || arg == "--live" {
			opts.Live = true
		} else if strings.HasPrefix(arg, "--sort=") || strings.HasPrefix(arg, "-s=") {
			parts := strings.SplitN(arg, "=", 2)
			opts.SortBy = strings.ToLower(strings.TrimSpace(parts[1]))
		} else if arg == "-s" || arg == "--sort" {
			if i+1 < len(args) {
				opts.SortBy = strings.ToLower(strings.TrimSpace(args[i+1]))
				i++
			}
		} else if strings.HasPrefix(arg, "--filter=") || strings.HasPrefix(arg, "-f=") {
			parts := strings.SplitN(arg, "=", 2)
			opts.Filter = strings.TrimSpace(parts[1])
		} else if arg == "-f" || arg == "--filter" {
			if i+1 < len(args) {
				opts.Filter = strings.TrimSpace(args[i+1])
				i++
			}
		} else if strings.HasPrefix(arg, "--limit=") || strings.HasPrefix(arg, "-n=") {
			parts := strings.SplitN(arg, "=", 2)
			if n, err := strconv.Atoi(parts[1]); err == nil && n > 0 {
				opts.Limit = n
			}
		} else if arg == "-n" || arg == "--limit" {
			if i+1 < len(args) {
				if n, err := strconv.Atoi(args[i+1]); err == nil && n > 0 {
					opts.Limit = n
				}
				i++
			}
		} else if strings.HasPrefix(arg, "--kill=") || strings.HasPrefix(arg, "-k=") {
			parts := strings.SplitN(arg, "=", 2)
			if pid, err := strconv.Atoi(parts[1]); err == nil {
				opts.KillPID = pid
			}
		} else if arg == "-k" || arg == "--kill" {
			if i+1 < len(args) {
				if pid, err := strconv.Atoi(args[i+1]); err == nil {
					opts.KillPID = pid
				}
				i++
			}
		} else if strings.HasPrefix(arg, "--signal=") {
			parts := strings.SplitN(arg, "=", 2)
			if sig, err := strconv.Atoi(parts[1]); err == nil {
				opts.Signal = sig
			}
		} else if strings.HasPrefix(arg, "--iter=") {
			parts := strings.SplitN(arg, "=", 2)
			if iter, err := strconv.Atoi(parts[1]); err == nil {
				opts.Iterations = iter
			}
		} else if strings.HasPrefix(arg, "--interval=") {
			parts := strings.SplitN(arg, "=", 2)
			if d, err := time.ParseDuration(parts[1]); err == nil {
				opts.Interval = d
			}
		} else if strings.HasPrefix(arg, "--theme=") || strings.HasPrefix(arg, "--color=") {
			// skip global flags
		} else if arg == "--theme" || arg == "--color" {
			i++
		} else if !strings.HasPrefix(arg, "-") && opts.Filter == "" {
			opts.Filter = strings.TrimSpace(arg)
		}
	}

	if opts.Limit <= 0 {
		opts.Limit = 15
	}
	switch opts.SortBy {
	case "cpu", "mem", "memory", "pid", "name":
		// valid
	default:
		opts.SortBy = "cpu"
	}

	return opts
}
