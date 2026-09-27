package clean

import (
	"encoding/json"
	"fmt"
	"strings"

	"nova/internal/command"
	"nova/internal/renderer"
	"nova/internal/theme"
)

// RenderJSON serializes cleanup results to JSON.
func RenderJSON(ctx *command.Context, res Results) error {
	data, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	ctx.Printer.Println(string(data))
	return nil
}

// RenderPlain outputs machine-parsable cleanup records.
func RenderPlain(ctx *command.Context, res Results) {
	for _, it := range res.Items {
		status := "candidate"
		if it.Deleted {
			status = "deleted"
		} else if it.DeleteError != "" {
			status = "error"
		}
		ctx.Printer.Printf("%s\t%s\t%d\t%s\n", it.Path, it.Category, it.Size, status)
	}
}

// RenderHuman renders an elegant framed table with summary badges.
func RenderHuman(ctx *command.Context, res Results) {
	th := ctx.Theme
	prof := ctx.Caps.ColorProfile
	unicode := ctx.Caps.UnicodeSupported
	w := ctx.Caps.Width
	if w <= 0 || w > 100 {
		w = 80
	}

	borderStyle := th.Style(theme.RoleAccent)

	if len(res.Items) == 0 {
		cleanMsg := []string{
			" " + th.Format(theme.RoleSuccess, "✔ Workspace is completely clean! No junk files detected.", prof),
			fmt.Sprintf("   Scanned %d directories with 0 disposable candidates.", res.Summary.ScannedDirs),
		}
		card := renderer.RenderCard("✨ Workspace Clean", cleanMsg, w, borderStyle, unicode, prof)
		for _, l := range card {
			ctx.Printer.Println(l)
		}
		return
	}

	var lines []string
	headerStr := fmt.Sprintf(" %-40s %-16s %10s",
		th.Format(theme.RoleMuted, "PATH", prof),
		th.Format(theme.RoleMuted, "CATEGORY", prof),
		th.Format(theme.RoleMuted, "SIZE", prof),
	)
	lines = append(lines, headerStr)
	lines = append(lines, strings.Repeat("─", w-4))

	maxShow := 15
	for i, it := range res.Items {
		if i >= maxShow {
			remaining := len(res.Items) - maxShow
			lines = append(lines, fmt.Sprintf(" … and %d more disposable items", remaining))
			break
		}

		sizeStr := renderer.FormatSize(it.Size, true)
		if it.IsDir {
			sizeStr = "<empty dir>"
		} else if it.IsBroken {
			sizeStr = "<broken>"
		}

		truncPath := renderer.Truncate(it.Path, 40, "…")
		role := theme.RoleWarning
		if it.Deleted {
			role = theme.RoleSuccess
		} else if it.DeleteError != "" {
			role = theme.RoleError
		}

		line := fmt.Sprintf(" %-40s %-16s %10s",
			th.Format(theme.RoleInfo, truncPath, prof),
			th.Format(role, it.Category, prof),
			th.Format(theme.RoleMuted, sizeStr, prof),
		)
		lines = append(lines, line)
	}

	lines = append(lines, strings.Repeat("─", w-4))

	reclaimedStr := renderer.FormatSize(res.Summary.ReclaimedBytes, true)
	if res.Summary.IsDryRun {
		warningBanner := fmt.Sprintf(" [DRY-RUN] Found %s items (%s reclaimable). Run with -f to delete.",
			th.Format(theme.RoleAccent, fmt.Sprintf("%d", res.Summary.TotalCandidates), prof),
			th.Format(theme.RoleWarning, reclaimedStr, prof),
		)
		lines = append(lines, warningBanner)
	} else {
		successBanner := fmt.Sprintf(" ✔ Deleted %s items. Reclaimed %s of storage space!",
			th.Format(theme.RoleSuccess, fmt.Sprintf("%d", res.Summary.TotalDeleted), prof),
			th.Format(theme.RoleSuccess, reclaimedStr, prof),
		)
		lines = append(lines, successBanner)
	}

	card := renderer.RenderCard("🧹 Workspace Cleanup", lines, w, borderStyle, unicode, prof)
	for _, l := range card {
		ctx.Printer.Println(l)
	}
}
