package notecmd

import (
	"fmt"
	"io"
	"strings"

	"nova/internal/command"
	"nova/internal/renderer"
	"nova/internal/theme"
)

// RenderNotesPlain outputs notes as tab-separated values for pipelines.
func RenderNotesPlain(w io.Writer, notes []*Note) {
	fmt.Fprintf(w, "ID\tTITLE\tTAGS\tDONE\tTOTAL\tUPDATED\n")
	for _, n := range notes {
		tagsStr := strings.Join(n.Tags, ",")
		updatedStr := n.Updated.Format("2006-01-02 15:04")
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%s\n", n.ID, n.Title, tagsStr, n.DoneCount, n.TotalTodos, updatedStr)
	}
}

// RenderNotesDashboard prints colorful note cards and summary.
func RenderNotesDashboard(w io.Writer, notes []*Note, ctx *command.Context) {
	th := ctx.Theme
	prof := ctx.Caps.ColorProfile
	width := ctx.Caps.Width
	if width < 70 {
		width = 70
	}
	if width > 100 {
		width = 100
	}

	header := fmt.Sprintf("╭ 📝 Developer Notes & Scratchpad %s╮", strings.Repeat("─", width-34))
	fmt.Fprintln(w, th.Format(theme.RoleAccent, header, prof))

	if len(notes) == 0 {
		fmt.Fprintln(w, th.Format(theme.RoleMuted, "  (No notes yet — create one with 'nova note add \"Title\" -c \"Content\"')", prof))
		bot := fmt.Sprintf("╰%s╯", strings.Repeat("─", width-2))
		fmt.Fprintln(w, th.Format(theme.RoleAccent, bot, prof))
		return
	}

	tableHeader := fmt.Sprintf("  %-16s %-24s %-16s %-16s %s", "ID", "TITLE", "TAGS", "TASKS", "UPDATED")
	fmt.Fprintln(w, th.Format(theme.RoleMuted, tableHeader, prof))
	fmt.Fprintln(w, th.Format(theme.RoleMuted, "  "+strings.Repeat("─", width-4), prof))

	totalTodos := 0
	totalDone := 0

	for _, n := range notes {
		totalTodos += n.TotalTodos
		totalDone += n.DoneCount

		idStr := renderer.Truncate(n.ID, 15, "…")
		titleStr := renderer.Truncate(n.Title, 23, "…")

		var tagList []string
		for _, tag := range n.Tags {
			tagList = append(tagList, "#"+tag)
		}
		tagsStr := renderer.Truncate(strings.Join(tagList, " "), 15, "…")
		if tagsStr == "" {
			tagsStr = "-"
		}

		taskStr := "-"
		if n.TotalTodos > 0 {
			pct := float64(n.DoneCount) / float64(n.TotalTodos) * 100.0
			taskStr = fmt.Sprintf("%d/%d (%.0f%%)", n.DoneCount, n.TotalTodos, pct)
		}

		relTime := renderer.FormatRelativeTime(n.Updated)

		row := fmt.Sprintf("  %-16s %-24s %-16s %-16s %s",
			th.Format(theme.RoleAccent, idStr, prof),
			th.Format(theme.RoleDocument, titleStr, prof),
			th.Format(theme.RoleInfo, tagsStr, prof),
			th.Format(theme.RoleSuccess, taskStr, prof),
			th.Format(theme.RoleDate, relTime, prof),
		)
		fmt.Fprintln(w, row)
	}

	fmt.Fprintln(w, th.Format(theme.RoleMuted, "  "+strings.Repeat("─", width-4), prof))

	summary := fmt.Sprintf("  Total Notes: %d  │  Tasks: %d/%d completed", len(notes), totalDone, totalTodos)
	if totalTodos > 0 {
		pct := float64(totalDone) / float64(totalTodos) * 100.0
		bar := renderBar(totalDone, totalTodos, 12)
		summary += fmt.Sprintf("  │  %s %.1f%%", bar, pct)
	}
	fmt.Fprintln(w, th.Format(theme.RoleMuted, summary, prof))

	bot := fmt.Sprintf("╰%s╯", strings.Repeat("─", width-2))
	fmt.Fprintln(w, th.Format(theme.RoleAccent, bot, prof))
}

// RenderNoteAdded prints visual confirmation when a note is created.
func RenderNoteAdded(w io.Writer, note *Note, ctx *command.Context) {
	th := ctx.Theme
	prof := ctx.Caps.ColorProfile

	fmt.Fprintln(w, th.Format(theme.RoleSuccess, fmt.Sprintf("✔ Note created successfully: %s", note.Title), prof))
	fmt.Fprintf(w, "  ID:      %s\n", th.Format(theme.RoleAccent, note.ID, prof))
	if len(note.Tags) > 0 {
		fmt.Fprintf(w, "  Tags:    %s\n", th.Format(theme.RoleInfo, strings.Join(note.Tags, ", "), prof))
	}
	if note.TotalTodos > 0 {
		fmt.Fprintf(w, "  Tasks:   %d item(s) detected\n", note.TotalTodos)
	}
	fmt.Fprintln(w, th.Format(theme.RoleMuted, "  View:    nova note show "+note.ID, prof))
}

