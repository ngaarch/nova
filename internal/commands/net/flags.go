package netcmd

import (
	"strconv"
	"strings"
	"time"
)

// Options holds CLI configuration for nova net.
type Options struct {
	Action   string        // "ping" (default), "scan", "dns"
	Target   string        // host or domain
	Count    int           // -c, --count: ping iterations (default: 4)
	Interval time.Duration // -i, --interval: delay between pings (default: 500ms)
	Timeout  time.Duration // -W, --timeout: socket timeout (default: 2s)
	Ports    []int         // -p, --ports: port list for port scan
	Plain    bool          // --plain
	JSON     bool          // --json
}

// DefaultOptions returns standard net command options.
func DefaultOptions() Options {
	return Options{
		Action:   "ping",
		Count:    4,
		Interval: 300 * time.Millisecond,
		Timeout:  2 * time.Second,
		Ports:    DefaultPorts(),
	}
}

// DefaultPorts returns top standard ports to scan.
func DefaultPorts() []int {
	return []int{
		21, 22, 25, 53, 80, 110, 143, 443, 465, 587,
		993, 995, 3000, 3306, 5432, 6379, 8000, 8080, 8443, 27017,
	}
}

// ParseFlags parses flags for nova net.
func ParseFlags(args []string) Options {
	opts := DefaultOptions()
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--plain" {
			opts.Plain = true
		} else if arg == "--json" {
			opts.JSON = true
		} else if strings.HasPrefix(arg, "-c=") || strings.HasPrefix(arg, "--count=") {
			parts := strings.SplitN(arg, "=", 2)
			if n, err := strconv.Atoi(parts[1]); err == nil && n > 0 {
				opts.Count = n
			}
		} else if arg == "-c" || arg == "--count" {
			if i+1 < len(args) {
				if n, err := strconv.Atoi(args[i+1]); err == nil && n > 0 {
					opts.Count = n
				}
				i++
			}
		} else if strings.HasPrefix(arg, "-W=") || strings.HasPrefix(arg, "--timeout=") {
			parts := strings.SplitN(arg, "=", 2)
			if d, err := time.ParseDuration(parts[1]); err == nil {
				opts.Timeout = d
			}
		} else if arg == "-W" || arg == "--timeout" {
			if i+1 < len(args) {
				if d, err := time.ParseDuration(args[i+1]); err == nil {
					opts.Timeout = d
				}
				i++
			}
		} else if strings.HasPrefix(arg, "-i=") || strings.HasPrefix(arg, "--interval=") {
			parts := strings.SplitN(arg, "=", 2)
			if d, err := time.ParseDuration(parts[1]); err == nil {
				opts.Interval = d
			}
		} else if arg == "-i" || arg == "--interval" {
			if i+1 < len(args) {
				if d, err := time.ParseDuration(args[i+1]); err == nil {
					opts.Interval = d
				}
				i++
			}
		} else if strings.HasPrefix(arg, "-p=") || strings.HasPrefix(arg, "--ports=") {
			parts := strings.SplitN(arg, "=", 2)
			opts.Ports = parsePortList(parts[1])
		} else if arg == "-p" || arg == "--ports" {
			if i+1 < len(args) {
				opts.Ports = parsePortList(args[i+1])
				i++
			}
		} else if arg == "ping" || arg == "scan" || arg == "dns" {
			opts.Action = arg
		} else if strings.HasPrefix(arg, "--theme=") || strings.HasPrefix(arg, "--color=") {
			// skip global flags
		} else if arg == "--theme" || arg == "--color" {
			i++
		} else if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
		}
	}

	if len(positional) > 0 {
		if opts.Target == "" {
			opts.Target = positional[0]
		}
	}

	if opts.Target == "" {
		opts.Target = "127.0.0.1"
	}

	return opts
}

func parsePortList(s string) []int {
	var result []int
	parts := strings.Split(s, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			rangeParts := strings.SplitN(part, "-", 2)
			start, e1 := strconv.Atoi(rangeParts[0])
			end, e2 := strconv.Atoi(rangeParts[1])
			if e1 == nil && e2 == nil && start <= end && start > 0 && end <= 65535 {
				for p := start; p <= end; p++ {
					result = append(result, p)
				}
			}
		} else if p, err := strconv.Atoi(part); err == nil && p > 0 && p <= 65535 {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		return DefaultPorts()
	}
	return result
}
