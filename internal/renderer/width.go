package renderer

import (
	"regexp"
	"strings"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// StripANSI removes all ANSI escape sequences from a string.
func StripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// RuneWidth returns the display cell width of a rune (0, 1, or 2).
func RuneWidth(r rune) int {
	if r == 0 || r < 32 || (r >= 0x7f && r < 0xa0) {
		return 0 // Control characters
	}

	// Zero-width combining characters and marks
	if (r >= 0x0300 && r <= 0x036F) ||
		(r >= 0x1DC0 && r <= 0x1DFF) ||
		(r >= 0x20D0 && r <= 0x20FF) ||
		(r >= 0xFE00 && r <= 0xFE0F) ||
		(r >= 0xFE20 && r <= 0xFE2F) {
		return 0
	}

	// Wide characters: East Asian Wide, CJK, and Emojis
	if (r >= 0x1100 && r <= 0x115F) ||
		(r >= 0x2E80 && r <= 0xA4CF) ||
		(r >= 0xAC00 && r <= 0xD7A3) ||
		(r >= 0xF900 && r <= 0xFAFF) ||
		(r >= 0xFE30 && r <= 0xFE6F) ||
		(r >= 0xFF00 && r <= 0xFF60) ||
		(r >= 0xFFE0 && r <= 0xFFE6) ||
		(r >= 0x1F300 && r <= 0x1F9FF) ||
		(r >= 0x20000 && r <= 0x2FA1F) {
		return 2
	}

	return 1
}

// VisibleWidth calculates the visual column width of a string on a terminal,
// ignoring ANSI escape codes and accounting for wide characters.
func VisibleWidth(s string) int {
	clean := StripANSI(s)
	width := 0
	for _, r := range clean {
		width += RuneWidth(r)
	}
	return width
}

// Truncate shortens a string so that its visible width does not exceed maxWidth.
// If truncation occurs, tail (e.g. "…") is appended within the maxWidth limit.
func Truncate(s string, maxWidth int, tail string) string {
	if maxWidth <= 0 {
		return ""
	}

	tailWidth := VisibleWidth(tail)
	if tailWidth > maxWidth {
		tail = ""
		tailWidth = 0
	}

	visWidth := VisibleWidth(s)
	if visWidth <= maxWidth {
		return s
	}

	targetContentWidth := maxWidth - tailWidth
	var b strings.Builder
	currentWidth := 0
	inEscape := false

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == 0x1b && i+1 < len(runes) && runes[i+1] == '[' {
			inEscape = true
			b.WriteRune(r)
			continue
		}
		if inEscape {
			b.WriteRune(r)
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}

		rw := RuneWidth(r)
		if currentWidth+rw > targetContentWidth {
			break
		}
		b.WriteRune(r)
		currentWidth += rw
	}

	b.WriteString(tail)
	if strings.Contains(s, "\x1b[") {
		b.WriteString("\x1b[0m")
	}

	return b.String()
}

// PadRight appends spaces to s until its visible width reaches targetWidth.
func PadRight(s string, targetWidth int) string {
	vw := VisibleWidth(s)
	if vw >= targetWidth {
		return s
	}
	return s + strings.Repeat(" ", targetWidth-vw)
}

// PadLeft prepends spaces to s until its visible width reaches targetWidth.
func PadLeft(s string, targetWidth int) string {
	vw := VisibleWidth(s)
	if vw >= targetWidth {
		return s
	}
	return strings.Repeat(" ", targetWidth-vw) + s
}

// Center pads s on both sides with spaces to center it within targetWidth.
func Center(s string, targetWidth int) string {
	vw := VisibleWidth(s)
	if vw >= targetWidth {
		return s
	}
	totalPad := targetWidth - vw
	leftPad := totalPad / 2
	rightPad := totalPad - leftPad
	return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
}
