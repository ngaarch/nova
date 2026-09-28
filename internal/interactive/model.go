package interactive

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	calccmd "nova/internal/commands/calc"
	"nova/internal/git"
	"nova/internal/terminal"
	"nova/internal/theme"
)

// Model represents the reactive state of the interactive navigator.
type Model struct {
	mu               sync.RWMutex
	CurrentDir       string
	Entries          []Entry
	Filtered         []Entry
	Cursor           int
	ScrollOffset     int
	FilterActive     bool
	FilterQuery      string
	HelpActive       bool
	InspectorActive  bool
	ShowHidden       bool
	Width            int
	Height           int
	Theme            *theme.Theme
	ThemeNames       []string
	ThemeIndex       int
	SortMode         string // "name", "size", "time", "ext"
	HexMode          bool
	DiffMode         bool
	Profile          terminal.ColorProfile
	UnicodeSupported bool
	PreviewLines     []string
	Git              *git.RepoStatus
	SelectedPaths    map[string]bool
	ActionMessage    string
	ActionTime       time.Time
	ConfirmDelete    bool
	RenameActive     bool
	RenameInput      string
	NewFileActive    bool
	NewFileInput     string
	NewFolderActive      bool
	NewFolderInput       string
	CommandPaletteActive bool
	CommandInput         string
	Bookmarks            []string
	BookmarkIndex        int
	PreviewCollapsed     bool
	Tick                 int
	FuzzyActive          bool
	FuzzyQuery           string
	FuzzyResults         []string
	FuzzyIndex           int
	EditorRunner         func(path string) error
}

// Lock acquires exclusive write lock on model state.
func (m *Model) Lock() {
	m.mu.Lock()
}

// Unlock releases exclusive write lock on model state.
func (m *Model) Unlock() {
	m.mu.Unlock()
}

// RLock acquires shared read lock on model state.
func (m *Model) RLock() {
	m.mu.RLock()
}

// RUnlock releases shared read lock on model state.
func (m *Model) RUnlock() {
	m.mu.RUnlock()
}

// NewModel constructs an initialized Model starting at dir.
func NewModel(dir string, width, height int, showHidden bool, th *theme.Theme, profile terminal.ColorProfile, unicodeSupported bool) (*Model, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve path %q: %w", dir, err)
	}

	fi, err := os.Stat(absDir)
	if err != nil {
		return nil, fmt.Errorf("stat path %q: %w", absDir, err)
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("path %q is not a directory", absDir)
	}

	if width < 20 {
		width = 80
	}
	if height < 8 {
		height = 24
	}

	allThemes := theme.ListThemes()
	thIdx := 0
	for i, name := range allThemes {
		if th != nil && name == th.Name {
			thIdx = i
			break
		}
	}

	m := &Model{
		CurrentDir:       absDir,
		ShowHidden:       showHidden,
		Width:            width,
		Height:           height,
		Theme:            th,
		ThemeNames:       allThemes,
		ThemeIndex:       thIdx,
		SortMode:         "name",
		Profile:          profile,
		UnicodeSupported: unicodeSupported,
		SelectedPaths:    make(map[string]bool),
	}

	if err := m.LoadCurrentDir(); err != nil {
		return nil, err
	}

	return m, nil
}

// ToggleSelect toggles selection of the current entry.
func (m *Model) ToggleSelect() {
	entry := m.CurrentEntry()
	if entry == nil {
		return
	}
	if m.SelectedPaths == nil {
		m.SelectedPaths = make(map[string]bool)
	}
	if m.SelectedPaths[entry.Path] {
		delete(m.SelectedPaths, entry.Path)
	} else {
		m.SelectedPaths[entry.Path] = true
	}
}

// ClearSelection unchecks all selected items.
func (m *Model) ClearSelection() {
	m.SelectedPaths = make(map[string]bool)
}

// SetActionMessage displays a temporary toast status message.
func (m *Model) SetActionMessage(msg string) {
	m.ActionMessage = msg
	m.ActionTime = time.Now()
}

