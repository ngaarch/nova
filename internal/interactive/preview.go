package interactive

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"nova/internal/commands/cat"
	"nova/internal/commands/cat/syntax"
	"nova/internal/renderer"
	"nova/internal/terminal"
	"nova/internal/theme"
)

// LoadPreview loads and formats a preview of the specified entry to fit within width and height.
func LoadPreview(entry *Entry, width, height int, th *theme.Theme, profile terminal.ColorProfile, forceHex bool) []string {
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

	return loadFilePreview(entry, width, height, th, profile, forceHex)
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

	relTime := renderer.FormatRelativeTime(entry.ModTime)
	meta := fmt.Sprintf("  Items: %d • Modified: %s", len(dirEntries), relTime)
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

func loadFilePreview(entry *Entry, width, height int, th *theme.Theme, profile terminal.ColorProfile, forceHex bool) []string {
	var lines []string

	// Guard against opening excessively large files
	const maxPreviewSize = 20 * 1024 * 1024 // 20 MB
	if entry.Size > maxPreviewSize {
		msg := fmt.Sprintf("  [File exceeds 20MB (%s) — preview omitted]", renderer.FormatSize(entry.Size, true))
		return []string{th.Format(theme.RoleWarning, msg, profile)}
	}

	// Archive preview for zip/jar files
	lowerName := strings.ToLower(entry.Name)
	if !forceHex && (strings.HasSuffix(lowerName, ".zip") || strings.HasSuffix(lowerName, ".jar")) {
		if zipLines, ok := loadZipPreview(entry, width, height, th, profile); ok {
			return zipLines
		}
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

	if forceHex || cat.IsBinary(sample) {
		modeStr := "Hex View"
		if forceHex {
			modeStr = "Hex Inspector (Forced)"
		}
		banner := fmt.Sprintf("⚡ Binary file: %s (%s) — %s", entry.Name, renderer.FormatSize(entry.Size, true), modeStr)
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
	relTime := renderer.FormatRelativeTime(entry.ModTime)
	sizeStr := renderer.FormatSize(entry.Size, true)

	badge := fmt.Sprintf("[%s] %s • %s", strings.ToUpper(lang), sizeStr, relTime)
	if lang == "markdown" {
		badge = fmt.Sprintf("[MARKDOWN DOC] %s • %s", sizeStr, relTime)
	}
	lines = append(lines, th.Format(theme.RoleAccent, renderer.Truncate(badge, width, "…"), profile))
	lines = append(lines, "")

	scanner := bufio.NewScanner(bytes.NewReader(sample))
	synState := &syntax.State{}

	lineNum := 1
	for scanner.Scan() && len(lines) < height {
		rawLine := scanner.Text()
		trimmed := strings.TrimSpace(rawLine)

		if lang == "markdown" {
			var formattedLine string
			if strings.HasPrefix(trimmed, "# ") {
				formattedLine = th.Format(theme.RoleAccent, "  # "+strings.ToUpper(trimmed[2:]), profile)
			} else if strings.HasPrefix(trimmed, "## ") {
				formattedLine = th.Format(theme.RoleDocument, "  ## "+trimmed[3:], profile)
			} else if strings.HasPrefix(trimmed, "### ") {
				formattedLine = th.Format(theme.RoleAccent, "  ### "+trimmed[4:], profile)
			} else if strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "* [x]") {
				formattedLine = "  " + th.Format(theme.RoleSuccess, "[✓]", profile) + " " + trimmed[5:]
			} else if strings.HasPrefix(trimmed, "- [ ]") || strings.HasPrefix(trimmed, "* [ ]") {
				formattedLine = "  " + th.Format(theme.RoleWarning, "[ ]", profile) + " " + trimmed[5:]
			} else if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
				formattedLine = "  " + th.Format(theme.RoleAccent, "•", profile) + " " + trimmed[2:]
			} else if strings.HasPrefix(trimmed, ">") {
				formattedLine = "  " + th.Format(theme.RoleAccent, "│", profile) + " " + strings.TrimSpace(trimmed[1:])
			} else if trimmed == "---" {
				formattedLine = "  " + th.Format(theme.RoleMuted, strings.Repeat("─", width-6), profile)
			} else {
				formattedLine = "  " + syntax.HighlightLine(rawLine, lang, synState, th, profile)
			}
			lines = append(lines, renderer.Truncate(formattedLine, width, "…"))
			lineNum++
			continue
		}

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

// LoadDiffPreview loads a git diff preview for the entry against HEAD.
func LoadDiffPreview(entry *Entry, width, height int, th *theme.Theme, profile terminal.ColorProfile) []string {
	if entry == nil {
		return []string{th.Format(theme.RoleMuted, "  (No selection)", profile)}
	}

	var lines []string
	header := fmt.Sprintf("🌿 Git Diff: %s", entry.Name)
	lines = append(lines, th.Format(theme.RoleAccent, renderer.Truncate(header, width, "…"), profile))
	lines = append(lines, "")

	cmd := exec.Command("git", "diff", "HEAD", "--", entry.Path)
	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		lines = append(lines, th.Format(theme.RoleMuted, "  (No uncommitted changes against HEAD)", profile))
		return lines
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() && len(lines) < height {
		line := scanner.Text()
		var styledLine string
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			styledLine = th.Format(theme.RoleSuccess, line, profile)
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			styledLine = th.Format(theme.RoleError, line, profile)
		} else if strings.HasPrefix(line, "@@") {
			styledLine = th.Format(theme.RoleAccent, line, profile)
		} else {
			styledLine = th.Format(theme.RoleMuted, line, profile)
		}
		lines = append(lines, renderer.Truncate(styledLine, width, "…"))
	}
	return lines
}

func loadZipPreview(entry *Entry, width, height int, th *theme.Theme, profile terminal.ColorProfile) ([]string, bool) {
	r, err := zip.OpenReader(entry.Path)
	if err != nil {
		return nil, false
	}
	defer r.Close()

	var lines []string
	var totalUncompressed uint64
	var totalCompressed uint64
	for _, f := range r.File {
		totalUncompressed += f.UncompressedSize64
		totalCompressed += f.CompressedSize64
	}

	savings := 0.0
	if totalUncompressed > 0 {
		savings = (1.0 - float64(totalCompressed)/float64(totalUncompressed)) * 100.0
		if savings < 0 {
			savings = 0
		}
	}

	banner := fmt.Sprintf("📦 Archive: %s (%d items, %s -> %s, %.1f%% saved)",
		entry.Name, len(r.File),
		renderer.FormatSize(int64(totalCompressed), true),
		renderer.FormatSize(int64(totalUncompressed), true),
		savings)
	lines = append(lines, th.Format(theme.RoleAccent, renderer.Truncate(banner, width, "…"), profile))
	lines = append(lines, "")

	if len(r.File) == 0 {
		lines = append(lines, th.Format(theme.RoleMuted, "  (empty archive)", profile))
		return lines, true
	}

	maxItems := height - len(lines)
	if maxItems < 1 {
		maxItems = 1
	}

	for i, f := range r.File {
		if i >= maxItems {
			remaining := len(r.File) - i
			lines = append(lines, th.Format(theme.RoleMuted, fmt.Sprintf("  … and %d more items", remaining), profile))
			break
		}

		icon := "📄 "
		role := theme.RoleRegularFile
		if f.FileInfo().IsDir() {
			icon = "📁 "
			role = theme.RoleDirectory
		}

		sizeStr := renderer.FormatSize(int64(f.UncompressedSize64), true)
		nameWidth := width - 16
		if nameWidth < 8 {
			nameWidth = 8
		}
		itemLine := fmt.Sprintf("  %s%-*s  %6s", icon, nameWidth, renderer.Truncate(f.Name, nameWidth, "…"), sizeStr)
		lines = append(lines, th.Format(role, renderer.Truncate(itemLine, width, "…"), profile))
	}

	return lines, true
}

