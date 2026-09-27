package touch

import "strings"

// Options represents parsed command-line flags for the touch command.
type Options struct {
	Parents   bool     `json:"parents"`
	NoCreate  bool     `json:"no_create"`
	Date      string   `json:"date,omitempty"`
	Reference string   `json:"reference,omitempty"`
	Plain     bool     `json:"plain"`
	JSON      bool     `json:"json"`
	Files     []string `json:"files"`
}

// ParseFlags parses command-line arguments for the touch command.
func ParseFlags(args []string) Options {
	opts := Options{}

	i := 0
	for i < len(args) {
		arg := args[i]

		if arg == "-p" || arg == "--parents" {
			opts.Parents = true
		} else if arg == "-c" || arg == "--no-create" {
			opts.NoCreate = true
		} else if arg == "--plain" {
			opts.Plain = true
		} else if arg == "--json" {
			opts.JSON = true
		} else if arg == "-d" || arg == "--date" {
			if i+1 < len(args) {
				opts.Date = args[i+1]
				i++
			}
		} else if strings.HasPrefix(arg, "--date=") {
			opts.Date = strings.TrimPrefix(arg, "--date=")
		} else if arg == "-r" || arg == "--reference" {
			if i+1 < len(args) {
				opts.Reference = args[i+1]
				i++
			}
		} else if strings.HasPrefix(arg, "--reference=") {
			opts.Reference = strings.TrimPrefix(arg, "--reference=")
		} else if !strings.HasPrefix(arg, "-") && arg != "" {
			opts.Files = append(opts.Files, arg)
		}
		i++
	}

	return opts
}
