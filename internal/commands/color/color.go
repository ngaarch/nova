package colorcmd

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"nova/internal/command"
	"nova/internal/output"
	"nova/internal/theme"
)

// ANSIColor represents an indexed or named terminal color.
type ANSIColor struct {
	Code int    `json:"code"`
	Name string `json:"name"`
	Hex  string `json:"hex"`
	R    uint8  `json:"r"`
	G    uint8  `json:"g"`
	B    uint8  `json:"b"`
}

// ContrastPair describes contrast accessibility between two colors.
type ContrastPair struct {
	FgName   string  `json:"fg_name"`
	FgHex    string  `json:"fg_hex"`
	BgName   string  `json:"bg_name"`
	BgHex    string  `json:"bg_hex"`
	Ratio    float64 `json:"ratio"`
	RatioStr string  `json:"ratio_str"`
	Level    string  `json:"level"`
	PassAA   bool    `json:"pass_aa"`
	PassAAA  bool    `json:"pass_aaa"`
}

// ColorReport contains the evaluated color palette and contrast metrics.
type ColorReport struct {
	Profile    string         `json:"profile"`
	ANSI16     []ANSIColor    `json:"ansi_16,omitempty"`
	Palette256 []ANSIColor    `json:"palette_256,omitempty"`
	Contrasts  []ContrastPair `json:"contrasts,omitempty"`
}

// Command returns the registered Command instance for color.
func Command() *command.Command {
	return &command.Command{
		Name:        "color",
		Aliases:     []string{"palette", "colors", "contrast"},
		Summary:     "Terminal color palette, contrast studio, and WCAG accessibility inspector",
		Usage:       "nova color [flags]",
		Description: "Inspect 16-color ANSI, 256-color palette, TrueColor gradients, and calculate WCAG 2.1 contrast ratios.",
		Phase:       32,
		Run:         Run,
	}
}

// Run executes the color command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)

	report := ColorReport{
		Profile: ctx.Caps.ColorProfile.String(),
	}

	if opts.Mode == "all" || opts.Mode == "16" {
		report.ANSI16 = GetANSI16Colors()
	}

	if opts.Mode == "all" || opts.Mode == "256" {
		report.Palette256 = Get256Colors()
	}

	if opts.Mode == "all" || opts.Mode == "contrast" {
		report.Contrasts = BuildContrastPairs(opts.Fg, opts.Bg)
	}

	if ctx.Printer.Mode == output.ModeJSON || opts.JSON {
		return ctx.Printer.PrintJSON(report)
	}

	if ctx.Printer.Mode == output.ModePlain || opts.Plain {
		RenderPlain(ctx.Stdout, report, opts.Mode)
	} else {
		RenderHuman(ctx.Stdout, report, ctx, opts)
	}

	return nil
}

// GetANSI16Colors returns the 16 standard ANSI colors.
func GetANSI16Colors() []ANSIColor {
	names := []string{
		"Black", "Red", "Green", "Yellow", "Blue", "Magenta", "Cyan", "White",
		"Bright Black", "Bright Red", "Bright Green", "Bright Yellow",
		"Bright Blue", "Bright Magenta", "Bright Cyan", "Bright White",
	}
	hexes := []string{
		"#000000", "#CD3131", "#0DBC79", "#E5E510", "#2472C8", "#BC3FBC", "#11A8CD", "#E5E5E5",
		"#666666", "#F14C4C", "#23D18B", "#F5F543", "#3B8EEA", "#D670D6", "#29B8DB", "#FFFFFF",
	}

	res := make([]ANSIColor, 16)
	for i := 0; i < 16; i++ {
		rgb, _ := ParseHexColor(hexes[i])
		res[i] = ANSIColor{
			Code: i,
			Name: names[i],
			Hex:  hexes[i],
			R:    rgb.R,
			G:    rgb.G,
			B:    rgb.B,
		}
	}
	return res
}

// Get256Colors returns the full 256 xterm color palette.
func Get256Colors() []ANSIColor {
	colors := make([]ANSIColor, 256)
	ansi16 := GetANSI16Colors()
	copy(colors[:16], ansi16)

	// 16-231: 6x6x6 color cube
	valMap := []uint8{0, 95, 175, 215, 255, 255}
	idx := 16
	for r := 0; r < 6; r++ {
		for g := 0; g < 6; g++ {
			for b := 0; b < 6; b++ {
				colors[idx] = ANSIColor{
					Code: idx,
					Name: fmt.Sprintf("Cube %d-%d-%d", r, g, b),
					Hex:  fmt.Sprintf("#%02X%02X%02X", valMap[r], valMap[g], valMap[b]),
					R:    valMap[r],
					G:    valMap[g],
					B:    valMap[b],
				}
				idx++
			}
		}
	}

	// 232-255: Grayscale ramp
	for i := 0; i < 24; i++ {
		gray := uint8(8 + i*10)
		colors[idx] = ANSIColor{
			Code: idx,
			Name: fmt.Sprintf("Gray %d", i+1),
			Hex:  fmt.Sprintf("#%02X%02X%02X", gray, gray, gray),
			R:    gray,
			G:    gray,
			B:    gray,
		}
		idx++
	}

	return colors
}

