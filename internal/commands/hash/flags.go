package hash

import (
	"strings"
)

// Options holds CLI configuration for nova hash.
type Options struct {
	Algorithm string   // sha256, sha512, sha1, md5, crc32
	Paths     []string // files or directories to hash
	CheckFile string   // -c, --check: path to checksum file to verify
	Recursive bool     // -r, --recursive
	Quiet     bool     // -q, --quiet: only report errors in check mode
	Plain     bool     // --plain
	JSON      bool     // --json
}

// DefaultOptions returns standard hash parameters.
func DefaultOptions() Options {
	return Options{
		Algorithm: "sha256",
		Paths:     nil,
	}
}

// ParseFlags parses flags for nova hash.
func ParseFlags(args []string) Options {
	opts := DefaultOptions()
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--plain" {
			opts.Plain = true
		} else if arg == "--json" {
			opts.JSON = true
		} else if arg == "-r" || arg == "--recursive" {
			opts.Recursive = true
		} else if arg == "-q" || arg == "--quiet" {
			opts.Quiet = true
		} else if strings.HasPrefix(arg, "-a=") || strings.HasPrefix(arg, "--algo=") {
			parts := strings.SplitN(arg, "=", 2)
			opts.Algorithm = strings.ToLower(parts[1])
		} else if arg == "-a" || arg == "--algo" {
			if i+1 < len(args) {
				opts.Algorithm = strings.ToLower(args[i+1])
				i++
			}
		} else if strings.HasPrefix(arg, "-c=") || strings.HasPrefix(arg, "--check=") {
			parts := strings.SplitN(arg, "=", 2)
			opts.CheckFile = parts[1]
		} else if arg == "-c" || arg == "--check" {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				opts.CheckFile = args[i+1]
				i++
			} else {
				opts.CheckFile = "-" // stdin
			}
		} else if strings.HasPrefix(arg, "--theme=") || strings.HasPrefix(arg, "--color=") || strings.HasPrefix(arg, "--config=") {
			// skip global flags
		} else if arg == "--theme" || arg == "--color" || arg == "--config" {
			i++
		} else if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
		}
	}

	opts.Paths = positional
	return opts
}
