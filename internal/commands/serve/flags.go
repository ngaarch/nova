package servecmd

import (
	"strconv"
	"strings"
)

// Options holds CLI configuration for nova serve.
type Options struct {
	Port   int    // -p, --port (default: 8080)
	Host   string // -h, --host (default: "127.0.0.1")
	Dir    string // -d, --dir (default: ".")
	SPA    bool   // -s, --spa (Single Page Application fallback to index.html)
	CORS   bool   // --cors (Enable CORS headers)
	QR     bool   // --qr (Render terminal QR code)
	Once   bool   // test mode: exit after 1 request or startup
	Plain  bool   // --plain
	JSON   bool   // --json
}

// DefaultOptions returns standard serve command options.
func DefaultOptions() Options {
	return Options{
		Port: 8080,
		Host: "127.0.0.1",
		Dir:  ".",
		QR:   true,
	}
}

// ParseFlags parses flags for nova serve.
func ParseFlags(args []string) Options {
	opts := DefaultOptions()
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--plain" {
			opts.Plain = true
		} else if arg == "--json" {
			opts.JSON = true
		} else if arg == "-s" || arg == "--spa" {
			opts.SPA = true
		} else if arg == "--cors" {
			opts.CORS = true
		} else if arg == "--no-qr" {
			opts.QR = false
		} else if arg == "--qr" {
			opts.QR = true
		} else if arg == "--once" {
			opts.Once = true
		} else if strings.HasPrefix(arg, "-p=") || strings.HasPrefix(arg, "--port=") {
			parts := strings.SplitN(arg, "=", 2)
			if p, err := strconv.Atoi(parts[1]); err == nil && p > 0 && p <= 65535 {
				opts.Port = p
			}
		} else if arg == "-p" || arg == "--port" {
			if i+1 < len(args) {
				if p, err := strconv.Atoi(args[i+1]); err == nil && p > 0 && p <= 65535 {
					opts.Port = p
				}
				i++
			}
		} else if strings.HasPrefix(arg, "--host=") {
			parts := strings.SplitN(arg, "=", 2)
			opts.Host = strings.TrimSpace(parts[1])
		} else if arg == "--host" {
			if i+1 < len(args) {
				opts.Host = strings.TrimSpace(args[i+1])
				i++
			}
		} else if strings.HasPrefix(arg, "-d=") || strings.HasPrefix(arg, "--dir=") {
			parts := strings.SplitN(arg, "=", 2)
			opts.Dir = strings.TrimSpace(parts[1])
		} else if arg == "-d" || arg == "--dir" {
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
		opts.Dir = positional[0]
	}

	return opts
}
