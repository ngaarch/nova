package cli

import (
	"fmt"
	"strings"

	"nova/internal/output"
)

// PrintHelp outputs formatted help to the printer according to its output mode.
func PrintHelp(ctx *Context, commands []*Command) error {
	if ctx.Printer.Mode == output.ModeJSON {
		type cmdJSON struct {
			Name    string   `json:"name"`
			Aliases []string `json:"aliases,omitempty"`
			Summary string   `json:"summary"`
			Phase   int      `json:"phase"`
		}
		var list []cmdJSON
		for _, cmd := range commands {
			list = append(list, cmdJSON{
				Name:    cmd.Name,
				Aliases: cmd.Aliases,
				Summary: cmd.Summary,
				Phase:   cmd.Phase,
			})
		}
		data := map[string]any{
			"name":        "nova",
			"description": "Modern, fast, beautiful terminal utility suite written in Go",
			"usage":       "nova [global-flags] <command> [args...]",
			"commands":    list,
		}
		return ctx.Printer.PrintJSON(data)
	}

	var b strings.Builder
	b.WriteString("nova — modern, fast, beautiful terminal utility suite\n\n")
	b.WriteString("Usage:\n")
	b.WriteString("  nova [global-flags] <command> [args...]\n")
	b.WriteString("  nova [global-flags]\n\n")

	b.WriteString("Commands:\n")
	for _, cmd := range commands {
		status := ""
		if cmd.Run == nil {
			status = fmt.Sprintf(" [Scheduled: Phase %d]", cmd.Phase)
		}
		b.WriteString(fmt.Sprintf("  %-14s %s%s\n", cmd.Name, cmd.Summary, status))
	}

	b.WriteString("\nGlobal Flags:\n")
	b.WriteString("  -h, --help            Show this help message\n")
	b.WriteString("  -v, --version         Show version information\n")
	b.WriteString("      --plain           Force unformatted, deterministic plain text\n")
	b.WriteString("      --json            Force structured JSON output\n")
	b.WriteString("      --color=<mode>    Color policy: auto, always, never\n")
	b.WriteString("      --theme=<name>    Theme: default, minimal, mono, nord, dracula, neon\n")
	b.WriteString("      --icons=<mode>    Icon policy: auto, always, never\n")
	b.WriteString("      --debug           Enable diagnostic debug logging to stderr\n\n")

	b.WriteString("Examples:\n")
	b.WriteString("  nova ls -l                    # Directory listing (Phase 3)\n")
	b.WriteString("  nova cat README.md            # Stream and view file (Phase 4)\n")
	b.WriteString("  nova tree -L 2                # Directory tree with depth 2 (Phase 5)\n")
	b.WriteString("  nova --json ls                # Machine-readable JSON output\n\n")
	b.WriteString("Documentation & Roadmap: https://github.com/nova-cli/nova (ROADMAP.md)\n")

	ctx.Printer.Print(b.String())
	return nil
}
