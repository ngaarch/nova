package interactive

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"nova/internal/terminal"
	"nova/internal/theme"
)

// Run launches the interactive terminal application on the controlling terminal.
func Run(initialDir string, showHidden bool, caps terminal.Capabilities, th *theme.Theme) error {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open /dev/tty: %w (interactive mode requires a controlling terminal)", err)
	}
	defer tty.Close()

	w, h := terminal.GetSize(tty, os.Getenv)
	if w <= 0 {
		w = caps.Width
	}
	if h <= 0 {
		h = caps.Height
	}

	model, err := NewModel(initialDir, w, h, showHidden, th, caps.ColorProfile, caps.UnicodeSupported)
	if err != nil {
		return err
	}

	// Put terminal in raw mode
	restore, err := terminal.MakeRaw(int(tty.Fd()))
	if err != nil {
		return fmt.Errorf("make raw: %w", err)
	}
	defer restore()

	// Switch to alternate screen and hide cursor
	_, _ = tty.WriteString("\x1b[?1049h\x1b[?25l")
	defer func() {
		_, _ = tty.WriteString("\x1b[?25h\x1b[?1049l")
	}()

	// Setup signal listener for dynamic terminal resizing
	sigChan, stopSignal := setupResizeSignal()
	defer stopSignal()

	done := make(chan struct{})
	defer close(done)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-sigChan:
				newW, newH := terminal.GetSize(tty, os.Getenv)
				if newW > 0 && newH > 0 {
					model.Resize(newW, newH)
					model.RLock()
					frame := Render(model)
					model.RUnlock()
					_, _ = tty.WriteString(frame)
				}
			}
		}
	}()

	return RunLoop(model, tty, tty)
}

// RunLoop executes the input processing and rendering loop.
func RunLoop(m *Model, in io.Reader, out io.Writer) error {
	reader := bufio.NewReader(in)

	for {
		// Render current state
		m.RLock()
		frame := Render(m)
		m.RUnlock()
		if _, err := io.WriteString(out, frame); err != nil {
			return err
		}

		// Read byte
		b, err := reader.ReadByte()
		if err != nil {
			return nil
		}

		// Escape sequences
		if b == 27 {
			if reader.Buffered() == 0 {
				// Standalone Esc
				if m.HelpActive {
					m.HelpActive = false
				} else if m.FilterActive {
					m.FilterActive = false
					m.FilterQuery = ""
					m.ApplyFilter()
				}
				continue
			}

			b2, err := reader.ReadByte()
			if err != nil {
				continue
			}

			if b2 == '[' {
				b3, err := reader.ReadByte()
				if err != nil {
					continue
				}

				switch b3 {
				case 'A': // Up arrow
					m.MoveCursor(-1)
				case 'B': // Down arrow
					m.MoveCursor(1)
				case 'C': // Right arrow
					_, _ = m.EnterSelected()
				case 'D': // Left arrow
					_ = m.GoToParent()
				case 'H': // Home
					m.MoveHome()
				case 'F': // End
					m.MoveEnd()
				case '5': // Page Up (often \x1b[5~)
					if reader.Buffered() > 0 {
						_, _ = reader.ReadByte() // Consume ~
					}
					m.PageMove(-1)
				case '6': // Page Down (often \x1b[6~)
					if reader.Buffered() > 0 {
						_, _ = reader.ReadByte() // Consume ~
					}
					m.PageMove(1)
				}
			}
			continue
		}

		// In Help modal: any key (q, ?, Esc, Enter) dismisses it
		if m.HelpActive {
			if b == '?' || b == 'q' || b == 27 || b == '\r' || b == '\n' {
				m.HelpActive = false
			}
			continue
		}

		// In Filter mode
		if m.FilterActive {
			if b == '\r' || b == '\n' {
				m.FilterActive = false
			} else if b == 127 || b == 8 { // Backspace
				if len(m.FilterQuery) > 0 {
					m.FilterQuery = m.FilterQuery[:len(m.FilterQuery)-1]
					m.ApplyFilter()
				} else {
					m.FilterActive = false
				}
			} else if b >= 32 && b <= 126 {
				m.FilterQuery += string(b)
				m.ApplyFilter()
			}
			continue
		}

		// Normal navigation mode
		switch b {
		case 'q', 3: // 'q' or Ctrl+C
			return nil
		case 'j':
			m.MoveCursor(1)
		case 'k':
			m.MoveCursor(-1)
		case 'h':
			_ = m.GoToParent()
		case 'l':
			_, _ = m.EnterSelected()
		case '\r', '\n':
			_, _ = m.EnterSelected()
		case 127, 8: // Backspace
			_ = m.GoToParent()
		case '/':
			m.FilterActive = true
			m.FilterQuery = ""
		case '.':
			_ = m.ToggleHidden()
		case '?':
			m.ToggleHelp()
		case 'g':
			m.MoveHome()
		case 'G':
			m.MoveEnd()
		}
	}
}
