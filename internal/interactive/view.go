package interactive

import (
	"fmt"
	"strings"

	"nova/internal/git"
	"nova/internal/renderer"
	"nova/internal/theme"
)

// Render generates the complete terminal screen frame for the model.
func Render(m *Model) string {
	width := m.Width
	height := m.Height
	if width < 20 || height < 6 {
		return "Terminal window too small\r\n"
	}

	var screen []string

	// 1. Header (Line 0)
	header := renderHeader(m, width)
	screen = append(screen, header)

	// 2. Pane column widths
	leftWidth := width/2 - 1
	if leftWidth < 10 {
		leftWidth = 10
	}
	rightWidth := width - leftWidth - 1 // 1 column for divider
	if rightWidth < 5 {
		rightWidth = 5
	}

	// Pane Headers (Line 1)
	leftHeader := renderer.PadRight("  NAME", leftWidth)
	rightHeader := renderer.PadRight(" PREVIEW", rightWidth)
	divChar := "│"
	if !m.UnicodeSupported {
		divChar = "|"
	}
	paneHeaderLine := m.Theme.Format(theme.RoleMuted, leftHeader+divChar+rightHeader, m.Profile)
	screen = append(screen, paneHeaderLine)

	// 3. Middle Content Rows (Lines 2 to height - 3)
	contentRows := height - 4
	if contentRows < 1 {
		contentRows = 1
	}

	for r := 0; r < contentRows; r++ {
		// Left pane cell
		leftCell := renderLeftCell(m, r, leftWidth)

		// Right pane cell
		rightCell := ""
		if r < len(m.PreviewLines) {
			rightCell = m.PreviewLines[r]
		}
		paddedRight := padOrTruncate(rightCell, rightWidth)

		rowLine := leftCell + m.Theme.Format(theme.RoleMuted, divChar, m.Profile) + paddedRight
		screen = append(screen, rowLine)
	}

	// 4. Status Bar (Line height - 2)
	statusBar := renderStatusBar(m, width)
	screen = append(screen, statusBar)

	// 5. Shortcut Footer (Line height - 1)
	footer := renderFooter(m, width)
	screen = append(screen, footer)

	// If Help modal is active, overlay it on top of the rendered screen
	if m.HelpActive {
		screen = overlayHelpModal(m, screen, width, height)
	}

	// Return full frame with ANSI home cursor
	return "\x1b[H" + strings.Join(screen, "\r\n")
}

func renderHeader(m *Model, width int) string {
	countStr := fmt.Sprintf("(%d items)", len(m.Filtered))
	if m.FilterQuery != "" {
		countStr = fmt.Sprintf("(%d/%d items, filter: %q)", len(m.Filtered), len(m.Entries), m.FilterQuery)
	}

	badge := m.Theme.Format(theme.RoleAccent, "[nova] ", m.Profile)
	pathStr := m.Theme.Format(theme.RoleDirectory, m.CurrentDir, m.Profile)
	itemsBadge := m.Theme.Format(theme.RoleMuted, " "+countStr, m.Profile)

	gitBadge := ""
	if m.Git != nil && m.Git.Branch != "" {
		gitBadge = " " + git.FormatBranch(m.Git.Branch, m.Git.IsDetached, m.UnicodeSupported, m.Theme, m.Profile)
	}

	raw := badge + pathStr + gitBadge + itemsBadge
	return padOrTruncate(raw, width)
}

func renderLeftCell(m *Model, row int, width int) string {
	itemIdx := m.ScrollOffset + row
	if itemIdx >= len(m.Filtered) {
		return strings.Repeat(" ", width)
	}

	entry := m.Filtered[itemIdx]
	isSelected := itemIdx == m.Cursor

	prefix := "  "
	if isSelected {
		prefix = "> "
	}

	role := theme.RoleRegularFile
	if entry.IsDir {
		role = theme.RoleDirectory
	} else if entry.IsSymlink {
		role = theme.RoleSymlink
	} else if entry.IsExec {
		role = theme.RoleExecutable
	}
	if entry.IsBroken {
		role = theme.RoleError
	}

	// Size indicator on right of left pane
	sizeStr := ""
	if !entry.IsDir {
		sizeStr = renderer.FormatSize(entry.Size, true)
	}

	// Git status indicator
	gitBadge := ""
	if entry.GitStatus != "" && entry.GitStatus != string(git.StatusClean) {
		badge := git.FormatStatusBadge(git.FileStatus(entry.GitStatus), m.Theme, m.Profile)
		gitBadge = badge + " "
	}

	// Available space for name
	availNameWidth := width - len(prefix) - len(gitBadge) - len(entry.Icon) - len(sizeStr) - 2
	if availNameWidth < 4 {
		availNameWidth = 4
	}

	truncatedName := renderer.Truncate(entry.Name, availNameWidth, "…")
	styledName := m.Theme.Format(role, truncatedName, m.Profile)

	// Combine components
	visibleLen := renderer.VisibleWidth(prefix + gitBadge + entry.Icon + truncatedName + sizeStr)
	paddingSpaces := width - visibleLen
	if paddingSpaces < 0 {
		paddingSpaces = 0
	}

	line := prefix + gitBadge + entry.Icon + styledName + strings.Repeat(" ", paddingSpaces) + m.Theme.Format(theme.RoleMuted, sizeStr, m.Profile)

	if isSelected {
		return m.Theme.Format(theme.RoleSelection, line, m.Profile)
	}
	return line
}

