package notecmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"nova/internal/command"
	"nova/internal/output"
	"nova/internal/theme"
)

// TodoItem represents an actionable task extracted from a note.
type TodoItem struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
	Done  bool   `json:"done"`
}

// Note represents a markdown note with extracted task items.
type Note struct {
	ID         string     `json:"id"`
	Title      string     `json:"title"`
	Tags       []string   `json:"tags,omitempty"`
	Content    string     `json:"content"`
	Created    time.Time  `json:"created"`
	Updated    time.Time  `json:"updated"`
	Todos      []TodoItem `json:"todos,omitempty"`
	DoneCount  int        `json:"done_count"`
	TotalTodos int        `json:"total_todos"`
}

// Command returns the registered Command instance for note.
func Command() *command.Command {
	return &command.Command{
		Name:        "note",
		Aliases:     []string{"notes", "todo", "memo"},
		Summary:     "Terminal notes, memo scratchpad, and TODO task tracker",
		Usage:       "nova note [list|add|show|todo|toggle|rm|search] [flags]",
		Description: "Capture notes, organize developer scratchpads, and track markdown todo checklists.",
		Phase:       28,
		Run:         Run,
	}
}

// Run executes the note command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)
	notesDir := resolveNotesDir(opts.Dir)
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		return fmt.Errorf("create notes dir %q: %w", notesDir, err)
	}

	switch opts.Action {
	case ActionAdd:
		return handleAdd(ctx, notesDir, opts)
	case ActionShow:
		return handleShow(ctx, notesDir, opts)
	case ActionTodo:
		return handleTodo(ctx, notesDir, opts)
	case ActionToggle:
		return handleToggle(ctx, notesDir, opts)
	case ActionDelete:
		return handleDelete(ctx, notesDir, opts)
	case ActionSearch:
		return handleSearch(ctx, notesDir, opts)
	case ActionList:
		fallthrough
	default:
		return handleList(ctx, notesDir, opts)
	}
}

func resolveNotesDir(override string) string {
	if override != "" {
		return override
	}
	// Check local workspace .nova/notes
	if fi, err := os.Stat(".nova/notes"); err == nil && fi.IsDir() {
		return ".nova/notes"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".config", "nova", "notes")
}

func generateID(title string) string {
	now := time.Now()
	slug := strings.ToLower(title)
	var sb strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
		} else if sb.Len() > 0 && sb.String()[sb.Len()-1] != '-' {
			sb.WriteRune('-')
		}
	}
	prefix := strings.Trim(sb.String(), "-")
	if len(prefix) > 20 {
		prefix = prefix[:20]
	}
	if prefix == "" {
		prefix = "note"
	}
	return fmt.Sprintf("%s-%s", prefix, now.Format("0102-150405"))
}

func extractTodos(content string) ([]TodoItem, int, int) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var todos []TodoItem
	idx := 1
	doneCount := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "- [ ] ") || strings.HasPrefix(line, "* [ ] ") {
			text := strings.TrimSpace(line[6:])
			todos = append(todos, TodoItem{
				Index: idx,
				Text:  text,
				Done:  false,
			})
			idx++
		} else if strings.HasPrefix(line, "- [x] ") || strings.HasPrefix(line, "* [x] ") ||
			strings.HasPrefix(line, "- [X] ") || strings.HasPrefix(line, "* [X] ") {
			text := strings.TrimSpace(line[6:])
			todos = append(todos, TodoItem{
				Index: idx,
				Text:  text,
				Done:  true,
			})
			doneCount++
			idx++
		}
	}

	return todos, doneCount, len(todos)
}

func loadNotes(dir string) ([]*Note, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var notes []*Note
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var note Note
		if err := json.Unmarshal(data, &note); err == nil {
			todos, done, total := extractTodos(note.Content)
			note.Todos = todos
			note.DoneCount = done
			note.TotalTodos = total
			notes = append(notes, &note)
		}
	}

	sort.Slice(notes, func(i, j int) bool {
		return notes[i].Updated.After(notes[j].Updated)
	})

	return notes, nil
}

func saveNote(dir string, note *Note) error {
	todos, done, total := extractTodos(note.Content)
	note.Todos = todos
	note.DoneCount = done
	note.TotalTodos = total

	data, err := json.MarshalIndent(note, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, note.ID+".json")
	return os.WriteFile(path, data, 0644)
}

