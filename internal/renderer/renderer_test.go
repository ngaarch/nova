package renderer

import (
	"strings"
	"testing"
)

func TestRuneWidth(t *testing.T) {
	cases := []struct {
		r        rune
		expected int
	}{
		{'a', 1},
		{'Z', 1},
		{'9', 1},
		{'\n', 0},
		{'\x00', 0},
		{'📁', 2},
		{'🐹', 2},
		{'🦀', 2},
		{'中', 2},
		{'文', 2},
		{'日', 2},
		{0x0300, 0},
	}

	for _, c := range cases {
		if got := RuneWidth(c.r); got != c.expected {
			t.Errorf("RuneWidth(%c / %U) = %d; want %d", c.r, c.r, got, c.expected)
		}
	}
}

func TestVisibleWidth(t *testing.T) {
	cases := []struct {
		input    string
		expected int
	}{
		{"hello", 5},
		{"\x1b[31mhello\x1b[0m", 5},
		{"\x1b[1;38;2;255;0;128mstyled\x1b[0m", 6},
		{"📁 dir", 6},
		{"🐹 go", 5},
		{"中文", 4},
	}

	for _, c := range cases {
		if got := VisibleWidth(c.input); got != c.expected {
			t.Errorf("VisibleWidth(%q) = %d; want %d", c.input, got, c.expected)
		}
	}
}

func TestStripANSI(t *testing.T) {
	styled := "\x1b[1;31mBold Red\x1b[0m and \x1b[38;2;100;200;50mRGB\x1b[0m"
	expected := "Bold Red and RGB"
	if got := StripANSI(styled); got != expected {
		t.Errorf("StripANSI() = %q; want %q", got, expected)
	}
}

func TestTruncate(t *testing.T) {
	s1 := "verylongfilename.txt"
	t1 := Truncate(s1, 10, "…")
	if VisibleWidth(t1) > 10 {
		t.Errorf("Truncate width %d exceeds 10 for %q", VisibleWidth(t1), t1)
	}
	if !strings.HasSuffix(t1, "…") {
		t.Errorf("expected suffix '…', got %q", t1)
	}

	styled := "\x1b[32mverylonggreenfilename.txt\x1b[0m"
	t2 := Truncate(styled, 12, "…")
	if VisibleWidth(t2) > 12 {
		t.Errorf("Truncate styled width %d exceeds 12 for %q", VisibleWidth(t2), t2)
	}
	if !strings.HasSuffix(t2, "\x1b[0m") {
		t.Errorf("expected ANSI reset suffix on truncated styled string, got %q", t2)
	}

	short := "short"
	if got := Truncate(short, 10, "…"); got != short {
		t.Errorf("expected %q, got %q", short, got)
	}
}

func TestPadUtilities(t *testing.T) {
	if p := PadRight("abc", 6); p != "abc   " || VisibleWidth(p) != 6 {
		t.Errorf("PadRight error: %q", p)
	}
	if p := PadLeft("abc", 6); p != "   abc" || VisibleWidth(p) != 6 {
		t.Errorf("PadLeft error: %q", p)
	}
	if p := Center("abc", 7); p != "  abc  " || VisibleWidth(p) != 7 {
		t.Errorf("Center error: %q", p)
	}
	wide := "📁"
	if p := PadRight(wide, 4); VisibleWidth(p) != 4 || p != "📁  " {
		t.Errorf("PadRight wide rune error: %q, width=%d", p, VisibleWidth(p))
	}
}

func TestGridRender(t *testing.T) {
	items := []string{
		"file1.txt", "file2.txt", "file3.txt", "file4.txt",
		"file5.txt", "file6.txt", "file7.txt", "file8.txt",
	}

	gridWide := NewGrid(160, 2)
	linesWide := gridWide.Render(items)
	if len(linesWide) > 2 {
		t.Errorf("expected at most 2 rows in wide terminal, got %d", len(linesWide))
	}

	gridStd := NewGrid(80, 2)
	linesStd := gridStd.Render(items)
	if len(linesStd) == 0 {
		t.Fatalf("expected non-empty output")
	}

	// Very narrow terminal (15 cols): items of width 9 cannot fit 2 cols (9+2+9=20 > 15), must degrade to 1 column (8 rows)
	gridNarrow := NewGrid(15, 2)
	linesNarrow := gridNarrow.Render(items)
	if len(linesNarrow) != len(items) {
		t.Errorf("expected 8 rows on narrow terminal, got %d", len(linesNarrow))
	}

	testItems := []string{"item0", "item1", "item2", "item3"}
	grid2Col := NewGrid(30, 2)
	lines2Col := grid2Col.Render(testItems)
	if len(lines2Col) == 2 {
		if !strings.HasPrefix(lines2Col[0], "item0") || !strings.Contains(lines2Col[0], "item2") {
			t.Errorf("row 0 should have item0 and item2, got: %s", lines2Col[0])
		}
		if !strings.HasPrefix(lines2Col[1], "item1") || !strings.Contains(lines2Col[1], "item3") {
			t.Errorf("row 1 should have item1 and item3, got: %s", lines2Col[1])
		}
	}
}

func TestTableRender(t *testing.T) {
	tbl := NewTable(80)
	tbl.AddColumn("Permissions", AlignLeft)
	tbl.AddColumn("Size", AlignRight)
	tbl.AddColumn("Name", AlignLeft)

	tbl.AddRow("drwxr-xr-x", "4.0 KB", "docs")
	tbl.AddRow("-rw-r--r--", "512 B", "README.md")
	tbl.AddRow("-rwxr-xr-x", "12.5 MB", "nova")

	lines := tbl.Render(false)
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(lines))
	}

	header := lines[0]
	if !strings.Contains(header, "Permissions") || !strings.Contains(header, "Size") || !strings.Contains(header, "Name") {
		t.Errorf("unexpected header: %s", header)
	}

	plainLines := tbl.Render(true)
	if len(plainLines) != 3 {
		t.Fatalf("expected 3 plain lines, got %d", len(plainLines))
	}
	if !strings.Contains(plainLines[0], "drwxr-xr-x\t4.0 KB\tdocs") {
		t.Errorf("expected tab-separated plain line, got %q", plainLines[0])
	}
}

func TestTableNarrowTerminalTruncation(t *testing.T) {
	tbl := NewTable(30)
	tbl.AddColumn("Perm", AlignLeft)
	tbl.AddColumn("Size", AlignRight)
	tbl.AddColumn("Path", AlignLeft)

	tbl.AddRow("-rw-r--r--", "10 KB", "a_very_long_path_that_exceeds_terminal_width.txt")

	lines := tbl.Render(false)
	for _, l := range lines {
		if VisibleWidth(l) > 30 {
			t.Errorf("table line width %d exceeds narrow terminal 30: %q", VisibleWidth(l), l)
		}
	}
}
