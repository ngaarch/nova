package qrcmd

import (
	"strings"
)

// Options holds CLI configuration for nova qr.
type Options struct {
	Text      string // content to encode into QR code
	File      string // -f, --file: read content from file
	Invert    bool   // -i, --invert: invert dark and light blocks
	QuietZone int    // -q, --quiet: border quiet zone width (default 2)
	Plain     bool   // --plain
	JSON      bool   // --json
}

// DefaultOptions returns standard QR options.
func DefaultOptions() Options {
	return Options{
		QuietZone: 2,
	}
}

// ParseFlags parses flags for nova qr.
func ParseFlags(args []string) Options {
	opts := DefaultOptions()
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--plain" {
			opts.Plain = true
		} else if arg == "--json" {
			opts.JSON = true
		} else if arg == "-i" || arg == "--invert" {
			opts.Invert = true
		} else if strings.HasPrefix(arg, "-f=") || strings.HasPrefix(arg, "--file=") {
			parts := strings.SplitN(arg, "=", 2)
			opts.File = parts[1]
		} else if arg == "-f" || arg == "--file" {
			if i+1 < len(args) {
				opts.File = args[i+1]
				i++
			}
		} else if strings.HasPrefix(arg, "-q=") || strings.HasPrefix(arg, "--quiet=") {
			parts := strings.SplitN(arg, "=", 2)
			if parts[1] == "0" {
				opts.QuietZone = 0
			} else if parts[1] == "1" {
				opts.QuietZone = 1
			} else if parts[1] == "2" {
				opts.QuietZone = 2
			} else if parts[1] == "4" {
				opts.QuietZone = 4
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
		opts.Text = strings.Join(positional, " ")
	}

	return opts
}
