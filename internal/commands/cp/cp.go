package cp

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

// Command returns the registered Command instance for the cp subcommand.
func Command() *command.Command {
	return &command.Command{
		Name:        "cp",
		Aliases:     []string{"copy"},
		Summary:     "Copy files and directories with progress, metadata preservation, and safety",
		Usage:       "nova cp [flags] <source...> <destination>",
		Description: "Copy files and directories with recursive support, overwrite protection, and metadata preservation.",
		Phase:       6,
		Run:         Run,
	}
}

// Run executes the cp command.
func Run(ctx *command.Context, args []string) error {
	opts := ParseFlags(args)

	if len(opts.Sources) == 0 || opts.Destination == "" {
		return command.NewUsageError("missing destination file operand", "Provide source and destination: nova cp [flags] <src...> <dst>")
	}

	destFi, destErr := os.Lstat(opts.Destination)
	destIsDir := destErr == nil && destFi.IsDir()

	if len(opts.Sources) > 1 && !destIsDir {
		return command.NewOpError(fmt.Sprintf("target %q is not a directory", opts.Destination), opts.Destination, nil, "When copying multiple sources, destination must be an existing directory.")
	}

	fsOpts := filesystem.CopyOptions{
		PreserveMetadata:    opts.Preserve,
		DereferenceSymlinks: opts.Dereference,
		Recursive:           opts.Recursive,
		NoClobber:           opts.NoClobber,
		Force:               opts.Force,
		DryRun:              opts.DryRun,
		Verbose:             opts.Verbose,
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

		srcFi, err := os.Lstat(src)
		if err != nil {
			return command.NewOpError(fmt.Sprintf("cannot stat %q", src), src, err, "Check file path spelling.")
		}

		if srcFi.IsDir() {
			if !opts.Recursive {
				return command.NewOpError(fmt.Sprintf("-r not specified; omitting directory %q", src), src, nil, "Use -r or -R to copy directories recursively.")
			}
			if err := filesystem.CopyDir(src, dstPath, fsOpts); err != nil {
				return command.NewOpError(fmt.Sprintf("failed to copy directory %q to %q", src, dstPath), src, err, "")
			}
		} else {
			if err := filesystem.CopyFile(src, dstPath, fsOpts); err != nil {
				return command.NewOpError(fmt.Sprintf("failed to copy file %q to %q", src, dstPath), src, err, "")
			}
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
