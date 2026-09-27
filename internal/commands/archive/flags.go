package archive

import (
	"strings"
)

// Mode represents the archive sub-operation.
type Mode string

const (
	ModePack   Mode = "pack"
	ModeUnpack Mode = "unpack"
	ModeList   Mode = "list"
)

// Options holds command configuration for archive.
type Options struct {
	Mode        Mode     // pack, unpack, list
	Output      string   // -o, --output
	Destination string   // -C, --dest
	ArchiveFile string   // target archive file
	Targets     []string // files/dirs to pack
	Plain       bool     // --plain
	JSON        bool     // --json
}

// ParseFlags extracts archive command options.
func ParseFlags(args []string) Options {
	var opts Options
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--plain" {
			opts.Plain = true
		} else if arg == "--json" {
			opts.JSON = true
		} else if (arg == "-o" || arg == "--output") && i+1 < len(args) {
			opts.Output = args[i+1]
			i++
		} else if (arg == "-C" || arg == "--dest") && i+1 < len(args) {
			opts.Destination = args[i+1]
			i++
		} else if strings.HasPrefix(arg, "--theme=") || strings.HasPrefix(arg, "--color=") || strings.HasPrefix(arg, "--config=") {
			// skip
		} else if arg == "--theme" || arg == "--color" || arg == "--config" {
			i++
		} else if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
		}
	}

	if len(positional) > 0 {
		cmd := strings.ToLower(positional[0])
		switch cmd {
		case "pack", "create", "zip":
			opts.Mode = ModePack
			if len(positional) > 1 {
				if opts.Output == "" && (strings.HasSuffix(positional[1], ".zip") || strings.HasSuffix(positional[1], ".tar.gz") || strings.HasSuffix(positional[1], ".tar") || strings.HasSuffix(positional[1], ".tgz")) {
					opts.Output = positional[1]
					opts.Targets = positional[2:]
				} else {
					opts.Targets = positional[1:]
				}
			}
		case "unpack", "extract", "unzip":
			opts.Mode = ModeUnpack
			if len(positional) > 1 {
				opts.ArchiveFile = positional[1]
			}
		case "list", "ls":
			opts.Mode = ModeList
			if len(positional) > 1 {
				opts.ArchiveFile = positional[1]
			}
		default:
			// Auto-detect mode based on file extension
			if strings.HasSuffix(cmd, ".zip") || strings.HasSuffix(cmd, ".tar.gz") || strings.HasSuffix(cmd, ".tar") || strings.HasSuffix(cmd, ".tgz") {
				opts.Mode = ModeList
				opts.ArchiveFile = cmd
			} else {
				opts.Mode = ModePack
				opts.Targets = positional
			}
		}
	}

	if opts.Destination == "" {
		opts.Destination = "."
	}

	return opts
}
