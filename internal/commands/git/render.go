package gitcmd

import (
	"fmt"
	"strings"

	"nova/internal/command"
	"nova/internal/output"
	"nova/internal/theme"
)

// RenderStatus outputs the Git status summary.
func RenderStatus(ctx *command.Context, data StatusData, _ Options) error {
	if ctx.Printer.Mode == output.ModeJSON {
		return ctx.Printer.PrintJSON(data)
	}

	if ctx.Printer.Mode == output.ModePlain {
		var lines []string
		lines = append(lines, fmt.Sprintf("branch: %s", data.Branch))
		lines = append(lines, fmt.Sprintf("ahead: %d, behind: %d", data.Ahead, data.Behind))
		for _, f := range data.Staged {
			lines = append(lines, fmt.Sprintf("staged: %s %s", f.Status, f.Path))
		}
		for _, f := range data.Unstaged {
			lines = append(lines, fmt.Sprintf("unstaged: %s %s", f.Status, f.Path))
		}
		for _, f := range data.Untracked {
			lines = append(lines, fmt.Sprintf("untracked: %s", f.Path))
		}
		if len(lines) > 0 {
			ctx.Printer.Print(strings.Join(lines, "\n") + "\n")
		}
		return nil
	}

	// Human mode
	prof := ctx.Caps.ColorProfile
	th := ctx.Theme

	title := "🌿 Git Repository Dashboard"
	header := fmt.Sprintf("╭ %s %s╮\n", title, strings.Repeat("─", max(2, 70-len(title))))
	var body strings.Builder
	body.WriteString(header)

	branchBadge := th.Format(theme.RoleAccent, data.Branch, prof)
	trackingStr := ""
	if data.Ahead > 0 || data.Behind > 0 {
		trackingStr = fmt.Sprintf(" [%s, %s]",
			th.Format(theme.RoleSuccess, fmt.Sprintf("ahead %d", data.Ahead), prof),
			th.Format(theme.RoleWarning, fmt.Sprintf("behind %d", data.Behind), prof),
		)
	}
	body.WriteString(fmt.Sprintf("│  Branch: %s%s\n", branchBadge, trackingStr))

	if data.Clean {
		body.WriteString(fmt.Sprintf("│ %s\n", strings.Repeat("─", 74)))
		cleanBadge := th.Format(theme.RoleSuccess, "✔ Working tree clean. Zero pending changes.", prof)
		body.WriteString(fmt.Sprintf("│  %s\n", cleanBadge))
	} else {
		if len(data.Staged) > 0 {
			body.WriteString(fmt.Sprintf("│ %s\n", strings.Repeat("─", 74)))
			stagedHeader := th.Format(theme.RoleSuccess, fmt.Sprintf("  [ Staged Changes (%d) ]", len(data.Staged)), prof)
			body.WriteString(fmt.Sprintf("│%s\n", stagedHeader))
			for _, f := range data.Staged {
				badge := th.Format(theme.RoleSuccess, "+", prof)
				body.WriteString(fmt.Sprintf("│    %s  %-10s %s\n", badge, f.Status, f.Path))
			}
		}

		if len(data.Unstaged) > 0 {
			body.WriteString(fmt.Sprintf("│ %s\n", strings.Repeat("─", 74)))
			unstagedHeader := th.Format(theme.RoleWarning, fmt.Sprintf("  [ Unstaged Changes (%d) ]", len(data.Unstaged)), prof)
			body.WriteString(fmt.Sprintf("│%s\n", unstagedHeader))
			for _, f := range data.Unstaged {
				badge := th.Format(theme.RoleWarning, "~", prof)
				body.WriteString(fmt.Sprintf("│    %s  %-10s %s\n", badge, f.Status, f.Path))
			}
		}

		if len(data.Untracked) > 0 {
			body.WriteString(fmt.Sprintf("│ %s\n", strings.Repeat("─", 74)))
			untrackedHeader := th.Format(theme.RoleInfo, fmt.Sprintf("  [ Untracked Files (%d) ]", len(data.Untracked)), prof)
			body.WriteString(fmt.Sprintf("│%s\n", untrackedHeader))
			for _, f := range data.Untracked {
				badge := th.Format(theme.RoleInfo, "?", prof)
				body.WriteString(fmt.Sprintf("│    %s  %s\n", badge, f.Path))
			}
		}
	}

	body.WriteString(fmt.Sprintf("│ %s\n", strings.Repeat("─", 74)))
	summaryText := fmt.Sprintf("Total modified items: %d", data.TotalFiles)
	body.WriteString(fmt.Sprintf("│  %s\n", th.Format(theme.RoleMuted, summaryText, prof)))
	body.WriteString(fmt.Sprintf("╰%s╯\n", strings.Repeat("─", 76)))

	ctx.Printer.Print(body.String())
	return nil
}

