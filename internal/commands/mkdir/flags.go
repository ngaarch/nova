package mkdir

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Options holds the parsed command-line flags for nova mkdir.
type Options struct {
	Parents bool
	Mode    os.FileMode
	Verbose bool
	DryRun  bool
	Help    bool
	Paths   []string
}

// ParseFlags parses command-line arguments for the mkdir command.
func ParseFlags(args []string) (Options, error) {
	opts := Options{
		Mode: 0755,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--" {
			opts.Paths = append(opts.Paths, args[i+1:]...)
			break
		}

		if strings.HasPrefix(arg, "--") {
			switch {
			case arg == "--parents":
				opts.Parents = true
			case arg == "--verbose":
				opts.Verbose = true
			case arg == "--dry-run":
				opts.DryRun = true
			case arg == "--help":
				opts.Help = true
			case strings.HasPrefix(arg, "--mode="):
				modeStr := strings.TrimPrefix(arg, "--mode=")
				m, err := parseMode(modeStr)
				if err != nil {
					return opts, err
				}
				opts.Mode = m
			case arg == "--mode":
				if i+1 < len(args) {
					i++
					m, err := parseMode(args[i])
					if err != nil {
						return opts, err
					}
					opts.Mode = m
				} else {
					return opts, fmt.Errorf("option '--mode' requires an argument")
				}
			default:
				opts.Paths = append(opts.Paths, arg)
			}
			continue
		}

		if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			// Single dash flags, could be combined like -pv
			var leftover []rune
			runes := []rune(arg[1:])
			for j := 0; j < len(runes); j++ {
				r := runes[j]
				switch r {
				case 'p':
					opts.Parents = true
				case 'v':
					opts.Verbose = true
				case 'h':
					opts.Help = true
				case 'm':
					// mode can follow immediately like -m0755 or as next arg
					if j+1 < len(runes) {
						modeStr := string(runes[j+1:])
						m, err := parseMode(modeStr)
						if err != nil {
							return opts, err
						}
						opts.Mode = m
						j = len(runes) // consumed remainder
					} else if i+1 < len(args) {
						i++
						m, err := parseMode(args[i])
						if err != nil {
							return opts, err
						}
						opts.Mode = m
					} else {
						return opts, fmt.Errorf("option '-m' requires an argument")
					}
				default:
					leftover = append(leftover, r)
				}
			}
			if len(leftover) > 0 {
				opts.Paths = append(opts.Paths, "-"+string(leftover))
			}
			continue
		}

		opts.Paths = append(opts.Paths, arg)
	}

	return opts, nil
}

func parseMode(s string) (os.FileMode, error) {
	val, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid mode %q: must be octal (e.g. 0755)", s)
	}
	return os.FileMode(val), nil
}
