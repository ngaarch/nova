package cli

import (
	"fmt"
	"strings"

	"nova/internal/output"
	"nova/internal/theme"
)

// PrintHelp outputs formatted help to the printer according to its output mode and theme.
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

	prof := ctx.Caps.ColorProfile
	th := ctx.Theme
	if ctx.Printer.Mode == output.ModePlain {
		prof = 0 // terminal.ColorNone
	}

	var b strings.Builder
	title := th.Format(theme.RoleAccent, "nova — modern, fast, beautiful terminal utility suite", prof)
	b.WriteString(title)
	b.WriteString("\n\n")

	b.WriteString(th.Format(theme.RoleAccent, "Usage:", prof))
	b.WriteString("\n")
	b.WriteString("  nova [global-flags] <command> [args...]\n")
	b.WriteString("  nova [global-flags]\n\n")

	b.WriteString(th.Format(theme.RoleAccent, "Commands:", prof))
	b.WriteString("\n")

	type categoryDef struct {
		Name     string
		Icon     string
		Commands []string
	}

	categories := []categoryDef{
		{
			Name:     "Files & Navigation",
			Icon:     "📁",
			Commands: []string{"interactive", "ls", "tree", "cat", "grep", "find", "stat", "du", "cp", "mv", "rm", "mkdir", "touch", "which", "clean"},
		},
		{
			Name:     "Network & Security",
			Icon:     "🔒",
			Commands: []string{"net", "http", "serve", "cert", "scan", "qr"},
		},
		{
			Name:     "System & Diagnostics",
			Icon:     "⚡",
			Commands: []string{"sysinfo", "top", "bench", "stress", "watch"},
		},
		{
			Name:     "Developer & Utilities",
			Icon:     "🛠️",
			Commands: []string{"git", "diff", "env", "calc", "hash", "hex", "archive", "run", "history", "md", "note", "color", "completion"},
		},
	}

	cmdMap := make(map[string]*Command, len(commands))
	for _, cmd := range commands {
		cmdMap[cmd.Name] = cmd
	}

	seen := make(map[string]bool)
	for _, cat := range categories {
		var matched []*Command
		for _, name := range cat.Commands {
			if c, ok := cmdMap[name]; ok {
				matched = append(matched, c)
				seen[name] = true
			}
		}
		if len(matched) == 0 {
			continue
		}
		catHeader := fmt.Sprintf("  %s %s", cat.Icon, cat.Name)
		b.WriteString(th.Format(theme.RoleAccent, catHeader, prof))
		b.WriteString("\n")
		for _, cmd := range matched {
			status := ""
			if cmd.Run == nil {
				status = " " + th.Format(theme.RoleMuted, fmt.Sprintf("[Scheduled: Phase %d]", cmd.Phase), prof)
			}
			cmdName := th.Format(theme.RoleExecutable, fmt.Sprintf("    %-14s", cmd.Name), prof)
			b.WriteString(fmt.Sprintf("%s %s%s\n", cmdName, cmd.Summary, status))
		}
		b.WriteString("\n")
	}

	// Remaining / other commands
	var others []*Command
	for _, cmd := range commands {
		if !seen[cmd.Name] {
			others = append(others, cmd)
		}
	}
	if len(others) > 0 {
		b.WriteString(th.Format(theme.RoleAccent, "  📦 Other Commands", prof))
		b.WriteString("\n")
		for _, cmd := range others {
			status := ""
			if cmd.Run == nil {
				status = " " + th.Format(theme.RoleMuted, fmt.Sprintf("[Scheduled: Phase %d]", cmd.Phase), prof)
			}
			cmdName := th.Format(theme.RoleExecutable, fmt.Sprintf("    %-14s", cmd.Name), prof)
			b.WriteString(fmt.Sprintf("%s %s%s\n", cmdName, cmd.Summary, status))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(th.Format(theme.RoleAccent, "Global Flags:", prof))
	b.WriteString("\n")
	flags := []struct {
		Flag string
		Desc string
	}{
		{"-h, --help", "Show this help message"},
		{"-v, --version", "Show version information"},
		{"    --plain", "Force unformatted, deterministic plain text"},
		{"    --json", "Force structured JSON output"},
		{"    --color=<mode>", "Color policy: auto, always, never"},
		{"    --theme=<name>", "Theme: default, minimal, mono, nord, dracula, neon, cyberpunk, synthwave, tokyo-night, catppuccin, gruvbox"},
		{"    --icons=<mode>", "Icon policy: auto, always, never"},
		{"    --debug", "Enable diagnostic debug logging to stderr"},
	}
	for _, f := range flags {
		flagFormatted := th.Format(theme.RoleInfo, fmt.Sprintf("%-18s", f.Flag), prof)
		b.WriteString(fmt.Sprintf("  %s %s\n", flagFormatted, f.Desc))
	}

	b.WriteString("\n")
	b.WriteString(th.Format(theme.RoleAccent, "Examples:", prof))
	b.WriteString("\n")
	examples := []struct {
		Cmd     string
		Comment string
	}{
		{"nova ls -l", "Directory listing (Phase 3)"},
		{"nova cat README.md", "Stream and view file (Phase 4)"},
		{"nova tree -L 2", "Directory tree with depth 2 (Phase 5)"},
		{"nova --json ls", "Machine-readable JSON output"},
	}
	for _, ex := range examples {
		cmdFormatted := th.Format(theme.RoleExecutable, fmt.Sprintf("%-28s", ex.Cmd), prof)
		commentFormatted := th.Format(theme.RoleMuted, "# "+ex.Comment, prof)
		b.WriteString(fmt.Sprintf("  %s %s\n", cmdFormatted, commentFormatted))
	}

	b.WriteString("\n")
	b.WriteString(th.Format(theme.RoleMuted, "Documentation & Roadmap: https://github.com/nova-cli/nova (ROADMAP.md)", prof))
	b.WriteString("\n")

	ctx.Printer.Print(b.String())
	return nil
}
