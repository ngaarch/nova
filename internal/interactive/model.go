package interactive

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

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
	ShowHidden       bool
	Width            int
	Height           int
	Theme            *theme.Theme
	Profile          terminal.ColorProfile
	UnicodeSupported bool
	PreviewLines     []string
	Git              *git.RepoStatus
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

	m := &Model{
		CurrentDir:       absDir,
		ShowHidden:       showHidden,
		Width:            width,
		Height:           height,
		Theme:            th,
		Profile:          profile,
		UnicodeSupported: unicodeSupported,
	}

	if err := m.LoadCurrentDir(); err != nil {
		return nil, err
	}

	return m, nil
}

// LoadCurrentDir reloads files in the current working directory.
func (m *Model) LoadCurrentDir() error {
	entries, err := ReadDir(m.CurrentDir, m.ShowHidden, m.UnicodeSupported)
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

	m.PreviewLines = LoadPreview(entry, paneWidth, paneHeight, m.Theme, m.Profile)
}
