package mdcmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"nova/internal/command"
	"nova/internal/commands/cat"
	"nova/internal/commands/cat/syntax"
	"nova/internal/output"
	"nova/internal/renderer"
	"nova/internal/terminal"
	"nova/internal/theme"
)

// HeadingInfo holds metadata about a document heading.
type HeadingInfo struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
	Line  int    `json:"line"`
}

// Document represents a parsed Markdown file.
type Document struct {
	Title     string        `json:"title"`
	FilePath  string        `json:"file_path,omitempty"`
	WordCount int           `json:"word_count"`
	LineCount int           `json:"line_count"`
	Headings  []HeadingInfo `json:"headings"`
	Rendered  []string      `json:"-"`
	Plain     []string      `json:"-"`
}

// Command returns the registered Command instance for md.
func Command() *command.Command {
	return &command.Command{
		Name:        "md",
		Aliases:     []string{"doc", "view", "markdown"},
		Summary:     "Terminal Markdown document renderer and viewer",
		Usage:       "nova md [file] [flags]",
		Description: "Render Markdown files with rich typography, syntax-highlighted code blocks, tables, and TOC.",
		Phase:       27,
		Run:         Run,
	}
}

// Run executes the md command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)

	// Determine terminal width
	renderWidth := opts.Width
	if ctx.Caps.Width > 20 && (opts.Width == 80 || opts.Width > ctx.Caps.Width) {
		renderWidth = ctx.Caps.Width - 4
		if renderWidth > 120 {
			renderWidth = 120
		}
	}
	opts.Width = renderWidth

	var rawContent string
	displayName := "stdin"

	if opts.FilePath != "" {
		data, err := os.ReadFile(opts.FilePath)
		if err != nil {
			return fmt.Errorf("read file %q: %w", opts.FilePath, err)
		}
		rawContent = string(data)
		displayName = opts.FilePath
	} else {
		// Read from stdin if not TTY or data available
		data, err := io.ReadAll(ctx.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		rawContent = string(data)
		if strings.TrimSpace(rawContent) == "" {
			return fmt.Errorf("no markdown content provided: specify a file or pipe markdown via stdin")
		}
	}

	doc := ParseMarkdown(rawContent, opts.Width, ctx.Theme, ctx.Caps.ColorProfile)
	doc.FilePath = displayName

	if ctx.Printer.Mode == output.ModeJSON || opts.JSON {
		return ctx.Printer.PrintJSON(doc)
	}

	if opts.TOC {
		tocLines := RenderTOC(doc, opts.Width, ctx.Theme, ctx.Caps.ColorProfile)
		if ctx.Printer.Mode == output.ModePlain || opts.Plain {
			for _, line := range tocLines {
				fmt.Fprintln(ctx.Stdout, line)
			}
		} else {
			for _, line := range tocLines {
				fmt.Fprintln(ctx.Stdout, line)
			}
		}
		return nil
	}

	linesToOutput := doc.Rendered
	if ctx.Printer.Mode == output.ModePlain || opts.Plain {
		linesToOutput = doc.Plain
	}

	if opts.Pager && ctx.Caps.IsTTY && ctx.Printer.Mode != output.ModePlain && !opts.Plain {
		pager := cat.NewPager(linesToOutput, displayName, ctx.Caps.Width, ctx.Caps.Height, ctx.Theme, ctx.Caps.ColorProfile)
		return pager.Run()
	}

	for _, line := range linesToOutput {
		fmt.Fprintln(ctx.Stdout, line)
	}
	return nil
}

