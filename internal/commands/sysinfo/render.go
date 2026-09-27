package sysinfo

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"nova/internal/command"
	"nova/internal/renderer"
	"nova/internal/theme"
)

// RenderJSON serializes telemetry info into structured JSON.
func RenderJSON(ctx *command.Context, info Info) error {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	ctx.Printer.Println(string(data))
	return nil
}

// RenderPlain prints key=value formatted diagnostics.
func RenderPlain(ctx *command.Context, info Info) {
	ctx.Printer.Printf("hostname=%s\n", info.Hostname)
	ctx.Printer.Printf("os=%s\n", info.OS)
	ctx.Printer.Printf("arch=%s\n", info.Arch)
	ctx.Printer.Printf("cpus=%d\n", info.CPUs)
	ctx.Printer.Printf("go_version=%s\n", info.GoVersion)
	ctx.Printer.Printf("goroutines=%d\n", info.Goroutines)
	ctx.Printer.Printf("memory_alloc_bytes=%d\n", info.MemoryAlloc)
	ctx.Printer.Printf("memory_sys_bytes=%d\n", info.MemorySys)
	ctx.Printer.Printf("disk_used_pct=%.1f\n", info.Disk.UsedPct)
	ctx.Printer.Printf("term_width=%d\n", info.TermWidth)
	ctx.Printer.Printf("term_height=%d\n", info.TermHeight)
	ctx.Printer.Printf("term_color=%s\n", info.TermColor)
	ctx.Printer.Printf("theme=%s\n", info.ThemeName)
	ctx.Printer.Printf("git_repo=%t\n", info.GitRepo)
	if info.GitRepo {
		ctx.Printer.Printf("git_branch=%s\n", info.GitBranch)
		ctx.Printer.Printf("git_clean=%t\n", info.GitClean)
	}
	ctx.Printer.Printf("path_total=%d\n", info.PathTotalCount)
	ctx.Printer.Printf("path_valid=%d\n", info.PathValidCount)
}

