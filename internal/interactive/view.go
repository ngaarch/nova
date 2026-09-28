package interactive

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	// 2. Pane layout
	divChar := "│"
	if !m.UnicodeSupported {
		divChar = "|"
	}

	contentRows := height - 4
	if contentRows < 1 {
		contentRows = 1
	}

	if m.PreviewCollapsed {
		leftWidth := width
		leftHeader := renderer.PadRight("  NAME", leftWidth)
		paneHeaderLine := m.Theme.Format(theme.RoleMuted, leftHeader, m.Profile)
		screen = append(screen, paneHeaderLine)

		for r := 0; r < contentRows; r++ {
			leftCell := renderLeftCell(m, r, leftWidth)
			screen = append(screen, leftCell)
		}
	} else {
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
		paneHeaderLine := m.Theme.Format(theme.RoleMuted, leftHeader+divChar+rightHeader, m.Profile)
		screen = append(screen, paneHeaderLine)

		for r := 0; r < contentRows; r++ {
			leftCell := renderLeftCell(m, r, leftWidth)
			rightCell := ""
			if r < len(m.PreviewLines) {
				rightCell = m.PreviewLines[r]
			}
			paddedRight := padOrTruncate(rightCell, rightWidth)
			rowLine := leftCell + m.Theme.Format(theme.RoleMuted, divChar, m.Profile) + paddedRight
			screen = append(screen, rowLine)
		}
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
	} else if m.InspectorActive {
		screen = overlayInspectorModal(m, screen, width, height)
	}

	// Return full frame with ANSI home cursor
	return "\x1b[H" + strings.Join(screen, "\r\n")
}

func renderHeader(m *Model, width int) string {
	badge := m.Theme.Format(theme.RoleAccent, "[nova] ", m.Profile)

	// Breadcrumb path styling
	home, _ := os.UserHomeDir()
	displayDir := m.CurrentDir
	if home != "" && strings.HasPrefix(displayDir, home) {
		displayDir = "~" + strings.TrimPrefix(displayDir, home)
	}

	sep := " › "
	if !m.UnicodeSupported {
		sep = " > "
	}
	parts := strings.Split(filepath.Clean(displayDir), string(filepath.Separator))
	var styledCrumbs []string
	for idx, part := range parts {
		if part == "" {
			if idx == 0 {
				styledCrumbs = append(styledCrumbs, "/")
			}
			continue
		}
		if idx == len(parts)-1 {
			styledCrumbs = append(styledCrumbs, m.Theme.Format(theme.RoleDirectory, part, m.Profile))
		} else {
			styledCrumbs = append(styledCrumbs, m.Theme.Format(theme.RoleMuted, part, m.Profile))
		}
	}
	crumbStr := strings.Join(styledCrumbs, m.Theme.Format(theme.RoleMuted, sep, m.Profile))

	countStr := fmt.Sprintf("(%d items)", len(m.Filtered))
	if m.FilterQuery != "" {
		countStr = fmt.Sprintf("(%d/%d, filter: %q)", len(m.Filtered), len(m.Entries), m.FilterQuery)
	}
	if len(m.SelectedPaths) > 0 {
		countStr += fmt.Sprintf(" [%d sel]", len(m.SelectedPaths))
	}
	itemsBadge := m.Theme.Format(theme.RoleMuted, " "+countStr, m.Profile)

	gitBadge := ""
	if m.Git != nil && m.Git.Branch != "" {
		gitBadge = " " + git.FormatBranch(m.Git.Branch, m.Git.IsDetached, m.UnicodeSupported, m.Theme, m.Profile)
	}

	sortBadge := m.Theme.Format(theme.RoleAccent, fmt.Sprintf(" [sort:%s]", m.SortMode), m.Profile)
	themeBadge := m.Theme.Format(theme.RoleAccent, fmt.Sprintf(" [%s]", m.Theme.Name), m.Profile)

	raw := badge + crumbStr + gitBadge + itemsBadge + sortBadge + themeBadge
	return padOrTruncate(raw, width)
}

