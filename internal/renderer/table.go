package renderer

import "strings"

// Alignment specifies text positioning within a table cell.
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignRight
	AlignCenter
)

// Column describes an individual table column.
type Column struct {
	Header string
	Align  Alignment
}

// Table renders formatted rows and columns, adapting dynamically to terminal constraints.
type Table struct {
	TermWidth int
	Columns   []Column
	Rows      [][]string
	Separator string
}

// NewTable constructs a Table with the specified terminal width constraint.
func NewTable(termWidth int) *Table {
	if termWidth <= 0 {
		termWidth = 80
	}
	return &Table{
		TermWidth: termWidth,
		Separator: "  ",
	}
}

// AddColumn appends a column definition to the table.
func (t *Table) AddColumn(header string, align Alignment) {
	t.Columns = append(t.Columns, Column{
		Header: header,
		Align:  align,
	})
}

// AddRow adds a row of cell values to the table.
func (t *Table) AddRow(cells ...string) {
	t.Rows = append(t.Rows, cells)
}

// Render formats the table rows. If plain is true, formats tab-delimited text without padding or headers.
func (t *Table) Render(plain bool) []string {
	if len(t.Rows) == 0 {
		return nil
	}

	numCols := len(t.Columns)
	if numCols == 0 {
		for _, row := range t.Rows {
			if len(row) > numCols {
				numCols = len(row)
			}
		}
		for i := 0; i < numCols; i++ {
			t.Columns = append(t.Columns, Column{Align: AlignLeft})
		}
	}

	if plain {
		var lines []string
		for _, row := range t.Rows {
			var unstyledRow []string
			for _, cell := range row {
				unstyledRow = append(unstyledRow, StripANSI(cell))
			}
			lines = append(lines, strings.Join(unstyledRow, "\t"))
		}
		return lines
	}

	colWidths := make([]int, numCols)
	for i, col := range t.Columns {
		if col.Header != "" {
			hw := VisibleWidth(col.Header)
			if hw > colWidths[i] {
				colWidths[i] = hw
			}
		}
	}

	for _, row := range t.Rows {
		for i := 0; i < numCols && i < len(row); i++ {
			cw := VisibleWidth(row[i])
			if cw > colWidths[i] {
				colWidths[i] = cw
			}
		}
	}

	sepWidth := VisibleWidth(t.Separator)
	totalWidth := 0
	for i, w := range colWidths {
		totalWidth += w
		if i < numCols-1 {
			totalWidth += sepWidth
		}
	}

	if totalWidth > t.TermWidth && numCols > 0 {
		overflow := totalWidth - t.TermWidth
		lastIdx := numCols - 1
		available := colWidths[lastIdx] - overflow
		if available < 10 {
			available = 10
		}
		colWidths[lastIdx] = available
	}

	var lines []string

	hasHeader := false
	for _, col := range t.Columns {
		if col.Header != "" {
			hasHeader = true
			break
		}
	}

	if hasHeader {
		var headerCells []string
		for i, col := range t.Columns {
			headerCells = append(headerCells, formatCell(col.Header, colWidths[i], col.Align))
		}
		lines = append(lines, strings.Join(headerCells, t.Separator))
	}

	for _, row := range t.Rows {
		var renderedRow []string
		for i := 0; i < numCols; i++ {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			if VisibleWidth(cell) > colWidths[i] {
				cell = Truncate(cell, colWidths[i], "…")
			}
			align := AlignLeft
			if i < len(t.Columns) {
				align = t.Columns[i].Align
			}
			renderedRow = append(renderedRow, formatCell(cell, colWidths[i], align))
		}
		lines = append(lines, strings.TrimRight(strings.Join(renderedRow, t.Separator), " "))
	}

	return lines
}

func formatCell(text string, width int, align Alignment) string {
	switch align {
	case AlignRight:
		return PadLeft(text, width)
	case AlignCenter:
		return Center(text, width)
	default:
		return PadRight(text, width)
	}
}
