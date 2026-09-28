package renderer

import (
	"fmt"
	"math"
	"strings"
	"time"

	"nova/internal/terminal"
	"nova/internal/theme"
)

// SpinnerFrames provides a modern sequence of Braille spinning dots for loaders.
var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// SpinnerFrame returns the spinner glyph for an arbitrary tick count.
func SpinnerFrame(tick int) string {
	idx := tick % len(SpinnerFrames)
	if idx < 0 {
		idx = 0
	}
	return SpinnerFrames[idx]
}

// SparklineBlocks provides 8 graduated vertical block levels.
var SparklineBlocks = []rune{' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// RenderSparkline generates a mini Unicode sparkline from a slice of values.
func RenderSparkline(values []float64) string {
	if len(values) == 0 {
		return ""
	}

	minVal := values[0]
	maxVal := values[0]
	for _, v := range values {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	rangeVal := maxVal - minVal
	var sb strings.Builder

	for _, v := range values {
		if rangeVal == 0 {
			sb.WriteRune(SparklineBlocks[0])
			continue
		}
		norm := (v - minVal) / rangeVal
		idx := int(math.Floor(norm * float64(len(SparklineBlocks)-1)))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(SparklineBlocks) {
			idx = len(SparklineBlocks) - 1
		}
		sb.WriteRune(SparklineBlocks[idx])
	}

	return sb.String()
}

// RenderProgressBar generates a colorful proportional progress bar with gradient styling.
func RenderProgressBar(width int, fraction float64, th *theme.Theme, profile terminal.ColorProfile) string {
	if fraction < 0 {
		fraction = 0
	}
	if fraction > 1 {
		fraction = 1
	}

	pctStr := fmt.Sprintf(" %5.1f%%", fraction*100)
	barWidth := width - len(pctStr) - 2
	if barWidth < 3 {
		barWidth = 3
	}

	filledLen := int(math.Round(fraction * float64(barWidth)))
	emptyLen := barWidth - filledLen
	if emptyLen < 0 {
		emptyLen = 0
	}

	filledStr := strings.Repeat("█", filledLen)
	emptyStr := strings.Repeat("░", emptyLen)

	role := theme.RoleSuccess
	if fraction > 0.85 {
		role = theme.RoleError
	} else if fraction > 0.65 {
		role = theme.RoleWarning
	}

	styledFilled := th.Format(role, filledStr, profile)
	styledEmpty := th.Format(theme.RoleMuted, emptyStr, profile)
	styledPct := th.Format(theme.RoleAccent, pctStr, profile)

	return "[" + styledFilled + styledEmpty + "]" + styledPct
}

// FormatRelativeTime converts a timestamp into an intuitive human relative string.
func FormatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < 0 {
		return "in future"
	}
	if diff < 10*time.Second {
		return "just now"
	}
	if diff < time.Minute {
		return fmt.Sprintf("%ds ago", int(diff.Seconds()))
	}
	if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	}
	days := int(diff.Hours() / 24)
	if days == 1 {
		return "yesterday"
	}
	if days < 7 {
		return fmt.Sprintf("%dd ago", days)
	}
	if days < 30 {
		return fmt.Sprintf("%dw ago", days/7)
	}
	if days < 365 {
		return fmt.Sprintf("%dmo ago", days/30)
	}
	return fmt.Sprintf("%dy ago", days/365)
}

