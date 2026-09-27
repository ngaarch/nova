package which

import (
	"encoding/json"
	"fmt"

	"nova/internal/command"
	"nova/internal/theme"
)

// RenderHuman formats resolved executables with icons, symlink targets, and size pills.
func RenderHuman(ctx *command.Context, results []Result) {
	th := ctx.Theme
	prof := ctx.Caps.ColorProfile
	unicode := ctx.Caps.UnicodeSupported

	for _, r := range results {
		if !r.Found {
			errIcon := "✖ "
			if !unicode {
				errIcon = "x "
			}
			msg := fmt.Sprintf("%s'%s' not found in $PATH", errIcon, r.Query)
			ctx.Printer.Println(th.Format(theme.RoleError, msg, prof))
			continue
		}

		icon := theme.LookupIcon(r.Query, theme.TypeExecutable, unicode)
		if icon != "" {
			icon += " "
		}

		pathStyled := th.Format(theme.RoleExecutable, r.Path, prof)
		meta := fmt.Sprintf("[%s • %s]", r.Permissions, r.HumanSize)
		metaStyled := th.Format(theme.RoleMuted, meta, prof)

		if r.IsSymlink && r.Target != "" {
			arrow := " -> "
			targetStyled := th.Format(theme.RoleSymlink, r.Target, prof)
			ctx.Printer.Println(fmt.Sprintf("%s%s%s%s  %s", icon, pathStyled, arrow, targetStyled, metaStyled))
		} else {
			ctx.Printer.Println(fmt.Sprintf("%s%s  %s", icon, pathStyled, metaStyled))
		}
	}
}

// RenderPlain prints plain paths for pipeline safety and scripts.
func RenderPlain(ctx *command.Context, results []Result) {
	for _, r := range results {
		if r.Found {
			ctx.Printer.Println(r.Path)
		}
	}
}

// RenderJSON serializes which results into structured JSON.
func RenderJSON(ctx *command.Context, results []Result) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	ctx.Printer.Println(string(data))
	return nil
}
