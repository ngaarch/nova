package netcmd

import (
	"fmt"
	"io"
	"strings"
	"time"

	"nova/internal/command"
	"nova/internal/renderer"
	"nova/internal/theme"
)

// RenderPingDashboard renders colorful TCP ping results.
func RenderPingDashboard(w io.Writer, res *PingResult, ctx *command.Context) {
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

	title := fmt.Sprintf(" 📡 TCP Ping Telemetry: %s (%s) ", res.Target, res.Address)
	topBorder := th.Format(theme.RoleAccent, box.TopLeft+title, profile) +
		th.Format(theme.RoleMuted, strings.Repeat(box.Horizontal, max(0, termWidth-renderer.VisibleWidth(title)-2))+box.TopRight, profile)
	fmt.Fprintln(w, topBorder)

	var latVals []float64
	for _, att := range res.Attempts {
		var statusStr string
		if att.Success {
			statusStr = th.Format(theme.RoleSuccess, "✔ SUCCESS", profile)
			latVals = append(latVals, float64(att.Latency.Microseconds()))
		} else {
			statusStr = th.Format(theme.RoleError, "✖ FAILED ", profile)
		}

		latRole := theme.RoleSuccess
		if att.Latency > 150*time.Millisecond {
			latRole = theme.RoleError
		} else if att.Latency > 60*time.Millisecond {
			latRole = theme.RoleWarning
		}

		latFormatted := formatLatency(att.Latency)
		row := fmt.Sprintf("  seq=%-2d  status=%s  latency=%s",
			att.Seq,
			statusStr,
			th.Format(latRole, fmt.Sprintf("%-10s", latFormatted), profile))

		if att.Error != "" {
			row += "  " + th.Format(theme.RoleMuted, "("+att.Error+")", profile)
		}
		fmt.Fprintln(w, row)
	}

	sep := th.Format(theme.RoleMuted, "  "+strings.Repeat("─", max(10, termWidth-4)), profile)
	fmt.Fprintln(w, sep)

	// Sparkline
	if len(latVals) > 0 {
		spark := renderer.RenderSparkline(latVals)
		fmt.Fprintf(w, "  %s  %s\n",
			th.Format(theme.RoleMuted, "Latency Sparkline:", profile),
			th.Format(theme.RoleAccent, spark, profile))
	}

	// Statistics
	lossRole := theme.RoleSuccess
	if res.LossPercent > 50 {
		lossRole = theme.RoleError
	} else if res.LossPercent > 0 {
		lossRole = theme.RoleWarning
	}

	fmt.Fprintf(w, "  Packets: %d sent, %d received (%s)\n",
		res.Transmitted, res.Received,
		th.Format(lossRole, fmt.Sprintf("%.0f%% packet loss", res.LossPercent), profile))

	if res.Received > 0 {
		fmt.Fprintf(w, "  rtt min/avg/max/mdev = %s / %s / %s / %s\n",
			th.Format(theme.RoleSuccess, formatLatency(res.MinLatency), profile),
			th.Format(theme.RoleAccent, formatLatency(res.AvgLatency), profile),
			th.Format(theme.RoleWarning, formatLatency(res.MaxLatency), profile),
			th.Format(theme.RoleDate, formatLatency(res.Jitter), profile))
	}

	botBorder := th.Format(theme.RoleMuted, box.BottomLeft+strings.Repeat(box.Horizontal, max(0, termWidth-2))+box.BottomRight, profile)
	fmt.Fprintln(w, botBorder)
}

// RenderPingPlain prints tab-delimited deterministic ping output.
func RenderPingPlain(w io.Writer, res *PingResult) {
	fmt.Fprintln(w, "SEQ\tSTATUS\tLATENCY_US\tERROR")
	for _, att := range res.Attempts {
		status := "FAIL"
		if att.Success {
			status = "OK"
		}
		fmt.Fprintf(w, "%d\t%s\t%d\t%s\n", att.Seq, status, att.Latency.Microseconds(), att.Error)
	}
	fmt.Fprintf(w, "SUMMARY\t%d\t%d\t%.1f%%\t%d\t%d\t%d\n",
		res.Transmitted, res.Received, res.LossPercent,
		res.MinLatency.Microseconds(), res.AvgLatency.Microseconds(), res.MaxLatency.Microseconds())
}

