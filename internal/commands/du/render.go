package du

import (
	"encoding/json"
	"fmt"
	"strings"

	"nova/internal/command"
	"nova/internal/renderer"
	"nova/internal/theme"
)

// Item represents a single measured entry in disk usage analysis.
type Item struct {
	Path      string  `json:"path"`
	Bytes     int64   `json:"bytes"`
	HumanSize string  `json:"human_size"`
	IsDir     bool    `json:"is_dir"`
	Files     int     `json:"files,omitempty"`
	Dirs      int     `json:"dirs,omitempty"`
	Percent   float64 `json:"percent"`
}

// RenderHuman renders disk usage with responsive usage bars and semantic colors.
func RenderHuman(ctx *command.Context, items []Item, totalBytes int64) {
	th := ctx.Theme
	profile := ctx.Caps.ColorProfile
	barWidth := 16

	for _, it := range items {
		// Calculate percentage relative to total
		pct := 0.0
		if totalBytes > 0 {
			pct = (float64(it.Bytes) / float64(totalBytes)) * 100.0
		}
		if pct > 100.0 {
			pct = 100.0
		}

		// Build visual bar
		filled := int((pct / 100.0) * float64(barWidth))
		if filled > barWidth {
			filled = barWidth
		}
		empty := barWidth - filled

		barStr := strings.Repeat("█", filled) + strings.Repeat("░", empty)
		barStyled := th.Format(theme.RoleAccent, barStr, profile)
		if pct > 75.0 {
			barStyled = th.Format(theme.RoleWarning, barStr, profile)
		}

		pctStr := fmt.Sprintf("%5.1f%%", pct)
		sizeStr := fmt.Sprintf("%10s", it.HumanSize)
		sizeStyled := th.Format(theme.RoleSize, sizeStr, profile)

		pathRole := theme.RoleRegularFile
		name := it.Path
		if it.IsDir {
			pathRole = theme.RoleDirectory
			if !strings.HasSuffix(name, "/") && name != "." {
				name += "/"
			}
		}
		pathStyled := th.Format(pathRole, name, profile)

		ctx.Printer.Println(fmt.Sprintf("%s  %s  %s  %s", barStyled, pctStr, sizeStyled, pathStyled))
	}
}

// RenderPlain prints tab-separated sizes and paths (standard du format).
func RenderPlain(ctx *command.Context, items []Item, human bool) {
	for _, it := range items {
		sizeStr := fmt.Sprintf("%d", it.Bytes)
		if human {
			sizeStr = it.HumanSize
		}
		ctx.Printer.Println(fmt.Sprintf("%s\t%s", sizeStr, it.Path))
	}
}

// RenderJSON serializes items to structured JSON.
func RenderJSON(ctx *command.Context, items []Item) error {
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	ctx.Printer.Println(string(data))
	return nil
}

func formatDUSize(bytes int64, human bool) string {
	return renderer.FormatSize(bytes, human)
}
