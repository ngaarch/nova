package watch

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"nova/internal/command"
	"nova/internal/output"
)

// EventType categorizes filesystem mutations.
type EventType string

const (
	EventCreate EventType = "CREATE"
	EventModify EventType = "MODIFY"
	EventDelete EventType = "DELETE"
)

// FileEvent represents a detected change.
type FileEvent struct {
	Type      EventType `json:"type"`
	Path      string    `json:"path"`
	Timestamp time.Time `json:"timestamp"`
}

// FileFingerprint caches file modification time and size.
type FileFingerprint struct {
	ModTime time.Time
	Size    int64
}

// Command returns the registered Command instance for the watch subcommand.
func Command() *command.Command {
	return &command.Command{
		Name:        "watch",
		Aliases:     []string{"monitor"},
		Summary:     "Watch files and directories for changes and execute commands automatically",
		Usage:       "nova watch [flags] [paths...] [-e <command>]",
		Description: "Live filesystem monitor with debounced change detection and task execution.",
		Phase:       16,
		Run:         Run,
	}
}

// Run executes the watch loop.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)
	if opts.Plain {
		ctx.Printer.Mode = output.ModePlain
	} else if opts.JSON {
		ctx.Printer.Mode = output.ModeJSON
	}

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return WatchLoop(sigCtx, ctx, opts)
}

// WatchLoop runs the main polling and change event dispatch loop.
func WatchLoop(sigCtx context.Context, ctx *command.Context, opts Options) error {
	state := make(map[string]FileFingerprint)

	// Initial scan to establish baseline
	scanFiles(opts.Paths, opts.Extensions, state)

	if ctx.Printer.Mode == output.ModeHuman {
		RenderWatchHeader(ctx, opts)
	}

	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	triggers := 0
	var pendingEvents []FileEvent
	var debounceTimer *time.Timer

	triggerAction := func() {
		if len(pendingEvents) == 0 {
			return
		}
		triggers++

		if ctx.Printer.Mode == output.ModeJSON {
			RenderJSONEvents(ctx, pendingEvents)
		} else if ctx.Printer.Mode == output.ModePlain {
			RenderPlainEvents(ctx, pendingEvents)
		} else {
			RenderHumanEvents(ctx, pendingEvents)
		}

		if opts.ExecCommand != "" {
			if opts.ClearScreen && ctx.Printer.Mode == output.ModeHuman {
				ctx.Printer.Print("\x1b[H\x1b[2J")
			}
			runCommand(ctx, opts.ExecCommand)
		}

		pendingEvents = nil
	}

	for {
		select {
		case <-sigCtx.Done():
			if ctx.Printer.Mode == output.ModeHuman {
				ctx.Printer.Println("\n  Stopped watching.")
			}
			return nil

		case <-ticker.C:
			current := make(map[string]FileFingerprint)
			scanFiles(opts.Paths, opts.Extensions, current)

			events := detectDiff(state, current)
			state = current

			if len(events) > 0 {
				pendingEvents = append(pendingEvents, events...)

				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(opts.Debounce, triggerAction)

				if opts.Iterations > 0 && triggers+1 >= opts.Iterations {
					time.Sleep(opts.Debounce + 50*time.Millisecond)
					return nil
				}
			}
		}
	}
}

func scanFiles(roots []string, exts []string, out map[string]FileFingerprint) {
	for _, root := range roots {
		info, err := os.Stat(root)
		if err != nil {
			continue
		}

		if !info.IsDir() {
			if matchesExt(root, exts) {
				out[root] = FileFingerprint{
					ModTime: info.ModTime(),
					Size:    info.Size(),
				}
			}
			continue
		}

		_ = filepath.Walk(root, func(path string, fi os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return nil
			}
			if fi.IsDir() {
				base := filepath.Base(path)
				if base == ".git" || base == "node_modules" || base == "vendor" {
					return filepath.SkipDir
				}
				return nil
			}

			if matchesExt(path, exts) {
				out[path] = FileFingerprint{
					ModTime: fi.ModTime(),
					Size:    fi.Size(),
				}
			}
			return nil
		})
	}
}

func matchesExt(path string, exts []string) bool {
	if len(exts) == 0 {
		return true
	}
	fileExt := strings.ToLower(filepath.Ext(path))
	for _, ext := range exts {
		if fileExt == ext {
			return true
		}
	}
	return false
}

func detectDiff(oldState, newState map[string]FileFingerprint) []FileEvent {
	var events []FileEvent
	now := time.Now()

	// Check additions and modifications
	for path, newFp := range newState {
		oldFp, exists := oldState[path]
		if !exists {
			events = append(events, FileEvent{
				Type:      EventCreate,
				Path:      path,
				Timestamp: now,
			})
		} else if !oldFp.ModTime.Equal(newFp.ModTime) || oldFp.Size != newFp.Size {
			events = append(events, FileEvent{
				Type:      EventModify,
				Path:      path,
				Timestamp: now,
			})
		}
	}

	// Check deletions
	for path := range oldState {
		if _, exists := newState[path]; !exists {
			events = append(events, FileEvent{
				Type:      EventDelete,
				Path:      path,
				Timestamp: now,
			})
		}
	}

	return events
}

func runCommand(ctx *command.Context, cmdStr string) {
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	_ = cmd.Run()
}