func renderStatusBar(m *Model, width int) string {
	if m.FilterActive {
		filterPrompt := fmt.Sprintf("/filter: %s_", m.FilterQuery)
		return padOrTruncate(m.Theme.Format(theme.RoleWarning, filterPrompt, m.Profile), width)
	}

	entry := m.CurrentEntry()
	if entry == nil {
		return padOrTruncate(m.Theme.Format(theme.RoleMuted, "  Ready", m.Profile), width)
	}

	details := fmt.Sprintf("  %s  %s  %s", entry.Mode.String(), renderer.FormatSize(entry.Size, true), entry.ModTime.Format("2006-01-02 15:04:05"))
	if entry.IsSymlink && entry.Target != "" {
		details += fmt.Sprintf(" -> %s", entry.Target)
	}

	return padOrTruncate(m.Theme.Format(theme.RoleInfo, details, m.Profile), width)
}

func renderFooter(m *Model, width int) string {
	hints := " [j/k,↑/↓] Move  [Enter/l] Open  [h/Bksp] Up  [/] Filter  [.] Hidden  [?] Help  [q] Quit"
	return padOrTruncate(m.Theme.Format(theme.RoleMuted, hints, m.Profile), width)
}

func padOrTruncate(s string, targetWidth int) string {
	vis := renderer.VisibleWidth(s)
	if vis > targetWidth {
		return renderer.Truncate(s, targetWidth, "")
	}
	if vis < targetWidth {
		return s + strings.Repeat(" ", targetWidth-vis)
	}
	return s
}

func overlayHelpModal(m *Model, screen []string, width, height int) []string {
	boxWidth := 46
	if boxWidth > width-4 {
		boxWidth = width - 4
	}
	if boxWidth < 20 {
		return screen
	}

	boxLines := []string{
		"┌──────────────────────────────────────────┐",
		"│         NOVA INTERACTIVE NAVIGATOR       │",
		"├──────────────────────────────────────────┤",
		"│  j, ↓         Move down                  │",
		"│  k, ↑         Move up                    │",
		"│  Enter, l     Enter directory            │",
		"│  h, Backspace Parent directory           │",
		"│  /            Search / filter items      │",
		"│  .            Toggle hidden files        │",
		"│  g / G        Jump to top / bottom       │",
		"│  PgUp / PgDn  Jump page up / down        │",
		"│  ?            Toggle this help modal     │",
		"│  Esc          Clear filter / Close help  │",
		"│  q            Quit interactive mode      │",
		"└──────────────────────────────────────────┘",
	}

	if !m.UnicodeSupported {
		for i, bl := range boxLines {
			boxLines[i] = strings.ReplaceAll(bl, "┌", "+")
			boxLines[i] = strings.ReplaceAll(boxLines[i], "┐", "+")
			boxLines[i] = strings.ReplaceAll(boxLines[i], "└", "+")
			boxLines[i] = strings.ReplaceAll(boxLines[i], "┘", "+")
			boxLines[i] = strings.ReplaceAll(boxLines[i], "├", "+")
			boxLines[i] = strings.ReplaceAll(boxLines[i], "┤", "+")
			boxLines[i] = strings.ReplaceAll(boxLines[i], "─", "-")
			boxLines[i] = strings.ReplaceAll(boxLines[i], "│", "|")
		}
	}

	boxHeight := len(boxLines)
	startY := (height - boxHeight) / 2
	startX := (width - boxWidth) / 2

	if startY < 0 {
		startY = 0
	}
	if startX < 0 {
		startX = 0
	}

	result := make([]string, len(screen))
	copy(result, screen)

	for i, bLine := range boxLines {
		y := startY + i
		if y >= len(result) {
			break
		}

		// Stamp box line at startX
		leftPad := strings.Repeat(" ", startX)
		styledBox := m.Theme.Format(theme.RoleAccent, bLine, m.Profile)
		remainder := width - startX - boxWidth
		if remainder < 0 {
			remainder = 0
		}
		rightPad := strings.Repeat(" ", remainder)

		result[y] = leftPad + styledBox + rightPad
	}

	return result
}
