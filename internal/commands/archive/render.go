package archive

import (
	"encoding/json"
	"fmt"
	"strings"

	"nova/internal/command"
	"nova/internal/output"
	"nova/internal/renderer"
	"nova/internal/theme"
)

// RenderPackResult renders archive creation metrics.
func RenderPackResult(ctx *command.Context, archivePath string, fileCount int, archiveSize int64) error {
	if ctx.Printer.Mode == output.ModeJSON {
		res := map[string]interface{}{
			"action":       "pack",
			"archive":      archivePath,
			"files_packed": fileCount,
			"archive_size": archiveSize,
		}
		data, _ := json.MarshalIndent(res, "", "  ")
		ctx.Printer.Println(string(data))
		return nil
	}

	if ctx.Printer.Mode == output.ModePlain {
		ctx.Printer.Printf("%s\t%d\t%d\n", archivePath, fileCount, archiveSize)
		return nil
	}

	th := ctx.Theme
	prof := ctx.Caps.ColorProfile
	unicode := ctx.Caps.UnicodeSupported
	w := ctx.Caps.Width
	if w <= 0 || w > 100 {
		w = 80
	}

	lines := []string{
		fmt.Sprintf(" %-18s %s", th.Format(theme.RoleMuted, "Archive:", prof), th.Format(theme.RoleSuccess, archivePath, prof)),
		fmt.Sprintf(" %-18s %s", th.Format(theme.RoleMuted, "Files Packed:", prof), th.Format(theme.RoleAccent, fmt.Sprintf("%d items", fileCount), prof)),
		fmt.Sprintf(" %-18s %s", th.Format(theme.RoleMuted, "Output Size:", prof), th.Format(theme.RoleInfo, renderer.FormatSize(archiveSize, true), prof)),
		strings.Repeat("─", w-4),
		" " + th.Format(theme.RoleSuccess, "✔ Archive successfully created and verified.", prof),
	}

	card := renderer.RenderCard("📦 Archive Created", lines, w, th.Style(theme.RoleAccent), unicode, prof)
	for _, l := range card {
		ctx.Printer.Println(l)
	}
	return nil
}

// RenderUnpackResult renders archive extraction metrics.
func RenderUnpackResult(ctx *command.Context, archivePath, destDir string, extractedCount int) error {
	if ctx.Printer.Mode == output.ModeJSON {
		res := map[string]interface{}{
			"action":          "unpack",
			"archive":         archivePath,
			"destination":     destDir,
			"extracted_files": extractedCount,
		}
		data, _ := json.MarshalIndent(res, "", "  ")
		ctx.Printer.Println(string(data))
		return nil
	}

	if ctx.Printer.Mode == output.ModePlain {
		ctx.Printer.Printf("%s\t%s\t%d\n", archivePath, destDir, extractedCount)
		return nil
	}

	th := ctx.Theme
	prof := ctx.Caps.ColorProfile
	unicode := ctx.Caps.UnicodeSupported
	w := ctx.Caps.Width
	if w <= 0 || w > 100 {
		w = 80
	}

	lines := []string{
		fmt.Sprintf(" %-18s %s", th.Format(theme.RoleMuted, "Archive:", prof), th.Format(theme.RoleInfo, archivePath, prof)),
		fmt.Sprintf(" %-18s %s", th.Format(theme.RoleMuted, "Extracted to:", prof), th.Format(theme.RoleSuccess, destDir, prof)),
		fmt.Sprintf(" %-18s %s", th.Format(theme.RoleMuted, "Files Extracted:", prof), th.Format(theme.RoleAccent, fmt.Sprintf("%d items", extractedCount), prof)),
		strings.Repeat("─", w-4),
		" " + th.Format(theme.RoleSuccess, "✔ Extraction completed safely (zip-slip protected).", prof),
	}

	card := renderer.RenderCard("📂 Archive Extracted", lines, w, th.Style(theme.RoleAccent), unicode, prof)
	for _, l := range card {
		ctx.Printer.Println(l)
	}
	return nil
}

// RenderListResult renders the contents and compression ratios of an archive.
func RenderListResult(ctx *command.Context, archivePath string, entries []ArchiveEntry) error {
	if ctx.Printer.Mode == output.ModeJSON {
		data, _ := json.MarshalIndent(entries, "", "  ")
		ctx.Printer.Println(string(data))
		return nil
	}

	if ctx.Printer.Mode == output.ModePlain {
		for _, e := range entries {
			ctx.Printer.Printf("%s\t%d\t%d\t%.1f%%\n", e.Name, e.Size, e.CompressedSize, e.CompressionRatio)
		}
		return nil
	}

	th := ctx.Theme
	prof := ctx.Caps.ColorProfile
	unicode := ctx.Caps.UnicodeSupported
	w := ctx.Caps.Width
	if w <= 0 || w > 100 {
		w = 80
	}

	var lines []string
	headerStr := fmt.Sprintf(" %-36s %10s %10s %8s",
		th.Format(theme.RoleMuted, "FILE", prof),
		th.Format(theme.RoleMuted, "ORIGINAL", prof),
		th.Format(theme.RoleMuted, "COMPRESSED", prof),
		th.Format(theme.RoleMuted, "SAVINGS", prof),
	)
	lines = append(lines, headerStr)
	lines = append(lines, strings.Repeat("─", w-4))

	var totalOriginal, totalCompressed int64
	maxItems := 18
	for i, e := range entries {
		totalOriginal += e.Size
		totalCompressed += e.CompressedSize

		if i >= maxItems {
			lines = append(lines, fmt.Sprintf(" … and %d more entries", len(entries)-maxItems))
			break
		}

		truncName := renderer.Truncate(e.Name, 36, "…")
		role := theme.RoleRegularFile
		if e.IsDir {
			role = theme.RoleDirectory
		}

		origStr := renderer.FormatSize(e.Size, true)
		compStr := renderer.FormatSize(e.CompressedSize, true)
		ratioStr := fmt.Sprintf("%5.1f%%", e.CompressionRatio)

		line := fmt.Sprintf(" %-36s %10s %10s %8s",
			th.Format(role, truncName, prof),
			th.Format(theme.RoleMuted, origStr, prof),
			th.Format(theme.RoleInfo, compStr, prof),
			th.Format(theme.RoleSuccess, ratioStr, prof),
		)
		lines = append(lines, line)
	}

	lines = append(lines, strings.Repeat("─", w-4))
	overallRatio := 0.0
	if totalOriginal > 0 {
		overallRatio = (1.0 - float64(totalCompressed)/float64(totalOriginal)) * 100.0
	}
	summaryLine := fmt.Sprintf(" Total: %d files | %s → %s (%.1f%% overall compression)",
		len(entries),
		renderer.FormatSize(totalOriginal, true),
		renderer.FormatSize(totalCompressed, true),
		overallRatio,
	)
	lines = append(lines, th.Format(theme.RoleAccent, summaryLine, prof))

	card := renderer.RenderCard("🗜️  Archive Contents: "+filepathBase(archivePath), lines, w, th.Style(theme.RoleAccent), unicode, prof)
	for _, l := range card {
		ctx.Printer.Println(l)
	}
	return nil
}

func filepathBase(p string) string {
	parts := strings.Split(strings.ReplaceAll(p, "\\", "/"), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return p
}
