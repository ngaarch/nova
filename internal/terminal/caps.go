package terminal

import (
	"os"
	"strconv"
	"strings"
)

// ColorProfile represents the color fidelity supported by the terminal.
type ColorProfile int

const (
	// ColorNone indicates plain monochrome text (no ANSI color codes).
	ColorNone ColorProfile = iota
	// Color16 indicates standard 16 ANSI colors (4-bit).
	Color16
	// Color256 indicates extended 256 colors (8-bit).
	Color256
	// ColorTrueColor indicates 24-bit direct RGB color.
	ColorTrueColor
)

// String returns a human-readable representation of the color profile.
func (c ColorProfile) String() string {
	switch c {
	case ColorTrueColor:
		return "truecolor"
	case Color256:
		return "256color"
	case Color16:
		return "16color"
	default:
		return "none"
	}
}

// Capabilities describes the operational capabilities of a terminal connection.
type Capabilities struct {
	IsTTY            bool
	Width            int
	Height           int
	ColorProfile     ColorProfile
	UnicodeSupported bool
}

// Detect inspects the given file descriptor and environment to determine terminal capabilities.
func Detect(file *os.File) Capabilities {
	return DetectWithEnv(file, os.Getenv)
}

// DetectWithEnv allows dependency-injected environment lookup for deterministic testing.
func DetectWithEnv(file *os.File, getenv func(string) string) Capabilities {
	isTTY := false
	if file != nil {
		isTTY = IsTerminal(file.Fd())
	}

	width, height := GetSize(file, getenv)

	colorProf := detectColor(isTTY, getenv)
	unicodeSupport := detectUnicode(getenv)

	return Capabilities{
		IsTTY:            isTTY,
		Width:            width,
		Height:           height,
		ColorProfile:     colorProf,
		UnicodeSupported: unicodeSupport,
	}
}

// detectColor determines the appropriate color profile based on TTY status and standard environment variables.
func detectColor(isTTY bool, getenv func(string) string) ColorProfile {
	clicolorForce := getenv("CLICOLOR_FORCE")
	forceColor := clicolorForce != "" && clicolorForce != "0"

	if !isTTY && !forceColor {
		return ColorNone
	}

	noColor := getenv("NO_COLOR")
	if noColor != "" && !forceColor {
		return ColorNone
	}

	clicolor := getenv("CLICOLOR")
	if clicolor == "0" && !forceColor {
		return ColorNone
	}

	term := strings.ToLower(getenv("TERM"))
	if term == "dumb" {
		if forceColor {
			return Color16
		}
		return ColorNone
	}

	colorTerm := strings.ToLower(getenv("COLORTERM"))
	if colorTerm == "truecolor" || colorTerm == "24bit" {
		return ColorTrueColor
	}

	if strings.Contains(term, "kitty") ||
		strings.Contains(term, "alacritty") ||
		strings.Contains(term, "wezterm") ||
		strings.Contains(term, "foot") ||
		strings.Contains(term, "ghostty") {
		return ColorTrueColor
	}

	if strings.Contains(term, "256color") || strings.Contains(term, "256-color") {
		return Color256
	}

	if term != "" {
		return Color16
	}

	return ColorNone
}

// detectUnicode determines whether the environment supports UTF-8 character encoding.
func detectUnicode(getenv func(string) string) bool {
	for _, envVar := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		val := strings.ToUpper(getenv(envVar))
		if strings.Contains(val, "UTF-8") || strings.Contains(val, "UTF8") {
			return true
		}
	}
	return false
}

// FallbackDimensions retrieves terminal dimensions from COLUMNS / LINES environment variables.
func FallbackDimensions(getenv func(string) string) (int, int) {
	width := 80
	height := 24

	if cols := getenv("COLUMNS"); cols != "" {
		if w, err := strconv.Atoi(cols); err == nil && w > 0 {
			width = w
		}
	}
	if lines := getenv("LINES"); lines != "" {
		if h, err := strconv.Atoi(lines); err == nil && h > 0 {
			height = h
		}
	}

	return width, height
}
