package touch

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"nova/internal/command"
	"nova/internal/output"
)

// Result records the touch outcome for an individual file target.
type Result struct {
	Path      string    `json:"path"`
	Created   bool      `json:"created"`
	Updated   bool      `json:"updated"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error,omitempty"`
}

// Command returns the registered Command instance for the touch subcommand.
func Command() *command.Command {
	return &command.Command{
		Name:        "touch",
		Aliases:     []string{},
		Summary:     "Create files or update file access and modification timestamps",
		Usage:       "nova touch [flags] <files...>",
		Description: "Create empty files or update timestamps with parent directory auto-creation.",
		Phase:       11,
		Run:         Run,
	}
}

// Run executes the touch command.
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

	if len(opts.Files) == 0 {
		return command.NewUsageError("missing file operand", "Specify one or more file paths: nova touch <files...>")
	}

	targetTime, err := resolveTimestamp(opts)
	if err != nil {
		return command.NewOpError("invalid timestamp", opts.Date, err, "Use format YYYY-MM-DD or YYYY-MM-DD HH:MM:SS or RFC3339.")
	}

	var results []Result
	var anyError bool

	for _, file := range opts.Files {
		res := touchFile(file, targetTime, opts)
		if res.Error != "" {
			anyError = true
		}
		results = append(results, res)
	}

	if ctx.Printer.Mode == output.ModeJSON {
		return RenderJSON(ctx, results)
	}

	if ctx.Printer.Mode == output.ModePlain {
		RenderPlain(ctx, results)
	} else {
		RenderHuman(ctx, results)
	}

	if anyError {
		return errors.New("one or more touch operations failed")
	}

	return nil
}

func resolveTimestamp(opts Options) (time.Time, error) {
	if opts.Reference != "" {
		fi, err := os.Stat(opts.Reference)
		if err != nil {
			return time.Time{}, fmt.Errorf("reference file %q: %w", opts.Reference, err)
		}
		return fi.ModTime(), nil
	}

	if opts.Date != "" {
		formats := []string{
			time.RFC3339,
			"2006-01-02 15:04:05",
			"2006-01-02 15:04",
			"2006-01-02",
		}
		for _, f := range formats {
			if t, err := time.ParseInLocation(f, opts.Date, time.Local); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("unable to parse date %q", opts.Date)
	}

	return time.Now(), nil
}

func touchFile(path string, targetTime time.Time, opts Options) Result {
	fi, err := os.Stat(path)
	exists := err == nil

	if !exists {
		if opts.NoCreate {
			return Result{
				Path:      path,
				Created:   false,
				Updated:   false,
				Timestamp: targetTime,
			}
		}

		if opts.Parents {
			dir := filepath.Dir(path)
			if dir != "." && dir != "" {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return Result{
						Path:  path,
						Error: fmt.Sprintf("cannot create parent directory: %v", err),
					}
				}
			}
		}

		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return Result{
				Path:  path,
				Error: err.Error(),
			}
		}
		_ = f.Close()

		_ = os.Chtimes(path, targetTime, targetTime)
		return Result{
			Path:      path,
			Created:   true,
			Updated:   true,
			Timestamp: targetTime,
		}
	}

	_ = fi
	if err := os.Chtimes(path, targetTime, targetTime); err != nil {
		return Result{
			Path:  path,
			Error: err.Error(),
		}
	}

	return Result{
		Path:      path,
		Created:   false,
		Updated:   true,
		Timestamp: targetTime,
	}
}
