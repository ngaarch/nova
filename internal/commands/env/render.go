package env

import (
	"fmt"
	"strings"

	"nova/internal/command"
	"nova/internal/output"
	"nova/internal/terminal"
	"nova/internal/theme"
)

// Render outputs the environment inspection according to output mode.
func Render(ctx *command.Context, res Result, opts Options) error {
	// Shell export mode
	if opts.Export != "" {
		return renderExport(ctx, res, opts.Export)
	}

	if ctx.Printer.Mode == output.ModeJSON {
		return ctx.Printer.PrintJSON(res)
	}

	if ctx.Printer.Mode == output.ModePlain {
		var lines []string
		for _, v := range res.Variables {
			val := v.Value
			if v.IsSecret && !opts.ShowSecrets {
				val = "********"
			}
			lines = append(lines, fmt.Sprintf("%s=%s", v.Key, val))
		}
		if len(lines) > 0 {
			ctx.Printer.Print(strings.Join(lines, "\n") + "\n")
		}
		return nil
	}

	// Human mode
	prof := ctx.Caps.ColorProfile
	th := ctx.Theme

	title := fmt.Sprintf("🌐 Environment Audit (%d variables)", res.Total)
	header := fmt.Sprintf("╭ %s %s╮\n", title, strings.Repeat("─", max(2, 70-len(title))))
	var body strings.Builder
	body.WriteString(header)

	if opts.Group {
		categories := []string{"Development Runtimes", "System & Shell", "Cloud & DevOps", "General"}
		varsByCategory := make(map[string][]Variable)
		for _, v := range res.Variables {
			varsByCategory[v.Category] = append(varsByCategory[v.Category], v)
		}

		for _, cat := range categories {
			vars := varsByCategory[cat]
			if len(vars) == 0 {
				continue
			}

			catHeader := th.Format(theme.RoleAccent, fmt.Sprintf("  [%s — %d items]", cat, len(vars)), prof)
			body.WriteString(fmt.Sprintf("│ %s\n", catHeader))

			for _, v := range vars {
				renderVarLine(&body, v, th, prof)
			}
		}
	} else {
		for _, v := range res.Variables {
			renderVarLine(&body, v, th, prof)
		}
	}

	body.WriteString(fmt.Sprintf("│ %s\n", strings.Repeat("─", 74)))
	secretsNote := ""
	if res.SecretsCount > 0 {
		if opts.ShowSecrets {
			secretsNote = th.Format(theme.RoleWarning, fmt.Sprintf("%d secrets revealed", res.SecretsCount), prof)
		} else {
			secretsNote = th.Format(theme.RoleSuccess, fmt.Sprintf("%d secrets masked (use -s to reveal)", res.SecretsCount), prof)
		}
	} else {
		secretsNote = th.Format(theme.RoleMuted, "no sensitive credentials detected", prof)
	}

	summaryText := fmt.Sprintf("Audited %d variables | %s", res.Total, secretsNote)
	body.WriteString(fmt.Sprintf("│  %s\n", summaryText))
	body.WriteString(fmt.Sprintf("╰%s╯\n", strings.Repeat("─", 76)))

	ctx.Printer.Print(body.String())
	return nil
}

func renderVarLine(b *strings.Builder, v Variable, th *theme.Theme, prof terminal.ColorProfile) {
	keyFormatted := th.Format(theme.RoleExecutable, v.Key, prof)
	eq := th.Format(theme.RoleMuted, "=", prof)

	valFormatted := ""
	if v.IsSecret && v.DisplayValue == "********" {
		maskedPill := th.Format(theme.RoleWarning, "[MASKED]", prof)
		valFormatted = fmt.Sprintf("%s %s", th.Format(theme.RoleMuted, "********", prof), maskedPill)
	} else {
		valFormatted = th.Format(theme.RoleRegularFile, v.DisplayValue, prof)
	}

	b.WriteString(fmt.Sprintf("│    %s %s %s\n", keyFormatted, eq, valFormatted))
}

func renderExport(ctx *command.Context, res Result, format string) error {
	var lines []string
	for _, v := range res.Variables {
		escapedVal := strings.ReplaceAll(v.Value, "\"", "\\\"")
		if format == "fish" {
			lines = append(lines, fmt.Sprintf("set -gx %s \"%s\"", v.Key, escapedVal))
		} else {
			lines = append(lines, fmt.Sprintf("export %s=\"%s\"", v.Key, escapedVal))
		}
	}
	if len(lines) > 0 {
		ctx.Printer.Print(strings.Join(lines, "\n") + "\n")
	}
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
