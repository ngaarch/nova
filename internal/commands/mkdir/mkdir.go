package mkdir

import (
	"fmt"

	"nova/internal/command"
	"nova/internal/filesystem"
	"nova/internal/theme"
)

// Command returns the registered Command instance for the mkdir subcommand.
func Command() *command.Command {
	return &command.Command{
		Name:        "mkdir",
		Aliases:     []string{"makedir"},
		Summary:     "Create directories with parent hierarchy creation and permission control",
		Usage:       "nova mkdir [flags] <directories...>",
		Description: "Create one or more directories with optional parent creation (-p), permission mode (-m), and dry-run safety.",
		Phase:       6,
		Run:         Run,
	}
}

// Run executes the mkdir command.
func Run(ctx *command.Context, args []string) error {
	opts, err := ParseFlags(args)
	if err != nil {
		return command.NewUsageError("invalid arguments for 'nova mkdir'", err.Error())
	}

	if opts.Help {
		ctx.Printer.Println(Command().Usage)
		return nil
	}

	if len(opts.Paths) == 0 {
		return command.NewUsageError("missing operand for 'nova mkdir'", "Provide one or more directories to create: nova mkdir [flags] <directories...>")
	}

	fsOpts := filesystem.MkdirOptions{
		Parents: opts.Parents,
		Mode:    opts.Mode,
		Verbose: opts.Verbose,
		DryRun:  opts.DryRun,
	}

	for _, path := range opts.Paths {
		if err := filesystem.MakeDir(path, fsOpts); err != nil {
			return command.NewOpError(fmt.Sprintf("cannot create directory %q", path), path, err, "")
		}

		if opts.Verbose || opts.DryRun {
			msg := fmt.Sprintf("created directory %q", path)
			if opts.DryRun {
				msg = "[dry-run] " + msg
			}
			ctx.Printer.Println(ctx.Theme.Format(theme.RoleSuccess, msg, ctx.Caps.ColorProfile))
		}
	}

	return nil
}
