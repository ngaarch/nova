package hash

import (
	"fmt"
	"strings"

	"nova/internal/command"
	"nova/internal/output"
	"nova/internal/theme"
)

// RenderCompute outputs the hash computation results.
func RenderCompute(ctx *command.Context, results []Result, opts Options) error {
	if ctx.Printer.Mode == output.ModeJSON {
		return ctx.Printer.PrintJSON(results)
	}

	if ctx.Printer.Mode == output.ModePlain {
		var lines []string
		for _, r := range results {
			if r.Error != "" {
				lines = append(lines, fmt.Sprintf("error: %s: %s", r.File, r.Error))
			} else {
				lines = append(lines, fmt.Sprintf("%s  %s", r.Hash, r.File))
			}
		}
		ctx.Printer.Print(strings.Join(lines, "\n") + "\n")
		return nil
	}

	// Human mode
	prof := ctx.Caps.ColorProfile
	th := ctx.Theme

	title := fmt.Sprintf("🔑 Cryptographic Hashes (%s)", strings.ToUpper(opts.Algorithm))
	header := fmt.Sprintf("╭ %s %s╮\n", title, strings.Repeat("─", max(2, 68-len(title))))
	var body strings.Builder
	body.WriteString(header)

	body.WriteString(fmt.Sprintf("│  %-30s %-10s %s\n",
		th.Format(theme.RoleMuted, "FILE", prof),
		th.Format(theme.RoleMuted, "SIZE", prof),
		th.Format(theme.RoleMuted, "HASH DIGEST", prof),
	))
	body.WriteString(fmt.Sprintf("│ %s\n", strings.Repeat("─", 74)))

	for _, r := range results {
		fileStr := r.File
		if len(fileStr) > 30 {
			fileStr = "..." + fileStr[len(fileStr)-27:]
		}

		if r.Error != "" {
			errText := th.Format(theme.RoleError, fmt.Sprintf("ERROR: %s", r.Error), prof)
			body.WriteString(fmt.Sprintf("│  %-30s %s\n", fileStr, errText))
		} else {
			sizeStr := formatBytes(r.Size)
			hashStr := th.Format(theme.RoleAccent, r.Hash, prof)
			body.WriteString(fmt.Sprintf("│  %-30s %-10s %s\n",
				th.Format(theme.RoleExecutable, fileStr, prof),
				th.Format(theme.RoleInfo, sizeStr, prof),
				hashStr,
			))
		}
	}

	body.WriteString(fmt.Sprintf("╰%s╯\n", strings.Repeat("─", 76)))
	ctx.Printer.Print(body.String())
	return nil
}

// RenderCheck outputs verification results.
func RenderCheck(ctx *command.Context, checks []CheckResult, summary CheckSummary, opts Options) error {
	if ctx.Printer.Mode == output.ModeJSON {
		return ctx.Printer.PrintJSON(map[string]any{
			"checks":  checks,
			"summary": summary,
		})
	}

	if ctx.Printer.Mode == output.ModePlain {
		var lines []string
		for _, c := range checks {
			if opts.Quiet && c.Status == "PASS" {
				continue
			}
			lines = append(lines, fmt.Sprintf("%s: %s", c.File, c.Status))
		}
		if len(lines) > 0 {
			ctx.Printer.Print(strings.Join(lines, "\n") + "\n")
		}
		if summary.Failed > 0 || summary.Missing > 0 {
			return fmt.Errorf("checksum verification failed: %d failed, %d missing", summary.Failed, summary.Missing)
		}
		return nil
	}

	// Human mode
	prof := ctx.Caps.ColorProfile
	th := ctx.Theme

	title := "🛡️  Checksum Verification"
	header := fmt.Sprintf("╭ %s %s╮\n", title, strings.Repeat("─", max(2, 68-len(title))))
	var body strings.Builder
	body.WriteString(header)

	for _, c := range checks {
		if opts.Quiet && c.Status == "PASS" {
			continue
		}

		var statusBadge string
		switch c.Status {
		case "PASS":
			statusBadge = th.Format(theme.RoleSuccess, "✔ PASS   ", prof)
			body.WriteString(fmt.Sprintf("│  %s  %s\n", statusBadge, c.File))
		case "FAILED":
			statusBadge = th.Format(theme.RoleError, "✖ FAILED ", prof)
			detail := th.Format(theme.RoleMuted, fmt.Sprintf("(expected %.8s... got %.8s...)", c.Expected, c.Actual), prof)
			body.WriteString(fmt.Sprintf("│  %s  %-28s %s\n", statusBadge, c.File, detail))
		case "MISSING":
			statusBadge = th.Format(theme.RoleWarning, "? MISSING", prof)
			detail := th.Format(theme.RoleMuted, fmt.Sprintf("(%s)", c.Error), prof)
			body.WriteString(fmt.Sprintf("│  %s  %-28s %s\n", statusBadge, c.File, detail))
		}
	}

	body.WriteString(fmt.Sprintf("│ %s\n", strings.Repeat("─", 74)))
	summaryText := fmt.Sprintf("Total: %d | Passed: %s | Failed: %s | Missing: %s",
		summary.Total,
		th.Format(theme.RoleSuccess, fmt.Sprintf("%d", summary.Passed), prof),
		th.Format(theme.RoleError, fmt.Sprintf("%d", summary.Failed), prof),
		th.Format(theme.RoleWarning, fmt.Sprintf("%d", summary.Missing), prof),
	)
	body.WriteString(fmt.Sprintf("│  %s\n", summaryText))
	body.WriteString(fmt.Sprintf("╰%s╯\n", strings.Repeat("─", 76)))

	ctx.Printer.Print(body.String())

	if summary.Failed > 0 || summary.Missing > 0 {
		return fmt.Errorf("checksum verification failed: %d failed, %d missing", summary.Failed, summary.Missing)
	}
	return nil
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
