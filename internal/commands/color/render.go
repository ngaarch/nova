package colorcmd

import (
	"fmt"
	"io"
	"math"
	"strings"

	"nova/internal/command"
	"nova/internal/renderer"
	"nova/internal/terminal"
	"nova/internal/theme"
)

// RenderPlain prints colors and contrasts in scriptable tab-separated format.
func RenderPlain(w io.Writer, report ColorReport, mode string) {
	if (mode == "all" || mode == "16") && len(report.ANSI16) > 0 {
		fmt.Fprintln(w, "CODE\tNAME\tHEX\tR\tG\tB")
		for _, c := range report.ANSI16 {
			fmt.Fprintf(w, "%d\t%s\t%s\t%d\t%d\t%d\n", c.Code, c.Name, c.Hex, c.R, c.G, c.B)
		}
	}

	if (mode == "all" || mode == "256") && len(report.Palette256) > 0 {
		if mode == "all" {
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, "CODE\tNAME\tHEX\tR\tG\tB")
		for _, c := range report.Palette256 {
			fmt.Fprintf(w, "%d\t%s\t%s\t%d\t%d\t%d\n", c.Code, c.Name, c.Hex, c.R, c.G, c.B)
		}
	}

	if (mode == "all" || mode == "contrast") && len(report.Contrasts) > 0 {
		if mode == "all" {
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, "FG_NAME\tFG_HEX\tBG_NAME\tBG_HEX\tRATIO\tLEVEL\tPASS_AA\tPASS_AAA")
		for _, p := range report.Contrasts {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%.2f\t%s\t%t\t%t\n",
				p.FgName, p.FgHex, p.BgName, p.BgHex, p.Ratio, p.Level, p.PassAA, p.PassAAA)
		}
	}
}

// RenderHuman prints rich, visual color palettes and contrast studios.
func RenderHuman(w io.Writer, report ColorReport, ctx *command.Context, opts Options) {
	th := ctx.Theme
	prof := ctx.Caps.ColorProfile
	width := ctx.Caps.Width
	if width < 76 {
		width = 76
	}
	if width > 100 {
		width = 100
	}

	repeatCount := width - 50
	if repeatCount < 2 {
		repeatCount = 2
	}
	header := fmt.Sprintf("╭ 🎨 Terminal Color Studio & Contrast Inspector %s╮", strings.Repeat("─", repeatCount))
	fmt.Fprintln(w, th.Format(theme.RoleAccent, renderer.Truncate(header, width, "─╮"), prof))

	fmt.Fprintf(w, "  Color Profile:  %s  |  Terminal Width: %d\n",
		th.Format(theme.RoleInfo, prof.String(), prof), ctx.Caps.Width)
	fmt.Fprintln(w, th.Format(theme.RoleMuted, "  "+strings.Repeat("─", width-4), prof))

	// 1. ANSI 16 Colors
	if (opts.Mode == "all" || opts.Mode == "16") && len(report.ANSI16) >= 16 {
		fmt.Fprintln(w, "  "+th.Format(theme.RoleAccent, "▸ 16 ANSI Standard Colors", prof))
		// Row 1: standard (0-7)
		fmt.Fprint(w, "    Standard: ")
		for i := 0; i < 8; i++ {
			c := report.ANSI16[i]
			printColorBlock(w, c, prof)
		}
		fmt.Fprintln(w)
		// Row 2: bright (8-15)
		fmt.Fprint(w, "    Bright:   ")
		for i := 8; i < 16; i++ {
			c := report.ANSI16[i]
			printColorBlock(w, c, prof)
		}
		fmt.Fprintln(w)
		if opts.Mode == "all" {
			fmt.Fprintln(w, th.Format(theme.RoleMuted, "  "+strings.Repeat("┈", width-4), prof))
		}
	}

	// 2. 256 Color Cube
	if (opts.Mode == "all" || opts.Mode == "256") && len(report.Palette256) == 256 {
		fmt.Fprintln(w, "  "+th.Format(theme.RoleAccent, "▸ 256 Color Palette (6x6x6 Cube & Grayscale)", prof))
		// Print 6x6x6 cube
		for r := 0; r < 6; r++ {
			fmt.Fprint(w, "    ")
			for g := 0; g < 6; g++ {
				for b := 0; b < 6; b++ {
					code := 16 + (r * 36) + (g * 6) + b
					c := report.Palette256[code]
					printMiniBlock(w, c, prof)
				}
				fmt.Fprint(w, " ")
			}
			fmt.Fprintln(w)
		}
		// Print Grayscale ramp
		fmt.Fprint(w, "    Grayscale: ")
		for i := 232; i < 256; i++ {
			c := report.Palette256[i]
			printMiniBlock(w, c, prof)
		}
		fmt.Fprintln(w)
		if opts.Mode == "all" {
			fmt.Fprintln(w, th.Format(theme.RoleMuted, "  "+strings.Repeat("┈", width-4), prof))
		}
	}

	// 3. 24-bit TrueColor Spectrum
	if opts.Mode == "all" || opts.Mode == "truecolor" {
		fmt.Fprintln(w, "  "+th.Format(theme.RoleAccent, "▸ 24-bit TrueColor Spectrum", prof))
		barWidth := width - 12
		if barWidth < 30 {
			barWidth = 30
		}
		if barWidth > 64 {
			barWidth = 64
		}

		fmt.Fprint(w, "    ")
		for i := 0; i < barWidth; i++ {
			ratio := float64(i) / float64(barWidth)
			rgb := rainbowColor(ratio)
			if prof == terminal.ColorTrueColor {
				fmt.Fprintf(w, "\x1b[48;2;%d;%d;%dm \x1b[0m", rgb.R, rgb.G, rgb.B)
			} else if prof == terminal.Color256 {
				fmt.Fprintf(w, "\x1b[48;5;%dm \x1b[0m", rgb.ANSI256())
			} else {
				fmt.Fprint(w, "━")
			}
		}
		fmt.Fprintln(w)
		if opts.Mode == "all" {
			fmt.Fprintln(w, th.Format(theme.RoleMuted, "  "+strings.Repeat("┈", width-4), prof))
		}
	}

	// 4. WCAG Contrast Studio
	if (opts.Mode == "all" || opts.Mode == "contrast") && len(report.Contrasts) > 0 {
		fmt.Fprintln(w, "  "+th.Format(theme.RoleAccent, "▸ WCAG 2.1 Accessibility & Contrast Studio", prof))
		for _, p := range report.Contrasts {
			badgeRole := theme.RoleError
			if p.PassAAA {
				badgeRole = theme.RoleSuccess
			} else if p.PassAA {
				badgeRole = theme.RoleWarning
			}

			badge := fmt.Sprintf("[%s]", p.Level)
			badgeFormatted := th.Format(badgeRole, fmt.Sprintf("%-9s", badge), prof)

			sample := " Preview Text "
			sampleFormatted := renderPreviewSample(sample, p.FgHex, p.BgHex, prof)

			fmt.Fprintf(w, "    %-15s on %-15s %s Ratio: %-7s %s\n",
				p.FgName, p.BgName, badgeFormatted, th.Format(theme.RoleAccent, p.RatioStr, prof), sampleFormatted)
		}
	}

	fmt.Fprintln(w, th.Format(theme.RoleMuted, "  "+strings.Repeat("─", width-4), prof))
	statusMsg := "✔ Color capabilities verified. WCAG compliance evaluated."
	fmt.Fprintln(w, "  "+th.Format(theme.RoleSuccess, statusMsg, prof))

	bot := fmt.Sprintf("╰%s╯", strings.Repeat("─", width-2))
	fmt.Fprintln(w, th.Format(theme.RoleAccent, bot, prof))
}

