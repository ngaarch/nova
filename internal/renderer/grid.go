package renderer

import "strings"

// Grid renders a list of items across multiple columns fitted to a terminal width.
type Grid struct {
	TermWidth int
	Gutter    int // Spacing between columns (default: 2)
}

// NewGrid constructs a Grid calculator with the specified terminal width and gutter.
func NewGrid(termWidth, gutter int) *Grid {
	if termWidth <= 0 {
		termWidth = 80
	}
	if gutter <= 0 {
		gutter = 2
	}
	return &Grid{
		TermWidth: termWidth,
		Gutter:    gutter,
	}
}

// Render formats a slice of items into multi-column layout strings.
// Arranges items in column-major order (standard Unix ls format).
func (g *Grid) Render(items []string) []string {
	if len(items) == 0 {
		return nil
	}

	widths := make([]int, len(items))
	maxWidth := 0
	for i, item := range items {
		w := VisibleWidth(item)
		widths[i] = w
		if w > maxWidth {
			maxWidth = w
		}
	}

	// If single item or terminal is narrower than max item, render single column
	if len(items) == 1 || g.TermWidth < (maxWidth+g.Gutter) {
		return items
	}

	maxCols := (g.TermWidth + g.Gutter) / (maxWidth + g.Gutter)
	if maxCols < 1 {
		maxCols = 1
	}
	if maxCols > len(items) {
		maxCols = len(items)
	}

	for cols := maxCols; cols >= 1; cols-- {
		rows := (len(items) + cols - 1) / cols

		colWidths := make([]int, cols)
		for c := 0; c < cols; c++ {
			for r := 0; r < rows; r++ {
				idx := c*rows + r
				if idx < len(items) {
					if widths[idx] > colWidths[c] {
						colWidths[c] = widths[idx]
					}
				}
			}
		}

		totalWidth := 0
		for c, cw := range colWidths {
			totalWidth += cw
			if c < cols-1 {
				totalWidth += g.Gutter
			}
		}

		if totalWidth <= g.TermWidth || cols == 1 {
			lines := make([]string, rows)
			for r := 0; r < rows; r++ {
				var lineBuilder strings.Builder
				for c := 0; c < cols; c++ {
					idx := c*rows + r
					if idx < len(items) {
						item := items[idx]
						if c < cols-1 {
							lineBuilder.WriteString(PadRight(item, colWidths[c]))
							lineBuilder.WriteString(strings.Repeat(" ", g.Gutter))
						} else {
							lineBuilder.WriteString(item)
						}
					}
				}
				lines[r] = strings.TrimRight(lineBuilder.String(), " ")
			}
			return lines
		}
	}

	return items
}