// RenderHuman prints rich, framed dashboard cards for system diagnostics.
func RenderHuman(ctx *command.Context, info Info, opts Options) {
	th := ctx.Theme
	prof := ctx.Caps.ColorProfile
	unicode := ctx.Caps.UnicodeSupported
	w := ctx.Caps.Width
	if w <= 0 || w > 100 {
		w = 80
	}

	borderStyle := th.Style(theme.RoleAccent)

	// 1. System & Architecture Card
	sysLines := []string{
		fmt.Sprintf(" %-16s %s", th.Format(theme.RoleMuted, "Host:", prof), th.Format(theme.RoleInfo, info.Hostname, prof)),
		fmt.Sprintf(" %-16s %s / %s", th.Format(theme.RoleMuted, "Platform:", prof), th.Format(theme.RoleInfo, info.OS, prof), th.Format(theme.RoleInfo, info.Arch, prof)),
		fmt.Sprintf(" %-16s %s cores", th.Format(theme.RoleMuted, "CPU:", prof), th.Format(theme.RoleAccent, fmt.Sprintf("%d", info.CPUs), prof)),
		fmt.Sprintf(" %-16s %s", th.Format(theme.RoleMuted, "Go Runtime:", prof), th.Format(theme.RoleSuccess, info.GoVersion, prof)),
		fmt.Sprintf(" %-16s %s active", th.Format(theme.RoleMuted, "Goroutines:", prof), th.Format(theme.RoleAccent, fmt.Sprintf("%d", info.Goroutines), prof)),
	}
	card1 := renderer.RenderCard("🖥️  System & Hardware", sysLines, w, borderStyle, unicode, prof)
	for _, l := range card1 {
		ctx.Printer.Println(l)
	}
	ctx.Printer.Println()

	// 2. Memory & Storage Card
	allocStr := renderer.FormatSize(int64(info.MemoryAlloc), true)
	sysMemStr := renderer.FormatSize(int64(info.MemorySys), true)
	memFraction := 0.0
	if info.MemorySys > 0 {
		memFraction = float64(info.MemoryAlloc) / float64(info.MemorySys)
	}
	memBar := renderer.RenderProgressBar(24, memFraction, th, prof)

	var diskLine string
	if info.Disk.Available {
		totalStr := renderer.FormatSize(int64(info.Disk.TotalBytes), true)
		usedStr := renderer.FormatSize(int64(info.Disk.UsedBytes), true)
		freeStr := renderer.FormatSize(int64(info.Disk.FreeBytes), true)
		diskFraction := info.Disk.UsedPct / 100.0
		diskBar := renderer.RenderProgressBar(24, diskFraction, th, prof)
		diskLine = fmt.Sprintf(" %-16s %s (%s / %s, %s free)",
			th.Format(theme.RoleMuted, "Disk Usage:", prof),
			diskBar,
			usedStr,
			totalStr,
			freeStr,
		)
	} else {
		diskLine = fmt.Sprintf(" %-16s %s", th.Format(theme.RoleMuted, "Disk Usage:", prof), th.Format(theme.RoleMuted, "n/a", prof))
	}

	memLines := []string{
		fmt.Sprintf(" %-16s %s (%s alloc / %s sys)", th.Format(theme.RoleMuted, "Go Memory:", prof), memBar, allocStr, sysMemStr),
		fmt.Sprintf(" %-16s %s cycles completed", th.Format(theme.RoleMuted, "Garbage Coll.:", prof), th.Format(theme.RoleAccent, fmt.Sprintf("%d", info.NumGC), prof)),
		diskLine,
		fmt.Sprintf(" %-16s %s", th.Format(theme.RoleMuted, "Working Dir:", prof), th.Format(theme.RoleInfo, info.Cwd, prof)),
	}
	card2 := renderer.RenderCard("⚡ Memory & Storage", memLines, w, borderStyle, unicode, prof)
	for _, l := range card2 {
		ctx.Printer.Println(l)
	}
	ctx.Printer.Println()

	// 3. Workspace, Terminal & Environment Card
	var gitStatusStr string
	if info.GitRepo {
		cleanBadge := th.Format(theme.RoleSuccess, "✔ clean", prof)
		if !info.GitClean {
			cleanBadge = th.Format(theme.RoleWarning, "● modified", prof)
		}
		gitStatusStr = fmt.Sprintf("%s [%s]",
			th.Format(theme.RoleAccent, info.GitBranch, prof),
			cleanBadge,
		)
	} else {
		gitStatusStr = th.Format(theme.RoleMuted, "not a git repository", prof)
	}

	pathSummary := fmt.Sprintf("%s entries (%s valid, %s missing)",
		th.Format(theme.RoleAccent, fmt.Sprintf("%d", info.PathTotalCount), prof),
		th.Format(theme.RoleSuccess, fmt.Sprintf("%d", info.PathValidCount), prof),
		th.Format(theme.RoleError, fmt.Sprintf("%d", info.PathTotalCount-info.PathValidCount), prof),
	)

	termLines := []string{
		fmt.Sprintf(" %-16s %s", th.Format(theme.RoleMuted, "Git Workspace:", prof), gitStatusStr),
		fmt.Sprintf(" %-16s %dx%d (TTY: %t, Profile: %s, Unicode: %t)",
			th.Format(theme.RoleMuted, "Terminal:", prof),
			info.TermWidth, info.TermHeight, info.TermIsTTY, info.TermColor, info.TermUnicode),
		fmt.Sprintf(" %-16s %s (Icons: %s)",
			th.Format(theme.RoleMuted, "Nova Theme:", prof),
			th.Format(theme.RoleAccent, info.ThemeName, prof),
			info.IconMode),
		fmt.Sprintf(" %-16s %s", th.Format(theme.RoleMuted, "System PATH:", prof), pathSummary),
	}

	if opts.Full && len(info.PathDirs) > 0 {
		termLines = append(termLines, strings.Repeat("─", w-4))
		termLines = append(termLines, th.Format(theme.RoleAccent, " PATH Directories:", prof))
		for idx, d := range info.PathDirs {
			bullet := "✔"
			role := theme.RoleSuccess
			if fi, err := os.Stat(d); err != nil || !fi.IsDir() {
				bullet = "✖"
				role = theme.RoleError
			}
			termLines = append(termLines, fmt.Sprintf("   %s [%d] %s", th.Format(role, bullet, prof), idx+1, d))
		}
	}

	card3 := renderer.RenderCard("🎨 Environment & Nova Context", termLines, w, borderStyle, unicode, prof)
	for _, l := range card3 {
		ctx.Printer.Println(l)
	}
}