// CycleTheme switches to the next available theme in real-time.
func (m *Model) CycleTheme() {
	if len(m.ThemeNames) == 0 {
		return
	}
	m.ThemeIndex = (m.ThemeIndex + 1) % len(m.ThemeNames)
	m.Theme = theme.Get(m.ThemeNames[m.ThemeIndex])
	m.SetActionMessage("Theme: " + m.Theme.Name)
	m.UpdatePreview()
}

// CycleSort cycles through sorting modes (name -> size -> time -> ext -> name).
func (m *Model) CycleSort() {
	switch m.SortMode {
	case "name":
		m.SortMode = "size"
	case "size":
		m.SortMode = "time"
	case "time":
		m.SortMode = "ext"
	default:
		m.SortMode = "name"
	}
	SortEntries(m.Entries, m.SortMode)
	m.ApplyFilter()
	m.SetActionMessage("Sort: " + m.SortMode)
}

// ToggleHex toggles hex dump mode in the preview pane.
func (m *Model) ToggleHex() {
	m.HexMode = !m.HexMode
	m.UpdatePreview()
	if m.HexMode {
		m.SetActionMessage("Hex inspector: ON")
	} else {
		m.SetActionMessage("Hex inspector: OFF")
	}
}

// AddBookmark adds the current directory to the quick bookmarks list.
func (m *Model) AddBookmark() {
	for _, b := range m.Bookmarks {
		if b == m.CurrentDir {
			m.SetActionMessage("Already bookmarked: " + filepath.Base(m.CurrentDir))
			return
		}
	}
	m.Bookmarks = append(m.Bookmarks, m.CurrentDir)
	m.SetActionMessage(fmt.Sprintf("Bookmarked [%d]: %s", len(m.Bookmarks), filepath.Base(m.CurrentDir)))
}

// JumpNextBookmark cycles navigation to the next saved bookmark.
func (m *Model) JumpNextBookmark() {
	if len(m.Bookmarks) == 0 {
		m.SetActionMessage("No bookmarks saved (press 'b' to add)")
		return
	}
	m.BookmarkIndex = (m.BookmarkIndex + 1) % len(m.Bookmarks)
	target := m.Bookmarks[m.BookmarkIndex]
	m.CurrentDir = target
	m.FilterQuery = ""
	m.FilterActive = false
	m.Cursor = 0
	m.ScrollOffset = 0
	_ = m.LoadCurrentDir()
	m.SetActionMessage(fmt.Sprintf("Jumped to bookmark [%d/%d]: %s", m.BookmarkIndex+1, len(m.Bookmarks), filepath.Base(target)))
}