// ParseMarkdown converts raw markdown into styled and plain document representations.
func ParseMarkdown(raw string, width int, th *theme.Theme, profile terminal.ColorProfile) *Document {
	scanner := bufio.NewScanner(strings.NewReader(raw))
	doc := &Document{}

	var rendered []string
	var plain []string

	inCodeBlock := false
	codeBlockLang := ""
	var codeLines []string

	inTable := false
	var tableRows [][]string

	lineNum := 0
	wordCount := 0

	flushTable := func() {
		if len(tableRows) > 0 {
			rLines, pLines := renderTable(tableRows, width, th, profile)
			rendered = append(rendered, rLines...)
			plain = append(plain, pLines...)
			tableRows = nil
		}
		inTable = false
	}

	flushCodeBlock := func() {
		if inCodeBlock {
			rLines, pLines := renderCodeBlock(codeLines, codeBlockLang, width, th, profile)
			rendered = append(rendered, rLines...)
			plain = append(plain, pLines...)
			codeLines = nil
			inCodeBlock = false
		}
	}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Count words
		words := strings.Fields(line)
		wordCount += len(words)

		trimmed := strings.TrimSpace(line)

		// Code block delimiter
		if strings.HasPrefix(trimmed, "```") {
			if inTable {
				flushTable()
			}
			if inCodeBlock {
				flushCodeBlock()
			} else {
				inCodeBlock = true
				codeBlockLang = strings.TrimPrefix(trimmed, "```")
				codeBlockLang = strings.TrimSpace(codeBlockLang)
				codeLines = nil
			}
			continue
		}

		if inCodeBlock {
			codeLines = append(codeLines, line)
			continue
		}

		// Table row detection
		if strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|") {
			// Skip markdown divider line |---|---|
			if isTableDivider(trimmed) {
				continue
			}
			inTable = true
			cols := parseTableRow(trimmed)
			tableRows = append(tableRows, cols)
			continue
		} else if inTable {
			flushTable()
		}

		// Horizontal rule
		if trimmed == "---" || trimmed == "***" || trimmed == "___" {
			rule := strings.Repeat("─", width)
			rendered = append(rendered, th.Format(theme.RoleMuted, rule, profile))
			plain = append(plain, rule)
			continue
		}

		// Heading detection
		if strings.HasPrefix(trimmed, "#") {
			level := 0
			for level < len(trimmed) && trimmed[level] == '#' {
				level++
			}
			if level <= 6 && level < len(trimmed) && trimmed[level] == ' ' {
				text := strings.TrimSpace(trimmed[level:])
				if doc.Title == "" {
					doc.Title = text
				}
				doc.Headings = append(doc.Headings, HeadingInfo{
					Level: level,
					Text:  text,
					Line:  lineNum,
				})

				rHead, pHead := renderHeading(text, level, width, th, profile)
				rendered = append(rendered, rHead...)
				plain = append(plain, pHead...)
				continue
			}
		}

		// Blockquote
		if strings.HasPrefix(trimmed, ">") {
			quoteText := strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
			bar := th.Format(theme.RoleAccent, "│ ", profile)
			styled := renderInline(quoteText, th, profile)
			rendered = append(rendered, bar+th.Format(theme.RoleMuted, styled, profile))
			plain = append(plain, "| "+quoteText)
			continue
		}

		// Checkbox list
		if strings.HasPrefix(trimmed, "- [ ] ") || strings.HasPrefix(trimmed, "* [ ] ") {
			item := trimmed[6:]
			box := th.Format(theme.RoleMuted, "☐ ", profile)
			rendered = append(rendered, "  "+box+renderInline(item, th, profile))
			plain = append(plain, "  [ ] "+item)
			continue
		}
		if strings.HasPrefix(trimmed, "- [x] ") || strings.HasPrefix(trimmed, "* [x] ") ||
			strings.HasPrefix(trimmed, "- [X] ") || strings.HasPrefix(trimmed, "* [X] ") {
			item := trimmed[6:]
			box := th.Format(theme.RoleSuccess, "☑ ", profile)
			rendered = append(rendered, "  "+box+renderInline(item, th, profile))
			plain = append(plain, "  [x] "+item)
			continue
		}

		// Bullet list
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ") {
			item := trimmed[2:]
			bullet := th.Format(theme.RoleAccent, "• ", profile)
			rendered = append(rendered, "  "+bullet+renderInline(item, th, profile))
			plain = append(plain, "  * "+item)
			continue
		}

		// Numbered list
		if isNumberedList(trimmed) {
			idx := strings.Index(trimmed, ". ")
			numStr := trimmed[:idx+1]
			item := trimmed[idx+2:]
			numBadge := th.Format(theme.RoleAccent, numStr+" ", profile)
			rendered = append(rendered, "  "+numBadge+renderInline(item, th, profile))
			plain = append(plain, "  "+numStr+" "+item)
			continue
		}

		// Empty line
		if trimmed == "" {
			rendered = append(rendered, "")
			plain = append(plain, "")
			continue
		}

		// Normal paragraph text with wrap
		wrapped := wrapParagraph(trimmed, width)
		for _, wLine := range wrapped {
			rendered = append(rendered, renderInline(wLine, th, profile))
			plain = append(plain, wLine)
		}
	}

	if inTable {
		flushTable()
	}
	if inCodeBlock {
		flushCodeBlock()
	}

	doc.LineCount = lineNum
	doc.WordCount = wordCount
	doc.Rendered = rendered
	doc.Plain = plain

	if doc.Title == "" {
		doc.Title = "Untitled Document"
	}

	return doc
}

