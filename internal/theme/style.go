package theme

import (
	"fmt"
	"strings"

	"nova/internal/terminal"
)

// RGB represents an 8-bit per channel RGB color value.
type RGB struct {
	R uint8
	G uint8
	B uint8
}

// HexRGB constructs an RGB value from a 6-digit hex color code (e.g. 0x50FA7B).
func HexRGB(hex uint32) RGB {
	return RGB{
		R: uint8((hex >> 16) & 0xFF),
		G: uint8((hex >> 8) & 0xFF),
		B: uint8(hex & 0xFF),
	}
}

// ANSI256 converts RGB to the closest standard ANSI 256-color palette index.
func (c RGB) ANSI256() int {
	if c.R == c.G && c.G == c.B {
		if c.R < 8 {
			return 16
		}
		if c.R > 248 {
			return 231
		}
		return 232 + int((float64(c.R)-8)/247.0*24.0)
	}

	rIdx := int(float64(c.R) / 255.0 * 5.0)
	gIdx := int(float64(c.G) / 255.0 * 5.0)
	bIdx := int(float64(c.B) / 255.0 * 5.0)
	return 16 + (36 * rIdx) + (6 * gIdx) + bIdx
}

// ANSI16 converts RGB to the closest 16-color ANSI code.
func (c RGB) ANSI16() int {
	bright := (int(c.R) + int(c.G) + int(c.B)) > 384
	var idx int

	if c.R > 128 && c.G <= 128 && c.B <= 128 {
		idx = 1 // Red
	} else if c.G > 128 && c.R <= 128 && c.B <= 128 {
		idx = 2 // Green
	} else if c.R > 128 && c.G > 128 && c.B <= 128 {
		idx = 3 // Yellow
	} else if c.B > 128 && c.R <= 128 && c.G <= 128 {
		idx = 4 // Blue
	} else if c.R > 128 && c.B > 128 && c.G <= 128 {
		idx = 5 // Magenta
	} else if c.G > 128 && c.B > 128 && c.R <= 128 {
		idx = 6 // Cyan
	} else if c.R > 128 && c.G > 128 && c.B > 128 {
		idx = 7 // White
	} else {
		idx = 0 // Black
	}

	if bright && idx > 0 {
		return idx + 8
	}
	return idx
}

// Style encapsulates foreground/background colors and text attributes.
type Style struct {
	Fg        *RGB
	Bg        *RGB
	Bold      bool
	Dim       bool
	Italic    bool
	Underline bool
}

// Format wraps text with appropriate ANSI escape sequences according to the terminal color profile.
func (s Style) Format(text string, profile terminal.ColorProfile) string {
	if text == "" {
		return ""
	}
	if profile == terminal.ColorNone {
		return text
	}

	var codes []string

	// Modifiers
	if s.Bold {
		codes = append(codes, "1")
	}
	if s.Dim {
		codes = append(codes, "2")
	}
	if s.Italic {
		codes = append(codes, "3")
	}
	if s.Underline {
		codes = append(codes, "4")
	}

	// Foreground color
	if s.Fg != nil {
		switch profile {
		case terminal.ColorTrueColor:
			codes = append(codes, fmt.Sprintf("38;2;%d;%d;%d", s.Fg.R, s.Fg.G, s.Fg.B))
		case terminal.Color256:
			codes = append(codes, fmt.Sprintf("38;5;%d", s.Fg.ANSI256()))
		case terminal.Color16:
			code16 := s.Fg.ANSI16()
			if code16 >= 8 {
				codes = append(codes, fmt.Sprintf("%d", 90+(code16-8)))
			} else {
				codes = append(codes, fmt.Sprintf("%d", 30+code16))
			}
		}
	}

	// Background color
	if s.Bg != nil {
		switch profile {
		case terminal.ColorTrueColor:
			codes = append(codes, fmt.Sprintf("48;2;%d;%d;%d", s.Bg.R, s.Bg.G, s.Bg.B))
		case terminal.Color256:
			codes = append(codes, fmt.Sprintf("48;5;%d", s.Bg.ANSI256()))
		case terminal.Color16:
			code16 := s.Bg.ANSI16()
			if code16 >= 8 {
				codes = append(codes, fmt.Sprintf("%d", 100+(code16-8)))
			} else {
				codes = append(codes, fmt.Sprintf("%d", 40+code16))
			}
		}
	}

	if len(codes) == 0 {
		return text
	}

	return fmt.Sprintf("\x1b[%sm%s\x1b[0m", strings.Join(codes, ";"), text)
}

// InterpolateRGB smoothly blends two RGB colors by parameter t (0.0 to 1.0).
func InterpolateRGB(from, to RGB, t float64) RGB {
	if t <= 0.0 {
		return from
	}
	if t >= 1.0 {
		return to
	}
	r := uint8(float64(from.R) + t*(float64(to.R)-float64(from.R)))
	g := uint8(float64(from.G) + t*(float64(to.G)-float64(from.G)))
	b := uint8(float64(from.B) + t*(float64(to.B)-float64(from.B)))
	return RGB{R: r, G: g, B: b}
}

// FormatGradient renders text with a smooth color gradient from startColor to endColor.
func FormatGradient(text string, from, to RGB, profile terminal.ColorProfile) string {
	if text == "" || profile == terminal.ColorNone {
		return text
	}
	runes := []rune(text)
	n := len(runes)
	if n <= 1 {
		st := Style{Fg: &from}
		return st.Format(text, profile)
	}

	var sb strings.Builder
	for i, r := range runes {
		t := float64(i) / float64(n-1)
		col := InterpolateRGB(from, to, t)
		st := Style{Fg: &col}
		sb.WriteString(st.Format(string(r), profile))
	}
	return sb.String()
}

// FormatMultiGradient renders text with a gradient spanning multiple RGB stops.
func FormatMultiGradient(text string, stops []RGB, profile terminal.ColorProfile) string {
	if len(stops) == 0 || text == "" || profile == terminal.ColorNone {
		return text
	}
	if len(stops) == 1 {
		st := Style{Fg: &stops[0]}
		return st.Format(text, profile)
	}

	runes := []rune(text)
	n := len(runes)
	if n <= 1 {
		st := Style{Fg: &stops[0]}
		return st.Format(text, profile)
	}

	var sb strings.Builder
	segments := float64(len(stops) - 1)
	for i, r := range runes {
		norm := float64(i) / float64(n-1) // 0.0 to 1.0
		scaled := norm * segments
		segIdx := int(scaled)
		if segIdx >= len(stops)-1 {
			segIdx = len(stops) - 2
		}
		localT := scaled - float64(segIdx)
		col := InterpolateRGB(stops[segIdx], stops[segIdx+1], localT)
		st := Style{Fg: &col}
		sb.WriteString(st.Format(string(r), profile))
	}
	return sb.String()
}
