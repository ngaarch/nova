package hex

import (
	"fmt"
	"io"
	"os"

	"nova/internal/command"
	"nova/internal/output"
)

// Row represents one formatted row of hex data.
type Row struct {
	Offset int64  `json:"offset"`
	Hex    string `json:"hex"`
	ASCII  string `json:"ascii"`
	Raw    []byte `json:"-"`
}

// DumpResult holds the dump rows and metadata for a target.
type DumpResult struct {
	Target     string `json:"target"`
	TotalBytes int64  `json:"total_bytes"`
	Start      int64  `json:"start_offset"`
	End        int64  `json:"end_offset"`
	Rows       []Row  `json:"rows"`
	Error      string `json:"error,omitempty"`
}

// Command returns the registered Command instance for hex.
func Command() *command.Command {
	return &command.Command{
		Name:        "hex",
		Aliases:     []string{"hexdump", "dump", "xxd"},
		Summary:     "Display modern, colorized hex dump inspection with byte categorization and ASCII preview",
		Usage:       "nova hex [flags] [files...]",
		Description: "Inspect binary files, code objects, and byte streams with color-coded offsets, grouping, and ASCII panel.",
		Phase:       17,
		Run:         Run,
	}
}

// Run executes the hex command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)
	if opts.Columns <= 0 {
		opts.Columns = 16
	}
	if opts.Grouping <= 0 {
		opts.Grouping = 1
	}

	targets := opts.Paths
	if len(targets) == 0 {
		targets = []string{"-"}
	}

	var results []DumpResult
	for _, target := range targets {
		res, err := dumpTarget(ctx, target, opts)
		if err != nil {
			res.Error = err.Error()
		}
		results = append(results, res)
	}

	return Render(ctx, results, opts)
}

func dumpTarget(ctx *command.Context, target string, opts Options) (DumpResult, error) {
	var r io.Reader
	var totalSize int64

	res := DumpResult{
		Target: target,
		Start:  opts.Skip,
	}

	if target == "-" {
		res.Target = "<stdin>"
		r = ctx.Stdin
		// For stdin stream, discard skip bytes
		if opts.Skip > 0 {
			_, err := io.CopyN(io.Discard, r, opts.Skip)
			if err != nil && err != io.EOF {
				return res, err
			}
		}
	} else {
		f, err := os.Open(target)
		if err != nil {
			return res, err
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			return res, err
		}
		if info.IsDir() {
			return res, fmt.Errorf("%s: is a directory", target)
		}
		totalSize = info.Size()

		if opts.Skip > 0 {
			if _, err := f.Seek(opts.Skip, io.SeekStart); err != nil {
				return res, err
			}
		}
		r = f
	}

	currOffset := opts.Skip
	buf := make([]byte, opts.Columns)
	var bytesRemaining int64 = -1
	if opts.Length >= 0 {
		bytesRemaining = opts.Length
	}

	for {
		readSize := opts.Columns
		if bytesRemaining >= 0 && bytesRemaining < int64(readSize) {
			readSize = int(bytesRemaining)
			if readSize == 0 {
				break
			}
		}

		n, err := io.ReadFull(r, buf[:readSize])
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])

			row := formatRow(currOffset, chunk, opts)
			res.Rows = append(res.Rows, row)
			currOffset += int64(n)
			if bytesRemaining >= 0 {
				bytesRemaining -= int64(n)
			}
		}

		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return res, err
		}
	}

	res.End = currOffset
	if totalSize > 0 {
		res.TotalBytes = totalSize
	} else {
		res.TotalBytes = currOffset - opts.Skip
	}

	return res, nil
}

func formatRow(offset int64, chunk []byte, opts Options) Row {
	// Build plain hex string and ASCII representation
	var hexParts []string
	var asciiRunes []rune

	for i, b := range chunk {
		hexParts = append(hexParts, fmt.Sprintf("%02x", b))
		if opts.Grouping > 1 && (i+1)%opts.Grouping == 0 && i+1 < len(chunk) {
			hexParts = append(hexParts, "") // extra separator
		}

		if b >= 32 && b <= 126 {
			asciiRunes = append(asciiRunes, rune(b))
		} else {
			asciiRunes = append(asciiRunes, '.')
		}
	}

	return Row{
		Offset: offset,
		Hex:    fmt.Sprintf("% x", chunk),
		ASCII:  string(asciiRunes),
		Raw:    chunk,
	}
}
