package watch

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nova/internal/command"
	"nova/internal/config"
	"nova/internal/logging"
	"nova/internal/output"
	"nova/internal/terminal"
	"nova/internal/theme"
)

func newTestContext(mode output.Mode) (*command.Context, *bytes.Buffer, *bytes.Buffer) {
	var in, stdout, stderr bytes.Buffer
	cfg := config.Default()
	caps := terminal.Capabilities{
		IsTTY:            mode == output.ModeHuman,
		Width:            80,
		Height:           24,
		ColorProfile:     terminal.ColorTrueColor,
		UnicodeSupported: true,
	}
	th := theme.Get("default")
	logger := logging.New(&stderr, false)
	ctx := command.NewContext(&in, &stdout, &stderr, cfg, caps, th, mode, logger)
	return ctx, &stdout, &stderr
}

func TestDetectDiff(t *testing.T) {
	now := time.Now()
	oldState := map[string]FileFingerprint{
		"file1.txt": {ModTime: now.Add(-time.Hour), Size: 100},
		"file2.txt": {ModTime: now.Add(-time.Hour), Size: 200},
	}

	newState := map[string]FileFingerprint{
		"file1.txt": {ModTime: now, Size: 150}, // modified
		"file3.txt": {ModTime: now, Size: 50},  // created
		// file2.txt is deleted
	}

	events := detectDiff(oldState, newState)
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	foundModify := false
	foundCreate := false
	foundDelete := false

	for _, ev := range events {
		if ev.Path == "file1.txt" && ev.Type == EventModify {
			foundModify = true
		}
		if ev.Path == "file3.txt" && ev.Type == EventCreate {
			foundCreate = true
		}
		if ev.Path == "file2.txt" && ev.Type == EventDelete {
			foundDelete = true
		}
	}

	if !foundModify || !foundCreate || !foundDelete {
		t.Errorf("expected modify, create, and delete events; got: %+v", events)
	}
}

func TestWatchLoopWithChanges(t *testing.T) {
	tmpDir := t.TempDir()
	trackedFile := filepath.Join(tmpDir, "tracked.txt")
	_ = os.WriteFile(trackedFile, []byte("initial"), 0644)

	ctx, stdout, _ := newTestContext(output.ModePlain)

	opts := Options{
		Paths:      []string{tmpDir},
		Interval:   50 * time.Millisecond,
		Debounce:   30 * time.Millisecond,
		Iterations: 1, // stop after 1 trigger
	}

	sigCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Trigger change after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = os.WriteFile(trackedFile, []byte("updated content!"), 0644)
	}()

	err := WatchLoop(sigCtx, ctx, opts)
	if err != nil {
		t.Fatalf("watch loop failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "tracked.txt") {
		t.Errorf("expected tracked.txt in watch events, got: %s", out)
	}
}

func TestRenderJSONEvents(t *testing.T) {
	ctx, stdout, _ := newTestContext(output.ModeJSON)
	events := []FileEvent{
		{Type: EventCreate, Path: "test.go", Timestamp: time.Now()},
	}

	RenderJSONEvents(ctx, events)
	if !strings.Contains(stdout.String(), `"type":"CREATE"`) {
		t.Errorf("expected JSON event with CREATE, got: %s", stdout.String())
	}
}
