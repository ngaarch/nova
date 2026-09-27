package diff

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"nova/internal/command"
	"nova/internal/output"
)

// Command returns the registered Command instance for the diff subcommand.
func Command() *command.Command {
	return &command.Command{
		Name:        "diff",
		Aliases:     []string{"compare"},
		Summary:     "Compare files with syntax-aware colorized unified diff and stats",
		Usage:       "nova diff [flags] <file1> <file2>",
		Description: "Compare two files with side-by-side or unified colored diff and change summary.",
		Phase:       11,
		Run:         Run,
	}
}

// Run executes the diff command.
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

	if opts.File1 == "" || opts.File2 == "" {
		return command.NewUsageError("missing file operands", "Usage: nova diff [flags] <file1> <file2>")
	}

	fi1, err1 := os.Stat(opts.File1)
	if err1 != nil {
		return command.NewOpError(fmt.Sprintf("cannot access %q", opts.File1), opts.File1, err1, "Check file path spelling and permissions.")
	}
	if fi1.IsDir() {
		return command.NewOpError(fmt.Sprintf("%q is a directory", opts.File1), opts.File1, nil, "Directories are not supported in basic file diff.")
	}

	fi2, err2 := os.Stat(opts.File2)
	if err2 != nil {
		return command.NewOpError(fmt.Sprintf("cannot access %q", opts.File2), opts.File2, err2, "Check file path spelling and permissions.")
	}
	if fi2.IsDir() {
		return command.NewOpError(fmt.Sprintf("%q is a directory", opts.File2), opts.File2, nil, "Directories are not supported in basic file diff.")
	}

	data1, err := os.ReadFile(opts.File1)
	if err != nil {
		return command.NewOpError(fmt.Sprintf("cannot read %q", opts.File1), opts.File1, err, "")
	}
	data2, err := os.ReadFile(opts.File2)
	if err != nil {
		return command.NewOpError(fmt.Sprintf("cannot read %q", opts.File2), opts.File2, err, "")
	}

	isBin1 := isBinaryData(data1)
	isBin2 := isBinaryData(data2)

	if isBin1 || isBin2 {
		identical := bytes.Equal(data1, data2)
		res := DiffResult{
			File1:     opts.File1,
			File2:     opts.File2,
			Time1:     fi1.ModTime(),
			Time2:     fi2.ModTime(),
			Identical: identical,
			IsBinary:  true,
		}
		if ctx.Printer.Mode == output.ModeJSON {
			return RenderJSON(ctx, res)
		}
		if ctx.Printer.Mode == output.ModePlain {
			RenderPlain(ctx, res, opts)
		} else {
			RenderHuman(ctx, res, opts)
		}
		return nil
	}

	lines1 := splitLines(string(data1))
	lines2 := splitLines(string(data2))

	res := ComputeDiff(lines1, lines2, opts)
	res.File1 = opts.File1
	res.File2 = opts.File2
	res.Time1 = fi1.ModTime()
	res.Time2 = fi2.ModTime()

	if ctx.Printer.Mode == output.ModeJSON {
		return RenderJSON(ctx, res)
	}

	if ctx.Printer.Mode == output.ModePlain {
		RenderPlain(ctx, res, opts)
	} else {
		RenderHuman(ctx, res, opts)
	}

	return nil
}

func isBinaryData(data []byte) bool {
	sample := data
	if len(sample) > 1024 {
		sample = sample[:1024]
	}
	for _, b := range sample {
		if b == 0 {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	// Normalize CRLF to LF
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimSuffix(s, "\n")
	return strings.Split(s, "\n")
}
