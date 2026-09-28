package colorcmd

import (
	"strings"
)

// Options holds configuration for the color command.
type Options struct {
	Mode  string `json:"mode"` // "all", "16", "256", "truecolor", "contrast"
	Fg    string `json:"fg"`
	Bg    string `json:"bg"`
	Plain bool   `json:"plain"`
	JSON  bool   `json:"json"`
}

// ParseFlags parses command line arguments for the color command.
func ParseFlags(args []string) Options {
	opts := Options{
		Mode: "all",
		Fg:   "#ABB2BF", // Default foreground (OneDark white/gray)
		Bg:   "#282C34", // Default background (OneDark dark)
	}

	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "--plain" {
			opts.Plain = true
			i++
		} else if arg == "--json" {
			opts.JSON = true
			i++
		} else if strings.HasPrefix(arg, "--mode=") {
			opts.Mode = strings.ToLower(strings.TrimPrefix(arg, "--mode="))
			i++
		} else if (arg == "--mode" || arg == "-m") && i+1 < len(args) {
			opts.Mode = strings.ToLower(args[i+1])
			i += 2
		} else if strings.HasPrefix(arg, "--fg=") {
			opts.Fg = strings.TrimPrefix(arg, "--fg=")
			i++
		} else if arg == "--fg" && i+1 < len(args) {
			opts.Fg = args[i+1]
			i += 2
		} else if strings.HasPrefix(arg, "--bg=") {
			opts.Bg = strings.TrimPrefix(arg, "--bg=")
			i++
		} else if arg == "--bg" && i+1 < len(args) {
			opts.Bg = args[i+1]
			i += 2
		} else if !strings.HasPrefix(arg, "-") {
			opts.Mode = strings.ToLower(arg)
			i++
		} else {
			i++
		}
	}

	switch opts.Mode {
	case "16", "256", "truecolor", "contrast":
		// valid modes
	default:
		opts.Mode = "all"
	}

	return opts
}
