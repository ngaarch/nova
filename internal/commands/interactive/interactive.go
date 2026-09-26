package interactive

import (
	"fmt"
	"strings"

	"nova/internal/command"
	"nova/internal/interactive"
)

// Options holds flags parsed for nova interactive.
type Options struct {
	ShowHidden bool
	Help       bool
	TargetDir  string
}

// ParseFlags parses arguments for the interactive command.
func ParseFlags(args []string) (Options, error) {
	opts := Options{
		TargetDir: ".",
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--help" || arg == "-h":
			opts.Help = true
		case arg == "--all" || arg == "-a":
			opts.ShowHidden = true
		case strings.HasPrefix(arg, "-"):
			return opts, fmt.Errorf("unknown flag: %s", arg)
		default:
			opts.TargetDir = arg
		}
	}

	return opts, nil
}

// Command returns the command specification for nova interactive.
func Command() *command.Command {
	return &command.Command{
		Name:        "interactive",
		Aliases:     []string{"ui", "tui"},
		Summary:     "Interactive terminal file navigator and previewer",
		Usage:       "nova interactive [flags] [path]",
		Description: "Explore, search, and preview filesystem entries with keyboard navigation in a dual-pane terminal interface.",
		Phase:       7,
		Run:         Run,
	}
}

// Run executes the interactive command.
func Run(ctx *command.Context, args []string) error {
	opts, err := ParseFlags(args)
	if err != nil {
		return command.NewUsageError("invalid arguments for 'nova interactive'", err.Error())
	}

	if opts.Help {
		ctx.Printer.Println(Command().Usage)
		return nil
	}

	if !ctx.Caps.IsTTY {
		return command.NewOpError(
			"interactive mode requires a controlling terminal TTY",
			"",
			nil,
			"Run 'nova interactive' in an interactive terminal or use 'nova ls' for pipeline output.",
		)
	}

	return interactive.Run(opts.TargetDir, opts.ShowHidden, ctx.Caps, ctx.Theme)
}
