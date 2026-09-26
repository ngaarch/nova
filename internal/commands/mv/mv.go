package mv

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nova/internal/command"
	"nova/internal/filesystem"
	"nova/internal/theme"
)

// Command returns the registered Command instance for the mv subcommand.
func Command() *command.Command {
	return &command.Command{
		Name:        "mv",
		Aliases:     []string{"move"},
		Summary:     "Move or rename files and directories safely and atomically",
		Usage:       "nova mv [flags] <source...> <destination>",
		Description: "Move or rename files and directories with atomic renames and cross-filesystem fallback.",
		Phase:       6,
		Run:         Run,
	}
}

// Run executes the mv command.
func Run(ctx *command.Context, args []string) error {
	opts := ParseFlags(args)

	if len(opts.Sources) == 0 || opts.Destination == "" {
		return command.NewUsageError("missing destination file operand", "Provide source and destination: nova mv [flags] <src...> <dst>")
	}

	destFi, destErr := os.Lstat(opts.Destination)
	destIsDir := destErr == nil && destFi.IsDir()

	if len(opts.Sources) > 1 && !destIsDir {
		return command.NewOpError(fmt.Sprintf("target %q is not a directory", opts.Destination), opts.Destination, nil, "When moving multiple sources, destination must be an existing directory.")
	}

	fsOpts := filesystem.MoveOptions{
		NoClobber: opts.NoClobber,
		Force:     opts.Force,
		DryRun:    opts.DryRun,
		Verbose:   opts.Verbose,
	}

	if opts.Interactive {
		inReader := bufio.NewReader(ctx.Stdin)
		fsOpts.Confirm = func(target string) (bool, error) {
			ctx.Printer.Printf("nova: overwrite %q? (y/n) [n]: ", target)
			line, err := inReader.ReadString('\n')
			if err != nil {
				return false, err
			}
			ans := strings.ToLower(strings.TrimSpace(line))
			return ans == "y" || ans == "yes", nil
		}
	}

	for _, src := range opts.Sources {
		dstPath := opts.Destination
		if destIsDir {
			dstPath = filepath.Join(opts.Destination, filepath.Base(src))
		}

		if err := filesystem.Move(src, dstPath, fsOpts); err != nil {
			return command.NewOpError(fmt.Sprintf("failed to move %q to %q", src, dstPath), src, err, "")
		}

		if opts.Verbose || opts.DryRun {
			msg := fmt.Sprintf("%s -> %s", src, dstPath)
			if opts.DryRun {
				msg = "[dry-run] " + msg
			}
			ctx.Printer.Println(ctx.Theme.Format(theme.RoleInfo, msg, ctx.Caps.ColorProfile))
		}
	}

	return nil
}
