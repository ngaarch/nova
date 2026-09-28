package scancmd

import (
	"fmt"
	"io"
	"strings"

	"nova/internal/command"
	"nova/internal/renderer"
	"nova/internal/theme"
)

// RenderPlain prints secret findings in tab-separated unix pipeline format.
func RenderPlain(w io.Writer, res *ScanResult) {
	fmt.Fprintf(w, "FILE\tLINE\tSEVERITY\tRULE\tMASKED_SECRET\n")
	for _, f := range res.Findings {
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\n", f.File, f.Line, f.Severity, f.RuleID, f.Masked)
	}
}

// RenderDashboard prints an informative, colored security audit card.
func RenderDashboard(w io.Writer, res *ScanResult, ctx *command.Context) {
	th := ctx.Theme
	prof := ctx.Caps.ColorProfile
	width := ctx.Caps.Width
	if width < 76 {
		width = 76
	}
	if width > 100 {
		width = 100
	}

	repeatCount := width - 44 - len(res.Target)
	if repeatCount < 2 {
		repeatCount = 2
	}
	header := fmt.Sprintf("╭ 🛡️ Developer Secret & Entropy Scanner: %s %s╮", res.Target, strings.Repeat("─", repeatCount))
	fmt.Fprintln(w, th.Format(theme.RoleAccent, renderer.Truncate(header, width, "─╮"), prof))

	fmt.Fprintf(w, "  Scanned:      %d file(s) in %s\n", res.FilesScanned, res.Duration)

	findingsSummary := fmt.Sprintf("%d detected  (Critical: %d, High: %d, Medium: %d)",
		res.TotalFindings, res.CriticalCount, res.HighCount, res.MediumCount)

	if res.TotalFindings == 0 {
		fmt.Fprintf(w, "  Findings:     %s\n", th.Format(theme.RoleSuccess, "0 detected (All clean)", prof))
		fmt.Fprintln(w, th.Format(theme.RoleMuted, "  "+strings.Repeat("─", width-4), prof))
		statusMsg := "✔ Clean! Zero leaked API keys, tokens, or private credentials detected."
		fmt.Fprintln(w, "  "+th.Format(theme.RoleSuccess, statusMsg, prof))
		bot := fmt.Sprintf("╰%s╯", strings.Repeat("─", width-2))
		fmt.Fprintln(w, th.Format(theme.RoleAccent, bot, prof))
		return
	}

	fmt.Fprintf(w, "  Findings:     %s\n", th.Format(theme.RoleError, findingsSummary, prof))
	fmt.Fprintln(w, th.Format(theme.RoleMuted, "  "+strings.Repeat("─", width-4), prof))

	for i, f := range res.Findings {
		sevRole := theme.RoleError
		if f.Severity == "HIGH" {
			sevRole = theme.RoleWarning
		} else if f.Severity == "MEDIUM" {
			sevRole = theme.RoleInfo
		}

		badge := fmt.Sprintf("[%s] %s", f.Severity, f.RuleName)
		fmt.Fprintln(w, "  "+th.Format(sevRole, badge, prof))

		loc := fmt.Sprintf("%s:%d", f.File, f.Line)
		fmt.Fprintf(w, "    Location:   %s\n", th.Format(theme.RoleAccent, loc, prof))
		fmt.Fprintf(w, "    Masked:     %s\n", th.Format(theme.RoleWarning, f.Masked, prof))
		if f.Entropy > 0 {
			fmt.Fprintf(w, "    Entropy:    %.2f bits/byte\n", f.Entropy)
		}
		if f.Snippet != "" {
			fmt.Fprintf(w, "    Snippet:    %s\n", th.Format(theme.RoleMuted, f.Snippet, prof))
		}

		if i < len(res.Findings)-1 {
			fmt.Fprintln(w, th.Format(theme.RoleMuted, "  "+strings.Repeat("┈", width-4), prof))
		}
	}

	fmt.Fprintln(w, th.Format(theme.RoleMuted, "  "+strings.Repeat("─", width-4), prof))
	statusMsg := "✖ Potential secret leak detected! Rotate and revoke compromised keys immediately."
	fmt.Fprintln(w, "  "+th.Format(theme.RoleError, statusMsg, prof))

	bot := fmt.Sprintf("╰%s╯", strings.Repeat("─", width-2))
	fmt.Fprintln(w, th.Format(theme.RoleAccent, bot, prof))
}