func findNote(dir, idOrTitle string) (*Note, error) {
	notes, err := loadNotes(dir)
	if err != nil {
		return nil, err
	}

	target := strings.ToLower(strings.TrimSpace(idOrTitle))
	for _, n := range notes {
		if strings.ToLower(n.ID) == target || strings.ToLower(n.Title) == target {
			return n, nil
		}
	}
	// Prefix match
	for _, n := range notes {
		if strings.HasPrefix(strings.ToLower(n.ID), target) || strings.Contains(strings.ToLower(n.Title), target) {
			return n, nil
		}
	}
	return nil, fmt.Errorf("note %q not found", idOrTitle)
}

func handleAdd(ctx *command.Context, dir string, opts Options) error {
	if strings.TrimSpace(opts.Title) == "" {
		return fmt.Errorf("note title required: e.g. nova note add \"Sprint Tasks\" -c \"- [ ] Write docs\"")
	}

	note := &Note{
		ID:      generateID(opts.Title),
		Title:   opts.Title,
		Tags:    opts.Tags,
		Content: opts.Content,
		Created: time.Now(),
		Updated: time.Now(),
	}

	if err := saveNote(dir, note); err != nil {
		return fmt.Errorf("save note: %w", err)
	}

	if ctx.Printer.Mode == output.ModeJSON || opts.JSON {
		return ctx.Printer.PrintJSON(note)
	}

	if ctx.Printer.Mode == output.ModePlain || opts.Plain {
		fmt.Fprintf(ctx.Stdout, "%s\t%s\t%s\n", note.ID, note.Title, strings.Join(note.Tags, ","))
	} else {
		RenderNoteAdded(ctx.Stdout, note, ctx)
	}
	return nil
}

func handleShow(ctx *command.Context, dir string, opts Options) error {
	if opts.ID == "" {
		return fmt.Errorf("note ID or title required: nova note show <id>")
	}
	note, err := findNote(dir, opts.ID)
	if err != nil {
		return err
	}

	if ctx.Printer.Mode == output.ModeJSON || opts.JSON {
		return ctx.Printer.PrintJSON(note)
	}

	if ctx.Printer.Mode == output.ModePlain || opts.Plain {
		fmt.Fprintf(ctx.Stdout, "ID:\t%s\nTitle:\t%s\nTags:\t%s\nContent:\n%s\n",
			note.ID, note.Title, strings.Join(note.Tags, ","), note.Content)
	} else {
		RenderNoteDetail(ctx.Stdout, note, ctx)
	}
	return nil
}

func handleList(ctx *command.Context, dir string, opts Options) error {
	notes, err := loadNotes(dir)
	if err != nil {
		return fmt.Errorf("load notes: %w", err)
	}

	if ctx.Printer.Mode == output.ModeJSON || opts.JSON {
		return ctx.Printer.PrintJSON(notes)
	}

	if ctx.Printer.Mode == output.ModePlain || opts.Plain {
		RenderNotesPlain(ctx.Stdout, notes)
	} else {
		RenderNotesDashboard(ctx.Stdout, notes, ctx)
	}
	return nil
}

func handleTodo(ctx *command.Context, dir string, opts Options) error {
	notes, err := loadNotes(dir)
	if err != nil {
		return fmt.Errorf("load notes: %w", err)
	}

	var allTodos []struct {
		NoteID    string   `json:"note_id"`
		NoteTitle string   `json:"note_title"`
		Item      TodoItem `json:"item"`
	}

	for _, n := range notes {
		for _, t := range n.Todos {
			allTodos = append(allTodos, struct {
				NoteID    string   `json:"note_id"`
				NoteTitle string   `json:"note_title"`
				Item      TodoItem `json:"item"`
			}{
				NoteID:    n.ID,
				NoteTitle: n.Title,
				Item:      t,
			})
		}
	}

	if ctx.Printer.Mode == output.ModeJSON || opts.JSON {
		return ctx.Printer.PrintJSON(allTodos)
	}

	if ctx.Printer.Mode == output.ModePlain || opts.Plain {
		for _, item := range allTodos {
			status := "[ ]"
			if item.Item.Done {
				status = "[x]"
			}
			fmt.Fprintf(ctx.Stdout, "%s\t%s\t%d\t%s\t%s\n", status, item.NoteID, item.Item.Index, item.NoteTitle, item.Item.Text)
		}
	} else {
		RenderTodoDashboard(ctx.Stdout, notes, ctx)
	}
	return nil
}