func renderLeftCell(m *Model, row int, width int) string {
	itemIdx := m.ScrollOffset + row
	if itemIdx >= len(m.Filtered) {
		return strings.Repeat(" ", width)
	}

	entry := m.Filtered[itemIdx]
	isSelected := itemIdx == m.Cursor

	prefix := "   "
	if isSelected {
		if m.UnicodeSupported {
			prefix = " ❯ "
		} else {
			prefix = " > "
		}
	}

	// Selection checkmark
	selectBadge := ""
	if m.SelectedPaths != nil && m.SelectedPaths[entry.Path] {
		selectBadge = m.Theme.Format(theme.RoleSuccess, "✔ ", m.Profile)
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

	// Size indicator on right of left pane with color thresholds
	sizeStr := ""
	if !entry.IsDir {
		sizeStr = renderer.FormatSize(entry.Size, true)
	}
	sizeRole := theme.RoleMuted
	if entry.Size > 50*1024*1024 {
		sizeRole = theme.RoleError
	} else if entry.Size > 1*1024*1024 {
		sizeRole = theme.RoleWarning
	} else if entry.Size > 10*1024 {
		sizeRole = theme.RoleInfo
	}

	// Git status indicator
	gitBadge := ""
	if entry.GitStatus != "" && entry.GitStatus != string(git.StatusClean) {
		badge := git.FormatStatusBadge(git.FileStatus(entry.GitStatus), m.Theme, m.Profile)
		gitBadge = badge + " "
	}

	// Available space for name
	availNameWidth := width - len(prefix) - len(selectBadge) - len(gitBadge) - len(entry.Icon) - len(sizeStr) - 2
	if availNameWidth < 4 {
		availNameWidth = 4
	}

	truncatedName := renderer.Truncate(entry.Name, availNameWidth, "…")

	// Match highlighting for fuzzy search
	var styledName string
	if m.FilterQuery != "" {
		styledName = renderer.HighlightFuzzyMatch(truncatedName, m.FilterQuery, m.Theme.Style(theme.RoleAccent), m.Theme.Style(role), m.Profile)
	} else {
		styledName = m.Theme.Format(role, truncatedName, m.Profile)
	}

	// Combine components
	visibleLen := renderer.VisibleWidth(prefix + selectBadge + gitBadge + entry.Icon + truncatedName + sizeStr)
	paddingSpaces := width - visibleLen
	if paddingSpaces < 0 {
		paddingSpaces = 0
	}

	line := prefix + selectBadge + gitBadge + entry.Icon + styledName + strings.Repeat(" ", paddingSpaces) + m.Theme.Format(sizeRole, sizeStr, m.Profile)

	if isSelected {
		return m.Theme.Format(theme.RoleSelection, line, m.Profile)
	}
	return line
}

func renderStatusBar(m *Model, width int) string {
	if m.ConfirmDelete {
		entry := m.CurrentEntry()
		name := ""
		if entry != nil {
			name = entry.Name
		}
		delMsg := fmt.Sprintf("  ⚠ Delete %q? Press [y] to confirm, [Esc/any] to cancel", name)
		return padOrTruncate(m.Theme.Format(theme.RoleError, delMsg, m.Profile), width)
	}

	if m.RenameActive {
		renamePrompt := fmt.Sprintf("  ✏ Rename to: %s_", m.RenameInput)
		return padOrTruncate(m.Theme.Format(theme.RoleWarning, renamePrompt, m.Profile), width)
	}

	if m.NewFileActive {
		filePrompt := fmt.Sprintf("  📄 New file name: %s_", m.NewFileInput)
		return padOrTruncate(m.Theme.Format(theme.RoleWarning, filePrompt, m.Profile), width)
	}

	if m.CommandPaletteActive {
		cmdPrompt := fmt.Sprintf("  %s_", m.CommandInput)
		return padOrTruncate(m.Theme.Format(theme.RoleAccent, cmdPrompt, m.Profile), width)
	}

	if m.NewFolderActive {
		folderPrompt := fmt.Sprintf("  📁 New folder name: %s_", m.NewFolderInput)
		return padOrTruncate(m.Theme.Format(theme.RoleWarning, folderPrompt, m.Profile), width)
	}

	if m.ActionMessage != "" && time.Since(m.ActionTime) < 3*time.Second {
		return padOrTruncate(m.Theme.Format(theme.RoleSuccess, "  "+m.ActionMessage, m.Profile), width)
	}

	if m.FilterActive {
		filterPrompt := fmt.Sprintf("/filter: %s_", m.FilterQuery)
		return padOrTruncate(m.Theme.Format(theme.RoleWarning, filterPrompt, m.Profile), width)
	}

	entry := m.CurrentEntry()
	if entry == nil {
		return padOrTruncate(m.Theme.Format(theme.RoleMuted, "  Ready", m.Profile), width)
	}

	relTime := renderer.FormatRelativeTime(entry.ModTime)
	details := fmt.Sprintf("  %s  %s  %s (%s)", entry.Mode.String(), renderer.FormatSize(entry.Size, true), relTime, entry.ModTime.Format("15:04:05"))
	if entry.IsSymlink && entry.Target != "" {
		details += fmt.Sprintf(" -> %s", entry.Target)
	}
	if len(m.SelectedPaths) > 0 {
		details += fmt.Sprintf("  [%d selected]", len(m.SelectedPaths))
	}
	if len(m.Bookmarks) > 0 {
		details += fmt.Sprintf("  [%d bm]", len(m.Bookmarks))
	}

	return padOrTruncate(m.Theme.Format(theme.RoleInfo, details, m.Profile), width)
}

func renderFooter(m *Model, width int) string {
	hints := " [j/k] Move  [:] Cmd  [s] Sort  [t] Theme  [x] Hex  [i] Inspect  [#] Hash  [b/B] Bm  [?] Help  [q] Quit"
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
	boxWidth := 48
	if boxWidth > width-4 {
		boxWidth = width - 4
	}
	if boxWidth < 20 {
		return screen
	}

	boxLines := []string{
		"╭──────────────────────────────────────────────╮",
		"│         NOVA INTERACTIVE NAVIGATOR           │",
		"├──────────────────────────────────────────────┤",
		"│  j, ↓         Move selection down            │",
		"│  k, ↑         Move selection up              │",
		"│  Enter, l     Enter directory / Open         │",
		"│  h, Backspace Go to parent directory         │",
		"│  :            Open command palette (:q, :w)  │",
		"│  b / B        Bookmark / cycle jumps         │",
		"│  Space        Toggle item selection (multi)  │",
		"│  s            Cycle sorting (name/size/time) │",
		"│  t            Cycle theme in real-time       │",
		"│  x, H         Toggle hex dump inspection     │",
		"│  i            Inspect file metadata          │",
		"│  #            Compute quick SHA-256 hash     │",
		"│  n            Create new file                │",
		"│  N            Create new folder              │",
		"│  p            Toggle preview pane collapse   │",
		"│  /            Search / fuzzy filter          │",
		"│  e            Edit file in $EDITOR           │",
		"│  d            Delete item (with confirm)     │",
		"│  r            Rename item                    │",
		"│  c            Copy file path to clipboard    │",
		"│  .            Toggle hidden / dot files      │",
		"│  g / G        Jump to top / bottom           │",
		"│  PgUp / PgDn  Jump page up / down            │",
		"│  ?            Toggle this help modal         │",
		"│  Esc          Cancel action / Clear filter   │",
		"│  q            Quit interactive mode          │",
		"╰──────────────────────────────────────────────╯",
	}

	if !m.UnicodeSupported {
		for i, bl := range boxLines {
			boxLines[i] = strings.ReplaceAll(bl, "╭", "+")
			boxLines[i] = strings.ReplaceAll(boxLines[i], "╮", "+")
			boxLines[i] = strings.ReplaceAll(boxLines[i], "╰", "+")
			boxLines[i] = strings.ReplaceAll(boxLines[i], "╯", "+")
			boxLines[i] = strings.ReplaceAll(boxLines[i], "┌", "+")
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

func overlayInspectorModal(m *Model, screen []string, width, height int) []string {
	entry := m.CurrentEntry()
	if entry == nil {
		return screen
	}

	boxWidth := 56
	if boxWidth > width-4 {
		boxWidth = width - 4
	}
	if boxWidth < 20 {
		return screen
	}

	fileType := "Regular File"
	if entry.IsDir {
		fileType = "Directory"
	} else if entry.IsSymlink {
		fileType = "Symbolic Link"
	}

	sizeStr := fmt.Sprintf("%d bytes (%s)", entry.Size, renderer.FormatSize(entry.Size, true))
	modeStr := fmt.Sprintf("%s (%04o)", entry.Mode.String(), entry.Mode.Perm())
	modStr := fmt.Sprintf("%s (%s)", entry.ModTime.Format("2006-01-02 15:04:05"), renderer.FormatRelativeTime(entry.ModTime))

	innerW := boxWidth - 4
	valW := innerW - 10
	if valW < 10 {
		valW = 10
	}

	nameLine := renderer.Truncate(entry.Name, valW, "...")
	pathLine := renderer.Truncate(entry.Path, valW, "...")

	boxLines := []string{
		"╭──────────────────────────────────────────────────────╮",
		"│               FILE METADATA INSPECTOR                │",
		"├──────────────────────────────────────────────────────┤",
		fmt.Sprintf("│  Name:     %-42s│", nameLine),
		fmt.Sprintf("│  Path:     %-42s│", pathLine),
		fmt.Sprintf("│  Type:     %-42s│", fileType),
		fmt.Sprintf("│  Size:     %-42s│", sizeStr),
		fmt.Sprintf("│  Mode:     %-42s│", modeStr),
		fmt.Sprintf("│  Modified: %-42s│", modStr),
	}
	if entry.IsSymlink && entry.Target != "" {
		targetLine := renderer.Truncate(entry.Target, valW, "...")
		boxLines = append(boxLines, fmt.Sprintf("│  Target:   %-42s│", targetLine))
	}
	boxLines = append(boxLines,
		"├──────────────────────────────────────────────────────┤",
		"│  Press [i], [Esc], or [q] to close inspector         │",
		"╰──────────────────────────────────────────────────────╯",
	)

	if !m.UnicodeSupported {
		for i, bl := range boxLines {
			boxLines[i] = strings.ReplaceAll(bl, "╭", "+")
			boxLines[i] = strings.ReplaceAll(boxLines[i], "╮", "+")
			boxLines[i] = strings.ReplaceAll(boxLines[i], "╰", "+")
			boxLines[i] = strings.ReplaceAll(boxLines[i], "╯", "+")
			boxLines[i] = strings.ReplaceAll(boxLines[i], "┌", "+")
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
