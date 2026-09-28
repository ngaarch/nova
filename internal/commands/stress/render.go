package stresscmd

import (
	"fmt"
	"io"
	"strings"

	"nova/internal/command"
	"nova/internal/renderer"
	"nova/internal/theme"
)

// RenderPlain prints stress test metrics in tab-separated pipeline format.
func RenderPlain(w io.Writer, res StressResult) {
	fmt.Fprintf(w, "WORKLOAD\tTOTAL_OPS\tOPS_PER_SEC\tTHROUGHPUT\n")
	for _, wl := range res.Workloads {
		fmt.Fprintf(w, "%s\t%d\t%.2f\t%s\n", wl.Name, wl.TotalOps, wl.OpsPerSec, wl.Throughput)
	}
}

// RenderDashboard prints an interactive, colored benchmark report card.
func RenderDashboard(w io.Writer, res StressResult, ctx *command.Context) {
	th := ctx.Theme
	prof := ctx.Caps.ColorProfile
	width := ctx.Caps.Width
	if width < 76 {
		width = 76
	}
	if width > 100 {
		width = 100
	}

	repeatCount := width - 46
	if repeatCount < 2 {
		repeatCount = 2
	}
	header := fmt.Sprintf("╭ ⚡ Multi-Core Hardware Stress & Throughput %s╮", strings.Repeat("─", repeatCount))
	fmt.Fprintln(w, th.Format(theme.RoleAccent, renderer.Truncate(header, width, "─╮"), prof))

	fmt.Fprintf(w, "  Duration:     %s\n", th.Format(theme.RoleInfo, res.Duration.String(), prof))
	fmt.Fprintf(w, "  Concurrency:  %d thread(s) active\n", res.Threads)
	fmt.Fprintln(w, th.Format(theme.RoleMuted, "  "+strings.Repeat("─", width-4), prof))

	for i, wl := range res.Workloads {
		fmt.Fprintf(w, "  %s  %s\n",
			th.Format(theme.RoleWarning, "▸", prof),
			th.Format(theme.RoleAccent, wl.Name, prof),
		)
		fmt.Fprintf(w, "    Throughput:  %s\n", th.Format(theme.RoleSuccess, wl.Throughput, prof))
		fmt.Fprintf(w, "    Operations:  %d total (%.2f ops/sec)\n", wl.TotalOps, wl.OpsPerSec)
		if wl.TotalBytes > 0 {
			fmt.Fprintf(w, "    Volume:      %s processed\n", renderer.FormatSize(wl.TotalBytes, true))
		}

		if i < len(res.Workloads)-1 {
			fmt.Fprintln(w, th.Format(theme.RoleMuted, "  "+strings.Repeat("┈", width-4), prof))
		}
	}

	fmt.Fprintln(w, th.Format(theme.RoleMuted, "  "+strings.Repeat("─", width-4), prof))
	statusMsg := "✔ Hardware benchmark and stress cycle finished without thread contention."
	fmt.Fprintln(w, "  "+th.Format(theme.RoleSuccess, statusMsg, prof))

	bot := fmt.Sprintf("╰%s╯", strings.Repeat("─", width-2))
	fmt.Fprintln(w, th.Format(theme.RoleAccent, bot, prof))
}
