package tree

import (
	"encoding/json"
	"fmt"
	"strings"

	"nova/internal/command"
	"nova/internal/filesystem"
	"nova/internal/renderer"
	"nova/internal/terminal"
	"nova/internal/theme"
)

// RenderTree renders the directory tree to the printer according to options.
func RenderTree(ctx *command.Context, root *filesystem.TreeNode, opts Options, showIcons bool) {
	isUnicode := ctx.Caps.UnicodeSupported && !opts.Plain
	profile := ctx.Caps.ColorProfile
	if opts.Plain {
		profile = terminal.ColorNone
	}

	// Print root header
	rootName := root.Entry.Name
	if root.Entry.IsDir {
		rootName += "/"
	}
	if showIcons && profile != terminal.ColorNone {
		icon := theme.LookupIcon(root.Entry.Name, root.Entry.EntityType, ctx.Caps.UnicodeSupported)
		rootName = icon + rootName
	}
	ctx.Printer.Println(ctx.Theme.Format(theme.RoleDirectory, rootName, profile))

	// Render descendants
	renderNodeChildren(ctx, root.Children, "", isUnicode, profile, opts, showIcons)

	// Summary footer
	var summaryParts []string
	summaryParts = append(summaryParts, fmt.Sprintf("%d directories", root.DirCount))
	if !opts.DirsOnly {
		summaryParts = append(summaryParts, fmt.Sprintf("%d files", root.FileCount))
	}
	if opts.Sizes {
		if opts.HumanReadable {
			summaryParts = append(summaryParts, fmt.Sprintf("total %s", formatSize(root.TotalSize, true)))
		} else {
			summaryParts = append(summaryParts, fmt.Sprintf("total %d bytes", root.TotalSize))
		}
	}

	ctx.Printer.Println()
	footer := strings.Join(summaryParts, ", ")
	ctx.Printer.Println(ctx.Theme.Format(theme.RoleMuted, footer, profile))
}

func renderNodeChildren(ctx *command.Context, children []*filesystem.TreeNode, prefix string, isUnicode bool, profile terminal.ColorProfile, opts Options, showIcons bool) {
	n := len(children)
	for i, child := range children {
		isLast := (i == n-1)

		// Glyphs
		var branch, nextPrefix string
		if isUnicode {
			if isLast {
				branch = "└── "
				nextPrefix = prefix + "    "
			} else {
				branch = "├── "
				nextPrefix = prefix + "│   "
			}
		} else {
			if isLast {
				branch = "\\-- "
				nextPrefix = prefix + "    "
			} else {
				branch = "|-- "
				nextPrefix = prefix + "|   "
			}
		}

		branchStyled := ctx.Theme.Format(theme.RoleMuted, branch, profile)
		prefixStyled := ctx.Theme.Format(theme.RoleMuted, prefix, profile)

		// Extra metadata items
		var metaParts []string
		if opts.Permissions {
			permStr := child.Entry.Permissions
			if profile != terminal.ColorNone {
				permStr = formatPermsStyled(child.Entry.Permissions, ctx.Theme, profile)
			}
			metaParts = append(metaParts, fmt.Sprintf("[%s]", permStr))
		}
		if opts.Sizes {
			sizeStr := formatSize(child.Entry.Size, opts.HumanReadable)
			if profile != terminal.ColorNone {
				sizeStr = ctx.Theme.Format(theme.RoleSize, sizeStr, profile)
			}
			metaParts = append(metaParts, fmt.Sprintf("[%s]", sizeStr))
		}

		// Icon
		iconPart := ""
		if showIcons && profile != terminal.ColorNone {
			icon := theme.LookupIcon(child.Entry.Name, child.Entry.EntityType, ctx.Caps.UnicodeSupported)
			iconPart = icon
		}

		// Name styling
		role := theme.RoleRegularFile
		name := child.Entry.Name
		if child.Entry.IsDir {
			role = theme.RoleDirectory
			name += "/"
		} else if child.Entry.IsBroken {
			role = theme.RoleBrokenSymlink
		} else if child.Entry.IsSymlink {
			role = theme.RoleSymlink
		} else if child.Entry.EntityType == theme.TypeExecutable {
			role = theme.RoleExecutable
		}

		nameStyled := ctx.Theme.Format(role, name, profile)

		// Symlink target info
		if child.Entry.IsSymlink {
			arrow := " -> "
			target := child.Entry.LinkTarget
			if child.Entry.IsBroken {
				target += " [broken link]"
			}
			nameStyled += ctx.Theme.Format(theme.RoleMuted, arrow, profile) + ctx.Theme.Format(theme.RoleSymlink, target, profile)
		}

		if child.IsCycle {
			nameStyled += ctx.Theme.Format(theme.RoleWarning, " [cycle detected]", profile)
		}

		var lineBuilder strings.Builder
		lineBuilder.WriteString(prefixStyled)
		lineBuilder.WriteString(branchStyled)
		for _, m := range metaParts {
			lineBuilder.WriteString(m + " ")
		}
		lineBuilder.WriteString(iconPart)
		lineBuilder.WriteString(nameStyled)

		ctx.Printer.Println(lineBuilder.String())

		// Recurse into children
		if len(child.Children) > 0 {
			renderNodeChildren(ctx, child.Children, nextPrefix, isUnicode, profile, opts, showIcons)
		}
	}
}

// RenderJSON serializes the tree structure to JSON.
func RenderJSON(ctx *command.Context, root *filesystem.TreeNode) error {
	data, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}
	ctx.Printer.Println(string(data))
	return nil
}

func formatSize(bytes int64, human bool) string {
	return renderer.FormatSize(bytes, human)
}

func formatPermsStyled(perms string, th *theme.Theme, profile terminal.ColorProfile) string {
	if len(perms) < 10 {
		return perms
	}
	var b strings.Builder
	b.WriteString(th.Format(theme.RoleMuted, string(perms[0]), profile))
	for i := 1; i < 10; i++ {
		c := perms[i]
		switch c {
		case 'r':
			b.WriteString(th.Format(theme.RolePermRead, "r", profile))
		case 'w':
			b.WriteString(th.Format(theme.RolePermWrite, "w", profile))
		case 'x', 's', 'S', 't', 'T':
			b.WriteString(th.Format(theme.RolePermExec, string(c), profile))
		default:
			b.WriteString(th.Format(theme.RoleMuted, "-", profile))
		}
	}
	return b.String()
}
