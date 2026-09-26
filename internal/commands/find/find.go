package find

import (
	"encoding/json"
	"path/filepath"

	"nova/internal/command"
	"nova/internal/filesystem"
	"nova/internal/output"
	"nova/internal/terminal"
	"nova/internal/theme"
)

// Command returns the registered Command instance for the find subcommand.
func Command() *command.Command {
	return &command.Command{
		Name:        "find",
		Summary:     "Search files across directories by predicates",
		Usage:       "nova find [path] [predicates...]",
		Description: "Search files across directories using predicates (name, size, type, mtime).",
		Phase:       5,
		Run:         Run,
	}
}

// Run executes the find command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	p := ParseFlags(args)
	if ctx.Printer.Mode == output.ModePlain {
		p.Plain = true
	} else if ctx.Printer.Mode == output.ModeJSON {
		p.JSON = true
	}

	walkOpts := filesystem.DefaultWalkOptions()
	walkOpts.MaxDepth = p.MaxDepth
	walkOpts.IncludeHidden = p.IncludeHidden

	var matches []*filesystem.Entry

	err := filesystem.Walk(p.RootPath, walkOpts, func(path string, entry *filesystem.Entry, depth int) error {
		relPath, _ := filepath.Rel(p.RootPath, path)
		if !Match(entry, relPath, depth, &p) {
			return nil
		}

		if p.JSON {
			cp := *entry
			matches = append(matches, &cp)
			return nil
		}

		if p.Plain || ctx.Printer.Mode == output.ModePlain {
			ctx.Printer.Println(entry.Path)
			return nil
		}

		// Human formatted mode
		role := theme.RoleRegularFile
		if entry.IsDir {
			role = theme.RoleDirectory
		} else if entry.IsBroken {
			role = theme.RoleBrokenSymlink
		} else if entry.IsSymlink {
			role = theme.RoleSymlink
		} else if entry.EntityType == theme.TypeExecutable {
			role = theme.RoleExecutable
		}

		icon := ""
		if ctx.Caps.ColorProfile != terminal.ColorNone {
			icon = theme.LookupIcon(entry.Name, entry.EntityType, ctx.Caps.UnicodeSupported)
		}
		pathStyled := ctx.Theme.Format(role, entry.Path, ctx.Caps.ColorProfile)
		ctx.Printer.Println(icon + pathStyled)
		return nil
	})

	if err != nil {
		return command.NewOpError("error during search traversal", p.RootPath, err, "")
	}

	if p.JSON {
		data, jErr := json.MarshalIndent(matches, "", "  ")
		if jErr != nil {
			return jErr
		}
		ctx.Printer.Println(string(data))
	}

	return nil
}
