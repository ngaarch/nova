package notecmd

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"nova/internal/command"
	"nova/internal/config"
	"nova/internal/logging"
	"nova/internal/output"
	"nova/internal/terminal"
	"nova/internal/theme"
)

func newTestContext(stdin string, stdout, stderr *bytes.Buffer) *command.Context {
	cfg := config.Config{
		Theme:     "default",
		ColorMode: "never",
	}
	caps := terminal.Capabilities{
		Width:        80,
		Height:       24,
		ColorProfile: terminal.ColorNone,
		IsTTY:        false,
	}
	th := theme.Get("default")
	logger := logging.New(stderr, false)
	return command.NewContext(
		strings.NewReader(stdin),
		stdout,
		stderr,
		cfg,
		caps,
		th,
		output.ModePlain,
		logger,
	)
}

func TestParseFlags(t *testing.T) {
	opts := ParseFlags([]string{"add", "Sprint Tasks", "-t", "dev,go", "-c", "- [ ] Task 1"})
	if opts.Action != ActionAdd {
		t.Errorf("expected ActionAdd, got %v", opts.Action)
	}
	if opts.Title != "Sprint Tasks" {
		t.Errorf("expected Title 'Sprint Tasks', got %q", opts.Title)
	}
	if len(opts.Tags) != 2 || opts.Tags[0] != "dev" || opts.Tags[1] != "go" {
		t.Errorf("unexpected tags: %+v", opts.Tags)
	}
	if opts.Content != "- [ ] Task 1" {
		t.Errorf("unexpected content: %q", opts.Content)
	}

	optsToggle := ParseFlags([]string{"toggle", "note-123", "2"})
	if optsToggle.Action != ActionToggle || optsToggle.ID != "note-123" || optsToggle.ItemIdx != 2 {
		t.Errorf("unexpected toggle opts: %+v", optsToggle)
	}
}

func TestNoteLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	notesDir := filepath.Join(tmpDir, "notes")

	var stdout, stderr bytes.Buffer
	ctx := newTestContext("", &stdout, &stderr)

	// 1. Add Note with tasks
	content := "# Goals\n- [ ] Write unit tests\n- [x] Initial design\n- [ ] Documentation\n"
	err := Run(ctx, []string{"add", "Sprint 1", "-t", "sprint,core", "-c", content, "--dir=" + notesDir, "--plain"})
	if err != nil {
		t.Fatalf("failed to add note: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "Sprint 1") {
		t.Errorf("expected note title in output, got: %s", out)
	}
	fields := strings.Fields(out)
	if len(fields) == 0 {
		t.Fatalf("no output from note add")
	}
	noteID := fields[0]

	// 2. List Notes (Dashboard mode)
	stdout.Reset()
	stderr.Reset()
	ctx.Printer.Mode = output.ModeHuman
	err = Run(ctx, []string{"list", "--dir=" + notesDir})
	if err != nil {
		t.Fatalf("failed to list notes: %v", err)
	}
	if !strings.Contains(stdout.String(), "Sprint 1") || !strings.Contains(stdout.String(), "#sprint") {
		t.Errorf("expected Sprint 1 and #sprint in dashboard, got: %s", stdout.String())
	}

	// 3. Show Note (JSON mode)
	stdout.Reset()
	stderr.Reset()
	ctx.Printer.Mode = output.ModeJSON
	err = Run(ctx, []string{"show", noteID, "--dir=" + notesDir})
	if err != nil {
		t.Fatalf("failed to show note: %v", err)
	}
	var loaded Note
	if err := json.Unmarshal(stdout.Bytes(), &loaded); err != nil {
		t.Fatalf("failed to parse JSON note: %v; raw: %s", err, stdout.String())
	}
	if loaded.Title != "Sprint 1" || loaded.TotalTodos != 3 || loaded.DoneCount != 1 {
		t.Errorf("unexpected note loaded: %+v", loaded)
	}

	// 4. Todo Dashboard
	stdout.Reset()
	stderr.Reset()
	ctx.Printer.Mode = output.ModePlain
	err = Run(ctx, []string{"todo", "--dir=" + notesDir, "--plain"})
	if err != nil {
		t.Fatalf("failed to run todo action: %v", err)
	}
	todoOut := stdout.String()
	if !strings.Contains(todoOut, "Write unit tests") || !strings.Contains(todoOut, "[x]") {
		t.Errorf("expected todo items in output, got: %s", todoOut)
	}

	// 5. Toggle Todo
	stdout.Reset()
	stderr.Reset()
	err = Run(ctx, []string{"toggle", noteID, "1", "--dir=" + notesDir, "--plain"})
	if err != nil {
		t.Fatalf("failed to toggle todo: %v", err)
	}

	// Verify toggled
	n, err := findNote(notesDir, noteID)
	if err != nil {
		t.Fatal(err)
	}
	if n.DoneCount != 2 {
		t.Errorf("expected DoneCount 2 after toggle, got %d", n.DoneCount)
	}

	// 6. Search
	stdout.Reset()
	stderr.Reset()
	err = Run(ctx, []string{"search", "unit tests", "--dir=" + notesDir, "--plain"})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "Sprint 1") {
		t.Errorf("expected Sprint 1 in search results, got: %s", stdout.String())
	}

	// 7. Delete Note
	stdout.Reset()
	stderr.Reset()
	err = Run(ctx, []string{"rm", noteID, "--dir=" + notesDir, "--plain"})
	if err != nil {
		t.Fatalf("delete note failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "Deleted") {
		t.Errorf("expected Deleted in output, got: %s", stdout.String())
	}

	// Verify deleted
	_, err = findNote(notesDir, noteID)
	if err == nil {
		t.Errorf("expected error finding deleted note")
	}
}
