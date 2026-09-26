package cat

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"nova/internal/command"
	"nova/internal/output"
	"nova/internal/terminal"
	"nova/internal/theme"
)

// Command returns the registered Command instance for the cat subcommand.
func Command() *command.Command {
	return &command.Command{
		Name:        "cat",
		Aliases:     []string{"view"},
		Summary:     "Stream, view, and inspect files with syntax highlighting and paging",
		Usage:       "nova cat [flags] [files...]",
		Description: "Stream, view, and inspect files with syntax highlighting, line numbers, and paging.",
		Phase:       4,
		Run:         Run,
	}
}

// Run executes the cat command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)
	if ctx.Printer.Mode == output.ModePlain {
		opts.Plain = true
		opts.NoPager = true
	}

	targets := opts.Files
	if len(targets) == 0 {
		targets = []string{"-"}
	}

	for _, target := range targets {
		if err := viewTarget(ctx, target, opts); err != nil {
			return err
		}
	}

	return nil
}

func viewTarget(ctx *command.Context, target string, opts Options) error {
	var r io.Reader
	var filename string
	var fileSize int64 = -1

	if target == "-" {
		r = ctx.Stdin
		filename = "stdin"
	} else {
		info, err := os.Stat(target)
		if err != nil {
			return command.NewOpError(fmt.Sprintf("cannot access %q", target), target, err, "Check file path spelling and permissions.")
		}
		if info.IsDir() {
			return command.NewOpError(fmt.Sprintf("%q is a directory", target), target, nil, "Use 'nova ls' to inspect directory contents.")
		}
		fileSize = info.Size()
		f, err := os.Open(target)
		if err != nil {
			return command.NewOpError(fmt.Sprintf("cannot open %q", target), target, err, "Check file read permissions.")
		}
		defer f.Close()
		r = f
		filename = target
	}

	// Read initial header buffer for detection (up to 1024 bytes)
	headerBuf := make([]byte, 1024)
	n, readErr := io.ReadFull(r, headerBuf)
	if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
		return command.NewOpError(fmt.Sprintf("failed to read from %q", filename), filename, readErr, "")
	}
	header := headerBuf[:n]

	// Combine header with rest of stream
	combinedReader := io.MultiReader(bytes.NewReader(header), r)

	// Hex dump mode requested explicitly
	if opts.HexDump {
		return streamHex(ctx, combinedReader)
	}

	// Binary file handling
	isBin := IsBinary(header)
	if isBin && !opts.ForceBinary {
		if opts.Plain || ctx.Printer.Mode == output.ModePlain {
			// Stream raw binary bytes in plain mode (pipeline safety)
			_, err := io.Copy(ctx.Stdout, combinedReader)
			return err
		}

		// Human mode warning banner
		warnMsg := fmt.Sprintf("[nova cat: binary file %q", filename)
		if fileSize >= 0 {
			warnMsg += fmt.Sprintf(" (%d bytes)", fileSize)
		}
		warnMsg += ". Use --binary to view raw bytes or --hex for hex dump]"
		ctx.Printer.Println(ctx.Theme.Format(theme.RoleWarning, warnMsg, ctx.Caps.ColorProfile))

		// Show 64-byte preview
		previewLen := 64
		if len(header) < previewLen {
			previewLen = len(header)
		}
		if previewLen > 0 {
			previewRows := FormatHexDump(header[:previewLen], 0)
			for _, row := range previewRows {
				ctx.Printer.Println(ctx.Theme.Format(theme.RoleMuted, "  "+row, ctx.Caps.ColorProfile))
			}
		}
		return nil
	}

	// Detect syntax language
	lang := opts.Language
	if lang == "" {
		lang = DetectLanguage(filename, header)
	}

	// JSON pretty formatting if requested
	if opts.FormatJSON || (lang == "json" && !opts.Plain && ctx.Printer.Mode == output.ModeHuman) {
		allData, err := io.ReadAll(combinedReader)
		if err == nil {
			formatted, jErr := FormatJSONData(allData)
			if jErr == nil {
				combinedReader = bytes.NewReader(formatted)
				lang = "json"
			} else {
				// Fallback to original data
				combinedReader = bytes.NewReader(allData)
			}
		}
	}

	profile := ctx.Caps.ColorProfile
	if opts.Plain || ctx.Printer.Mode == output.ModePlain {
		profile = terminal.ColorNone
	}

	proc := NewLineProcessor(opts, ctx.Theme, profile, lang)

	// Paging decision
	shouldPage := ctx.Printer.Mode == output.ModeHuman &&
		!opts.NoPager &&
		opts.Paging != "never" &&
		ctx.Caps.Height > 0

	if shouldPage {
		// Read lines up to terminal height to check if paging is needed
		var lines []string
		streamReader := bufio.NewReader(combinedReader)
		for {
			rawLine, err := streamReader.ReadBytes('\n')
			if len(rawLine) > 0 {
				formatted, emit := proc.ProcessLine(rawLine)
				if emit {
					lines = append(lines, formatted)
				}
			}
			if err != nil {
				break
			}
			// If we know opts.Paging is auto and lines already exceed terminal height, we must page
			if opts.Paging == "auto" && len(lines) >= ctx.Caps.Height {
				// Read remaining lines
				for {
					rawRem, remErr := streamReader.ReadBytes('\n')
					if len(rawRem) > 0 {
						remFmt, remEmit := proc.ProcessLine(rawRem)
						if remEmit {
							lines = append(lines, remFmt)
						}
					}
					if remErr != nil {
						break
					}
				}
				pager := NewPager(lines, filename, ctx.Caps.Width, ctx.Caps.Height, ctx.Theme, profile)
				return pager.Run()
			}
		}

		if opts.Paging == "always" && len(lines) > 0 {
			pager := NewPager(lines, filename, ctx.Caps.Width, ctx.Caps.Height, ctx.Theme, profile)
			return pager.Run()
		}

		// Content fits inside terminal height: print directly!
		for _, line := range lines {
			ctx.Printer.Println(line)
		}
		return nil
	}

	// Non-paging streaming directly to writer
	return StreamDirect(combinedReader, ctx.Stdout, proc)
}

func streamHex(ctx *command.Context, r io.Reader) error {
	buf := make([]byte, 16)
	var offset int64 = 0

	for {
		n, err := io.ReadFull(r, buf)
		if n > 0 {
			rows := FormatHexDump(buf[:n], offset)
			for _, row := range rows {
				ctx.Printer.Println(row)
			}
			offset += int64(n)
		}
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return err
		}
	}
	return nil
}
