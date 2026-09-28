package mdcmd

import (
	"strconv"
	"strings"
)

// Options holds configuration for markdown rendering.
type Options struct {
	FilePath string `json:"file_path"`
	Width    int    `json:"width"`
	Pager    bool   `json:"pager"`
	TOC      bool   `json:"toc"`
	Plain    bool   `json:"plain"`
	JSON     bool   `json:"json"`
}

// ParseFlags parses command line arguments for the md command.
func ParseFlags(args []string) Options {
	opts := Options{
		Width: 80,
	}

	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "--pager" || arg == "-p" {
			opts.Pager = true
			i++
		} else if arg == "--toc" || arg == "-t" {
			opts.TOC = true
			i++
		} else if arg == "--plain" {
			opts.Plain = true
			i++
		} else if arg == "--json" {
			opts.JSON = true
			i++
		} else if strings.HasPrefix(arg, "--width=") {
			if w, err := strconv.Atoi(strings.TrimPrefix(arg, "--width=")); err == nil && w > 10 {
				opts.Width = w
			}
			i++
		} else if (arg == "--width" || arg == "-w") && i+1 < len(args) {
			if w, err := strconv.Atoi(args[i+1]); err == nil && w > 10 {
				opts.Width = w
			}
			i += 2
		} else if !strings.HasPrefix(arg, "-") && opts.FilePath == "" {
			opts.FilePath = arg
			i++
		} else {
			i++
		}
	}

	return opts
}