// RenderLog outputs the commit log.
func RenderLog(ctx *command.Context, commits []CommitItem, _ Options) error {
	if ctx.Printer.Mode == output.ModeJSON {
		return ctx.Printer.PrintJSON(commits)
	}

	if ctx.Printer.Mode == output.ModePlain {
		var lines []string
		for _, c := range commits {
			lines = append(lines, fmt.Sprintf("%s %s (%s) %s", c.Hash, c.RelTime, c.Author, c.Message))
		}
		if len(lines) > 0 {
			ctx.Printer.Print(strings.Join(lines, "\n") + "\n")
		}
		return nil
	}

	// Human mode
	prof := ctx.Caps.ColorProfile
	th := ctx.Theme

	title := fmt.Sprintf("📜 Git Recent Commits (%d)", len(commits))
	header := fmt.Sprintf("╭ %s %s╮\n", title, strings.Repeat("─", max(2, 70-len(title))))
	var body strings.Builder
	body.WriteString(header)

	for _, c := range commits {
		hashFormatted := th.Format(theme.RoleAccent, c.Hash, prof)
		timeFormatted := th.Format(theme.RoleMuted, fmt.Sprintf("%-12s", c.RelTime), prof)
		authorFormatted := th.Format(theme.RoleInfo, fmt.Sprintf("<%s>", c.Author), prof)
		refStr := ""
		if c.Ref != "" {
			refStr = " " + th.Format(theme.RoleWarning, c.Ref, prof)
		}

		body.WriteString(fmt.Sprintf("│  * %s %s %s%s\n", hashFormatted, timeFormatted, c.Message, refStr))
		body.WriteString(fmt.Sprintf("│    %s\n", authorFormatted))
	}

	body.WriteString(fmt.Sprintf("╰%s╯\n", strings.Repeat("─", 76)))
	ctx.Printer.Print(body.String())
	return nil
}

// RenderBranch outputs the branch list.
func RenderBranch(ctx *command.Context, branches []BranchItem, _ Options) error {
	if ctx.Printer.Mode == output.ModeJSON {
		return ctx.Printer.PrintJSON(branches)
	}

	if ctx.Printer.Mode == output.ModePlain {
		var lines []string
		for _, b := range branches {
			prefix := "  "
			if b.IsCurrent {
				prefix = "* "
			}
			lines = append(lines, prefix+b.Name)
		}
		if len(lines) > 0 {
			ctx.Printer.Print(strings.Join(lines, "\n") + "\n")
		}
		return nil
	}

	// Human mode
	prof := ctx.Caps.ColorProfile
	th := ctx.Theme

	title := "🌿 Git Branches"
	header := fmt.Sprintf("╭ %s %s╮\n", title, strings.Repeat("─", max(2, 70-len(title))))
	var body strings.Builder
	body.WriteString(header)

	for _, b := range branches {
		if b.IsCurrent {
			currentName := th.Format(theme.RoleAccent, "* "+b.Name+" (active)", prof)
			body.WriteString(fmt.Sprintf("│  %s\n", currentName))
		} else if b.IsRemote {
			remoteName := th.Format(theme.RoleMuted, "  "+b.Name, prof)
			body.WriteString(fmt.Sprintf("│  %s\n", remoteName))
		} else {
			body.WriteString(fmt.Sprintf("│    %s\n", b.Name))
		}
	}

	body.WriteString(fmt.Sprintf("╰%s╯\n", strings.Repeat("─", 76)))
	ctx.Printer.Print(body.String())
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
