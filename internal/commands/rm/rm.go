package rm

import (
	"bufio"
	"fmt"
	"strings"

	"nova/internal/command"
	"nova/internal/filesystem"
	"nova/internal/theme"
)

// Command returns the registered Command instance for the rm subcommand.
func Command() *command.Command {
	return &command.Command{
		Name:        "rm",
		Aliases:     []string{"remove"},
		Summary:     "Remove files and directories with root protection and interactive safety",
		Usage:       "nova rm [flags] <targets...>",
		Description: "Remove files and directories with root protection, recursive confirmation, and safety controls.",
		Phase:       6,
		Run:         Run,
	}
}

// Run executes the rm command.
func Run(ctx *command.Context, args []string) error {
	opts := ParseFlags(args)

	if len(opts.Targets) == 0 {
		return command.NewUsageError("missing operand for 'nova rm'", "Provide one or more files or directories to remove: nova rm [flags] <targets...>")
	}

	fsOpts := filesystem.RemoveOptions{
		Recursive:   opts.Recursive,
		Force:       opts.Force,
		DryRun:      opts.DryRun,
		Verbose:     opts.Verbose,
		ProtectRoot: true,
	}

	if opts.Interactive {
		inReader := bufio.NewReader(ctx.Stdin)
		fsOpts.Confirm = func(target string) (bool, error) {
			ctx.Printer.Printf("nova: remove %q? (y/n) [n]: ", target)
			line, err := inReader.ReadString('\n')
			if err != nil {
				return false, err
			}
			ans := strings.ToLower(strings.TrimSpace(line))
			return ans == "y" || ans == "yes", nil
		}
	}

	for _, target := range opts.Targets {
		if err := filesystem.Remove(target, fsOpts); err != nil {
			return command.NewOpError(fmt.Sprintf("cannot remove %q", target), target, err, "")
		}

		if opts.Verbose || opts.DryRun {
			msg := fmt.Sprintf("removed %s", target)
			if opts.DryRun {
				msg = "[dry-run] " + msg
			}
			ctx.Printer.Println(ctx.Theme.Format(theme.RoleInfo, msg, ctx.Caps.ColorProfile))
		}
	}

	return nil
}
