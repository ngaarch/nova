package terminal

import (
	"os"
	"testing"
)

func TestColorProfileString(t *testing.T) {
	cases := []struct {
		profile  ColorProfile
		expected string
	}{
		{ColorNone, "none"},
		{Color16, "16color"},
		{Color256, "256color"},
		{ColorTrueColor, "truecolor"},
		{ColorProfile(99), "none"},
	}

	for _, c := range cases {
		if got := c.profile.String(); got != c.expected {
			t.Errorf("ColorProfile(%d).String() = %q; want %q", c.profile, got, c.expected)
		}
	}
}

func TestDetectColor(t *testing.T) {
	tests := []struct {
		name     string
		isTTY    bool
		env      map[string]string
		expected ColorProfile
	}{
		{
			name:     "non-tty without force returns none",
			isTTY:    false,
			env:      map[string]string{"TERM": "xterm-256color"},
			expected: ColorNone,
		},
		{
			name:     "non-tty with CLICOLOR_FORCE returns color",
			isTTY:    false,
			env:      map[string]string{"TERM": "xterm-256color", "CLICOLOR_FORCE": "1"},
			expected: Color256,
		},
		{
			name:     "NO_COLOR set disables color",
			isTTY:    true,
			env:      map[string]string{"TERM": "xterm-256color", "NO_COLOR": "1"},
			expected: ColorNone,
		},
		{
			name:     "CLICOLOR=0 disables color",
			isTTY:    true,
			env:      map[string]string{"TERM": "xterm-256color", "CLICOLOR": "0"},
			expected: ColorNone,
		},
		{
			name:     "TERM=dumb disables color",
			isTTY:    true,
			env:      map[string]string{"TERM": "dumb"},
			expected: ColorNone,
		},
		{
			name:     "COLORTERM=truecolor enables truecolor",
			isTTY:    true,
			env:      map[string]string{"TERM": "xterm", "COLORTERM": "truecolor"},
			expected: ColorTrueColor,
		},
		{
			name:     "Modern terminal (kitty) enables truecolor",
			isTTY:    true,
			env:      map[string]string{"TERM": "xterm-kitty"},
			expected: ColorTrueColor,
		},
		{
			name:     "256color terminal enables 256color",
			isTTY:    true,
			env:      map[string]string{"TERM": "xterm-256color"},
			expected: Color256,
		},
		{
			name:     "Standard term enables 16 colors",
			isTTY:    true,
			env:      map[string]string{"TERM": "ansi"},
			expected: Color16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getenv := func(key string) string {
				return tt.env[key]
			}
			got := detectColor(tt.isTTY, getenv)
			if got != tt.expected {
				t.Errorf("detectColor() = %v; want %v", got, tt.expected)
			}
		})
	}
}

func TestDetectUnicode(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		expected bool
	}{
		{
			name:     "LC_ALL with UTF-8",
			env:      map[string]string{"LC_ALL": "en_US.UTF-8"},
			expected: true,
		},
		{
			name:     "LANG with utf8 lowercase",
			env:      map[string]string{"LANG": "en_GB.utf8"},
			expected: true,
		},
		{
			name:     "C locale without utf8",
			env:      map[string]string{"LANG": "C", "LC_ALL": "C"},
			expected: false,
		},
		{
			name:     "Empty env vars",
			env:      map[string]string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getenv := func(key string) string {
				return tt.env[key]
			}
			got := detectUnicode(getenv)
			if got != tt.expected {
				t.Errorf("detectUnicode() = %v; want %v", got, tt.expected)
			}
		})
	}
}

func TestFallbackDimensions(t *testing.T) {
	tests := []struct {
		name       string
		env        map[string]string
		wantWidth  int
		wantHeight int
	}{
		{
			name:       "defaults when empty",
			env:        map[string]string{},
			wantWidth:  80,
			wantHeight: 24,
		},
		{
			name:       "custom COLUMNS and LINES",
			env:        map[string]string{"COLUMNS": "120", "LINES": "40"},
			wantWidth:  120,
			wantHeight: 40,
		},
		{
			name:       "invalid non-integer falls back",
			env:        map[string]string{"COLUMNS": "abc", "LINES": "-5"},
			wantWidth:  80,
			wantHeight: 24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getenv := func(key string) string {
				return tt.env[key]
			}
			w, h := FallbackDimensions(getenv)
			if w != tt.wantWidth || h != tt.wantHeight {
				t.Errorf("FallbackDimensions() = (%d, %d); want (%d, %d)", w, h, tt.wantWidth, tt.wantHeight)
			}
		})
	}
}

func TestDetectWithEnv(t *testing.T) {
	env := map[string]string{
		"COLUMNS":   "100",
		"LINES":     "50",
		"LANG":      "en_US.UTF-8",
		"COLORTERM": "truecolor",
	}
	getenv := func(key string) string {
		return env[key]
	}

	caps := DetectWithEnv(nil, getenv)
	if caps.IsTTY {
		t.Errorf("expected IsTTY = false for nil file")
	}
	if caps.Width != 100 || caps.Height != 50 {
		t.Errorf("expected 100x50, got %dx%d", caps.Width, caps.Height)
	}
	if !caps.UnicodeSupported {
		t.Errorf("expected UnicodeSupported = true")
	}
	if caps.ColorProfile != ColorNone {
		t.Errorf("expected ColorNone when not TTY and no force, got %v", caps.ColorProfile)
	}
}

func TestIsTerminalOnDevNull(t *testing.T) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Skip("unable to open os.DevNull")
	}
	defer f.Close()

	if IsTerminal(f.Fd()) {
		t.Errorf("IsTerminal(/dev/null) should be false")
	}
}
