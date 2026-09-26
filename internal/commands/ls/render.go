package ls

import (
	"fmt"

	"nova/internal/command"
	"nova/internal/filesystem"
	"nova/internal/renderer"
	"nova/internal/theme"
)

// RenderCompact prints entries formatted across responsive columns.
func RenderCompact(ctx *command.Context, entries []filesystem.Entry, showIcons bool) {
	if len(entries) == 0 {
		return
	}

	th := ctx.Theme
	prof := ctx.Caps.ColorProfile
	items := make([]string, len(entries))

	for i, entry := range entries {
		icon := ""
		if showIcons {
			icon = theme.LookupIcon(entry.Name, entry.EntityType, ctx.Caps.UnicodeSupported)
		}
		items[i] = FormatName(entry, icon, th, prof)
	}

	grid := renderer.NewGrid(ctx.Caps.Width, 2)
	lines := grid.Render(items)
	for _, line := range lines {
		ctx.Printer.Println(line)
	}
}

// RenderLong prints detailed file metadata using responsive tabular layout.
func RenderLong(ctx *command.Context, entries []filesystem.Entry, showIcons bool, humanSize bool) {
	if len(entries) == 0 {
		return
	}

	th := ctx.Theme
	prof := ctx.Caps.ColorProfile
	tbl := renderer.NewTable(ctx.Caps.Width)

	tbl.AddColumn("Permissions", renderer.AlignLeft)
	tbl.AddColumn("Owner", renderer.AlignLeft)
	tbl.AddColumn("Group", renderer.AlignLeft)
	tbl.AddColumn("Size", renderer.AlignRight)
	tbl.AddColumn("Date", renderer.AlignLeft)
	tbl.AddColumn("Name", renderer.AlignLeft)

	var totalSize int64
	var numDirs, numFiles int

	for _, entry := range entries {
		if entry.IsDir {
			numDirs++
		} else {
			numFiles++
			totalSize += entry.Size
		}

		icon := ""
		if showIcons {
			icon = theme.LookupIcon(entry.Name, entry.EntityType, ctx.Caps.UnicodeSupported)
		}

		permStr := FormatPermissions(entry.Permissions, th, prof)
		ownerStr := th.Format(theme.RoleUser, entry.Owner, prof)
		groupStr := th.Format(theme.RoleGroup, entry.Group, prof)
		sizeStr := th.Format(theme.RoleSize, FormatSize(entry.Size, humanSize), prof)
		dateStr := th.Format(theme.RoleDate, FormatTime(entry.ModTime), prof)
		nameStr := FormatName(entry, icon, th, prof)

		tbl.AddRow(permStr, ownerStr, groupStr, sizeStr, dateStr, nameStr)
	}

	lines := tbl.Render(false)
	for _, line := range lines {
		ctx.Printer.Println(line)
	}

	summary := fmt.Sprintf("%d items (%d directories, %d files) • %s total",
		len(entries), numDirs, numFiles, FormatSize(totalSize, humanSize))
	summaryStyled := th.Format(theme.RoleMuted, summary, prof)
	ctx.Printer.Println(summaryStyled)
}

// RenderPlain prints plain text listings for scripts or pipelines.
func RenderPlain(ctx *command.Context, entries []filesystem.Entry, long bool, humanSize bool) {
	if !long {
		for _, e := range entries {
			ctx.Printer.Println(e.Name)
		}
		return
	}

	tbl := renderer.NewTable(ctx.Caps.Width)
	for _, entry := range entries {
		sizeStr := FormatSize(entry.Size, humanSize)
		dateStr := FormatTime(entry.ModTime)
		name := entry.Name
		if entry.IsSymlink && entry.LinkTarget != "" {
			name += " -> " + entry.LinkTarget
		}
		tbl.AddRow(entry.Permissions, entry.Owner, entry.Group, sizeStr, dateStr, name)
	}

	for _, line := range tbl.Render(true) {
		ctx.Printer.Println(line)
	}
}

// RenderJSON serializes entries as a structured JSON array.
func RenderJSON(ctx *command.Context, entries []filesystem.Entry) error {
	return ctx.Printer.PrintJSON(entries)
}
