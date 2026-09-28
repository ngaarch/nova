package colorcmd

import (
	"bytes"
	"math"
	"strings"
	"testing"

	"nova/internal/command"
	"nova/internal/config"
	"nova/internal/logging"
	"nova/internal/output"
	"nova/internal/terminal"
	"nova/internal/theme"
)

func newTestContext(stdin string, stdout, stderr *bytes.Buffer, mode output.Mode, prof terminal.ColorProfile) *command.Context {
	cfg := config.Config{
		Theme:     "default",
		ColorMode: "auto",
	}
	caps := terminal.Capabilities{
		Width:        80,
		Height:       24,
		ColorProfile: prof,
		IsTTY:        true,
	}
	th := theme.Get("default")
	logger := logging.New(stderr, false)
	return command.NewContext(
		strings.NewReader(stdin),
		stdout,
		stderr,
		cfg,
		caps,
		th,
		mode,
		logger,
	)
}

func TestParseFlags(t *testing.T) {
	opts := ParseFlags([]string{"--mode=256", "--fg=#123456", "--bg=#ABCDEF", "--plain", "--json"})
	if opts.Mode != "256" {
		t.Errorf("expected mode 256, got %s", opts.Mode)
	}
	if opts.Fg != "#123456" {
		t.Errorf("expected fg #123456, got %s", opts.Fg)
	}
	if opts.Bg != "#ABCDEF" {
		t.Errorf("expected bg #ABCDEF, got %s", opts.Bg)
	}
	if !opts.Plain || !opts.JSON {
		t.Errorf("expected plain and json to be true")
	}

	opts2 := ParseFlags([]string{"-m", "16", "--fg", "FFFFFF", "--bg", "000000"})
	if opts2.Mode != "16" || opts2.Fg != "FFFFFF" || opts2.Bg != "000000" {
		t.Errorf("unexpected parsed flags: %+v", opts2)
	}

	optsDefault := ParseFlags([]string{})
	if optsDefault.Mode != "all" {
		t.Errorf("expected default mode all, got %s", optsDefault.Mode)
	}
}

func TestParseHexColor(t *testing.T) {
	c, err := ParseHexColor("#FF00AA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.R != 255 || c.G != 0 || c.B != 170 {
		t.Errorf("expected (255, 0, 170), got (%d, %d, %d)", c.R, c.G, c.B)
	}

	cShort, err := ParseHexColor("#F0A")
	if err != nil {
		t.Fatalf("unexpected error for short hex: %v", err)
	}
	if cShort.R != 255 || cShort.G != 0 || cShort.B != 170 {
		t.Errorf("expected short hex expansion, got (%d, %d, %d)", cShort.R, cShort.G, cShort.B)
	}

	_, errInvalid := ParseHexColor("invalid")
	if errInvalid == nil {
		t.Errorf("expected error for invalid hex")
	}
}

func TestRelativeLuminanceAndContrast(t *testing.T) {
	blackLum := RelativeLuminance(0, 0, 0)
	if blackLum != 0.0 {
		t.Errorf("expected black luminance 0.0, got %f", blackLum)
	}

	whiteLum := RelativeLuminance(255, 255, 255)
	if math.Abs(whiteLum-1.0) > 0.001 {
		t.Errorf("expected white luminance 1.0, got %f", whiteLum)
	}

	ratio := ContrastRatio(whiteLum, blackLum)
	if math.Abs(ratio-21.0) > 0.01 {
		t.Errorf("expected contrast ratio 21:1, got %f", ratio)
	}

	if WCAGLevel(21.0) != "AAA" {
		t.Errorf("expected AAA for 21.0")
	}
	if WCAGLevel(5.0) != "AA" {
		t.Errorf("expected AA for 5.0")
	}
	if WCAGLevel(3.5) != "AA Large" {
		t.Errorf("expected AA Large for 3.5")
	}
	if WCAGLevel(2.0) != "Fail" {
		t.Errorf("expected Fail for 2.0")
	}
}

func TestGetANSI16And256(t *testing.T) {
	ansi16 := GetANSI16Colors()
	if len(ansi16) != 16 {
		t.Fatalf("expected 16 colors, got %d", len(ansi16))
	}
	for i, c := range ansi16 {
		if c.Code != i {
			t.Errorf("expected code %d, got %d", i, c.Code)
		}
	}

	pal256 := Get256Colors()
	if len(pal256) != 256 {
		t.Fatalf("expected 256 colors, got %d", len(pal256))
	}
}

func TestColorRunJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ctx := newTestContext("", &stdout, &stderr, output.ModeJSON, terminal.ColorTrueColor)

	err := Run(ctx, []string{"--json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, `"ansi_16"`) || !strings.Contains(out, `"palette_256"`) || !strings.Contains(out, `"contrasts"`) {
		t.Fatalf("expected complete json color report, got: %s", out)
	}
}

func TestColorRunPlain(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ctx := newTestContext("", &stdout, &stderr, output.ModePlain, terminal.ColorNone)

	err := Run(ctx, []string{"--mode=contrast", "--plain"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "FG_NAME") || !strings.Contains(out, "PASS_AA") {
		t.Fatalf("expected plain contrast output, got: %s", out)
	}
}

func TestColorRunHuman(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ctx := newTestContext("", &stdout, &stderr, output.ModeHuman, terminal.ColorTrueColor)

	err := Run(ctx, []string{"--mode=all"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "Terminal Color Studio") || !strings.Contains(out, "16 ANSI Standard Colors") {
		t.Fatalf("expected human color studio output, got: %s", out)
	}
}

func TestColorCommandMetadata(t *testing.T) {
	cmd := Command()
	if cmd.Name != "color" {
		t.Errorf("expected name color, got %s", cmd.Name)
	}
	if cmd.Phase != 32 {
		t.Errorf("expected phase 32, got %d", cmd.Phase)
	}
}