// ExecuteCommand runs a command entered via the command palette.
func (m *Model) ExecuteCommand(rawCmd string) error {
	trimmed := strings.TrimSpace(rawCmd)
	if trimmed == "" {
		return nil
	}
	parts := strings.Fields(trimmed)
	cmd := strings.ToLower(parts[0])

	cleanCmd := strings.TrimPrefix(cmd, ":")
	switch cleanCmd {
	case "q", "quit", "exit":
		return fmt.Errorf("QUIT")
	case "w", "reload", "r":
		_ = m.LoadCurrentDir()
		m.SetActionMessage("Reloaded directory")
	case "help", "?":
		m.ToggleHelp()
	case "theme":
		if len(parts) > 1 {
			thName := parts[1]
			m.Theme = theme.Get(thName)
			m.SetActionMessage("Theme set to: " + m.Theme.Name)
			m.UpdatePreview()
		} else {
			m.CycleTheme()
		}
	case ":sort", "sort":
		if len(parts) > 1 {
			m.SortMode = parts[1]
			SortEntries(m.Entries, m.SortMode)
			m.ApplyFilter()
			m.SetActionMessage("Sort set to: " + m.SortMode)
		} else {
			m.CycleSort()
		}
	case ":mkdir", "mkdir":
		if len(parts) > 1 {
			newPath := filepath.Join(m.CurrentDir, parts[1])
			if err := os.MkdirAll(newPath, 0755); err != nil {
				m.SetActionMessage("Error: " + err.Error())
			} else {
				m.SetActionMessage("Created: " + parts[1])
				_ = m.LoadCurrentDir()
			}
		}
	case ":touch", "touch":
		if len(parts) > 1 {
			newPath := filepath.Join(m.CurrentDir, parts[1])
			if err := os.WriteFile(newPath, []byte(""), 0644); err != nil {
				m.SetActionMessage("Error: " + err.Error())
			} else {
				m.SetActionMessage("Created: " + parts[1])
				_ = m.LoadCurrentDir()
			}
		}
	case "inspect", "info", "i":
		m.ToggleInspector()
	case "hash", "sha256":
		m.QuickHash()
	case "hex", "dump":
		m.ToggleHex()
	case "diff":
		m.ToggleDiff()
	case "find", "fuzzy":
		m.OpenFuzzyFinder()
	case "bookmark", "bm":
		m.AddBookmark()
	case "jump":
		m.JumpNextBookmark()
	case "top", "proc":
		m.SetActionMessage("Tip: run 'nova top' in terminal for real-time process monitor")
	case "net", "ping":
		m.SetActionMessage("Tip: run 'nova net' in terminal for network latency diagnostics")
	case "run", "tasks":
		m.SetActionMessage("Tip: run 'nova run' in terminal to orchestrate project tasks")
	case "calc", "math", "eval", "expr":
		if len(parts) > 1 {
			expr := strings.TrimSpace(trimmed[len(parts[0]):])
			val, err := calccmd.Evaluate(expr)
			if err != nil {
				m.SetActionMessage("Calc error: " + err.Error())
			} else {
				res := calccmd.FormatResult(expr, val, calccmd.Options{Precision: 6})
				extra := ""
				if res.IsInteger && res.Hex != "" {
					extra = fmt.Sprintf(" [hex: %s | bin: %s]", res.Hex, res.Binary)
					if res.HumanBytes != "" {
						extra += fmt.Sprintf(" (%s)", res.HumanBytes)
					}
				}
				m.SetActionMessage(fmt.Sprintf("%s = %s%s", expr, res.Formatted, extra))
			}
		} else {
			m.SetActionMessage("Usage: :calc <expression> (e.g. :calc 2^16 + 1024)")
		}
	case "history", "hist":
		m.SetActionMessage("Tip: run 'nova history' in terminal for shell productivity analytics")
	case "serve", "http":
		m.SetActionMessage("Tip: run 'nova serve' in terminal for zero-config web server with QR code")
	default:
		m.SetActionMessage("Unknown command: " + parts[0])
	}
	return nil
}

// LoadCurrentDir reloads files in the current working directory.
func (m *Model) LoadCurrentDir() error {
	entries, err := ReadDir(m.CurrentDir, m.ShowHidden, m.UnicodeSupported, m.SortMode)
	if err != nil {
		return err
	}

	m.Git, _ = git.GetRepoStatus(m.CurrentDir, 50*time.Millisecond)
	if m.Git != nil {
		for i := range entries {
			st := m.Git.GetStatus(entries[i].Path)
			if st != git.StatusClean {
				entries[i].GitStatus = string(st)
			}
		}
	}

	m.Entries = entries
	m.ApplyFilter()
	return nil
}

// ApplyFilter filters the loaded entries based on FilterQuery.
func (m *Model) ApplyFilter() {
	m.Filtered = FilterEntries(m.Entries, m.FilterQuery)

	if m.Cursor >= len(m.Filtered) {
		m.Cursor = len(m.Filtered) - 1
	}
	if m.Cursor < 0 {
		m.Cursor = 0
	}

	m.AdjustScroll()
	m.UpdatePreview()
}

// AdjustScroll keeps the cursor visible within the left list pane.
func (m *Model) AdjustScroll() {
	listHeight := m.ListHeight()
	if listHeight <= 0 {
		return
	}

	if m.Cursor < m.ScrollOffset {
		m.ScrollOffset = m.Cursor
	} else if m.Cursor >= m.ScrollOffset+listHeight {
		m.ScrollOffset = m.Cursor - listHeight + 1
	}

	if m.ScrollOffset < 0 {
		m.ScrollOffset = 0
	}
}

