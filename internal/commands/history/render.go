package historycmd

import (
	"fmt"
	"io"
	"math"
	"strings"

	"nova/internal/command"
	"nova/internal/renderer"
	"nova/internal/theme"
)

// RenderDashboard renders the history intelligence dashboard.
func RenderDashboard(w io.Writer, rep *Report, ctx *command.Context) {
	th := ctx.Theme
	if th == nil {
		th = theme.Get("default")
	}
	profile := ctx.Caps.ColorProfile
	termWidth := ctx.Caps.Width
	if termWidth <= 0 {
		termWidth = 80
	}

	box := renderer.RoundedBoxStyle

	title := " 📜 Shell Command History Analytics "
	topBorder := th.Format(theme.RoleAccent, box.TopLeft+title, profile) +
		th.Format(theme.RoleMuted, strings.Repeat(box.Horizontal, max(0, termWidth-renderer.VisibleWidth(title)-2))+box.TopRight, profile)
	fmt.Fprintln(w, topBorder)

	// Summary line
	fileDisp := rep.HistoryFile
	if len(fileDisp) > 28 {
		fileDisp = "…" + fileDisp[len(fileDisp)-27:]
	}
	summary := fmt.Sprintf("  %s %s  │  %s %s  │  %s %s",
		th.Format(theme.RoleMuted, "File:", profile),
		th.Format(theme.RoleDate, fileDisp, profile),
		th.Format(theme.RoleMuted, "Total:", profile),
		th.Format(theme.RoleAccent, fmt.Sprintf("%d", rep.TotalCommands), profile),
		th.Format(theme.RoleMuted, "Unique:", profile),
		th.Format(theme.RoleSuccess, fmt.Sprintf("%d", rep.UniqueCommands), profile))
	fmt.Fprintln(w, summary)

	sep := th.Format(theme.RoleMuted, "  "+strings.Repeat("─", max(10, termWidth-4)), profile)
	fmt.Fprintln(w, sep)

	if len(rep.TopCommands) == 0 {
		fmt.Fprintln(w, "  "+th.Format(theme.RoleMuted, "No matching commands found in history", profile))
	} else {
		// Table header
		header := fmt.Sprintf("  %-5s %-14s %-12s %-20s %s",
			th.Format(theme.RoleMuted, "RANK", profile),
			th.Format(theme.RoleMuted, "COMMAND", profile),
			th.Format(theme.RoleMuted, "CATEGORY", profile),
			th.Format(theme.RoleMuted, "DISTRIBUTION", profile),
			th.Format(theme.RoleMuted, "COUNT", profile))
		fmt.Fprintln(w, header)

		barWidth := 16

		for idx, item := range rep.TopCommands {
			rankStr := fmt.Sprintf("#%d", idx+1)
			filled := int(math.Round((item.Percentage / 100.0) * float64(barWidth)))
			if filled < 1 && item.Percentage > 0 {
				filled = 1
			}
			if filled > barWidth {
				filled = barWidth
			}
			empty := barWidth - filled

			barFilled := th.Format(theme.RoleSuccess, strings.Repeat("█", filled), profile)
			barEmpty := th.Format(theme.RoleMuted, strings.Repeat("░", empty), profile)
			barChart := barFilled + barEmpty

			catBadge := fmt.Sprintf("[%s]", item.Category)
			countStr := fmt.Sprintf("%d (%.1f%%)", item.Count, item.Percentage)

			row := fmt.Sprintf("  %-5s %-14s %-12s %s  %s",
				th.Format(theme.RoleMuted, rankStr, profile),
				th.Format(theme.RoleAccent, item.Command, profile),
				th.Format(theme.RoleInfo, catBadge, profile),
				barChart,
				th.Format(theme.RoleDate, countStr, profile))
			fmt.Fprintln(w, row)
		}
	}

	if len(rep.Categories) > 0 {
		fmt.Fprintln(w, sep)
		var catParts []string
		for _, c := range rep.Categories {
			catParts = append(catParts, fmt.Sprintf("%s: %.1f%%",
				th.Format(theme.RoleMuted, c.Category, profile),
				c.Percentage))
		}
		catLine := "  " + th.Format(theme.RoleAccent, "Categories: ", profile) + strings.Join(catParts, "  ")
		fmt.Fprintln(w, catLine)
	}

	botBorder := th.Format(theme.RoleMuted, box.BottomLeft+strings.Repeat(box.Horizontal, max(0, termWidth-2))+box.BottomRight, profile)
	fmt.Fprintln(w, botBorder)
}

// RenderPlain prints tab-delimited history analytics.
func RenderPlain(w io.Writer, rep *Report) {
	fmt.Fprintln(w, "RANK\tCOMMAND\tCATEGORY\tCOUNT\tPERCENTAGE")
	for idx, item := range rep.TopCommands {
		fmt.Fprintf(w, "%d\t%s\t%s\t%d\t%.2f%%\n", idx+1, item.Command, item.Category, item.Count, item.Percentage)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
