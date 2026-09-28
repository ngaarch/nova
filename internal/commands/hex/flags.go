package hex

import (
	"strconv"
	"strings"
)

// Options holds CLI configuration for nova hex.
type Options struct {
	Paths    []string // files to inspect
	Skip     int64    // -s, --skip: byte offset to start dump
	Length   int64    // -n, --length: number of bytes to dump (-1 for all)
	Grouping int      // -g, --group: group size in bytes (1, 2, 4, 8)
	Columns  int      // -c, --cols: bytes per row (default 16)
	Find     string   // --find: byte or ascii string pattern to highlight
	Plain    bool     // --plain
	JSON     bool     // --json
}

// DefaultOptions returns standard hex dump parameters.
func DefaultOptions() Options {
	return Options{
		Skip:     0,
		Length:   -1,
		Grouping: 1,
		Columns:  16,
	}
}

// ParseFlags parses flags for nova hex.
func ParseFlags(args []string) Options {
	opts := DefaultOptions()
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--plain" {
			opts.Plain = true
		} else if arg == "--json" {
			opts.JSON = true
		} else if strings.HasPrefix(arg, "-s=") || strings.HasPrefix(arg, "--skip=") {
			parts := strings.SplitN(arg, "=", 2)
			opts.Skip = parseOffset(parts[1])
		} else if arg == "-s" || arg == "--skip" {
			if i+1 < len(args) {
				opts.Skip = parseOffset(args[i+1])
				i++
			}
		} else if strings.HasPrefix(arg, "-n=") || strings.HasPrefix(arg, "--length=") {
			parts := strings.SplitN(arg, "=", 2)
			opts.Length = parseOffset(parts[1])
		} else if arg == "-n" || arg == "--length" {
			if i+1 < len(args) {
				opts.Length = parseOffset(args[i+1])
				i++
			}
		} else if strings.HasPrefix(arg, "-g=") || strings.HasPrefix(arg, "--group=") {
			parts := strings.SplitN(arg, "=", 2)
			if g, err := strconv.Atoi(parts[1]); err == nil && g > 0 {
				opts.Grouping = g
			}
		} else if arg == "-g" || arg == "--group" {
			if i+1 < len(args) {
				if g, err := strconv.Atoi(args[i+1]); err == nil && g > 0 {
					opts.Grouping = g
				}
				i++
			}
		} else if strings.HasPrefix(arg, "-c=") || strings.HasPrefix(arg, "--cols=") {
			parts := strings.SplitN(arg, "=", 2)
			if c, err := strconv.Atoi(parts[1]); err == nil && c > 0 {
				opts.Columns = c
			}
		} else if arg == "-c" || arg == "--cols" {
			if i+1 < len(args) {
				if c, err := strconv.Atoi(args[i+1]); err == nil && c > 0 {
					opts.Columns = c
				}
				i++
			}
		} else if strings.HasPrefix(arg, "--find=") {
			parts := strings.SplitN(arg, "=", 2)
			opts.Find = parts[1]
		} else if arg == "--find" {
			if i+1 < len(args) {
				opts.Find = args[i+1]
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

	opts.Paths = positional
	return opts
}

func parseOffset(s string) int64 {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		val, err := strconv.ParseInt(s[2:], 16, 64)
		if err == nil {
			return val
		}
	}
	// Unit suffixes
	multiplier := int64(1)
	if strings.HasSuffix(s, "K") || strings.HasSuffix(s, "k") {
		multiplier = 1024
		s = s[:len(s)-1]
	} else if strings.HasSuffix(s, "M") || strings.HasSuffix(s, "m") {
		multiplier = 1024 * 1024
		s = s[:len(s)-1]
	} else if strings.HasSuffix(s, "G") || strings.HasSuffix(s, "g") {
		multiplier = 1024 * 1024 * 1024
		s = s[:len(s)-1]
	}

	val, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return val * multiplier
	}
	return 0
}