// ListHeight returns the number of visible rows available for directory items.
func (m *Model) ListHeight() int {
	// Top header: 2 lines
	// Bottom status/filter: 2 lines
	// Borders: 2 lines
	h := m.Height - 6
	if h < 2 {
		return 2
	}
	return h
}

// MoveCursor adjusts the cursor by delta rows.
func (m *Model) MoveCursor(delta int) {
	if len(m.Filtered) == 0 {
		return
	}

	newCursor := m.Cursor + delta
	if newCursor < 0 {
		newCursor = 0
	}
	if newCursor >= len(m.Filtered) {
		newCursor = len(m.Filtered) - 1
	}

	if newCursor != m.Cursor {
		m.Cursor = newCursor
		m.AdjustScroll()
		m.UpdatePreview()
	}
}

// PageMove jumps the cursor by an entire page.
func (m *Model) PageMove(delta int) {
	pageSize := m.ListHeight()
	m.MoveCursor(delta * pageSize)
}

// MoveHome jumps cursor to top.
func (m *Model) MoveHome() {
	m.Cursor = 0
	m.AdjustScroll()
	m.UpdatePreview()
}

// MoveEnd jumps cursor to bottom.
func (m *Model) MoveEnd() {
	if len(m.Filtered) > 0 {
		m.Cursor = len(m.Filtered) - 1
		m.AdjustScroll()
		m.UpdatePreview()
	}
}

// CurrentEntry returns the currently highlighted entry or nil.
func (m *Model) CurrentEntry() *Entry {
	if len(m.Filtered) == 0 || m.Cursor < 0 || m.Cursor >= len(m.Filtered) {
		return nil
	}
	return &m.Filtered[m.Cursor]
}

// EnterSelected navigates into the highlighted item if it is a directory.
func (m *Model) EnterSelected() (bool, error) {
	entry := m.CurrentEntry()
	if entry == nil || !entry.IsDir {
		return false, nil
	}

	m.CurrentDir = entry.Path
	m.FilterQuery = ""
	m.FilterActive = false
	m.Cursor = 0
	m.ScrollOffset = 0

	err := m.LoadCurrentDir()
	return true, err
}

// GoToParent navigates to the parent directory.
func (m *Model) GoToParent() error {
	parent := filepath.Dir(m.CurrentDir)
	if parent == m.CurrentDir {
		// Already at root filesystem
		return nil
	}

	oldDirName := filepath.Base(m.CurrentDir)
	m.CurrentDir = parent
	m.FilterQuery = ""
	m.FilterActive = false
	m.Cursor = 0
	m.ScrollOffset = 0

	if err := m.LoadCurrentDir(); err != nil {
		return err
	}

	// Try to restore cursor position on the directory we just navigated out of
	for i, e := range m.Filtered {
		if e.Name == oldDirName {
			m.Cursor = i
			break
		}
	}
	m.AdjustScroll()
	m.UpdatePreview()
	return nil
}

// ToggleHelp switches the help modal visibility.
func (m *Model) ToggleHelp() {
	m.HelpActive = !m.HelpActive
}

// ToggleHidden toggles display of hidden/dot files.
func (m *Model) ToggleHidden() error {
	m.ShowHidden = !m.ShowHidden
	return m.LoadCurrentDir()
}

// Resize updates the terminal dimensions.
func (m *Model) Resize(w, h int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if w < 20 {
		w = 20
	}
	if h < 8 {
		h = 8
	}
	m.Width = w
	m.Height = h
	m.AdjustScroll()
	m.UpdatePreview()
}

// UpdatePreview generates the preview lines for the currently selected entry.
func (m *Model) UpdatePreview() {
	entry := m.CurrentEntry()
	// Right pane width is roughly 50% of screen width minus divider
	paneWidth := m.Width/2 - 2
	paneHeight := m.ListHeight() + 2

	if m.DiffMode {
		m.PreviewLines = LoadDiffPreview(entry, paneWidth, paneHeight, m.Theme, m.Profile)
	} else {
		m.PreviewLines = LoadPreview(entry, paneWidth, paneHeight, m.Theme, m.Profile, m.HexMode)
	}
}

