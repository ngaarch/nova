package servecmd

import (
	"fmt"
	"io"
	"strings"
	"time"

	"nova/internal/command"
	"nova/internal/renderer"
	"nova/internal/theme"
)

// RenderStartupBanner renders the server launch card.
func RenderStartupBanner(w io.Writer, info *ServerInfo, qrLines []string, opts Options, ctx *command.Context) {
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

	title := " 🚀 Nova Development Web Server "
	topBorder := th.Format(theme.RoleAccent, box.TopLeft+title, profile) +
		th.Format(theme.RoleMuted, strings.Repeat(box.Horizontal, max(0, termWidth-renderer.VisibleWidth(title)-2))+box.TopRight, profile)
	fmt.Fprintln(w, topBorder)

	fmt.Fprintf(w, "  %s  %s\n",
		th.Format(theme.RoleMuted, "Serving:  ", profile),
		th.Format(theme.RoleRegularFile, info.Dir, profile))

	fmt.Fprintf(w, "  %s  %s\n",
		th.Format(theme.RoleMuted, "Local URL:", profile),
		th.Format(theme.RoleSuccess, info.LocalURL, profile))

	modeFlags := fmt.Sprintf("SPA: %t, CORS: %t", info.SPA, info.CORS)
	fmt.Fprintf(w, "  %s  %s\n",
		th.Format(theme.RoleMuted, "Options:  ", profile),
		th.Format(theme.RoleInfo, modeFlags, profile))

	if len(qrLines) > 0 {
		sep := th.Format(theme.RoleMuted, "  "+strings.Repeat("─", max(10, termWidth-4)), profile)
		fmt.Fprintln(w, sep)
		fmt.Fprintln(w, "  "+th.Format(theme.RoleDate, "Scan QR with mobile device on local network:", profile))
		for _, ql := range qrLines {
			fmt.Fprintln(w, "  "+ql)
		}
	}

	sep := th.Format(theme.RoleMuted, "  "+strings.Repeat("─", max(10, termWidth-4)), profile)
	fmt.Fprintln(w, sep)
	fmt.Fprintln(w, "  "+th.Format(theme.RoleMuted, "Press Ctrl+C to terminate server.", profile))

	botBorder := th.Format(theme.RoleMuted, box.BottomLeft+strings.Repeat(box.Horizontal, max(0, termWidth-2))+box.BottomRight, profile)
	fmt.Fprintln(w, botBorder)
}

// RenderRequestLog prints a color-coded request telemetry line.
func RenderRequestLog(w io.Writer, log RequestLog, ctx *command.Context) {
	th := ctx.Theme
	if th == nil {
		th = theme.Get("default")
	}
	profile := ctx.Caps.ColorProfile

	statusRole := theme.RoleSuccess
	statusGlyph := "✔"
	if log.Status >= 500 {
		statusRole = theme.RoleError
		statusGlyph = "✖"
	} else if log.Status >= 400 {
		statusRole = theme.RoleWarning
		statusGlyph = "✖"
	} else if log.Status >= 300 {
		statusRole = theme.RoleInfo
		statusGlyph = "→"
	}

	statusStr := fmt.Sprintf("%s %-3d", statusGlyph, log.Status)
	methodStr := fmt.Sprintf("%-6s", log.Method)
	latStr := formatLatency(log.Duration)
	sizeStr := renderer.FormatSize(log.Bytes, true)

	row := fmt.Sprintf("  %s  %s  %-24s  %8s  %8s",
		th.Format(statusRole, statusStr, profile),
		th.Format(theme.RoleAccent, methodStr, profile),
		th.Format(theme.RoleRegularFile, log.Path, profile),
		th.Format(theme.RoleDate, latStr, profile),
		th.Format(theme.RoleSize, sizeStr, profile))

	fmt.Fprintln(w, row)
}

func formatLatency(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%d µs", d.Microseconds())
	}
	return fmt.Sprintf("%.2f ms", float64(d.Microseconds())/1000.0)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