func handleToggle(ctx *command.Context, dir string, opts Options) error {
	if opts.ID == "" {
		return fmt.Errorf("note ID required: nova note toggle <id> [task-num]")
	}
	note, err := findNote(dir, opts.ID)
	if err != nil {
		return err
	}

	targetIdx := opts.ItemIdx
	if targetIdx <= 0 {
		targetIdx = 1
	}

	// Toggle in content
	lines := strings.Split(note.Content, "\n")
	currIdx := 1
	toggled := false
	newStatus := false

	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "- [ ] ") || strings.HasPrefix(trimmed, "* [ ] ") {
			if currIdx == targetIdx {
				lines[i] = strings.Replace(l, "[ ]", "[x]", 1)
				toggled = true
				newStatus = true
				break
			}
			currIdx++
		} else if strings.HasPrefix(trimmed, "- [x] ") || strings.HasPrefix(trimmed, "* [x] ") ||
			strings.HasPrefix(trimmed, "- [X] ") || strings.HasPrefix(trimmed, "* [X] ") {
			if currIdx == targetIdx {
				lines[i] = strings.Replace(l, "[x]", "[ ]", 1)
				lines[i] = strings.Replace(lines[i], "[X]", "[ ]", 1)
				toggled = true
				newStatus = false
				break
			}
			currIdx++
		}
	}

	if !toggled {
		return fmt.Errorf("task #%d not found in note %q (note has %d tasks)", targetIdx, note.Title, len(note.Todos))
	}

	note.Content = strings.Join(lines, "\n")
	note.Updated = time.Now()
	if err := saveNote(dir, note); err != nil {
		return fmt.Errorf("save updated note: %w", err)
	}

	statusStr := "completed"
	if !newStatus {
		statusStr = "pending"
	}

	if ctx.Printer.Mode == output.ModePlain || opts.Plain {
		fmt.Fprintf(ctx.Stdout, "Toggled task #%d to %s in %s\n", targetIdx, statusStr, note.ID)
	} else {
		fmt.Fprintln(ctx.Stdout, ctx.Theme.Format(theme.RoleSuccess, fmt.Sprintf("✔ Toggled task #%d to %s in %q", targetIdx, statusStr, note.Title), ctx.Caps.ColorProfile))
	}
	return nil
}

func handleDelete(ctx *command.Context, dir string, opts Options) error {
	if opts.ID == "" {
		return fmt.Errorf("note ID required: nova note rm <id>")
	}
	note, err := findNote(dir, opts.ID)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, note.ID+".json")
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete note: %w", err)
	}

	if ctx.Printer.Mode == output.ModePlain || opts.Plain {
		fmt.Fprintf(ctx.Stdout, "Deleted %s\n", note.ID)
	} else {
		fmt.Fprintln(ctx.Stdout, ctx.Theme.Format(theme.RoleSuccess, fmt.Sprintf("✔ Deleted note %q (%s)", note.Title, note.ID), ctx.Caps.ColorProfile))
	}
	return nil
}

func handleSearch(ctx *command.Context, dir string, opts Options) error {
	notes, err := loadNotes(dir)
	if err != nil {
		return fmt.Errorf("load notes: %w", err)
	}

	query := strings.ToLower(opts.Query)
	var matches []*Note
	for _, n := range notes {
		if strings.Contains(strings.ToLower(n.Title), query) ||
			strings.Contains(strings.ToLower(n.Content), query) {
			matches = append(matches, n)
			continue
		}
		for _, tag := range n.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				matches = append(matches, n)
				break
			}
		}
	}

	if ctx.Printer.Mode == output.ModeJSON || opts.JSON {
		return ctx.Printer.PrintJSON(matches)
	}

	if ctx.Printer.Mode == output.ModePlain || opts.Plain {
		RenderNotesPlain(ctx.Stdout, matches)
	} else {
		RenderNotesDashboard(ctx.Stdout, matches, ctx)
	}
	return nil
}
