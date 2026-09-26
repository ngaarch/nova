package cat

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// IsBinary determines whether the byte slice represents binary content rather than text.
// It checks the first 1024 bytes for null bytes or a high proportion of control codes.
func IsBinary(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	limit := len(data)
	if limit > 1024 {
		limit = 1024
	}
	sample := data[:limit]

	// Immediate indicator: null byte
	if bytes.IndexByte(sample, 0) != -1 {
		return true
	}

	// Check for non-printable control characters
	nonPrintable := 0
	for _, b := range sample {
		if b < 0x20 && b != '\t' && b != '\n' && b != '\r' && b != '\f' && b != '\b' {
			nonPrintable++
		}
	}

	// If more than 30% are control chars, treat as binary
	if float64(nonPrintable)/float64(limit) > 0.30 {
		return true
	}

	// Validate UTF-8 encoding
	if !utf8.Valid(sample) {
		// If it's invalid UTF-8 and contains non-trivial control characters, treat as binary
		if nonPrintable > 0 {
			return true
		}
	}

	return false
}

// DetectLanguage resolves the programming language / format for syntax highlighting.
func DetectLanguage(filename string, header []byte) string {
	base := filepath.Base(filename)
	ext := strings.ToLower(filepath.Ext(filename))

	// Exact filenames
	switch strings.ToLower(base) {
	case "dockerfile", "containerfile":
		return "dockerfile"
	case "makefile", "gnumakefile":
		return "makefile"
	case "gemfile", "rakefile":
		return "ruby"
	case "go.mod", "go.sum":
		return "go"
	}

	// Extensions
	switch ext {
	case ".go":
		return "go"
	case ".py", ".pyw":
		return "python"
	case ".js", ".mjs", ".cjs":
		return "javascript"
	case ".ts", ".mts", ".cts":
		return "typescript"
	case ".jsx", ".tsx":
		return "typescript"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	case ".md", ".markdown":
		return "markdown"
	case ".sh", ".bash", ".zsh":
		return "bash"
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".cxx", ".hpp", ".hxx":
		return "cpp"
	case ".rs":
		return "rust"
	case ".html", ".htm":
		return "html"
	case ".xml", ".svg":
		return "xml"
	case ".sql":
		return "sql"
	case ".css", ".scss", ".sass":
		return "css"
	case ".diff", ".patch":
		return "diff"
	}

	// Inspect header for shebang or markers
	if len(header) > 0 {
		lines := bytes.SplitN(header, []byte{'\n'}, 2)
		firstLine := strings.TrimSpace(string(lines[0]))

		if strings.HasPrefix(firstLine, "#!") {
			lower := strings.ToLower(firstLine)
			switch {
			case strings.Contains(lower, "python"):
				return "python"
			case strings.Contains(lower, "bash"), strings.Contains(lower, "sh"), strings.Contains(lower, "zsh"):
				return "bash"
			case strings.Contains(lower, "node"):
				return "javascript"
			case strings.Contains(lower, "ruby"):
				return "ruby"
			case strings.Contains(lower, "perl"):
				return "perl"
			}
		}

		if strings.HasPrefix(firstLine, "<?xml") {
			return "xml"
		}
		if strings.HasPrefix(strings.ToLower(firstLine), "<!doctype html") || strings.HasPrefix(firstLine, "<html") {
			return "html"
		}
	}

	return "text"
}

// FormatHexDump formats raw bytes into canonical hex dump rows:
// 00000000  7f 45 4c 46 02 01 01 00  00 00 00 00 00 00 00 00  |.ELF............|
func FormatHexDump(data []byte, startOffset int64) []string {
	var rows []string
	n := len(data)

	for i := 0; i < n; i += 16 {
		end := i + 16
		if end > n {
			end = n
		}
		chunk := data[i:end]

		// Address offset (8 hex digits)
		addr := fmt.Sprintf("%08x", startOffset+int64(i))

		// Hex bytes (16 bytes, split 8 and 8)
		var hexPart strings.Builder
		for j := 0; j < 16; j++ {
			if j == 8 {
				hexPart.WriteString(" ")
			}
			if j < len(chunk) {
				hexPart.WriteString(fmt.Sprintf("%02x ", chunk[j]))
			} else {
				hexPart.WriteString("   ")
			}
		}

		// ASCII representation
		var asciiPart strings.Builder
		for _, b := range chunk {
			if b >= 0x20 && b <= 0x7e {
				asciiPart.WriteByte(b)
			} else {
				asciiPart.WriteByte('.')
			}
		}

		rows = append(rows, fmt.Sprintf("%s  %s |%s|", addr, strings.TrimRight(hexPart.String(), " "), asciiPart.String()))
	}

	return rows
}