// RenderScanDashboard renders port scan findings.
func RenderScanDashboard(w io.Writer, res *PortScanResult, ctx *command.Context) {
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

	title := fmt.Sprintf(" 🔍 Port Scanner: %s ", res.Target)
	topBorder := th.Format(theme.RoleAccent, box.TopLeft+title, profile) +
		th.Format(theme.RoleMuted, strings.Repeat(box.Horizontal, max(0, termWidth-renderer.VisibleWidth(title)-2))+box.TopRight, profile)
	fmt.Fprintln(w, topBorder)

	summary := fmt.Sprintf("  Scanned %d ports in %s  │  Open: %s, Closed/Filtered: %s",
		res.TotalScanned,
		res.Duration.Round(time.Millisecond),
		th.Format(theme.RoleSuccess, fmt.Sprintf("%d", res.OpenCount), profile),
		th.Format(theme.RoleMuted, fmt.Sprintf("%d", res.ClosedCount), profile))
	fmt.Fprintln(w, summary)

	sep := th.Format(theme.RoleMuted, "  "+strings.Repeat("─", max(10, termWidth-4)), profile)
	fmt.Fprintln(w, sep)

	header := fmt.Sprintf("  %-7s %-10s %-16s %s",
		th.Format(theme.RoleMuted, "PORT", profile),
		th.Format(theme.RoleMuted, "STATE", profile),
		th.Format(theme.RoleMuted, "SERVICE", profile),
		th.Format(theme.RoleMuted, "LATENCY", profile))
	fmt.Fprintln(w, header)

	for _, p := range res.Ports {
		// In full view or if open, show
		if p.State != "open" && res.OpenCount > 0 {
			// If there are open ports, only show open ports or first few closed to avoid clutter
			continue
		}

		var stateStyled string
		switch p.State {
		case "open":
			stateStyled = th.Format(theme.RoleSuccess, "✔ open", profile)
		case "timeout":
			stateStyled = th.Format(theme.RoleWarning, "⚠ timeout", profile)
		default:
			stateStyled = th.Format(theme.RoleMuted, "✖ closed", profile)
		}

		row := fmt.Sprintf("  %-7d %-10s %-16s %s",
			p.Port,
			stateStyled,
			th.Format(theme.RoleAccent, p.Service, profile),
			th.Format(theme.RoleDate, formatLatency(p.Latency), profile))
		fmt.Fprintln(w, row)
	}

	botBorder := th.Format(theme.RoleMuted, box.BottomLeft+strings.Repeat(box.Horizontal, max(0, termWidth-2))+box.BottomRight, profile)
	fmt.Fprintln(w, botBorder)
}

// RenderScanPlain prints plain tab-delimited scan data.
func RenderScanPlain(w io.Writer, res *PortScanResult) {
	fmt.Fprintln(w, "PORT\tSTATE\tSERVICE\tLATENCY_US")
	for _, p := range res.Ports {
		fmt.Fprintf(w, "%d\t%s\t%s\t%d\n", p.Port, p.State, p.Service, p.Latency.Microseconds())
	}
}

// RenderDNSDashboard renders DNS inspection data.
func RenderDNSDashboard(w io.Writer, res *DNSResult, ctx *command.Context) {
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

	title := fmt.Sprintf(" 🌐 DNS Records: %s ", res.Domain)
	topBorder := th.Format(theme.RoleAccent, box.TopLeft+title, profile) +
		th.Format(theme.RoleMuted, strings.Repeat(box.Horizontal, max(0, termWidth-renderer.VisibleWidth(title)-2))+box.TopRight, profile)
	fmt.Fprintln(w, topBorder)

	if len(res.A) > 0 {
		fmt.Fprintf(w, "  %s  %s\n",
			th.Format(theme.RoleMuted, "A (IPv4):     ", profile),
			th.Format(theme.RoleSuccess, strings.Join(res.A, ", "), profile))
	}
	if len(res.AAAA) > 0 {
		fmt.Fprintf(w, "  %s  %s\n",
			th.Format(theme.RoleMuted, "AAAA (IPv6):  ", profile),
			th.Format(theme.RoleInfo, strings.Join(res.AAAA, ", "), profile))
	}
	if res.CNAME != "" {
		fmt.Fprintf(w, "  %s  %s\n",
			th.Format(theme.RoleMuted, "CNAME:        ", profile),
			th.Format(theme.RoleAccent, res.CNAME, profile))
	}
	if len(res.MX) > 0 {
		fmt.Fprintf(w, "  %s  %s\n",
			th.Format(theme.RoleMuted, "MX (Mail):    ", profile),
			th.Format(theme.RoleDate, strings.Join(res.MX, ", "), profile))
	}
	if len(res.TXT) > 0 {
		for _, txt := range res.TXT {
			fmt.Fprintf(w, "  %s  %s\n",
				th.Format(theme.RoleMuted, "TXT:          ", profile),
				txt)
		}
	}

	fmt.Fprintf(w, "  %s  %s\n",
		th.Format(theme.RoleMuted, "Lookup Time:  ", profile),
		res.Duration.Round(time.Microsecond))

	botBorder := th.Format(theme.RoleMuted, box.BottomLeft+strings.Repeat(box.Horizontal, max(0, termWidth-2))+box.BottomRight, profile)
	fmt.Fprintln(w, botBorder)
}

// RenderDNSPlain prints plain DNS records.
func RenderDNSPlain(w io.Writer, res *DNSResult) {
	fmt.Fprintln(w, "TYPE\tVALUE")
	for _, a := range res.A {
		fmt.Fprintf(w, "A\t%s\n", a)
	}
	for _, aaaa := range res.AAAA {
		fmt.Fprintf(w, "AAAA\t%s\n", aaaa)
	}
	if res.CNAME != "" {
		fmt.Fprintf(w, "CNAME\t%s\n", res.CNAME)
	}
	for _, mx := range res.MX {
		fmt.Fprintf(w, "MX\t%s\n", mx)
	}
	for _, txt := range res.TXT {
		fmt.Fprintf(w, "TXT\t%s\n", txt)
	}
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
