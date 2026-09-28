package calccmd

import (
	"fmt"
	"io"
	"strings"

	"nova/internal/command"
	"nova/internal/renderer"
	"nova/internal/theme"
)

// RenderDashboard renders a card with the expression, result, and multi-base breakdown.
func RenderDashboard(w io.Writer, res *CalcResult, ctx *command.Context) {
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

	title := " 🧮 Calculator Evaluation "
	topBorder := th.Format(theme.RoleAccent, box.TopLeft+title, profile) +
		th.Format(theme.RoleMuted, strings.Repeat(box.Horizontal, max(0, termWidth-renderer.VisibleWidth(title)-2))+box.TopRight, profile)
	fmt.Fprintln(w, topBorder)

	fmt.Fprintf(w, "  %s  %s\n",
		th.Format(theme.RoleMuted, "Expression: ", profile),
		th.Format(theme.RoleRegularFile, res.Expression, profile))

	fmt.Fprintf(w, "  %s  %s\n",
		th.Format(theme.RoleMuted, "Result:     ", profile),
		th.Format(theme.RoleSuccess, res.Formatted, profile))

	if res.IsInteger {
		sep := th.Format(theme.RoleMuted, "  "+strings.Repeat("─", max(10, termWidth-4)), profile)
		fmt.Fprintln(w, sep)

		fmt.Fprintf(w, "  %s  %s\n",
			th.Format(theme.RoleMuted, "Hex:        ", profile),
			th.Format(theme.RoleAccent, res.Hex, profile))
		fmt.Fprintf(w, "  %s  %s\n",
			th.Format(theme.RoleMuted, "Binary:     ", profile),
			th.Format(theme.RoleInfo, res.Binary, profile))
		fmt.Fprintf(w, "  %s  %s\n",
			th.Format(theme.RoleMuted, "Octal:      ", profile),
			th.Format(theme.RoleDate, res.Octal, profile))

		if res.HumanBytes != "" {
			fmt.Fprintf(w, "  %s  %s\n",
				th.Format(theme.RoleMuted, "Byte Units: ", profile),
				th.Format(theme.RoleSize, res.HumanBytes, profile))
		}
	}

	botBorder := th.Format(theme.RoleMuted, box.BottomLeft+strings.Repeat(box.Horizontal, max(0, termWidth-2))+box.BottomRight, profile)
	fmt.Fprintln(w, botBorder)
}

// RenderPlain prints plain text number.
func RenderPlain(w io.Writer, res *CalcResult, opts Options) {
	switch opts.Base {
	case 16:
		if res.Hex != "" {
			fmt.Fprintln(w, res.Hex)
			return
		}
	case 2:
		if res.Binary != "" {
			fmt.Fprintln(w, res.Binary)
			return
		}
	case 8:
		if res.Octal != "" {
			fmt.Fprintln(w, res.Octal)
			return
		}
	}
	fmt.Fprintln(w, res.Formatted)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