func printColorBlock(w io.Writer, c ANSIColor, prof terminal.ColorProfile) {
	if prof == terminal.ColorTrueColor {
		fmt.Fprintf(w, "\x1b[48;2;%d;%d;%dm\x1b[38;2;%d;%d;%dm %02d \x1b[0m ",
			c.R, c.G, c.B, contrastTextByte(c.R, c.G, c.B), contrastTextByte(c.R, c.G, c.B), contrastTextByte(c.R, c.G, c.B), c.Code)
	} else if prof == terminal.Color256 || prof == terminal.Color16 {
		fmt.Fprintf(w, "\x1b[48;5;%dm %02d \x1b[0m ", c.Code, c.Code)
	} else {
		fmt.Fprintf(w, "[%02d] ", c.Code)
	}
}

func printMiniBlock(w io.Writer, c ANSIColor, prof terminal.ColorProfile) {
	if prof == terminal.ColorTrueColor {
		fmt.Fprintf(w, "\x1b[48;2;%d;%d;%dm \x1b[0m", c.R, c.G, c.B)
	} else if prof == terminal.Color256 {
		fmt.Fprintf(w, "\x1b[48;5;%dm \x1b[0m", c.Code)
	} else {
		fmt.Fprint(w, "·")
	}
}

func contrastTextByte(r, g, b uint8) uint8 {
	lum := RelativeLuminance(r, g, b)
	if lum > 0.4 {
		return 0 // Black text on bright backgrounds
	}
	return 255 // White text on dark backgrounds
}

func rainbowColor(t float64) theme.RGB {
	// Simple smooth hue to RGB mapping
	h := t * 6.0
	c := 1.0
	x := c * (1.0 - math.Abs(math.Mod(h, 2.0)-1.0))

	var r, g, b float64
	switch {
	case h < 1.0:
		r, g, b = c, x, 0
	case h < 2.0:
		r, g, b = x, c, 0
	case h < 3.0:
		r, g, b = 0, c, x
	case h < 4.0:
		r, g, b = 0, x, c
	case h < 5.0:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return theme.RGB{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
	}
}

func renderPreviewSample(text, fgHex, bgHex string, prof terminal.ColorProfile) string {
	if prof == terminal.ColorNone {
		return fmt.Sprintf("[%s]", text)
	}
	fgRGB, errFg := ParseHexColor(fgHex)
	bgRGB, errBg := ParseHexColor(bgHex)
	if errFg != nil || errBg != nil {
		return text
	}

	if prof == terminal.ColorTrueColor {
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm%s\x1b[0m",
			fgRGB.R, fgRGB.G, fgRGB.B, bgRGB.R, bgRGB.G, bgRGB.B, text)
	}
	if prof == terminal.Color256 {
		return fmt.Sprintf("\x1b[38;5;%dm\x1b[48;5;%dm%s\x1b[0m",
			fgRGB.ANSI256(), bgRGB.ANSI256(), text)
	}
	return text
}
