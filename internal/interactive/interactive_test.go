package interactive

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nova/internal/terminal"
	"nova/internal/theme"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	// Subdirectories
	subA := filepath.Join(tmpDir, "sub_a")
	subB := filepath.Join(tmpDir, "sub_b")
	_ = os.Mkdir(subA, 0755)
	_ = os.Mkdir(subB, 0755)

	// Files inside subA
	_ = os.WriteFile(filepath.Join(subA, "leaf.txt"), []byte("leaf content"), 0644)

	// Files inside root
	_ = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "notes.txt"), []byte("Meeting notes\nLine 2\nLine 3\n"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "binary.bin"), []byte{0x7f, 0x45, 0x4c, 0x46, 0x00, 0x01, 0x02, 0x03}, 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte("secret"), 0644)

	// Symlinks
	_ = os.Symlink(filepath.Join(tmpDir, "main.go"), filepath.Join(tmpDir, "main_link"))
	_ = os.Symlink(filepath.Join(tmpDir, "non_existent"), filepath.Join(tmpDir, "broken_link"))

	return tmpDir
}

func TestInteractiveReadDir(t *testing.T) {
	tmpDir := setupTestDir(t)

	// Without hidden files
	entries, err := ReadDir(tmpDir, false, true)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	hasSubA := false
	hasSubB := false
	hasMain := false
	hasHidden := false

	for _, e := range entries {
		if e.Name == "sub_a" {
			hasSubA = true
			if !e.IsDir {
				t.Errorf("sub_a should be marked as directory")
			}
		}
		if e.Name == "sub_b" {
			hasSubB = true
		}
		if e.Name == "main.go" {
			hasMain = true
		}
		if e.Name == ".hidden" {
			hasHidden = true
		}
	}

	if !hasSubA || !hasSubB || !hasMain {
		t.Errorf("expected sub_a, sub_b, and main.go to be loaded")
	}
	if hasHidden {
		t.Errorf("expected .hidden to be omitted when showHidden is false")
	}

	// Verify directories come first
	dirSeen := true
	for _, e := range entries {
		if !e.IsDir {
			dirSeen = false
		} else if !dirSeen {
			t.Errorf("directory %s found after non-directory files", e.Name)
		}
	}

	// With hidden files
	hiddenEntries, err := ReadDir(tmpDir, true, true)
	if err != nil {
		t.Fatalf("ReadDir with hidden failed: %v", err)
	}
	foundHidden := false
	for _, e := range hiddenEntries {
		if e.Name == ".hidden" {
			foundHidden = true
			break
		}
	}
	if !foundHidden {
		t.Errorf("expected .hidden to be present when showHidden is true")
	}
}

func TestInteractiveFilter(t *testing.T) {
	entries := []Entry{
		{Name: "app.go"},
		{Name: "app_test.go"},
		{Name: "README.md"},
		{Name: "main.go"},
		{Name: "go.mod"},
	}

	// Exact / substring match
	res := FilterEntries(entries, "app")
	if len(res) != 2 {
		t.Errorf("expected 2 matches for 'app', got %d", len(res))
	}

	// Case-insensitive match
	res = FilterEntries(entries, "readme")
	if len(res) != 1 || res[0].Name != "README.md" {
		t.Errorf("expected README.md match for 'readme', got %v", res)
	}

	// Fuzzy match
	res = FilterEntries(entries, "mgo") // matches main.go
	if len(res) < 1 {
		t.Errorf("expected fuzzy match for 'mgo'")
	}

	// Empty query returns all
	res = FilterEntries(entries, "")
	if len(res) != len(entries) {
		t.Errorf("expected all entries for empty query")
	}
}

func TestInteractivePreview(t *testing.T) {
	tmpDir := setupTestDir(t)
	th := theme.Get("default")
	profile := terminal.ColorNone

	entries, err := ReadDir(tmpDir, false, true)
	if err != nil {
		t.Fatal(err)
	}

	// 1. Directory preview
	var dirEntry *Entry
	var textEntry *Entry
	var binEntry *Entry
	var brokenEntry *Entry

	for i := range entries {
		switch entries[i].Name {
		case "sub_a":
			dirEntry = &entries[i]
		case "main.go":
			textEntry = &entries[i]
		case "binary.bin":
			binEntry = &entries[i]
		case "broken_link":
			brokenEntry = &entries[i]
		}
	}

	if dirEntry == nil || textEntry == nil || binEntry == nil || brokenEntry == nil {
		t.Fatalf("missing expected entries in test directory")
	}

	// Directory preview
	dirLines := LoadPreview(dirEntry, 40, 15, th, profile)
	if len(dirLines) == 0 || !strings.Contains(dirLines[0], "Directory: sub_a") {
		t.Errorf("unexpected dir preview: %v", dirLines)
	}

	// Text file preview
	textLines := LoadPreview(textEntry, 40, 15, th, profile)
	if len(textLines) == 0 || !strings.Contains(textLines[0], "package main") {
		t.Errorf("unexpected text preview: %v", textLines)
	}

	// Binary file preview
	binLines := LoadPreview(binEntry, 40, 15, th, profile)
	if len(binLines) == 0 || !strings.Contains(binLines[0], "Binary file") {
		t.Errorf("unexpected binary preview: %v", binLines)
	}

	// Broken link preview
	brokenLines := LoadPreview(brokenEntry, 40, 15, th, profile)
	if len(brokenLines) == 0 || !strings.Contains(brokenLines[0], "Broken symlink") {
		t.Errorf("unexpected broken link preview: %v", brokenLines)
	}
}

