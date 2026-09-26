package cat

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"nova/internal/renderer"
	"nova/internal/terminal"
	"nova/internal/theme"
)

// Pager manages interactive viewport display for long files.
type Pager struct {
	Lines         []string
	Title         string
	Width         int
	Height        int
	TopLine       int
	SearchQuery   string
	SearchMatches []int
	CurrentMatch  int
	Theme         *theme.Theme
	Profile       terminal.ColorProfile
	InSearch      bool
	SearchInput   string
	ShowHelp      bool
}

// NewPager constructs an initialized pager.
func NewPager(lines []string, title string, width, height int, th *theme.Theme, profile terminal.ColorProfile) *Pager {
	if height < 5 {
		height = 24
	}
	if width < 10 {
		width = 80
	}
	return &Pager{
		Lines:   lines,
		Title:   title,
		Width:   width,
		Height:  height,
		TopLine: 0,
		Theme:   th,
		Profile: profile,
	}
}

// Run executes the interactive event loop on the controlling terminal.
func (p *Pager) Run() error {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		// No controlling terminal available; fallback to standard print
		for _, line := range p.Lines {
			fmt.Println(line)
		}
		return nil
	}
	defer tty.Close()

	restore, err := terminal.MakeRaw(int(tty.Fd()))
	if err != nil {
		for _, line := range p.Lines {
			fmt.Println(line)
		}
		return nil
	}
	defer restore()

	// Switch to alternate screen and hide cursor
	tty.WriteString("\x1b[?1049h\x1b[?25l")
	defer tty.WriteString("\x1b[?25h\x1b[?1049l")

	reader := bufio.NewReader(tty)

	for {
		p.render(tty)

		// Read key input
		b, err := reader.ReadByte()
		if err != nil {
			break
		}

		if p.InSearch {
			if b == '\r' || b == '\n' {
				p.InSearch = false
				p.SearchQuery = p.SearchInput
				p.performSearch()
			} else if b == 27 { // Escape
				p.InSearch = false
				p.SearchInput = ""
			} else if b == 127 || b == 8 { // Backspace
				if len(p.SearchInput) > 0 {
					p.SearchInput = p.SearchInput[:len(p.SearchInput)-1]
				}
			} else if b >= 32 && b <= 126 {
				p.SearchInput += string(b)
			}
			continue
		}

		if p.ShowHelp {
			p.ShowHelp = false
			continue
		}

		// Escape sequence handling
		if b == 27 {
			if reader.Buffered() > 0 {
				b2, _ := reader.ReadByte()
				if b2 == '[' {
					b3, _ := reader.ReadByte()
					switch b3 {
					case 'A': // Up
						p.scrollUp(1)
					case 'B': // Down
						p.scrollDown(1)
					case '5': // PageUp ([5~)
						if reader.Buffered() > 0 {
							reader.ReadByte()
						}
						p.scrollUp(p.Height - 2)
					case '6': // PageDown ([6~)
						if reader.Buffered() > 0 {
							reader.ReadByte()
						}
						p.scrollDown(p.Height - 2)
					case 'H': // Home
						p.TopLine = 0
					case 'F': // End
						p.scrollToBottom()
					}
				}
			} else {
				// Standalone escape quits
				break
			}
			continue
		}

		// Normal key bindings
		switch b {
		case 'q', 'Q', 3: // q, Q, Ctrl+C
			return nil
		case 'j', '\r', '\n': // Down
			p.scrollDown(1)
		case 'k': // Up
			p.scrollUp(1)
		case 'f', ' ': // PageDown / Space
			p.scrollDown(p.Height - 2)
		case 'b': // PageUp
			p.scrollUp(p.Height - 2)
		case 'g': // Top
			p.TopLine = 0
		case 'G': // Bottom
			p.scrollToBottom()
		case '/': // Search
			p.InSearch = true
			p.SearchInput = ""
		case 'n': // Next search match
			p.nextMatch()
		case 'N': // Previous search match
			p.prevMatch()
		case 'h', '?': // Help
			p.ShowHelp = true
		}
	}

	return nil
}

func (p *Pager) scrollDown(n int) {
	maxTop := len(p.Lines) - (p.Height - 1)
	if maxTop < 0 {
		maxTop = 0
	}
	p.TopLine += n
	if p.TopLine > maxTop {
		p.TopLine = maxTop
	}
}

func (p *Pager) scrollUp(n int) {
	p.TopLine -= n
	if p.TopLine < 0 {
		p.TopLine = 0
	}
}

func (p *Pager) scrollToBottom() {
	maxTop := len(p.Lines) - (p.Height - 1)
	if maxTop < 0 {
		maxTop = 0
	}
	p.TopLine = maxTop
}

