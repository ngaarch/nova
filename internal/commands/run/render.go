package runcmd

import (
	"fmt"
	"io"
	"strings"
	"time"

	"nova/internal/command"
	"nova/internal/renderer"
	"nova/internal/theme"
)

// RenderTasksDashboard renders a card list of all discovered tasks.
func RenderTasksDashboard(w io.Writer, tasks []Task, ctx *command.Context) {
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

	title := fmt.Sprintf(" 🛠 Project Tasks & Scripts (%d detected) ", len(tasks))
	topBorder := th.Format(theme.RoleAccent, box.TopLeft+title, profile) +
		th.Format(theme.RoleMuted, strings.Repeat(box.Horizontal, max(0, termWidth-renderer.VisibleWidth(title)-2))+box.TopRight, profile)
	fmt.Fprintln(w, topBorder)

	if len(tasks) == 0 {
		fmt.Fprintln(w, "  "+th.Format(theme.RoleMuted, "No tasks detected in workspace (package.json, Makefile, go.mod, etc.)", profile))
	} else {
		// Header
		header := fmt.Sprintf("  %-16s %-8s %-24s %s",
			th.Format(theme.RoleMuted, "TASK", profile),
			th.Format(theme.RoleMuted, "SOURCE", profile),
			th.Format(theme.RoleMuted, "COMMAND", profile),
			th.Format(theme.RoleMuted, "DESCRIPTION", profile))
		fmt.Fprintln(w, header)
		sep := th.Format(theme.RoleMuted, "  "+strings.Repeat("─", max(10, termWidth-4)), profile)
		fmt.Fprintln(w, sep)

		for _, t := range tasks {
			cmdStr := t.Command
			if len(cmdStr) > 22 {
				cmdStr = cmdStr[:21] + "…"
			}

			descStr := t.Description
			descWidth := termWidth - 54
			if descWidth > 10 && len(descStr) > descWidth {
				descStr = descStr[:descWidth-1] + "…"
			}

			sourceBadge := fmt.Sprintf("[%s]", t.Runtime)

			row := fmt.Sprintf("  %-16s %-8s %-24s %s",
				th.Format(theme.RoleSuccess, t.Name, profile),
				th.Format(theme.RoleInfo, sourceBadge, profile),
				th.Format(theme.RoleCode, cmdStr, profile),
				th.Format(theme.RoleMuted, descStr, profile))
			fmt.Fprintln(w, row)
		}
	}

	fmt.Fprintln(w, "  "+th.Format(theme.RoleDate, "Tip: Run a task using 'nova run <task-name>'", profile))

	botBorder := th.Format(theme.RoleMuted, box.BottomLeft+strings.Repeat(box.Horizontal, max(0, termWidth-2))+box.BottomRight, profile)
	fmt.Fprintln(w, botBorder)
}

// RenderTasksPlain renders tab-delimited tasks.
func RenderTasksPlain(w io.Writer, tasks []Task) {
	fmt.Fprintln(w, "NAME\tRUNTIME\tSOURCE\tCOMMAND\tDESCRIPTION")
	for _, t := range tasks {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", t.Name, t.Runtime, t.Source, t.Command, t.Description)
	}
}

// RenderTaskStartBanner prints a task launch banner.
func RenderTaskStartBanner(w io.Writer, task *Task, extraArgs []string, ctx *command.Context) {
	th := ctx.Theme
	if th == nil {
		th = theme.Get("default")
	}
	profile := ctx.Caps.ColorProfile

	cmdDisplay := task.Command
	if len(extraArgs) > 0 {
		cmdDisplay += " " + strings.Join(extraArgs, " ")
	}

	fmt.Fprintf(w, "%s %s %s\n",
		th.Format(theme.RoleAccent, "▶ [nova run]", profile),
		th.Format(theme.RoleSuccess, task.Name, profile),
		th.Format(theme.RoleMuted, "("+cmdDisplay+")", profile))
}

// RenderTaskEndBanner prints task completion outcome.
func RenderTaskEndBanner(w io.Writer, res *ExecutionResult, ctx *command.Context) {
	th := ctx.Theme
	if th == nil {
		th = theme.Get("default")
	}
	profile := ctx.Caps.ColorProfile

	durStr := res.Duration.Round(time.Millisecond).String()

	if res.Success {
		fmt.Fprintf(w, "%s %s %s\n",
			th.Format(theme.RoleSuccess, "✔", profile),
			th.Format(theme.RoleSuccess, fmt.Sprintf("Task '%s' completed successfully", res.Task.Name), profile),
			th.Format(theme.RoleDate, fmt.Sprintf("(%s)", durStr), profile))
	} else {
		fmt.Fprintf(w, "%s %s %s\n",
			th.Format(theme.RoleError, "✖", profile),
			th.Format(theme.RoleError, fmt.Sprintf("Task '%s' failed (exit code %d)", res.Task.Name, res.ExitCode), profile),
			th.Format(theme.RoleDate, fmt.Sprintf("(%s)", durStr), profile))
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
