package interactive

import (
	"archive/zip"
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
	entries, err := ReadDir(tmpDir, false, true, "name")
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
	hiddenEntries, err := ReadDir(tmpDir, true, true, "name")
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

	entries, err := ReadDir(tmpDir, false, true, "name")
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
	dirLines := LoadPreview(dirEntry, 40, 15, th, profile, false)
	if len(dirLines) == 0 || !strings.Contains(dirLines[0], "Directory: sub_a") {
		t.Errorf("unexpected dir preview: %v", dirLines)
	}

	// Text file preview
	textLines := LoadPreview(textEntry, 40, 15, th, profile, false)
	if len(textLines) == 0 || !strings.Contains(textLines[0], "[GO]") || !strings.Contains(strings.Join(textLines, "\n"), "package main") {
		t.Errorf("unexpected text preview: %v", textLines)
	}

	// Binary file preview
	binLines := LoadPreview(binEntry, 40, 15, th, profile, false)
	if len(binLines) == 0 || !strings.Contains(binLines[0], "Binary file") {
		t.Errorf("unexpected binary preview: %v", binLines)
	}

	// Broken link preview
	brokenLines := LoadPreview(brokenEntry, 40, 15, th, profile, false)
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

func TestInteractiveActions(t *testing.T) {
	tmpDir := setupTestDir(t)
	th := theme.Get("default")
	profile := terminal.ColorNone

	m, err := NewModel(tmpDir, 80, 24, false, th, profile, true)
	if err != nil {
		t.Fatal(err)
	}

	// 1. Test space (toggle selection)
	in1 := bytes.NewReader([]byte{' ', 'c', 'q'})
	var out1 bytes.Buffer
	if err := RunLoop(m, in1, &out1); err != nil {
		t.Fatal(err)
	}
	if len(m.SelectedPaths) != 1 {
		t.Errorf("expected 1 selected path, got %d", len(m.SelectedPaths))
	}
	if !strings.Contains(m.ActionMessage, "Copied") {
		t.Errorf("expected copied action message, got %q", m.ActionMessage)
	}

	// 2. Test delete cancel
	in2 := bytes.NewReader([]byte{'d', 'n', 'q'})
	var out2 bytes.Buffer
	if err := RunLoop(m, in2, &out2); err != nil {
		t.Fatal(err)
	}
	if m.ConfirmDelete {
		t.Errorf("expected ConfirmDelete to be false after cancel")
	}
	if m.ActionMessage != "Delete canceled" {
		t.Errorf("expected 'Delete canceled', got %q", m.ActionMessage)
	}

	// 3. Test rename
	// Create a test file specifically for rename
	testFile := filepath.Join(tmpDir, "to_rename.txt")
	_ = os.WriteFile(testFile, []byte("test"), 0644)
	_ = m.LoadCurrentDir()

	// Find the index of to_rename.txt
	for i, e := range m.Filtered {
		if e.Name == "to_rename.txt" {
			m.Cursor = i
			break
		}
	}

	// Trigger rename: 'r', then type 'n', 'e', 'w', '.', 't', 'x', 't', then Enter '\n', then 'q'
	renameSeq := append([]byte{'r', 27}, 'q') // cancel first
	in3 := bytes.NewReader(renameSeq)
	var out3 bytes.Buffer
	_ = RunLoop(m, in3, &out3)
	if m.ActionMessage != "Rename canceled" {
		t.Errorf("expected 'Rename canceled', got %q", m.ActionMessage)
	}

	// 4. Test new file creation: 'n', type 'c', 'r', 'e', 'a', 't', 'e', 'd', '.', 't', 'x', 't', Enter, 'q'
	newFileSeq := []byte{'n', 'c', 'r', 'e', 'a', 't', 'e', 'd', '.', 't', 'x', 't', '\n', 'q'}
	in4 := bytes.NewReader(newFileSeq)
	var out4 bytes.Buffer
	_ = RunLoop(m, in4, &out4)
	if _, err := os.Stat(filepath.Join(tmpDir, "created.txt")); os.IsNotExist(err) {
		t.Errorf("expected created.txt to be created via 'n' action")
	}

	// 5. Test preview collapse: 'p'
	in5 := bytes.NewReader([]byte{'p', 'q'})
	var out5 bytes.Buffer
	_ = RunLoop(m, in5, &out5)
	if !m.PreviewCollapsed {
		t.Errorf("expected PreviewCollapsed to be true after pressing 'p'")
	}
}

func TestInteractiveThemeSortHex(t *testing.T) {
	tmpDir := setupTestDir(t)
	th := theme.Get("default")
	m, err := NewModel(tmpDir, 80, 24, false, th, terminal.ColorTrueColor, true)
	if err != nil {
		t.Fatal(err)
	}

	// 1. Test Theme Cycling: 't'
	initialTheme := m.Theme.Name
	in1 := bytes.NewReader([]byte{'t', 'q'})
	var out1 bytes.Buffer
	_ = RunLoop(m, in1, &out1)
	if m.Theme.Name == initialTheme {
		t.Errorf("expected theme to change from %s, got %s", initialTheme, m.Theme.Name)
	}

	// 2. Test Sort Mode Cycling: 's'
	initialSort := m.SortMode
	in2 := bytes.NewReader([]byte{'s', 'q'})
	var out2 bytes.Buffer
	_ = RunLoop(m, in2, &out2)
	if m.SortMode == initialSort {
		t.Errorf("expected sort mode to cycle, got %s", m.SortMode)
	}

	// 3. Test Hex Toggle: 'x'
	in3 := bytes.NewReader([]byte{'x', 'q'})
	var out3 bytes.Buffer
	_ = RunLoop(m, in3, &out3)
	if !m.HexMode {
		t.Errorf("expected HexMode to be true after pressing 'x'")
	}
}

func TestInteractivePaletteAndBookmarks(t *testing.T) {
	tmpDir := setupTestDir(t)
	th := theme.Get("default")
	m, err := NewModel(tmpDir, 80, 24, false, th, terminal.ColorTrueColor, true)
	if err != nil {
		t.Fatal(err)
	}

	// 1. Test Bookmark adding: 'b'
	in1 := bytes.NewReader([]byte{'b', 'q'})
	var out1 bytes.Buffer
	_ = RunLoop(m, in1, &out1)
	if len(m.Bookmarks) != 1 {
		t.Errorf("expected 1 bookmark, got %d", len(m.Bookmarks))
	}

	// 2. Test Command Palette execution: ':', type 'r', 'e', 'l', 'o', 'a', 'd', Enter, 'q'
	cmdSeq := append([]byte{':', 'r', 'e', 'l', 'o', 'a', 'd', '\n'}, 'q')
	in2 := bytes.NewReader(cmdSeq)
	var out2 bytes.Buffer
	_ = RunLoop(m, in2, &out2)
	if !strings.Contains(m.ActionMessage, "Reloaded") {
		t.Errorf("expected 'Reloaded' message from :reload, got: %s", m.ActionMessage)
	}
}

func TestInteractiveInspectorAndQuickHash(t *testing.T) {
	tmpDir := setupTestDir(t)
	th := theme.Get("default")
	m, err := NewModel(tmpDir, 80, 24, false, th, terminal.ColorTrueColor, true)
	if err != nil {
		t.Fatal(err)
	}

	// 1. Test Inspector toggle: 'i'
	in1 := bytes.NewReader([]byte{'i', 'q'})
	var out1 bytes.Buffer
	_ = RunLoop(m, in1, &out1)
	if !m.InspectorActive {
		t.Errorf("expected InspectorActive to be true after pressing 'i'")
	}

	// 2. Test Inspector close: 'i' again
	in2 := bytes.NewReader([]byte{'i', 'q'})
	var out2 bytes.Buffer
	_ = RunLoop(m, in2, &out2)
	if m.InspectorActive {
		t.Errorf("expected InspectorActive to be false after toggling 'i' again")
	}

	// 3. Test QuickHash: '#'
	// Select a file first
	for i, e := range m.Entries {
		if !e.IsDir {
			m.Cursor = i
			break
		}
	}
	in3 := bytes.NewReader([]byte{'#', 'q'})
	var out3 bytes.Buffer
	_ = RunLoop(m, in3, &out3)
	if !strings.Contains(m.ActionMessage, "SHA256:") {
		t.Errorf("expected ActionMessage to contain SHA256:, got: %s", m.ActionMessage)
	}
}

func TestInteractiveDiffMode(t *testing.T) {
	tmpDir := setupTestDir(t)
	th := theme.Get("default")
	m, err := NewModel(tmpDir, 80, 24, false, th, terminal.ColorTrueColor, true)
	if err != nil {
		t.Fatal(err)
	}

	// Test 'D' toggles DiffMode
	in := bytes.NewReader([]byte{'D', 'q'})
	var out bytes.Buffer
	_ = RunLoop(m, in, &out)
	if !m.DiffMode {
		t.Errorf("expected DiffMode to be true after pressing 'D'")
	}
}

func TestInteractiveFuzzyFinder(t *testing.T) {
	tmpDir := setupTestDir(t)
	th := theme.Get("default")
	m, err := NewModel(tmpDir, 80, 24, false, th, terminal.ColorNone, true)
	if err != nil {
		t.Fatal(err)
	}

	// 1. Test opening fuzzy finder: 'f'
	in := bytes.NewReader([]byte{'f', 'q'})
	var out bytes.Buffer
	_ = RunLoop(m, in, &out)
	if !m.FuzzyActive {
		t.Errorf("expected FuzzyActive true after pressing 'f'")
	}

	// 2. Render while FuzzyActive
	frame := Render(m)
	if !strings.Contains(frame, "Fuzzy File Finder") {
		t.Errorf("expected Fuzzy File Finder in frame, got: %s", frame)
	}

	// 3. Test typing and selection jump
	m.FuzzyQuery = "sub"
	m.UpdateFuzzyResults()
	if len(m.FuzzyResults) == 0 {
		t.Errorf("expected matches for 'sub'")
	}

	m.SelectFuzzyResult()
	if m.FuzzyActive {
		t.Errorf("expected FuzzyActive false after SelectFuzzyResult")
	}
	if !strings.Contains(m.ActionMessage, "Jumped to:") {
		t.Errorf("expected ActionMessage to contain Jumped to:, got: %s", m.ActionMessage)
	}
}

func TestInteractiveZipPreview(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "bundle.zip")

	// Create a test zip file
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w1, err := zw.Create("hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w1.Write([]byte("Hello, world! Welcome to Nova interactive zip preview!"))
	w2, err := zw.Create("folder/test.go")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w2.Write([]byte("package test\n"))
	_ = zw.Close()
	_ = f.Close()

	fi, err := os.Stat(zipPath)
	if err != nil {
		t.Fatal(err)
	}

	entry := &Entry{
		Name:    "bundle.zip",
		Path:    zipPath,
		Size:    fi.Size(),
		ModTime: fi.ModTime(),
	}

	th := theme.Get("default")
	lines := LoadPreview(entry, 80, 20, th, terminal.ColorNone, false)
	joined := strings.Join(lines, "\n")

	if !strings.Contains(joined, "Archive: bundle.zip") {
		t.Errorf("expected zip archive banner in preview, got: %s", joined)
	}
	if !strings.Contains(joined, "hello.txt") || !strings.Contains(joined, "test.go") {
		t.Errorf("expected zip file contents in preview, got: %s", joined)
	}
}

func TestInteractiveCalcCommand(t *testing.T) {
	tmpDir := setupTestDir(t)
	th := theme.Get("default")
	m, err := NewModel(tmpDir, 80, 24, false, th, terminal.ColorNone, true)
	if err != nil {
		t.Fatal(err)
	}

	// 1. Valid calc expression
	err = m.ExecuteCommand(":calc 2^8 + 100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(m.ActionMessage, "356") || !strings.Contains(m.ActionMessage, "0x164") {
		t.Errorf("expected 356 and 0x164 in ActionMessage, got: %q", m.ActionMessage)
	}

	// 2. Syntax error handling
	err = m.ExecuteCommand(":calc 10 + * 2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(m.ActionMessage, "Calc error:") {
		t.Errorf("expected Calc error in ActionMessage, got: %q", m.ActionMessage)
	}

	// 3. History and Serve command palette tips
	_ = m.ExecuteCommand(":history")
	if !strings.Contains(m.ActionMessage, "nova history") {
		t.Errorf("expected history tip, got: %q", m.ActionMessage)
	}

	_ = m.ExecuteCommand(":serve")
	if !strings.Contains(m.ActionMessage, "nova serve") {
		t.Errorf("expected serve tip, got: %q", m.ActionMessage)
	}
}

func TestInteractiveMarkdownPreview(t *testing.T) {
	tmpDir := t.TempDir()
	mdPath := filepath.Join(tmpDir, "docs.md")
	content := "# Project Roadmap\n\n## Core Engine\n- [x] Task done\n- [ ] Task pending\n- Bullet item\n> Note here\n"
	if err := os.WriteFile(mdPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	fi, err := os.Stat(mdPath)
	if err != nil {
		t.Fatal(err)
	}

	entry := &Entry{
		Name:    "docs.md",
		Path:    mdPath,
		Size:    fi.Size(),
		ModTime: fi.ModTime(),
	}

	th := theme.Get("default")
	lines := LoadPreview(entry, 80, 20, th, terminal.ColorNone, false)
	joined := strings.Join(lines, "\n")

	if !strings.Contains(joined, "MARKDOWN DOC") {
		t.Errorf("expected MARKDOWN DOC badge in preview, got: %s", joined)
	}
	if !strings.Contains(joined, "ROADMAP") || !strings.Contains(joined, "Core Engine") {
		t.Errorf("expected headings in markdown preview, got: %s", joined)
	}
	if !strings.Contains(joined, "[✓]") || !strings.Contains(joined, "[ ]") {
		t.Errorf("expected checkboxes in markdown preview, got: %s", joined)
	}
}

func TestInteractiveBookmarkNumberJumpAndPalette(t *testing.T) {
	tmpDir := setupTestDir(t)
	th := theme.Get("default")
	m, err := NewModel(tmpDir, 80, 24, false, th, terminal.ColorNone, true)
	if err != nil {
		t.Fatal(err)
	}

	// 1. Add bookmarks
	m.AddBookmark()
	subA := filepath.Join(tmpDir, "sub_a")
	m.CurrentDir = subA
	m.AddBookmark()

	if len(m.Bookmarks) != 2 {
		t.Fatalf("expected 2 bookmarks, got %d", len(m.Bookmarks))
	}

	// 2. Jump to bookmark 1
	m.JumpBookmarkIndex(0)
	if m.CurrentDir != tmpDir {
		t.Errorf("expected CurrentDir %q after jump 0, got %q", tmpDir, m.CurrentDir)
	}

	// 3. Jump to bookmark 2
	m.JumpBookmarkIndex(1)
	if m.CurrentDir != subA {
		t.Errorf("expected CurrentDir %q after jump 1, got %q", subA, m.CurrentDir)
	}

	// 4. Test v1.9 command palette tips
	_ = m.ExecuteCommand(":md")
	if !strings.Contains(m.ActionMessage, "nova md") {
		t.Errorf("expected md tip, got: %q", m.ActionMessage)
	}

	_ = m.ExecuteCommand(":note")
	if !strings.Contains(m.ActionMessage, "nova note") {
		t.Errorf("expected note tip, got: %q", m.ActionMessage)
	}

	_ = m.ExecuteCommand(":cert")
	if !strings.Contains(m.ActionMessage, "nova cert") {
		t.Errorf("expected cert tip, got: %q", m.ActionMessage)
	}
}