// ToggleDiff toggles git diff preview mode.
func (m *Model) ToggleDiff() {
	m.DiffMode = !m.DiffMode
	if m.DiffMode {
		m.SetActionMessage("Git diff preview enabled (press D to toggle back)")
	} else {
		m.SetActionMessage("Standard preview enabled")
	}
	m.UpdatePreview()
}

// ToggleInspector toggles the metadata inspector modal.
func (m *Model) ToggleInspector() {
	m.InspectorActive = !m.InspectorActive
	if m.InspectorActive {
		m.HelpActive = false
		m.SetActionMessage("Inspector opened (press i or Esc to close)")
	} else {
		m.SetActionMessage("")
	}
}

// QuickHash calculates SHA-256 checksum for the currently focused file and shows it in status.
func (m *Model) QuickHash() {
	entry := m.CurrentEntry()
	if entry == nil || entry.IsDir {
		m.SetActionMessage("Cannot hash directory")
		return
	}
	f, err := os.Open(entry.Path)
	if err != nil {
		m.SetActionMessage("Hash error: " + err.Error())
		return
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		m.SetActionMessage("Hash error: " + err.Error())
		return
	}
	digest := hex.EncodeToString(h.Sum(nil))
	m.SetActionMessage(fmt.Sprintf("SHA256: %.16s... (%s)", digest, entry.Name))
}

// OpenFuzzyFinder opens the fuzzy file finder overlay.
func (m *Model) OpenFuzzyFinder() {
	m.FuzzyActive = true
	m.FuzzyQuery = ""
	m.FuzzyIndex = 0
	m.UpdateFuzzyResults()
	m.HelpActive = false
	m.InspectorActive = false
}

// UpdateFuzzyResults searches current directory recursively up to depth 3 for matching files.
func (m *Model) UpdateFuzzyResults() {
	query := strings.ToLower(m.FuzzyQuery)
	var matches []string

	_ = filepath.Walk(m.CurrentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		rel, err := filepath.Rel(m.CurrentDir, path)
		if err != nil || rel == "." {
			return nil
		}
		if !m.ShowHidden && strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.Count(rel, string(filepath.Separator)) > 3 {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if query == "" || strings.Contains(strings.ToLower(rel), query) {
			matches = append(matches, rel)
			if len(matches) >= 30 {
				return filepath.SkipAll
			}
		}
		return nil
	})

	m.FuzzyResults = matches
	if m.FuzzyIndex >= len(m.FuzzyResults) {
		if len(m.FuzzyResults) > 0 {
			m.FuzzyIndex = len(m.FuzzyResults) - 1
		} else {
			m.FuzzyIndex = 0
		}
	}
}

// SelectFuzzyResult jumps to the focused fuzzy search result.
func (m *Model) SelectFuzzyResult() {
	if len(m.FuzzyResults) == 0 || m.FuzzyIndex < 0 || m.FuzzyIndex >= len(m.FuzzyResults) {
		m.FuzzyActive = false
		return
	}
	targetRel := m.FuzzyResults[m.FuzzyIndex]
	targetAbs := filepath.Join(m.CurrentDir, targetRel)
	m.FuzzyActive = false

	fi, err := os.Stat(targetAbs)
	if err != nil {
		m.SetActionMessage("Jump error: " + err.Error())
		return
	}

	if fi.IsDir() {
		m.CurrentDir = targetAbs
		_ = m.LoadCurrentDir()
	} else {
		parent := filepath.Dir(targetAbs)
		if parent != m.CurrentDir {
			m.CurrentDir = parent
			_ = m.LoadCurrentDir()
		}
		baseName := filepath.Base(targetAbs)
		for i, e := range m.Filtered {
			if e.Name == baseName {
				m.Cursor = i
				m.AdjustScroll()
				m.UpdatePreview()
				break
			}
		}
	}
	m.SetActionMessage("Jumped to: " + targetRel)
}