// ParseHexColor parses a hex color string like "#ABB2BF", "ABB2BF", "#FFF", or "FFF".
func ParseHexColor(s string) (theme.RGB, error) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(s) == 3 {
		// Expand short hex (e.g. "F0A" -> "FF00AA")
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return theme.RGB{}, fmt.Errorf("invalid hex color length: %q", s)
	}

	val, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return theme.RGB{}, fmt.Errorf("invalid hex color: %w", err)
	}

	return theme.HexRGB(uint32(val)), nil
}

// RelativeLuminance calculates WCAG 2.1 relative luminance for an sRGB color.
func RelativeLuminance(r, g, b uint8) float64 {
	toLinear := func(c uint8) float64 {
		v := float64(c) / 255.0
		if v <= 0.04045 {
			return v / 12.92
		}
		return math.Pow((v+0.055)/1.055, 2.4)
	}

	rLin := toLinear(r)
	gLin := toLinear(g)
	bLin := toLinear(b)

	return 0.2126*rLin + 0.7152*gLin + 0.0722*bLin
}

// ContrastRatio calculates the WCAG 2.1 contrast ratio between two relative luminance values.
func ContrastRatio(lum1, lum2 float64) float64 {
	l1 := lum1
	l2 := lum2
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

// WCAGLevel returns compliance level string ("AAA", "AA", "AA Large", or "Fail").
func WCAGLevel(ratio float64) string {
	if ratio >= 7.0 {
		return "AAA"
	}
	if ratio >= 4.5 {
		return "AA"
	}
	if ratio >= 3.0 {
		return "AA Large"
	}
	return "Fail"
}

// BuildContrastPairs creates contrast evaluations for custom fg/bg and common UI pairings.
func BuildContrastPairs(customFgHex, customBgHex string) []ContrastPair {
	pairs := make([]ContrastPair, 0, 8)

	// User-specified custom pair
	fgRGB, errFg := ParseHexColor(customFgHex)
	bgRGB, errBg := ParseHexColor(customBgHex)
	if errFg == nil && errBg == nil {
		lumFg := RelativeLuminance(fgRGB.R, fgRGB.G, fgRGB.B)
		lumBg := RelativeLuminance(bgRGB.R, bgRGB.G, bgRGB.B)
		cr := ContrastRatio(lumFg, lumBg)
		lvl := WCAGLevel(cr)
		pairs = append(pairs, ContrastPair{
			FgName:   "Custom FG",
			FgHex:    fmt.Sprintf("#%02X%02X%02X", fgRGB.R, fgRGB.G, fgRGB.B),
			BgName:   "Custom BG",
			BgHex:    fmt.Sprintf("#%02X%02X%02X", bgRGB.R, bgRGB.G, bgRGB.B),
			Ratio:    math.Round(cr*100) / 100,
			RatioStr: fmt.Sprintf("%.2f:1", cr),
			Level:    lvl,
			PassAA:   cr >= 4.5,
			PassAAA:  cr >= 7.0,
		})
	}

	// Standard terminal preset pairings
	presets := []struct {
		fgName, fgHex string
		bgName, bgHex string
	}{
		{"Bright White", "#FFFFFF", "Dark Canvas", "#1E1E1E"},
		{"OneDark Text", "#ABB2BF", "OneDark BG", "#282C34"},
		{"Nord Frost", "#88C0D0", "Nord Polar", "#2E3440"},
		{"Dracula Green", "#50FA7B", "Dracula BG", "#282A36"},
		{"Solarized Text", "#657B83", "Solarized Base", "#FDF6E3"},
		{"Dim / Muted", "#5C6370", "Dark Canvas", "#1E1E1E"},
	}

	for _, p := range presets {
		fRGB, _ := ParseHexColor(p.fgHex)
		bRGB, _ := ParseHexColor(p.bgHex)
		lFg := RelativeLuminance(fRGB.R, fRGB.G, fRGB.B)
		lBg := RelativeLuminance(bRGB.R, bRGB.G, bRGB.B)
		cr := ContrastRatio(lFg, lBg)
		pairs = append(pairs, ContrastPair{
			FgName:   p.fgName,
			FgHex:    p.fgHex,
			BgName:   p.bgName,
			BgHex:    p.bgHex,
			Ratio:    math.Round(cr*100) / 100,
			RatioStr: fmt.Sprintf("%.2f:1", cr),
			Level:    WCAGLevel(cr),
			PassAA:   cr >= 4.5,
			PassAAA:  cr >= 7.0,
		})
	}

	return pairs
}