func (p *Pager) performSearch() {
	p.SearchMatches = nil
	p.CurrentMatch = 0
	if p.SearchQuery == "" {
		return
	}

	query := strings.ToLower(p.SearchQuery)
	for idx, line := range p.Lines {
		plainLine := strings.ToLower(renderer.StripANSI(line))
		if strings.Contains(plainLine, query) {
			p.SearchMatches = append(p.SearchMatches, idx)
		}
	}

	// Jump to first match at or after TopLine
	for i, matchLine := range p.SearchMatches {
		if matchLine >= p.TopLine {
			p.CurrentMatch = i
			p.TopLine = matchLine
			return
		}
	}

	if len(p.SearchMatches) > 0 {
		p.CurrentMatch = 0
		p.TopLine = p.SearchMatches[0]
	}
}

func (p *Pager) nextMatch() {
	if len(p.SearchMatches) == 0 {
		return
	}
	p.CurrentMatch = (p.CurrentMatch + 1) % len(p.SearchMatches)
	p.TopLine = p.SearchMatches[p.CurrentMatch]
}

func (p *Pager) prevMatch() {
	if len(p.SearchMatches) == 0 {
		return
	}
	p.CurrentMatch = (p.CurrentMatch - 1 + len(p.SearchMatches)) % len(p.SearchMatches)
	p.TopLine = p.SearchMatches[p.CurrentMatch]
}

func (p *Pager) render(w io.Writer) {
	var b strings.Builder
	// Move cursor to top-left and clear screen
	b.WriteString("\x1b[H\x1b[2J")

	if p.ShowHelp {
		b.WriteString(p.renderHelp())
		w.Write([]byte(b.String()))
		return
	}

	contentHeight := p.Height - 1
	for i := 0; i < contentHeight; i++ {
		lineIdx := p.TopLine + i
		if lineIdx < len(p.Lines) {
			line := p.Lines[lineIdx]
			// Highlight search matches
			if p.SearchQuery != "" {
				line = highlightMatchInLine(line, p.SearchQuery, p.Theme, p.Profile)
			}
			b.WriteString(line)
		}
		b.WriteString("\r\n")
	}

	// Status bar at the bottom
	statusBar := p.renderStatusBar()
	b.WriteString(statusBar)

	w.Write([]byte(b.String()))
}

func (p *Pager) renderStatusBar() string {
	if p.InSearch {
		prompt := fmt.Sprintf("/%s_", p.SearchInput)
		return p.Theme.Format(theme.RoleSelection, prompt, p.Profile)
	}

	total := len(p.Lines)
	currentEnd := p.TopLine + p.Height - 1
	if currentEnd > total {
		currentEnd = total
	}
	percent := 0
	if total > 0 {
		percent = (currentEnd * 100) / total
	}

	status := fmt.Sprintf(" %s │ Lines %d-%d of %d (%d%%)", p.Title, p.TopLine+1, currentEnd, total, percent)
	if len(p.SearchMatches) > 0 {
		status += fmt.Sprintf(" │ Match %d/%d (%q)", p.CurrentMatch+1, len(p.SearchMatches), p.SearchQuery)
	} else if p.SearchQuery != "" {
		status += fmt.Sprintf(" │ Pattern not found: %q", p.SearchQuery)
	}
	status += " │ Press '?' for help, 'q' to quit "

	// Pad or truncate to terminal width
	visWidth := renderer.VisibleWidth(status)
	if visWidth < p.Width {
		status += strings.Repeat(" ", p.Width-visWidth)
	}

	return p.Theme.Format(theme.RoleSelection, status, p.Profile)
}

func (p *Pager) renderHelp() string {
	helpText := []string{
		"  nova cat — Interactive Controls:",
		"  ──────────────────────────────────────────",
		"  ↑ / k           Scroll up 1 line",
		"  ↓ / j           Scroll down 1 line",
		"  PageDown / Space Scroll down 1 page",
		"  PageUp / b      Scroll up 1 page",
		"  g / Home        Jump to beginning of file",
		"  G / End         Jump to end of file",
		"  /               Search pattern",
		"  n               Next search match",
		"  N               Previous search match",
		"  q / Ctrl+C      Exit pager",
		"  ──────────────────────────────────────────",
		"  Press any key to return to document...",
	}

	var b strings.Builder
	for _, l := range helpText {
		b.WriteString(l + "\r\n")
	}
	return b.String()
}

func highlightMatchInLine(line string, query string, th *theme.Theme, profile terminal.ColorProfile) string {
	if query == "" || profile == terminal.ColorNone {
		return line
	}
	lowerLine := strings.ToLower(line)
	lowerQuery := strings.ToLower(query)
	idx := strings.Index(lowerLine, lowerQuery)
	if idx == -1 {
		return line
	}

	// Simple match highlight
	match := line[idx : idx+len(query)]
	highlighted := th.Format(theme.RoleSelection, match, profile)
	return line[:idx] + highlighted + line[idx+len(query):]
}
