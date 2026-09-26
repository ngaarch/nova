package ls

import (
	"fmt"
	"path/filepath"

	"nova/internal/command"
	"nova/internal/filesystem"
	"nova/internal/output"
	"nova/internal/theme"
)

// Command returns the registered Command instance for the ls subcommand.
func Command() *command.Command {
	return &command.Command{
		Name:        "ls",
		Aliases:     []string{"list"},
		Summary:     "List directory contents with modern layout, colors, and icons",
		Usage:       "nova ls [flags] [paths...]",
		Description: "List files and directories with semantic colors, icons, and responsive layouts.",
		Phase:       3,
		Run:         Run,
	}
}

// Run executes the ls command within the provided context and argument slice.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)

	// Determine icon display policy
	showIcons := false
	if opts.Icons == "always" {
		showIcons = true
	} else if opts.Icons == "never" {
		showIcons = false
	} else {
		// auto: display icons in human mode
		showIcons = ctx.Printer.Mode == output.ModeHuman
	}

	visitedDirs := make(map[string]bool)

	for i, path := range opts.Paths {
		if len(opts.Paths) > 1 && ctx.Printer.Mode == output.ModeHuman {
			header := ctx.Theme.Format(theme.RoleAccent, path+":", ctx.Caps.ColorProfile)
			if i > 0 {
				ctx.Printer.Println()
			}
			ctx.Printer.Println(header)
		}

		if err := listPath(ctx, path, opts, showIcons, visitedDirs); err != nil {
			return err
		}
	}

	return nil
}

func listPath(ctx *command.Context, path string, opts Options, showIcons bool, visitedDirs map[string]bool) error {
	needDetails := opts.Long || opts.Sort == "size" || opts.Sort == "time"

	entries, err := filesystem.ReadDir(path, opts.All, needDetails)
	if err != nil {
		return command.NewOpError(fmt.Sprintf("cannot access %q", path), path, err, "Check file permissions and path spelling.")
	}

	SortEntries(entries, opts.Sort, opts.Reverse, opts.DirsFirst)

	switch ctx.Printer.Mode {
	case output.ModeJSON:
		if err := RenderJSON(ctx, entries); err != nil {
			return command.NewOpError("failed to serialize JSON output", path, err, "")
		}
	case output.ModePlain:
		RenderPlain(ctx, entries, opts.Long, opts.HumanReadable)
	default: // output.ModeHuman
		if opts.Long {
			RenderLong(ctx, entries, showIcons, opts.HumanReadable)
		} else {
			RenderCompact(ctx, entries, showIcons)
		}
	}

	// Recursive traversal (-R)
	if opts.Recursive {
		absPath, _ := filepath.Abs(path)
		visitedDirs[absPath] = true

		for _, entry := range entries {
			if entry.IsDir && !entry.IsSymlink && entry.Name != "." && entry.Name != ".." {
				subPath := filepath.Join(path, entry.Name)
				subAbs, _ := filepath.Abs(subPath)
				if visitedDirs[subAbs] {
					continue
				}

				if ctx.Printer.Mode == output.ModeHuman {
					ctx.Printer.Println()
					header := ctx.Theme.Format(theme.RoleAccent, subPath+":", ctx.Caps.ColorProfile)
					ctx.Printer.Println(header)
				}
				if err := listPath(ctx, subPath, opts, showIcons, visitedDirs); err != nil {
					ctx.Printer.Errorf("nova: ls: %v", err)
				}
			}
		}
	}

	return nil
}
