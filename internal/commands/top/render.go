package topcmd

import (
	"fmt"
	"io"
	"strings"

	"nova/internal/command"
	"nova/internal/renderer"
	"nova/internal/theme"
)

// RenderDashboard renders the modern top telemetry dashboard.
func RenderDashboard(w io.Writer, stats *SystemStats, opts Options, ctx *command.Context) {
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

	topBorder := th.Format(theme.RoleAccent, box.TopLeft+" 📊 System & Process Monitor ", profile) +
		th.Format(theme.RoleMuted, strings.Repeat(box.Horizontal, max(0, termWidth-VisibleLen(" 📊 System & Process Monitor ")-2))+box.TopRight, profile)
	fmt.Fprintln(w, topBorder)

	// Summary line 1: LoadAvg & Uptime
	loadStr := fmt.Sprintf("%.2f, %.2f, %.2f", stats.LoadAvg[0], stats.LoadAvg[1], stats.LoadAvg[2])
	uptimeStr := stats.Uptime
	if uptimeStr == "" {
		uptimeStr = "n/a"
	}
	line1 := fmt.Sprintf("  %s %s  │  %s %s",
		th.Format(theme.RoleMuted, "Load:", profile),
		th.Format(theme.RoleAccent, loadStr, profile),
		th.Format(theme.RoleMuted, "Uptime:", profile),
		th.Format(theme.RoleDate, uptimeStr, profile))
	fmt.Fprintln(w, line1)

	// Summary line 2: CPU & Memory progress bars
	cpuBar := renderer.RenderProgressBar(24, stats.CPUUsage/100.0, th, profile)
	memBar := renderer.RenderProgressBar(24, stats.MemPercent/100.0, th, profile)
	line2 := fmt.Sprintf("  %s %s  │  %s %s (%s / %s)",
		th.Format(theme.RoleMuted, "CPU:", profile),
		cpuBar,
		th.Format(theme.RoleMuted, "Mem:", profile),
		memBar,
		renderer.FormatSize(int64(stats.MemUsed), true),
		renderer.FormatSize(int64(stats.MemTotal), true))
	fmt.Fprintln(w, line2)

	// Summary line 3: Process counts
	line3 := fmt.Sprintf("  %s %s total, %s running, %s sleeping, %s zombie",
		th.Format(theme.RoleMuted, "Tasks:", profile),
		th.Format(theme.RoleAccent, fmt.Sprintf("%d", stats.TotalProcs), profile),
		th.Format(theme.RoleSuccess, fmt.Sprintf("%d", stats.RunningProcs), profile),
		th.Format(theme.RoleInfo, fmt.Sprintf("%d", stats.SleepProcs), profile),
		th.Format(theme.RoleError, fmt.Sprintf("%d", stats.ZombieProcs), profile))
	fmt.Fprintln(w, line3)

	sep := th.Format(theme.RoleMuted, "  "+strings.Repeat("─", max(10, termWidth-4)), profile)
	fmt.Fprintln(w, sep)

	// Process Table Header
	header := fmt.Sprintf("  %-7s %-12s %-5s %7s %7s %10s %5s  %s",
		th.Format(theme.RoleMuted, "PID", profile),
		th.Format(theme.RoleMuted, "USER", profile),
		th.Format(theme.RoleMuted, "STAT", profile),
		th.Format(theme.RoleMuted, "CPU%", profile),
		th.Format(theme.RoleMuted, "MEM%", profile),
		th.Format(theme.RoleMuted, "RSS", profile),
		th.Format(theme.RoleMuted, "TH", profile),
		th.Format(theme.RoleMuted, "COMMAND", profile))
	fmt.Fprintln(w, header)

	for _, p := range stats.Processes {
		var stateStyled string
		switch p.State {
		case "R":
			stateStyled = th.Format(theme.RoleSuccess, p.State, profile)
		case "S", "I":
			stateStyled = th.Format(theme.RoleInfo, p.State, profile)
		case "D":
			stateStyled = th.Format(theme.RoleWarning, p.State, profile)
		case "Z":
			stateStyled = th.Format(theme.RoleError, p.State, profile)
		default:
			stateStyled = th.Format(theme.RoleMuted, p.State, profile)
		}

		cpuRole := theme.RoleRegularFile
		if p.CPUPercent > 50.0 {
			cpuRole = theme.RoleError
		} else if p.CPUPercent > 15.0 {
			cpuRole = theme.RoleWarning
		}

		memRole := theme.RoleRegularFile
		if p.MemPercent > 30.0 {
			memRole = theme.RoleError
		} else if p.MemPercent > 10.0 {
			memRole = theme.RoleWarning
		}

		userDisplay := p.User
		if len(userDisplay) > 12 {
			userDisplay = userDisplay[:11] + "…"
		}

		cmdDisplay := p.CommandLine
		cmdWidth := termWidth - 56
		if cmdWidth > 10 && len(cmdDisplay) > cmdWidth {
			cmdDisplay = cmdDisplay[:cmdWidth-1] + "…"
		}

		row := fmt.Sprintf("  %-7d %-12s %-5s %7s %7s %10s %5d  %s",
			p.PID,
			th.Format(theme.RoleUser, userDisplay, profile),
			stateStyled,
			th.Format(cpuRole, fmt.Sprintf("%.1f", p.CPUPercent), profile),
			th.Format(memRole, fmt.Sprintf("%.1f", p.MemPercent), profile),
			th.Format(theme.RoleSize, renderer.FormatSize(int64(p.RSS), true), profile),
			p.Threads,
			th.Format(theme.RoleExecutable, cmdDisplay, profile))

		fmt.Fprintln(w, row)
	}

	botBorder := th.Format(theme.RoleMuted, box.BottomLeft+strings.Repeat(box.Horizontal, max(0, termWidth-2))+box.BottomRight, profile)
	fmt.Fprintln(w, botBorder)
}

// RenderPlain renders tab-delimited deterministic output.
func RenderPlain(w io.Writer, stats *SystemStats, opts Options) {
	fmt.Fprintln(w, "PID\tUSER\tSTATE\tCPU%\tMEM%\tRSS\tTHREADS\tCOMMAND")
	for _, p := range stats.Processes {
		fmt.Fprintf(w, "%d\t%s\t%s\t%.1f\t%.1f\t%d\t%d\t%s\n",
			p.PID, p.User, p.State, p.CPUPercent, p.MemPercent, p.RSS, p.Threads, p.CommandLine)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func VisibleLen(s string) int {
	return renderer.VisibleWidth(s)
}