func isTableDivider(line string) bool {
	clean := strings.ReplaceAll(line, "|", "")
	clean = strings.ReplaceAll(clean, "-", "")
	clean = strings.ReplaceAll(clean, ":", "")
	clean = strings.ReplaceAll(clean, " ", "")
	return clean == ""
}

func parseTableRow(line string) []string {
	parts := strings.Split(line, "|")
	if len(parts) >= 2 {
		parts = parts[1 : len(parts)-1]
	}
	var cols []string
	for _, p := range parts {
		cols = append(cols, strings.TrimSpace(p))
	}
	return cols
}

func isNumberedList(s string) bool {
	idx := strings.Index(s, ". ")
	if idx <= 0 || idx > 4 {
		return false
	}
	for i := 0; i < idx; i++ {
		if !unicode.IsDigit(rune(s[i])) {
			return false
		}
	}
	return true
}

func wrapParagraph(text string, maxWidth int) []string {
	if maxWidth <= 10 {
		return []string{text}
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	currLine := words[0]

	for _, w := range words[1:] {
		if len(currLine)+1+len(w) <= maxWidth {
			currLine += " " + w
		} else {
			lines = append(lines, currLine)
			currLine = w
		}
	}
	lines = append(lines, currLine)
	return lines
}

func renderHeading(text string, level, width int, th *theme.Theme, profile terminal.ColorProfile) ([]string, []string) {
	var rLines []string
	var pLines []string

	rLines = append(rLines, "")
	pLines = append(pLines, "")

	switch level {
	case 1:
		border := strings.Repeat("═", width)
		titleLine := fmt.Sprintf("  # %s", strings.ToUpper(text))
		rLines = append(rLines, th.Format(theme.RoleAccent, border, profile))
		rLines = append(rLines, th.Format(theme.RoleDocument, titleLine, profile))
		rLines = append(rLines, th.Format(theme.RoleAccent, border, profile))

		pLines = append(pLines, border)
		pLines = append(pLines, titleLine)
		pLines = append(pLines, border)
	case 2:
		titleLine := fmt.Sprintf("## %s", text)
		div := strings.Repeat("─", width)
		rLines = append(rLines, th.Format(theme.RoleDocument, titleLine, profile))
		rLines = append(rLines, th.Format(theme.RoleMuted, div, profile))

		pLines = append(pLines, titleLine)
		pLines = append(pLines, div)
	case 3:
		titleLine := fmt.Sprintf("### %s", text)
		rLines = append(rLines, th.Format(theme.RoleAccent, titleLine, profile))
		pLines = append(pLines, titleLine)
	default:
		prefix := strings.Repeat("#", level) + " "
		titleLine := prefix + text
		rLines = append(rLines, th.Format(theme.RoleAccent, titleLine, profile))
		pLines = append(pLines, titleLine)
	}

	rLines = append(rLines, "")
	pLines = append(pLines, "")
	return rLines, pLines
}

func renderCodeBlock(lines []string, lang string, width int, th *theme.Theme, profile terminal.ColorProfile) ([]string, []string) {
	var rLines []string
	var pLines []string

	badge := " Code "
	if lang != "" {
		badge = fmt.Sprintf(" %s ", strings.ToUpper(lang))
	}
	topBar := fmt.Sprintf("┌─%s%s┐", badge, strings.Repeat("─", width-len(badge)-3))
	botBar := fmt.Sprintf("└%s┘", strings.Repeat("─", width-2))

	rLines = append(rLines, th.Format(theme.RoleMuted, topBar, profile))
	pLines = append(pLines, topBar)

	synState := &syntax.State{}
	for _, l := range lines {
		highlighted := l
		if lang != "" {
			highlighted = syntax.HighlightLine(l, lang, synState, th, profile)
		}
		avail := width - 4
		if avail < 5 {
			avail = 5
		}
		bar := th.Format(theme.RoleMuted, "│ ", profile)
		rLines = append(rLines, bar+renderer.Truncate(highlighted, avail, "…"))
		pLines = append(pLines, "│ "+l)
	}

	rLines = append(rLines, th.Format(theme.RoleMuted, botBar, profile))
	pLines = append(pLines, botBar)

	rLines = append(rLines, "")
	pLines = append(pLines, "")
	return rLines, pLines
}

func renderTable(rows [][]string, width int, th *theme.Theme, profile terminal.ColorProfile) ([]string, []string) {
	if len(rows) == 0 {
		return nil, nil
	}

	colCount := 0
	for _, r := range rows {
		if len(r) > colCount {
			colCount = len(r)
		}
	}
	if colCount == 0 {
		return nil, nil
	}

	colWidths := make([]int, colCount)
	for _, r := range rows {
		for i, c := range r {
			if len(c) > colWidths[i] {
				colWidths[i] = len(c)
			}
		}
	}

	// Bound maximum total width
	for i := range colWidths {
		if colWidths[i] < 4 {
			colWidths[i] = 4
		}
		if colWidths[i] > 30 {
			colWidths[i] = 30
		}
	}

	var rLines []string
	var pLines []string

	for rIdx, r := range rows {
		var cells []string
		for cIdx := 0; cIdx < colCount; cIdx++ {
			val := ""
			if cIdx < len(r) {
				val = r[cIdx]
			}
			w := colWidths[cIdx]
			cells = append(cells, fmt.Sprintf("%-*s", w, renderer.Truncate(val, w, "…")))
		}

		line := "| " + strings.Join(cells, " | ") + " |"
		if rIdx == 0 {
			rLines = append(rLines, th.Format(theme.RoleDocument, line, profile))
			pLines = append(pLines, line)

			// Separator
			var divs []string
			for _, w := range colWidths {
				divs = append(divs, strings.Repeat("-", w))
			}
			sep := "|-" + strings.Join(divs, "-|-") + "-|"
			rLines = append(rLines, th.Format(theme.RoleMuted, sep, profile))
			pLines = append(pLines, sep)
		} else {
			rLines = append(rLines, line)
			pLines = append(pLines, line)
		}
	}

	rLines = append(rLines, "")
	pLines = append(pLines, "")
	return rLines, pLines
}

func renderInline(text string, th *theme.Theme, profile terminal.ColorProfile) string {
	// Simple inline token replacement: **bold**, `code`, [text](url)
	res := text

	// Code `...`
	for {
		start := strings.Index(res, "`")
		if start == -1 {
			break
		}
		end := strings.Index(res[start+1:], "`")
		if end == -1 {
			break
		}
		end = start + 1 + end
		snippet := res[start+1 : end]
		styled := th.Format(theme.RoleAccent, snippet, profile)
		res = res[:start] + styled + res[end+1:]
	}

	// Bold **...**
	for {
		start := strings.Index(res, "**")
		if start == -1 {
			break
		}
		end := strings.Index(res[start+2:], "**")
		if end == -1 {
			break
		}
		end = start + 2 + end
		boldText := res[start+2 : end]
		styled := th.Format(theme.RoleDocument, boldText, profile)
		res = res[:start] + styled + res[end+2:]
	}

	return res
}

// RenderTOC formats the table of contents from extracted headings.
func RenderTOC(doc *Document, width int, th *theme.Theme, profile terminal.ColorProfile) []string {
	var lines []string
	cardWidth := width
	if cardWidth < 40 {
		cardWidth = 40
	}

	header := fmt.Sprintf("╭ 📑 Table of Contents: %s %s╮", doc.Title, strings.Repeat("─", cardWidth-26-len(doc.Title)))
	lines = append(lines, th.Format(theme.RoleAccent, renderer.Truncate(header, cardWidth, "─╮"), profile))

	if len(doc.Headings) == 0 {
		lines = append(lines, th.Format(theme.RoleMuted, "  (No headings found in document)", profile))
	} else {
		for _, h := range doc.Headings {
			indent := strings.Repeat("  ", h.Level-1)
			bullet := "• "
			if h.Level > 1 {
				bullet = "└─ "
			}
			item := fmt.Sprintf("  %s%s%s (L%d)", indent, bullet, h.Text, h.Line)
			role := theme.RoleDocument
			if h.Level > 2 {
				role = theme.RoleAccent
			}
			lines = append(lines, th.Format(role, renderer.Truncate(item, cardWidth-2, "…"), profile))
		}
	}

	meta := fmt.Sprintf("  Total Headings: %d │ Words: %d │ Lines: %d", len(doc.Headings), doc.WordCount, doc.LineCount)
	lines = append(lines, th.Format(theme.RoleMuted, strings.Repeat("─", cardWidth-2), profile))
	lines = append(lines, th.Format(theme.RoleMuted, meta, profile))
	lines = append(lines, th.Format(theme.RoleAccent, "╰"+strings.Repeat("─", cardWidth-2)+"╯", profile))

	return lines
}
