package touch

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"nova/internal/command"
	"nova/internal/theme"
)

// RenderHuman prints user-friendly status indicators for touch operations.
func RenderHuman(ctx *command.Context, results []Result) {
	th := ctx.Theme
	prof := ctx.Caps.ColorProfile
	unicode := ctx.Caps.UnicodeSupported

	for _, r := range results {
		if r.Error != "" {
			errIcon := "✖ "
			if !unicode {
				errIcon = "x "
			}
			msg := fmt.Sprintf("%sFailed to touch %s: %s", errIcon, r.Path, r.Error)
			ctx.Printer.Println(th.Format(theme.RoleError, msg, prof))
			continue
		}

		icon := theme.LookupIcon(filepath.Base(r.Path), theme.TypeRegular, unicode)
		if icon != "" {
			icon += " "
		}

		timeStr := th.Format(theme.RoleMuted, "["+r.Timestamp.Format("2006-01-02 15:04:05")+"]", prof)
		pathStr := th.Format(theme.RoleRegularFile, r.Path, prof)

		if r.Created {
			tag := th.Format(theme.RoleSuccess, "✔ Created:", prof)
			if !unicode {
				tag = th.Format(theme.RoleSuccess, "+ Created:", prof)
			}
			ctx.Printer.Println(fmt.Sprintf("%s %s%s  %s", tag, icon, pathStr, timeStr))
		} else if r.Updated {
			tag := th.Format(theme.RoleAccent, "✔ Updated:", prof)
			if !unicode {
				tag = th.Format(theme.RoleAccent, "* Updated:", prof)
			}
			ctx.Printer.Println(fmt.Sprintf("%s %s%s  %s", tag, icon, pathStr, timeStr))
		}
	}
}

// RenderPlain prints paths of processed files.
func RenderPlain(ctx *command.Context, results []Result) {
	for _, r := range results {
		if r.Error == "" && (r.Created || r.Updated) {
			ctx.Printer.Println(r.Path)
		}
	}
}

// RenderJSON outputs structured JSON array of touch results.
func RenderJSON(ctx *command.Context, results []Result) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	ctx.Printer.Println(string(data))
	return nil
}
