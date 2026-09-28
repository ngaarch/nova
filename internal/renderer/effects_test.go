package renderer

import (
	"strings"
	"testing"
	"time"

	"nova/internal/terminal"
	"nova/internal/theme"
)

func TestSpinnerFrame(t *testing.T) {
	f0 := SpinnerFrame(0)
	f1 := SpinnerFrame(1)
	if f0 == "" || f1 == "" {
		t.Fatalf("expected non-empty spinner frames")
	}
	if f0 == f1 {
		t.Errorf("expected different frames for index 0 and 1")
	}
}

func TestRenderSparkline(t *testing.T) {
	vals := []float64{10, 20, 50, 80, 100}
	spark := RenderSparkline(vals)
	if len([]rune(spark)) != len(vals) {
		t.Errorf("expected %d sparkline glyphs, got %d", len(vals), len([]rune(spark)))
	}
}

func TestRenderProgressBar(t *testing.T) {
	th := theme.Get("default")
	bar := RenderProgressBar(30, 0.5, th, terminal.ColorNone)
	if !strings.Contains(bar, "50.0%") {
		t.Errorf("expected 50.0%% in progress bar, got: %s", bar)
	}
	if !strings.Contains(bar, "█") {
		t.Errorf("expected filled block glyphs in progress bar, got: %s", bar)
	}
}

func TestFormatRelativeTime(t *testing.T) {
	now := time.Now()
	if s := FormatRelativeTime(now.Add(-5 * time.Second)); s != "just now" {
		t.Errorf("expected 'just now', got %q", s)
	}
	if s := FormatRelativeTime(now.Add(-30 * time.Second)); !strings.HasSuffix(s, "s ago") {
		t.Errorf("expected seconds ago, got %q", s)
	}
	if s := FormatRelativeTime(now.Add(-10 * time.Minute)); s != "10m ago" {
		t.Errorf("expected '10m ago', got %q", s)
	}
	if s := FormatRelativeTime(now.Add(-3 * time.Hour)); s != "3h ago" {
		t.Errorf("expected '3h ago', got %q", s)
	}
	if s := FormatRelativeTime(now.Add(-25 * time.Hour)); s != "yesterday" {
		t.Errorf("expected 'yesterday', got %q", s)
	}
}

func TestHighlightFuzzyMatch(t *testing.T) {
	matchSt := theme.Style{Bold: true}
	normSt := theme.Style{}

	highlighted := HighlightFuzzyMatch("document.pdf", "doc", matchSt, normSt, terminal.ColorTrueColor)
	if !strings.Contains(highlighted, "\x1b[") {
		t.Errorf("expected ANSI formatting on match, got %q", highlighted)
	}
}

func TestRenderCard(t *testing.T) {
	st := theme.Style{}
	lines := []string{"hello", "world"}
	card := RenderCard("My Card", lines, 20, st, true, terminal.ColorNone)

	if len(card) != 4 {
		t.Fatalf("expected 4 card lines (top, 2 content, bot), got %d", len(card))
	}
	if !strings.Contains(card[0], "╭") || !strings.Contains(card[0], "My Card") {
		t.Errorf("card top border mismatch: %q", card[0])
	}
	if !strings.Contains(card[3], "╰") {
		t.Errorf("card bottom border mismatch: %q", card[3])
	}
}

func TestInterpolateRGBAndGradient(t *testing.T) {
	c1 := theme.HexRGB(0x000000)
	c2 := theme.HexRGB(0xFFFFFF)

	mid := InterpolateRGB(c1, c2, 0.5)
	if mid.R != 127 || mid.G != 127 || mid.B != 127 {
		t.Errorf("expected ~127, got (%d, %d, %d)", mid.R, mid.G, mid.B)
	}

	gradNone := RenderGradientText("NOVA SUITE", c1, c2, true, terminal.ColorNone)
	if gradNone != "NOVA SUITE" {
		t.Errorf("expected plain text when ColorNone, got %q", gradNone)
	}

	gradColor := RenderGradientText("NOVA SUITE", c1, c2, true, terminal.ColorTrueColor)
	if !strings.Contains(gradColor, "\x1b[") {
		t.Errorf("expected ANSI escapes for gradient text, got %q", gradColor)
	}
}
