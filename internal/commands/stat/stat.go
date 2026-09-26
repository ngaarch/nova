package stat

import (
	"fmt"
	"strings"

	"nova/internal/command"
	"nova/internal/filesystem"
	"nova/internal/output"
)

// Command returns the registered Command instance for the stat subcommand.
func Command() *command.Command {
	return &command.Command{
		Name:        "stat",
		Summary:     "Display structured file status and filesystem metadata",
		Usage:       "nova stat [flags] <paths...>",
		Description: "Display detailed file metadata, allocation blocks, permissions, and timestamps.",
		Phase:       5,
		Run:         Run,
	}
}

// Run executes the stat command.
func Run(ctx *command.Context, args []string) error {
	var isPlain, isJSON bool
	var paths []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--plain" {
			isPlain = true
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			isJSON = true
			ctx.Printer.Mode = output.ModeJSON
		} else if strings.HasPrefix(arg, "--theme=") || strings.HasPrefix(arg, "--color=") || strings.HasPrefix(arg, "--config=") {
			continue
		} else if arg == "--theme" || arg == "--color" || arg == "--config" {
			i++
			continue
		} else if !strings.HasPrefix(arg, "-") {
			paths = append(paths, arg)
		}
	}

	if ctx.Printer.Mode == output.ModePlain {
		isPlain = true
	} else if ctx.Printer.Mode == output.ModeJSON {
		isJSON = true
	}

	if len(paths) == 0 {
		return command.NewUsageError("missing file operand for 'nova stat'", "Provide one or more file or directory paths: nova stat <file>")
	}

	for i, path := range paths {
		st, err := filesystem.GetDetailedStat(path)
		if err != nil {
			return command.NewOpError(fmt.Sprintf("cannot stat %q", path), path, err, "Check file path spelling and permissions.")
		}

		if isJSON {
			if err := RenderJSON(ctx, st); err != nil {
				return err
			}
		} else if isPlain {
			if i > 0 {
				ctx.Printer.Println("---")
			}
			RenderPlain(ctx, st)
		} else {
			if i > 0 {
				ctx.Printer.Println()
			}
			RenderHuman(ctx, st)
		}
	}

	return nil
}