// RenderNoteDetail prints a full formatted view of a single note.
func RenderNoteDetail(w io.Writer, note *Note, ctx *command.Context) {
	th := ctx.Theme
	prof := ctx.Caps.ColorProfile
	width := ctx.Caps.Width
	if width < 70 {
		width = 70
	}
	if width > 100 {
		width = 100
	}

	header := fmt.Sprintf("╭ 📝 Note: %s %s╮", note.Title, strings.Repeat("─", width-11-len(note.Title)))
	fmt.Fprintln(w, th.Format(theme.RoleAccent, renderer.Truncate(header, width, "─╮"), prof))

	fmt.Fprintf(w, "  ID:       %s\n", th.Format(theme.RoleAccent, note.ID, prof))
	if len(note.Tags) > 0 {
		fmt.Fprintf(w, "  Tags:     %s\n", th.Format(theme.RoleInfo, strings.Join(note.Tags, ", "), prof))
	}
	relCreated := renderer.FormatRelativeTime(note.Created)
	relUpdated := renderer.FormatRelativeTime(note.Updated)
	fmt.Fprintf(w, "  Created:  %s  │  Updated: %s\n", relCreated, relUpdated)

	if note.TotalTodos > 0 {
		pct := float64(note.DoneCount) / float64(note.TotalTodos) * 100.0
		bar := renderBar(note.DoneCount, note.TotalTodos, 16)
		fmt.Fprintf(w, "  Progress: %s %d/%d (%.1f%%)\n", bar, note.DoneCount, note.TotalTodos, pct)
	}

	fmt.Fprintln(w, th.Format(theme.RoleMuted, "  "+strings.Repeat("─", width-4), prof))
	fmt.Fprintln(w, "  Content:")

	if strings.TrimSpace(note.Content) == "" {
		fmt.Fprintln(w, th.Format(theme.RoleMuted, "    (empty content)", prof))
	} else {
		for _, line := range strings.Split(note.Content, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "* [x]") {
				fmt.Fprintln(w, "    "+th.Format(theme.RoleSuccess, "[✓] "+trimmed[5:], prof))
			} else if strings.HasPrefix(trimmed, "- [ ]") || strings.HasPrefix(trimmed, "* [ ]") {
				fmt.Fprintln(w, "    "+th.Format(theme.RoleWarning, "[ ] "+trimmed[5:], prof))
			} else {
				fmt.Fprintln(w, "    "+line)
			}
		}
	}

	bot := fmt.Sprintf("╰%s╯", strings.Repeat("─", width-2))
	fmt.Fprintln(w, th.Format(theme.RoleAccent, bot, prof))
}

// RenderTodoDashboard displays a consolidated view of all tasks across notes.
func RenderTodoDashboard(w io.Writer, notes []*Note, ctx *command.Context) {
	th := ctx.Theme
	prof := ctx.Caps.ColorProfile
	width := ctx.Caps.Width
	if width < 70 {
		width = 70
	}
	if width > 100 {
		width = 100
	}

	header := fmt.Sprintf("╭ 📋 TODO Checklist Tracker %s╮", strings.Repeat("─", width-28))
	fmt.Fprintln(w, th.Format(theme.RoleAccent, header, prof))

	totalTodos := 0
	totalDone := 0
	for _, n := range notes {
		totalTodos += n.TotalTodos
		totalDone += n.DoneCount
	}

	if totalTodos == 0 {
		fmt.Fprintln(w, th.Format(theme.RoleMuted, "  (No TODO items found across any notes)", prof))
		bot := fmt.Sprintf("╰%s╯", strings.Repeat("─", width-2))
		fmt.Fprintln(w, th.Format(theme.RoleAccent, bot, prof))
		return
	}

	pct := float64(totalDone) / float64(totalTodos) * 100.0
	bar := renderBar(totalDone, totalTodos, 16)
	meta := fmt.Sprintf("  Overall: %s %d/%d completed (%.1f%%)", bar, totalDone, totalTodos, pct)
	fmt.Fprintln(w, meta)
	fmt.Fprintln(w, th.Format(theme.RoleMuted, "  "+strings.Repeat("─", width-4), prof))

	for _, n := range notes {
		if len(n.Todos) == 0 {
			continue
		}

		noteHeader := fmt.Sprintf("  📄 %s (%s):", n.Title, n.ID)
		fmt.Fprintln(w, th.Format(theme.RoleDocument, noteHeader, prof))

		for _, t := range n.Todos {
			if t.Done {
				icon := th.Format(theme.RoleSuccess, "[✓]", prof)
				taskLine := fmt.Sprintf("    %s #%d %s", icon, t.Index, t.Text)
				fmt.Fprintln(w, th.Format(theme.RoleMuted, taskLine, prof))
			} else {
				icon := th.Format(theme.RoleWarning, "[ ]", prof)
				taskLine := fmt.Sprintf("    %s #%d %s", icon, t.Index, t.Text)
				fmt.Fprintln(w, taskLine)
			}
		}
		fmt.Fprintln(w, "")
	}

	tip := "  Tip: Use 'nova note toggle <id> <task#>' to check/uncheck items."
	fmt.Fprintln(w, th.Format(theme.RoleMuted, tip, prof))

	bot := fmt.Sprintf("╰%s╯", strings.Repeat("─", width-2))
	fmt.Fprintln(w, th.Format(theme.RoleAccent, bot, prof))
}

func renderBar(done, total, barWidth int) string {
	if total == 0 {
		return "[" + strings.Repeat("░", barWidth) + "]"
	}
	filled := int(float64(done) / float64(total) * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	if filled < 0 {
		filled = 0
	}
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled) + "]"
}