func TestInteractiveModelNavigation(t *testing.T) {
	tmpDir := setupTestDir(t)
	th := theme.Get("default")
	profile := terminal.ColorNone

	m, err := NewModel(tmpDir, 80, 24, false, th, profile, true)
	if err != nil {
		t.Fatalf("NewModel failed: %v", err)
	}

	if len(m.Filtered) == 0 {
		t.Fatalf("expected filtered entries")
	}

	// Cursor down
	initialCursor := m.Cursor
	m.MoveCursor(1)
	if m.Cursor != initialCursor+1 {
		t.Errorf("expected cursor %d, got %d", initialCursor+1, m.Cursor)
	}

	// Cursor up
	m.MoveCursor(-1)
	if m.Cursor != initialCursor {
		t.Errorf("expected cursor to return to %d, got %d", initialCursor, m.Cursor)
	}

	// Clamping at top
	m.MoveCursor(-10)
	if m.Cursor != 0 {
		t.Errorf("expected cursor clamped to 0, got %d", m.Cursor)
	}

	// Home / End
	m.MoveEnd()
	if m.Cursor != len(m.Filtered)-1 {
		t.Errorf("expected cursor at end (%d), got %d", len(m.Filtered)-1, m.Cursor)
	}
	m.MoveHome()
	if m.Cursor != 0 {
		t.Errorf("expected cursor at top (0), got %d", m.Cursor)
	}

	// Enter directory (first item should be sub_a or sub_b)
	entered, err := m.EnterSelected()
	if err != nil {
		t.Fatalf("EnterSelected failed: %v", err)
	}
	if !entered {
		t.Fatalf("expected to enter directory")
	}
	if m.CurrentDir == tmpDir {
		t.Errorf("expected CurrentDir to change")
	}

	// Go to parent
	err = m.GoToParent()
	if err != nil {
		t.Fatalf("GoToParent failed: %v", err)
	}
	if m.CurrentDir != tmpDir {
		t.Errorf("expected CurrentDir to return to tmpDir, got %s", m.CurrentDir)
	}

	// Filtering
	m.FilterActive = true
	m.FilterQuery = "main"
	m.ApplyFilter()
	if len(m.Filtered) == 0 {
		t.Errorf("expected filtered items for query 'main'")
	}
	for _, e := range m.Filtered {
		if !strings.Contains(strings.ToLower(e.Name), "main") {
			t.Errorf("unexpected entry %s in filtered list for 'main'", e.Name)
		}
	}

	// Clear filter
	m.FilterQuery = ""
	m.ApplyFilter()
	if len(m.Filtered) != len(m.Entries) {
		t.Errorf("expected all entries restored after filter clear")
	}

	// Help toggle
	m.ToggleHelp()
	if !m.HelpActive {
		t.Errorf("expected HelpActive true")
	}
	m.ToggleHelp()
	if m.HelpActive {
		t.Errorf("expected HelpActive false")
	}
}

func TestInteractiveRenderResponsive(t *testing.T) {
	tmpDir := setupTestDir(t)
	th := theme.Get("default")
	profile := terminal.ColorNone

	// Standard terminal (80x24)
	mStandard, err := NewModel(tmpDir, 80, 24, false, th, profile, true)
	if err != nil {
		t.Fatal(err)
	}
	standardFrame := Render(mStandard)
	if !strings.Contains(standardFrame, "[nova]") {
		t.Errorf("expected header banner in standard frame")
	}
	if !strings.Contains(standardFrame, "PREVIEW") {
		t.Errorf("expected PREVIEW header in standard frame")
	}

	// Narrow terminal (40x10)
	mNarrow, err := NewModel(tmpDir, 40, 10, false, th, profile, true)
	if err != nil {
		t.Fatal(err)
	}
	narrowFrame := Render(mNarrow)
	if !strings.Contains(narrowFrame, "[nova]") {
		t.Errorf("expected header banner in narrow frame")
	}

	// Ultrawide terminal (200x50)
	mWide, err := NewModel(tmpDir, 200, 50, false, th, profile, true)
	if err != nil {
		t.Fatal(err)
	}
	wideFrame := Render(mWide)
	if !strings.Contains(wideFrame, "[nova]") {
		t.Errorf("expected header banner in wide frame")
	}

	// Help modal overlay
	mStandard.ToggleHelp()
	helpFrame := Render(mStandard)
	if !strings.Contains(helpFrame, "NOVA INTERACTIVE NAVIGATOR") {
		t.Errorf("expected help modal in rendered frame")
	}
}

func TestInteractiveRunLoop(t *testing.T) {
	tmpDir := setupTestDir(t)
	th := theme.Get("default")
	profile := terminal.ColorNone

	m, err := NewModel(tmpDir, 80, 24, false, th, profile, true)
	if err != nil {
		t.Fatal(err)
	}

	// Simulate sequence:
	// 'j' (down), 'k' (up), '?' (open help), '?' (close help), '/' (start filter), 'm', 'a', 'i', 'n', '\r' (commit filter), 'q' (quit)
	inputSequence := []byte{'j', 'k', '?', '?', '/', 'm', 'a', 'i', 'n', '\r', 'q'}
	in := bytes.NewReader(inputSequence)
	var out bytes.Buffer

	err = RunLoop(m, in, &out)
	if err != nil {
		t.Fatalf("RunLoop failed: %v", err)
	}

	// Assert that filter was applied and model is in consistent state
	if m.FilterQuery != "main" {
		t.Errorf("expected filter query 'main', got %q", m.FilterQuery)
	}
	if out.Len() == 0 {
		t.Errorf("expected rendered frames in output")
	}
}
