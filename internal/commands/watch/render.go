package watch

import (
	"encoding/json"
	"fmt"
	"strings"

	"nova/internal/command"
	"nova/internal/renderer"
	"nova/internal/theme"
)

// RenderWatchHeader prints the initial monitoring dashboard banner.
func RenderWatchHeader(ctx *command.Context, opts Options) {
	th := ctx.Theme
	prof := ctx.Caps.ColorProfile
	unicode := ctx.Caps.UnicodeSupported
	w := ctx.Caps.Width
	if w <= 0 || w > 100 {
		w = 80
	}

	targets := strings.Join(opts.Paths, ", ")
	execCmd := opts.ExecCommand
	if execCmd == "" {
		execCmd = "(log events only)"
	}

	lines := []string{
		fmt.Sprintf(" %-16s %s", th.Format(theme.RoleMuted, "Watching:", prof), th.Format(theme.RoleInfo, targets, prof)),
		fmt.Sprintf(" %-16s %s (interval: %v, debounce: %v)", th.Format(theme.RoleMuted, "Action:", prof), th.Format(theme.RoleAccent, execCmd, prof), opts.Interval, opts.Debounce),
		strings.Repeat("─", w-4),
		" " + th.Format(theme.RoleSuccess, "⚡ Live file watcher active. Press Ctrl+C to terminate.", prof),
	}

	card := renderer.RenderCard("👀 Nova Watch Monitor", lines, w, th.Style(theme.RoleAccent), unicode, prof)
	for _, l := range card {
		ctx.Printer.Println(l)
	}
	ctx.Printer.Println()
}

// RenderHumanEvents prints live formatted change event rows.
func RenderHumanEvents(ctx *command.Context, events []FileEvent) {
	th := ctx.Theme
	prof := ctx.Caps.ColorProfile

	for _, ev := range events {
		timeStr := th.Format(theme.RoleMuted, ev.Timestamp.Format("15:04:05.000"), prof)
		var badge string
		switch ev.Type {
		case EventCreate:
			badge = th.Format(theme.RoleSuccess, "[CREATE]", prof)
		case EventModify:
			badge = th.Format(theme.RoleWarning, "[MODIFY]", prof)
		case EventDelete:
			badge = th.Format(theme.RoleError, "[DELETE]", prof)
		}
		pathStr := th.Format(theme.RoleInfo, ev.Path, prof)
		ctx.Printer.Printf("  %s  %s  %s\n", timeStr, badge, pathStr)
	}
}

// RenderPlainEvents outputs tab-delimited events for pipelines.
func RenderPlainEvents(ctx *command.Context, events []FileEvent) {
	for _, ev := range events {
		ctx.Printer.Printf("%s\t%s\t%s\n", ev.Timestamp.Format("15:04:05.000"), ev.Type, ev.Path)
	}
}

// RenderJSONEvents streams JSON records for events.
func RenderJSONEvents(ctx *command.Context, events []FileEvent) {
	for _, ev := range events {
		data, _ := json.Marshal(ev)
		ctx.Printer.Println(string(data))
	}
}
