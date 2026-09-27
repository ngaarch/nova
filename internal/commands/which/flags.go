package which

import "strings"

// Options represents parsed flags for the which command.
type Options struct {
	All       bool     `json:"all"`
	Silent    bool     `json:"silent"`
	Plain     bool     `json:"plain"`
	JSON      bool     `json:"json"`
	ExecNames []string `json:"exec_names"`
}

// ParseFlags parses command-line arguments for the which command.
func ParseFlags(args []string) Options {
	opts := Options{}

	for _, arg := range args {
		if arg == "-a" || arg == "--all" {
			opts.All = true
		} else if arg == "-s" || arg == "--silent" {
			opts.Silent = true
		} else if arg == "--plain" {
			opts.Plain = true
		} else if arg == "--json" {
			opts.JSON = true
		} else if !strings.HasPrefix(arg, "-") && arg != "" {
			opts.ExecNames = append(opts.ExecNames, arg)
		}
	}

	return opts
}
