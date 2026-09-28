package hex

import (
	"fmt"
	"strings"

	"nova/internal/command"
	"nova/internal/output"
	"nova/internal/theme"
)

// Render outputs the hex dump results according to output mode.
func Render(ctx *command.Context, results []DumpResult, opts Options) error {
	if ctx.Printer.Mode == output.ModeJSON {
		return ctx.Printer.PrintJSON(results)
	}

	if ctx.Printer.Mode == output.ModePlain {
		var lines []string
		for _, r := range results {
			if r.Error != "" {
				lines = append(lines, fmt.Sprintf("error: %s: %s", r.Target, r.Error))
				continue
			}
			for _, row := range r.Rows {
				lines = append(lines, formatPlainRow(row, opts.Columns))
			}
		}
		if len(lines) > 0 {
			ctx.Printer.Print(strings.Join(lines, "\n") + "\n")
		}
		return nil
	}

	// Human mode
	prof := ctx.Caps.ColorProfile
	th := ctx.Theme

	for i, r := range results {
		if r.Error != "" {
			errText := th.Format(theme.RoleError, fmt.Sprintf("✖ Error dumping %s: %s\n", r.Target, r.Error), prof)
			ctx.Printer.Print(errText)
			continue
		}

		if i > 0 {
			ctx.Printer.Print("\n")
		}

		title := fmt.Sprintf("🔍 Hex Dump: %s", r.Target)
		header := fmt.Sprintf("╭ %s %s╮\n", title, strings.Repeat("─", max(2, 70-len(title))))
		var body strings.Builder
		body.WriteString(header)

		// Column numbers header (for 16 cols default)
		if opts.Columns == 16 {
			colHeader := fmt.Sprintf("│  %s   %s  %s   %s\n",
				th.Format(theme.RoleMuted, "OFFSET  ", prof),
				th.Format(theme.RoleMuted, "00 01 02 03 04 05 06 07", prof),
				th.Format(theme.RoleMuted, "08 09 0A 0B 0C 0D 0E 0F", prof),
				th.Format(theme.RoleMuted, "ASCII           ", prof),
			)
			body.WriteString(colHeader)
			body.WriteString(fmt.Sprintf("│ %s\n", strings.Repeat("─", 74)))
		}

		for _, row := range r.Rows {
			body.WriteString("│  ")
			// Offset
			offsetStr := th.Format(theme.RoleMuted, fmt.Sprintf("%08x", row.Offset), prof)
			body.WriteString(offsetStr)
			body.WriteString("  ")

			// Hex bytes with color categorization
			for bIdx := 0; bIdx < opts.Columns; bIdx++ {
				if bIdx == 8 && opts.Columns == 16 {
					body.WriteString(" ")
				}

				if bIdx < len(row.Raw) {
					b := row.Raw[bIdx]
					hexByte := fmt.Sprintf("%02x", b)
					var coloredByte string
					if b == 0 {
						coloredByte = th.Format(theme.RoleMuted, hexByte, prof)
					} else if b >= 32 && b <= 126 {
						coloredByte = th.Format(theme.RoleSuccess, hexByte, prof)
					} else if b < 32 || b == 127 {
						coloredByte = th.Format(theme.RoleAccent, hexByte, prof)
					} else {
						coloredByte = th.Format(theme.RoleWarning, hexByte, prof)
					}
					body.WriteString(coloredByte)
				} else {
					body.WriteString("  ")
				}
				body.WriteString(" ")
			}

			// ASCII panel
			body.WriteString(" │")
			for _, b := range row.Raw {
				if b >= 32 && b <= 126 {
					body.WriteString(th.Format(theme.RoleExecutable, string(rune(b)), prof))
				} else {
					body.WriteString(th.Format(theme.RoleMuted, ".", prof))
				}
			}
			// Pad remaining ASCII
			if len(row.Raw) < opts.Columns {
				body.WriteString(strings.Repeat(" ", opts.Columns-len(row.Raw)))
			}
			body.WriteString("│\n")
		}

		body.WriteString(fmt.Sprintf("│ %s\n", strings.Repeat("─", 74)))
		bytesCount := r.End - r.Start
		summaryText := fmt.Sprintf("Displayed %d bytes (offset 0x%x - 0x%x)", bytesCount, r.Start, r.End)
		body.WriteString(fmt.Sprintf("│  %s\n", th.Format(theme.RoleInfo, summaryText, prof)))
		body.WriteString(fmt.Sprintf("╰%s╯\n", strings.Repeat("─", 76)))

		ctx.Printer.Print(body.String())
	}

	return nil
}

func formatPlainRow(row Row, maxCols int) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%08x  ", row.Offset))

	for i := 0; i < maxCols; i++ {
		if i == 8 && maxCols == 16 {
			b.WriteString(" ")
		}
		if i < len(row.Raw) {
			b.WriteString(fmt.Sprintf("%02x ", row.Raw[i]))
		} else {
			b.WriteString("   ")
		}
	}

	b.WriteString(" |")
	b.WriteString(row.ASCII)
	b.WriteString("|")
	return b.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
