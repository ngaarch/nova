package theme

import (
	"strings"
	"testing"

	"nova/internal/terminal"
)

func TestHexRGB(t *testing.T) {
	c := HexRGB(0xFF5533)
	if c.R != 0xFF || c.G != 0x55 || c.B != 0x33 {
		t.Errorf("HexRGB(0xFF5533) = RGB(%d, %d, %d); want (255, 85, 51)", c.R, c.G, c.B)
	}
}

func TestStyleFormatNone(t *testing.T) {
	st := Style{
		Fg:   rgbPtr(HexRGB(0xFF0000)),
		Bold: true,
	}
	formatted := st.Format("hello", terminal.ColorNone)
	if formatted != "hello" {
		t.Errorf("expected unmodified text for ColorNone, got %q", formatted)
	}
}

func TestStyleFormatTrueColor(t *testing.T) {
	st := Style{
		Fg:   rgbPtr(HexRGB(0xFF0080)),
		Bold: true,
	}
	formatted := st.Format("test", terminal.ColorTrueColor)
	if !strings.HasPrefix(formatted, "\x1b[1;38;2;255;0;128m") || !strings.HasSuffix(formatted, "\x1b[0m") {
		t.Errorf("unexpected truecolor format: %q", formatted)
	}
}

func TestStyleFormat256(t *testing.T) {
	st := Style{
		Fg: rgbPtr(HexRGB(0xFF0000)),
	}
	formatted := st.Format("test", terminal.Color256)
	if !strings.Contains(formatted, "38;5;") {
		t.Errorf("expected 256-color escape code, got %q", formatted)
	}
}

func TestStyleFormat16(t *testing.T) {
	st := Style{
		Fg: rgbPtr(HexRGB(0x00FF00)), // Green
	}
	formatted := st.Format("test", terminal.Color16)
	if !strings.HasPrefix(formatted, "\x1b[") || !strings.Contains(formatted, "mtest\x1b[0m") {
		t.Errorf("expected 16-color escape code, got %q", formatted)
	}
}

func TestThemeRegistry(t *testing.T) {
	themes := []string{"default", "minimal", "mono", "nord", "dracula", "neon"}
	for _, name := range themes {
		th := Get(name)
		if th == nil {
			t.Errorf("theme %q returned nil", name)
		} else if th.Name != name {
			t.Errorf("theme name = %q; want %q", th.Name, name)
		}
	}

	fallback := Get("nonexistent_theme")
	if fallback == nil || fallback.Name != "default" {
		t.Errorf("expected fallback to default, got %v", fallback)
	}
}

func TestThemeFormat(t *testing.T) {
	dracula := Get("dracula")
	formatted := dracula.Format(RoleDirectory, "mydir", terminal.ColorTrueColor)
	if !strings.Contains(formatted, "mydir") || !strings.Contains(formatted, "\x1b[") {
		t.Errorf("expected formatted directory, got %q", formatted)
	}

	mono := Get("mono")
	monoFormatted := mono.Format(RoleDirectory, "mydir", terminal.ColorTrueColor)
	if monoFormatted != "mydir" {
		t.Errorf("expected plain text for mono theme, got %q", monoFormatted)
	}
}

func TestLookupIcon(t *testing.T) {
	if icon := LookupIcon("dir", TypeDirectory, true); icon != "📁 " {
		t.Errorf("expected directory unicode icon, got %q", icon)
	}
	if icon := LookupIcon("main.go", TypeRegular, true); icon != "🐹 " {
		t.Errorf("expected go unicode icon, got %q", icon)
	}
	if icon := LookupIcon("script.py", TypeRegular, true); icon != "🐍 " {
		t.Errorf("expected python unicode icon, got %q", icon)
	}
	if icon := LookupIcon("sym", TypeSymlink, true); icon != "🔗 " {
		t.Errorf("expected symlink unicode icon, got %q", icon)
	}

	if icon := LookupIcon("dir", TypeDirectory, false); icon != "/ " {
		t.Errorf("expected directory ASCII icon, got %q", icon)
	}
	if icon := LookupIcon("sym", TypeSymlink, false); icon != "@ " {
		t.Errorf("expected symlink ASCII icon, got %q", icon)
	}
	if icon := LookupIcon("exec", TypeExecutable, false); icon != "* " {
		t.Errorf("expected exec ASCII icon, got %q", icon)
	}
	if icon := LookupIcon("main.go", TypeRegular, false); icon != "  " {
		t.Errorf("expected regular file ASCII icon, got %q", icon)
	}
}
