package interactive

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"nova/internal/commands/cat"
	"nova/internal/commands/cat/syntax"
	"nova/internal/renderer"
	"nova/internal/terminal"
	"nova/internal/theme"
)

// LoadPreview loads and formats a preview of the specified entry to fit within width and height.
func LoadPreview(entry *Entry, width, height int, th *theme.Theme, profile terminal.ColorProfile) []string {
	if entry == nil {
		return []string{th.Format(theme.RoleMuted, "  (No selection)", profile)}
	}

	if width < 10 {
		width = 10
	}
	if height < 2 {
		height = 2
	}

	if entry.IsBroken {
		return []string{
			th.Format(theme.RoleError, fmt.Sprintf("  Broken symlink: %s -> %s", entry.Name, entry.Target), profile),
		}
	}

	if entry.IsDir {
		return loadDirPreview(entry, width, height, th, profile)
	}

	return loadFilePreview(entry, width, height, th, profile)
}

func loadDirPreview(entry *Entry, width, height int, th *theme.Theme, profile terminal.ColorProfile) []string {
	var lines []string

	header := fmt.Sprintf("📁 Directory: %s", entry.Name)
	lines = append(lines, th.Format(theme.RoleDirectory, renderer.Truncate(header, width, "…"), profile))

	dirEntries, err := os.ReadDir(entry.Path)
	if err != nil {
		lines = append(lines, th.Format(theme.RoleError, fmt.Sprintf("  error reading directory: %v", err), profile))
		return lines
	}

	meta := fmt.Sprintf("  Items: %d", len(dirEntries))
	lines = append(lines, th.Format(theme.RoleMuted, meta, profile))
	lines = append(lines, "")

	if len(dirEntries) == 0 {
		lines = append(lines, th.Format(theme.RoleMuted, "  (empty directory)", profile))
		return lines
	}

	maxItems := height - len(lines)
	if maxItems < 1 {
		maxItems = 1
	}

	for i, de := range dirEntries {
		if i >= maxItems {
			remaining := len(dirEntries) - i
			lines = append(lines, th.Format(theme.RoleMuted, fmt.Sprintf("  … and %d more items", remaining), profile))
			break
		}

		info, err := de.Info()
		sizeStr := ""
		if err == nil && !de.IsDir() {
			sizeStr = fmt.Sprintf(" (%s)", renderer.FormatSize(info.Size(), true))
		}

		prefix := "  📄 "
		role := theme.RoleRegularFile
		if de.IsDir() {
			prefix = "  📁 "
			role = theme.RoleDirectory
		}

		itemLine := fmt.Sprintf("%s%s%s", prefix, de.Name(), sizeStr)
		lines = append(lines, th.Format(role, renderer.Truncate(itemLine, width, "…"), profile))
	}

	return lines
}

func loadFilePreview(entry *Entry, width, height int, th *theme.Theme, profile terminal.ColorProfile) []string {
	var lines []string

	// Guard against opening excessively large files
	const maxPreviewSize = 20 * 1024 * 1024 // 20 MB
	if entry.Size > maxPreviewSize {
		msg := fmt.Sprintf("  [File exceeds 20MB (%s) — preview omitted]", renderer.FormatSize(entry.Size, true))
		return []string{th.Format(theme.RoleWarning, msg, profile)}
	}

	f, err := os.Open(entry.Path)
	if err != nil {
		return []string{th.Format(theme.RoleError, fmt.Sprintf("  cannot open file: %v", err), profile)}
	}
	defer f.Close()

	// Read up to 32KB sample for detection and preview
	sample := make([]byte, 32*1024)
	n, err := io.ReadFull(f, sample)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return []string{th.Format(theme.RoleError, fmt.Sprintf("  cannot read file: %v", err), profile)}
	}
	sample = sample[:n]

	if len(sample) == 0 {
		return []string{th.Format(theme.RoleMuted, "  (empty file)", profile)}
	}

	if cat.IsBinary(sample) {
		banner := fmt.Sprintf("⚡ Binary file (%s) — Hex View", renderer.FormatSize(entry.Size, true))
		lines = append(lines, th.Format(theme.RoleAccent, renderer.Truncate(banner, width, "…"), profile))
		lines = append(lines, "")

		hexRows := cat.FormatHexDump(sample, 0)
		maxRows := height - len(lines)
		if maxRows < 1 {
			maxRows = 1
		}
		for i, row := range hexRows {
			if i >= maxRows {
				break
			}
			lines = append(lines, th.Format(theme.RoleMuted, renderer.Truncate(row, width, "…"), profile))
		}
		return lines
	}

	// Text file: syntax highlight with line numbers
	lang := cat.DetectLanguage(entry.Name, sample)
	scanner := bufio.NewScanner(bytes.NewReader(sample))
	synState := &syntax.State{}

	lineNum := 1
	for scanner.Scan() && len(lines) < height {
		rawLine := scanner.Text()
		highlighted := syntax.HighlightLine(rawLine, lang, synState, th, profile)

		numPrefix := th.Format(theme.RoleMuted, fmt.Sprintf("%3d │ ", lineNum), profile)
		availWidth := width - 6
		if availWidth < 5 {
			availWidth = 5
		}
		truncatedContent := renderer.Truncate(highlighted, availWidth, "…")
		lines = append(lines, numPrefix+truncatedContent)
		lineNum++
	}

	return lines
}