// HighlightFuzzyMatch highlights the matching characters in a text string based on a fuzzy search query.
func HighlightFuzzyMatch(text, query string, matchStyle, normalStyle theme.Style, profile terminal.ColorProfile) string {
	if query == "" || profile == terminal.ColorNone {
		return normalStyle.Format(text, profile)
	}

	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)

	matchedIndices := make(map[int]bool)
	tIdx := 0
	for _, qChar := range lowerQuery {
		found := false
		for ; tIdx < len(lowerText); tIdx++ {
			if rune(lowerText[tIdx]) == qChar {
				matchedIndices[tIdx] = true
				tIdx++
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	var sb strings.Builder
	runes := []rune(text)
	for i, r := range runes {
		if matchedIndices[i] {
			sb.WriteString(matchStyle.Format(string(r), profile))
		} else {
			sb.WriteString(normalStyle.Format(string(r), profile))
		}
	}

	return sb.String()
}

// BoxStyle defines border characters for framed cards and dialogs.
type BoxStyle struct {
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
	Horizontal  string
	Vertical    string
	TDown       string
	TUp         string
}

// RoundedBoxStyle provides modern rounded corners (╭ ╮ ╯ ╰).
var RoundedBoxStyle = BoxStyle{
	TopLeft:     "╭",
	TopRight:    "╮",
	BottomLeft:  "╰",
	BottomRight: "╯",
	Horizontal:  "─",
	Vertical:    "│",
	TDown:       "┬",
	TUp:         "┴",
}

// ASCIIBoxStyle provides clean ASCII fallback borders (+ - |).
var ASCIIBoxStyle = BoxStyle{
	TopLeft:     "+",
	TopRight:    "+",
	BottomLeft:  "+",
	BottomRight: "+",
	Horizontal:  "-",
	Vertical:    "|",
	TDown:       "+",
	TUp:         "+",
}

// RenderCard wraps content lines in a stylish framed card with title.
func RenderCard(title string, lines []string, width int, borderStyle theme.Style, unicode bool, profile terminal.ColorProfile) []string {
	box := RoundedBoxStyle
	if !unicode {
		box = ASCIIBoxStyle
	}

	if width < 10 {
		width = 10
	}

	innerW := width - 2
	var output []string

	// Header
	headerTitle := ""
	if title != "" {
		headerTitle = " " + title + " "
	}
	headerTitleLen := VisibleWidth(headerTitle)
	remainingH := innerW - headerTitleLen
	if remainingH < 0 {
		remainingH = 0
	}

	topBorder := box.TopLeft + headerTitle + strings.Repeat(box.Horizontal, remainingH) + box.TopRight
	output = append(output, borderStyle.Format(topBorder, profile))

	// Content lines
	for _, line := range lines {
		vLen := VisibleWidth(line)
		pad := innerW - vLen
		if pad < 0 {
			pad = 0
		}
		row := box.Vertical + " " + line + strings.Repeat(" ", pad-1) + box.Vertical
		output = append(output, borderStyle.Format(row, profile))
	}

	// Bottom border
	botBorder := box.BottomLeft + strings.Repeat(box.Horizontal, innerW) + box.BottomRight
	output = append(output, borderStyle.Format(botBorder, profile))

	return output
}

// InterpolateRGB smoothly blends two RGB colors based on factor t (0.0 to 1.0).
func InterpolateRGB(start, end theme.RGB, t float64) theme.RGB {
	if t <= 0 {
		return start
	}
	if t >= 1 {
		return end
	}
	r := uint8(float64(start.R) + t*float64(int(end.R)-int(start.R)))
	g := uint8(float64(start.G) + t*float64(int(end.G)-int(start.G)))
	b := uint8(float64(start.B) + t*float64(int(end.B)-int(start.B)))
	return theme.RGB{R: r, G: g, B: b}
}

// RenderGradientText renders text smoothly transitioning across a linear RGB color gradient.
func RenderGradientText(text string, start, end theme.RGB, bold bool, profile terminal.ColorProfile) string {
	if text == "" || profile == terminal.ColorNone {
		return text
	}

	runes := []rune(text)
	total := len(runes)
	if total <= 1 {
		c := start
		st := theme.Style{Fg: &c, Bold: bold}
		return st.Format(text, profile)
	}

	var sb strings.Builder
	for i, r := range runes {
		t := float64(i) / float64(total-1)
		col := InterpolateRGB(start, end, t)
		st := theme.Style{Fg: &col, Bold: bold}
		sb.WriteString(st.Format(string(r), profile))
	}
	return sb.String()
}
