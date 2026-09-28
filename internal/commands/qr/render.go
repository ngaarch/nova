package qrcmd

import (
	"fmt"
	"strings"

	"nova/internal/command"
	"nova/internal/output"
	"nova/internal/theme"
)

// Render outputs the generated QR code.
func Render(ctx *command.Context, code Code, opts Options) error {
	if ctx.Printer.Mode == output.ModeJSON {
		return ctx.Printer.PrintJSON(code)
	}

	if ctx.Printer.Mode == output.ModePlain {
		// Output standard ASCII grid
		var b strings.Builder
		for r := 0; r < code.Size; r++ {
			for c := 0; c < code.Size; c++ {
				if code.Matrix[r][c] != opts.Invert {
					b.WriteString("##")
				} else {
					b.WriteString("  ")
				}
			}
			b.WriteString("\n")
		}
		ctx.Printer.Print(b.String())
		return nil
	}

	// Human mode: Unicode half-block characters
	prof := ctx.Caps.ColorProfile
	th := ctx.Theme

	title := fmt.Sprintf("📱 Terminal QR Code (v%d - %dx%d)", code.Version, code.Size, code.Size)
	header := fmt.Sprintf("╭ %s %s╮\n", title, strings.Repeat("─", max(2, code.Size+opts.QuietZone*2-len(title)+4)))
	var body strings.Builder
	body.WriteString(header)

	// Quiet zone padding
	qPad := strings.Repeat(" ", opts.QuietZone)

	// Top quiet zone (1 half-block row = 2 module quiet zone)
	if opts.QuietZone > 0 {
		topPad := strings.Repeat(" ", code.Size+opts.QuietZone*2)
		body.WriteString(fmt.Sprintf("│  %s  │\n", topPad))
	}

	// Scan through matrix 2 rows at a time
	for r := 0; r < code.Size; r += 2 {
		var rowStr strings.Builder
		rowStr.WriteString("│  ")
		rowStr.WriteString(qPad)

		for c := 0; c < code.Size; c++ {
			top := code.Matrix[r][c]
			bottom := false
			if r+1 < code.Size {
				bottom = code.Matrix[r+1][c]
			}

			if opts.Invert {
				top = !top
				bottom = !bottom
			}

			// Render half block
			// Note: On dark background terminals, bright block (white) is foreground.
			if top && bottom {
				rowStr.WriteString("█")
			} else if top && !bottom {
				rowStr.WriteString("▀")
			} else if !top && bottom {
				rowStr.WriteString("▄")
			} else {
				rowStr.WriteString(" ")
			}
		}

		rowStr.WriteString(qPad)
		rowStr.WriteString("  │\n")
		body.WriteString(rowStr.String())
	}

	// Bottom quiet zone
	if opts.QuietZone > 0 {
		botPad := strings.Repeat(" ", code.Size+opts.QuietZone*2)
		body.WriteString(fmt.Sprintf("│  %s  │\n", botPad))
	}

	previewText := code.Text
	if len(previewText) > 40 {
		previewText = previewText[:37] + "..."
	}
	summaryLine := fmt.Sprintf("Data: %s (%d chars)", previewText, len(code.Text))
	body.WriteString(fmt.Sprintf("│  %s\n", th.Format(theme.RoleInfo, summaryLine, prof)))
	body.WriteString(fmt.Sprintf("╰%s╯\n", strings.Repeat("─", max(2, code.Size+opts.QuietZone*2+6))))

	ctx.Printer.Print(body.String())
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
