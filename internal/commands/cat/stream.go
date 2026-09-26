package cat

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"nova/internal/commands/cat/syntax"
	"nova/internal/terminal"
	"nova/internal/theme"
)

// LineProcessor processes and formats a single line of input.
type LineProcessor struct {
	Options         Options
	Theme           *theme.Theme
	Profile         terminal.ColorProfile
	Language        string
	SyntaxState     *syntax.State
	LineNumber      int
	PrevLineBlank   bool
	MaxNumberWidth  int
}

// NewLineProcessor creates an initialized processor for streaming lines.
func NewLineProcessor(opts Options, th *theme.Theme, profile terminal.ColorProfile, lang string) *LineProcessor {
	return &LineProcessor{
		Options:        opts,
		Theme:          th,
		Profile:        profile,
		Language:       lang,
		SyntaxState:    &syntax.State{},
		LineNumber:     0,
		PrevLineBlank:  false,
		MaxNumberWidth: 4,
	}
}

// ProcessLine formats a single raw line according to flags and syntax rules.
// Returns formatted string and a boolean indicating whether the line should be emitted (false if squeezed).
func (p *LineProcessor) ProcessLine(raw []byte) (string, bool) {
	// Strip trailing newline character
	hasCR := false
	if len(raw) > 0 && raw[len(raw)-1] == '\n' {
		raw = raw[:len(raw)-1]
		if len(raw) > 0 && raw[len(raw)-1] == '\r' {
			raw = raw[:len(raw)-1]
			hasCR = true
		}
	}

	isBlank := len(bytes.TrimSpace(raw)) == 0

	// Squeeze blank lines (-s)
	if p.Options.SqueezeBlank && isBlank && p.PrevLineBlank {
		return "", false
	}
	p.PrevLineBlank = isBlank

	// Clean / sanitize invalid UTF-8
	lineStr := sanitizeUTF8(raw)

	// Replace tabs if requested (-T)
	if p.Options.ShowTabs {
		if p.Profile != terminal.ColorNone {
			tabRep := p.Theme.Format(theme.RoleMuted, "^I", p.Profile)
			lineStr = strings.ReplaceAll(lineStr, "\t", tabRep)
		} else {
			lineStr = strings.ReplaceAll(lineStr, "\t", "^I")
		}
	}

	// Show carriage returns and control codes if requested (-A / -v)
	if p.Options.ShowAll && hasCR {
		if p.Profile != terminal.ColorNone {
			crRep := p.Theme.Format(theme.RoleMuted, "^M", p.Profile)
			lineStr += crRep
		} else {
			lineStr += "^M"
		}
	}

	// Apply syntax highlighting
	var content string
	if !p.Options.Plain && p.Profile != terminal.ColorNone && p.Language != "text" && p.Language != "" {
		content = syntax.HighlightLine(lineStr, p.Language, p.SyntaxState, p.Theme, p.Profile)
	} else {
		content = lineStr
	}

	// Show line ends (-E)
	if p.Options.ShowEnds {
		if p.Profile != terminal.ColorNone {
			content += p.Theme.Format(theme.RoleMuted, "$", p.Profile)
		} else {
			content += "$"
		}
	}

	// Line numbering (-n or -b)
	if p.Options.Number {
		if p.Options.NumberNonblank && isBlank {
			// Don't number blank line with -b, but pad to keep column alignment
			prefix := strings.Repeat(" ", p.MaxNumberWidth) + " │ "
			if p.Profile != terminal.ColorNone {
				prefix = p.Theme.Format(theme.RoleMuted, prefix, p.Profile)
			}
			return prefix + content, true
		}

		p.LineNumber++
		numStr := fmt.Sprintf("%*d", p.MaxNumberWidth, p.LineNumber)
		sep := " │ "
		if p.Profile != terminal.ColorNone {
			prefix := p.Theme.Format(theme.RoleMuted, numStr+sep, p.Profile)
			return prefix + content, true
		}
		return numStr + sep + content, true
	}

	return content, true
}

// sanitizeUTF8 converts invalid UTF-8 byte sequences into safe representations.
func sanitizeUTF8(data []byte) string {
	if utf8.Valid(data) {
		return string(data)
	}

	var b strings.Builder
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r == utf8.RuneError && size == 1 {
			b.WriteString(fmt.Sprintf("\\x%02x", data[0]))
			data = data[1:]
		} else {
			b.WriteRune(r)
			data = data[size:]
		}
	}
	return b.String()
}

// StreamDirect streams lines directly to an output writer with constant memory.
func StreamDirect(r io.Reader, out io.Writer, proc *LineProcessor) error {
	reader := bufio.NewReaderSize(r, 64*1024)
	writer := bufio.NewWriterSize(out, 64*1024)
	defer writer.Flush()

	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			formatted, emit := proc.ProcessLine(line)
			if emit {
				if _, wErr := writer.WriteString(formatted + "\n"); wErr != nil {
					return wErr
				}
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	return nil
}

// FormatJSONData pretty-prints JSON if valid, or returns error.
func FormatJSONData(raw []byte) ([]byte, error) {
	var val interface{}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&val); err != nil {
		return nil, err
	}
	return json.MarshalIndent(val, "", "  ")
}
