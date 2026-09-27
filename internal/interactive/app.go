package interactive

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	defer func() {
		if restore != nil {
			restore()
		}
	}()

	model.EditorRunner = func(filePath string) error {
		// Exit alternate screen and show cursor
		_, _ = tty.WriteString("\x1b[?25h\x1b[?1049l")
		if restore != nil {
			restore()
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = os.Getenv("VISUAL")
		}
		if editor == "" {
			for _, ed := range []string{"nano", "vim", "vi"} {
				if _, err := exec.LookPath(ed); err == nil {
					editor = ed
					break
				}
			}
		}
		if editor == "" {
			editor = "vi"
		}

		cmd := exec.Command(editor, filePath)
		cmd.Stdin = tty
		cmd.Stdout = tty
		cmd.Stderr = tty
		_ = cmd.Run()

		// Restore raw mode and re-enter alternate screen
		var rawErr error
		restore, rawErr = terminal.MakeRaw(int(tty.Fd()))
		_, _ = tty.WriteString("\x1b[?1049h\x1b[?25l")
		return rawErr
	}

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
				} else if m.ConfirmDelete {
					m.ConfirmDelete = false
					m.SetActionMessage("Delete canceled")
				} else if m.RenameActive {
					m.RenameActive = false
					m.SetActionMessage("Rename canceled")
				} else if m.NewFileActive {
					m.NewFileActive = false
					m.SetActionMessage("New file canceled")
				} else if m.NewFolderActive {
					m.NewFolderActive = false
					m.SetActionMessage("New folder canceled")
				} else if len(m.SelectedPaths) > 0 {
					m.ClearSelection()
					m.SetActionMessage("Selection cleared")
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
			} else {
				_ = reader.UnreadByte()
				if m.HelpActive {
					m.HelpActive = false
				} else if m.FilterActive {
					m.FilterActive = false
					m.FilterQuery = ""
					m.ApplyFilter()
				} else if m.ConfirmDelete {
					m.ConfirmDelete = false
					m.SetActionMessage("Delete canceled")
				} else if m.RenameActive {
					m.RenameActive = false
					m.SetActionMessage("Rename canceled")
				} else if m.NewFileActive {
					m.NewFileActive = false
					m.SetActionMessage("New file canceled")
				} else if m.NewFolderActive {
					m.NewFolderActive = false
					m.SetActionMessage("New folder canceled")
				} else if len(m.SelectedPaths) > 0 {
					m.ClearSelection()
					m.SetActionMessage("Selection cleared")
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

		// In Confirm Delete mode
		if m.ConfirmDelete {
			if b == 'y' || b == 'Y' {
				if len(m.SelectedPaths) > 0 {
					count := 0
					for p := range m.SelectedPaths {
						if err := os.RemoveAll(p); err == nil {
							count++
						}
					}
					m.ClearSelection()
					m.SetActionMessage(fmt.Sprintf("Deleted %d items", count))
				} else if entry := m.CurrentEntry(); entry != nil {
					if err := os.RemoveAll(entry.Path); err != nil {
						m.SetActionMessage("Error deleting: " + err.Error())
					} else {
						m.SetActionMessage("Deleted " + entry.Name)
					}
				}
				_ = m.LoadCurrentDir()
			} else {
				m.SetActionMessage("Delete canceled")
			}
			m.ConfirmDelete = false
			continue
		}

		// In Rename mode
		if m.RenameActive {
			if b == '\r' || b == '\n' {
				entry := m.CurrentEntry()
				newName := strings.TrimSpace(m.RenameInput)
				if entry != nil && newName != "" && newName != entry.Name {
					oldPath := entry.Path
					newPath := filepath.Join(m.CurrentDir, newName)
					if err := os.Rename(oldPath, newPath); err != nil {
						m.SetActionMessage("Rename failed: " + err.Error())
					} else {
						m.SetActionMessage("Renamed to " + newName)
						_ = m.LoadCurrentDir()
					}
				} else {
					m.SetActionMessage("Rename canceled")
				}
				m.RenameActive = false
			} else if b == 127 || b == 8 { // Backspace
				if len(m.RenameInput) > 0 {
					m.RenameInput = m.RenameInput[:len(m.RenameInput)-1]
				}
			} else if b == 27 {
				m.RenameActive = false
				m.SetActionMessage("Rename canceled")
			} else if b >= 32 && b <= 126 {
				m.RenameInput += string(b)
			}
			continue
		}

		// In New File mode
		if m.NewFileActive {
			if b == '\r' || b == '\n' {
				fileName := strings.TrimSpace(m.NewFileInput)
				if fileName != "" {
					newPath := filepath.Join(m.CurrentDir, fileName)
					if err := os.WriteFile(newPath, []byte(""), 0644); err != nil {
						m.SetActionMessage("Create file failed: " + err.Error())
					} else {
						m.SetActionMessage("Created file: " + fileName)
						_ = m.LoadCurrentDir()
					}
				} else {
					m.SetActionMessage("New file canceled")
				}
				m.NewFileActive = false
			} else if b == 127 || b == 8 { // Backspace
				if len(m.NewFileInput) > 0 {
					m.NewFileInput = m.NewFileInput[:len(m.NewFileInput)-1]
				}
			} else if b == 27 {
				m.NewFileActive = false
				m.SetActionMessage("New file canceled")
			} else if b >= 32 && b <= 126 {
				m.NewFileInput += string(b)
			}
			continue
		}

		// In New Folder mode
		if m.NewFolderActive {
			if b == '\r' || b == '\n' {
				dirName := strings.TrimSpace(m.NewFolderInput)
				if dirName != "" {
					newPath := filepath.Join(m.CurrentDir, dirName)
					if err := os.MkdirAll(newPath, 0755); err != nil {
						m.SetActionMessage("Create folder failed: " + err.Error())
					} else {
						m.SetActionMessage("Created folder: " + dirName)
						_ = m.LoadCurrentDir()
					}
				} else {
					m.SetActionMessage("New folder canceled")
				}
				m.NewFolderActive = false
			} else if b == 127 || b == 8 { // Backspace
				if len(m.NewFolderInput) > 0 {
					m.NewFolderInput = m.NewFolderInput[:len(m.NewFolderInput)-1]
				}
			} else if b == 27 {
				m.NewFolderActive = false
				m.SetActionMessage("New folder canceled")
			} else if b >= 32 && b <= 126 {
				m.NewFolderInput += string(b)
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
		case ' ':
			m.ToggleSelect()
		case 'n':
			m.NewFileActive = true
			m.NewFileInput = ""
		case 'N':
			m.NewFolderActive = true
			m.NewFolderInput = ""
		case 'p':
			m.PreviewCollapsed = !m.PreviewCollapsed
			m.UpdatePreview()
		case 'c':
			if len(m.SelectedPaths) > 0 {
				m.SetActionMessage(fmt.Sprintf("Copied %d path(s)", len(m.SelectedPaths)))
			} else if entry := m.CurrentEntry(); entry != nil {
				m.SetActionMessage("Copied: " + entry.Path)
			}
		case 'd':
			if len(m.SelectedPaths) > 0 {
				m.ConfirmDelete = true
				m.SetActionMessage(fmt.Sprintf("Delete %d selected item(s)? (y/n)", len(m.SelectedPaths)))
			} else if entry := m.CurrentEntry(); entry != nil {
				m.ConfirmDelete = true
				m.SetActionMessage(fmt.Sprintf("Delete '%s'? (y/n)", entry.Name))
			}
		case 'r':
			if entry := m.CurrentEntry(); entry != nil {
				m.RenameActive = true
				m.RenameInput = entry.Name
				m.SetActionMessage("Rename: " + entry.Name)
			}
		case 'e':
			if entry := m.CurrentEntry(); entry != nil && !entry.IsDir {
				if m.EditorRunner != nil {
					_ = m.EditorRunner(entry.Path)
					_ = m.LoadCurrentDir()
				} else {
					m.SetActionMessage("Editor: " + entry.Name)
				}
			}
		case '/':
			m.FilterActive = true
			m.FilterQuery = ""
		case '.':
			_ = m.ToggleHidden()
		case '?':
			m.ToggleHelp()
		case 't':
			m.CycleTheme()
		case 's':
			m.CycleSort()
		case 'x':
			m.ToggleHex()
		case 'g':
			m.MoveHome()
		case 'G':
			m.MoveEnd()
		}
	}
}
