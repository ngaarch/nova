package mdcmd

import (
	"bytes"
	"encoding/json"
	"os"
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
	opts := ParseFlags([]string{"README.md", "-w", "100", "--toc", "--pager"})
	if opts.FilePath != "README.md" {
		t.Errorf("expected FilePath README.md, got %q", opts.FilePath)
	}
	if opts.Width != 100 {
		t.Errorf("expected Width 100, got %d", opts.Width)
	}
	if !opts.TOC {
		t.Errorf("expected TOC true")
	}
	if !opts.Pager {
		t.Errorf("expected Pager true")
	}
}

func TestParseMarkdown_HeadingsAndLists(t *testing.T) {
	sample := `# Document Title
Some intro paragraph with **bold** text and ` + "`code`" + `.

## Section One
- [ ] Task 1
- [x] Task 2 completed
* Regular bullet
1. Numbered item 1
2. Numbered item 2

> Important blockquote note

---
`

	th := theme.Get("default")
	doc := ParseMarkdown(sample, 80, th, terminal.ColorNone)

	if doc.Title != "Document Title" {
		t.Errorf("expected title 'Document Title', got %q", doc.Title)
	}
	if len(doc.Headings) != 2 {
		t.Fatalf("expected 2 headings, got %d", len(doc.Headings))
	}
	if doc.Headings[0].Text != "Document Title" || doc.Headings[0].Level != 1 {
		t.Errorf("unexpected heading 0: %+v", doc.Headings[0])
	}
	if doc.Headings[1].Text != "Section One" || doc.Headings[1].Level != 2 {
		t.Errorf("unexpected heading 1: %+v", doc.Headings[1])
	}

	joined := strings.Join(doc.Plain, "\n")
	if !strings.Contains(joined, "[ ] Task 1") {
		t.Errorf("expected uncompleted task in plain text, got:\n%s", joined)
	}
	if !strings.Contains(joined, "[x] Task 2 completed") {
		t.Errorf("expected completed task in plain text, got:\n%s", joined)
	}
	if !strings.Contains(joined, "| Important blockquote note") {
		t.Errorf("expected blockquote in plain text, got:\n%s", joined)
	}
}

func TestParseMarkdown_TableAndCodeBlock(t *testing.T) {
	sample := `### Code & Tables

` + "```go\npackage main\n\nfunc main() {}\n```" + `

| Name | Role | Status |
|---|---|---|
| Alice | Admin | Active |
| Bob | User | Pending |
`

	th := theme.Get("default")
	doc := ParseMarkdown(sample, 80, th, terminal.ColorNone)

	joined := strings.Join(doc.Plain, "\n")
	if !strings.Contains(joined, "package main") {
		t.Errorf("expected code content in output, got:\n%s", joined)
	}
	if !strings.Contains(joined, "Alice") || !strings.Contains(joined, "Bob") {
		t.Errorf("expected table cells in output, got:\n%s", joined)
	}
}

func TestRenderTOC(t *testing.T) {
	doc := &Document{
		Title: "Project Guide",
		Headings: []HeadingInfo{
			{Level: 1, Text: "Introduction", Line: 1},
			{Level: 2, Text: "Getting Started", Line: 10},
			{Level: 3, Text: "Prerequisites", Line: 15},
		},
		WordCount: 500,
		LineCount: 40,
	}

	th := theme.Get("default")
	lines := RenderTOC(doc, 80, th, terminal.ColorNone)
	joined := strings.Join(lines, "\n")

	if !strings.Contains(joined, "Table of Contents: Project Guide") {
		t.Errorf("expected TOC header, got:\n%s", joined)
	}
	if !strings.Contains(joined, "Introduction") || !strings.Contains(joined, "Prerequisites") {
		t.Errorf("expected headings listed in TOC, got:\n%s", joined)
	}
}

func TestRunCommand_FileAndStdin(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "sample.md")
	content := "# Heading\n\nThis is sample markdown.\n"
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// 1. Read from file
	var stdout, stderr bytes.Buffer
	ctx := newTestContext("", &stdout, &stderr)
	err := Run(ctx, []string{filePath, "--plain"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "HEADING") {
		t.Errorf("expected 'HEADING' in output, got: %s", stdout.String())
	}

	// 2. Read from stdin
	stdout.Reset()
	stderr.Reset()
	ctxStdin := newTestContext("## Stdin Header\n\nInline text.\n", &stdout, &stderr)
	err = Run(ctxStdin, []string{"--plain"})
	if err != nil {
		t.Fatalf("unexpected error reading stdin: %v", err)
	}
	if !strings.Contains(stdout.String(), "Stdin Header") {
		t.Errorf("expected 'Stdin Header' in output, got: %s", stdout.String())
	}

	// 3. JSON Output
	stdout.Reset()
	stderr.Reset()
	ctxJSON := newTestContext(content, &stdout, &stderr)
	err = Run(ctxJSON, []string{"--json"})
	if err != nil {
		t.Fatalf("unexpected error for JSON mode: %v", err)
	}
	var resDoc Document
	if err := json.Unmarshal(stdout.Bytes(), &resDoc); err != nil {
		t.Fatalf("failed to decode JSON output: %v; raw: %s", err, stdout.String())
	}
	if resDoc.Title != "Heading" {
		t.Errorf("expected Title 'Heading', got %q", resDoc.Title)
	}
}
