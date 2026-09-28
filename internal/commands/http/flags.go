package httpcmd

import (
	"strings"
	"time"
)

// Options holds CLI configuration for nova http.
type Options struct {
	URL             string
	Method          string        // -X, --method (GET, POST, PUT, DELETE, HEAD, PATCH)
	Headers         []string      // -H, --header "Key: Value"
	Data            string        // -d, --data
	DataFile        string        // --data-file
	Timeout         time.Duration // -t, --timeout (default 10s)
	FollowRedirects bool          // -L, --location (default true)
	Insecure        bool          // -k, --insecure
	IncludeHeaders  bool          // -i, --include
	HeadOnly        bool          // -I, --head
	OutputFile      string        // -o, --output
	Timing          bool          // --timing (default true in human mode)
	Plain           bool          // --plain
	JSON            bool          // --json
}

// DefaultOptions returns standard HTTP options.
func DefaultOptions() Options {
	return Options{
		Method:          "GET",
		Timeout:         10 * time.Second,
		FollowRedirects: true,
		Timing:          true,
	}
}

// ParseFlags parses flags for nova http.
func ParseFlags(args []string) Options {
	opts := DefaultOptions()
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--plain" {
			opts.Plain = true
		} else if arg == "--json" {
			opts.JSON = true
		} else if arg == "-k" || arg == "--insecure" {
			opts.Insecure = true
		} else if arg == "-i" || arg == "--include" {
			opts.IncludeHeaders = true
		} else if arg == "-I" || arg == "--head" {
			opts.HeadOnly = true
			opts.Method = "HEAD"
		} else if arg == "-L" || arg == "--location" {
			opts.FollowRedirects = true
		} else if arg == "--no-timing" {
			opts.Timing = false
		} else if strings.HasPrefix(arg, "-X=") || strings.HasPrefix(arg, "--method=") {
			parts := strings.SplitN(arg, "=", 2)
			opts.Method = strings.ToUpper(parts[1])
		} else if arg == "-X" || arg == "--method" {
			if i+1 < len(args) {
				opts.Method = strings.ToUpper(args[i+1])
				i++
			}
		} else if strings.HasPrefix(arg, "-H=") || strings.HasPrefix(arg, "--header=") {
			parts := strings.SplitN(arg, "=", 2)
			opts.Headers = append(opts.Headers, parts[1])
		} else if arg == "-H" || arg == "--header" {
			if i+1 < len(args) {
				opts.Headers = append(opts.Headers, args[i+1])
				i++
			}
		} else if strings.HasPrefix(arg, "-d=") || strings.HasPrefix(arg, "--data=") {
			parts := strings.SplitN(arg, "=", 2)
			opts.Data = parts[1]
			if opts.Method == "GET" {
				opts.Method = "POST"
			}
		} else if arg == "-d" || arg == "--data" {
			if i+1 < len(args) {
				opts.Data = args[i+1]
				if opts.Method == "GET" {
					opts.Method = "POST"
				}
				i++
			}
		} else if strings.HasPrefix(arg, "-o=") || strings.HasPrefix(arg, "--output=") {
			parts := strings.SplitN(arg, "=", 2)
			opts.OutputFile = parts[1]
		} else if arg == "-o" || arg == "--output" {
			if i+1 < len(args) {
				opts.OutputFile = args[i+1]
				i++
			}
		} else if strings.HasPrefix(arg, "-t=") || strings.HasPrefix(arg, "--timeout=") {
			parts := strings.SplitN(arg, "=", 2)
			if d, err := time.ParseDuration(parts[1]); err == nil {
				opts.Timeout = d
			}
		} else if arg == "-t" || arg == "--timeout" {
			if i+1 < len(args) {
				if d, err := time.ParseDuration(args[i+1]); err == nil {
					opts.Timeout = d
				}
				i++
			}
		} else if strings.HasPrefix(arg, "--theme=") || strings.HasPrefix(arg, "--color=") || strings.HasPrefix(arg, "--config=") {
			// skip global flags
		} else if arg == "--theme" || arg == "--color" || arg == "--config" {
			i++
		} else if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
		}
	}

	if len(positional) > 0 {
		opts.URL = positional[0]
	}

	return opts
}
