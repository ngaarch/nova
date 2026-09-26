package cat

import (
	"strings"
)

// Options holds configuration options parsed from command-line arguments for cat.
type Options struct {
	Number         bool   // -n, --number, --numbers
	NumberNonblank bool   // -b, --number-nonblank
	SqueezeBlank   bool   // -s, --squeeze-blank
	ShowEnds       bool   // -E, --show-ends
	ShowTabs       bool   // -T, --show-tabs
	ShowAll        bool   // -A, -v, --show-all
	Plain          bool   // -p, --plain
	NoPager        bool   // --no-pager
	Paging         string // "auto", "always", "never"
	Language       string // -l, --lang, --language=<lang>
	FormatJSON     bool   // --json
	ForceBinary    bool   // --binary
	HexDump        bool   // --hex
	Files          []string
}

// DefaultOptions returns default options for cat.
func DefaultOptions() Options {
	return Options{
		Number:         false,
		NumberNonblank: false,
		SqueezeBlank:   false,
		ShowEnds:       false,
		ShowTabs:       false,
		ShowAll:        false,
		Plain:          false,
		NoPager:        false,
		Paging:         "auto",
		Language:       "",
		FormatJSON:     false,
		ForceBinary:    false,
		HexDump:        false,
		Files:          nil,
	}
}

// ParseFlags parses flags specifically for the cat subcommand.
func ParseFlags(args []string) Options {
	opts := DefaultOptions()
	var files []string

	i := 0
	for i < len(args) {
		arg := args[i]

		// End of flags delimiter
		if arg == "--" {
			i++
			for i < len(args) {
				files = append(files, args[i])
				i++
			}
			break
		}

		// Skip global flags
		if arg == "--plain" {
			opts.Plain = true
			i++
			continue
		}
		if arg == "--json" {
			opts.FormatJSON = true
			i++
			continue
		}
		if strings.HasPrefix(arg, "--theme=") || strings.HasPrefix(arg, "--color=") || strings.HasPrefix(arg, "--config=") {
			i++
			continue
		}
		if arg == "--theme" || arg == "--color" || arg == "--config" {
			i += 2
			continue
		}

		if arg == "-n" || arg == "--number" || arg == "--numbers" {
			opts.Number = true
			i++
		} else if arg == "-b" || arg == "--number-nonblank" {
			opts.NumberNonblank = true
			opts.Number = true
			i++
		} else if arg == "-s" || arg == "--squeeze-blank" {
			opts.SqueezeBlank = true
			i++
		} else if arg == "-E" || arg == "--show-ends" {
			opts.ShowEnds = true
			i++
		} else if arg == "-T" || arg == "--show-tabs" {
			opts.ShowTabs = true
			i++
		} else if arg == "-A" || arg == "-v" || arg == "--show-all" {
			opts.ShowAll = true
			opts.ShowEnds = true
			opts.ShowTabs = true
			i++
		} else if arg == "-p" {
			opts.Plain = true
			i++
		} else if arg == "--no-pager" {
			opts.NoPager = true
			opts.Paging = "never"
			i++
		} else if strings.HasPrefix(arg, "--paging=") {
			opts.Paging = strings.TrimPrefix(arg, "--paging=")
			if opts.Paging == "never" {
				opts.NoPager = true
			}
			i++
		} else if arg == "--paging" && i+1 < len(args) {
			opts.Paging = args[i+1]
			if opts.Paging == "never" {
				opts.NoPager = true
			}
			i += 2
		} else if strings.HasPrefix(arg, "--lang=") || strings.HasPrefix(arg, "--language=") {
			idx := strings.Index(arg, "=")
			opts.Language = arg[idx+1:]
			i++
		} else if (arg == "-l" || arg == "--lang" || arg == "--language") && i+1 < len(args) {
			opts.Language = args[i+1]
			i += 2
		} else if arg == "--binary" {
			opts.ForceBinary = true
			i++
		} else if arg == "--hex" {
			opts.HexDump = true
			i++
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 && !strings.HasPrefix(arg, "--") {
			// Bundled short flags like -ns, -bE, etc.
			for _, r := range arg[1:] {
				switch r {
				case 'n':
					opts.Number = true
				case 'b':
					opts.NumberNonblank = true
					opts.Number = true
				case 's':
					opts.SqueezeBlank = true
				case 'E':
					opts.ShowEnds = true
				case 'T':
					opts.ShowTabs = true
				case 'A', 'v':
					opts.ShowAll = true
					opts.ShowEnds = true
					opts.ShowTabs = true
				case 'p':
					opts.Plain = true
				}
			}
			i++
		} else {
			files = append(files, arg)
			i++
		}
	}

	opts.Files = files
	return opts
}
