package completion

import (
	"bytes"
	"strings"
	"testing"

	"nova/internal/command"
	"nova/internal/config"
	"nova/internal/logging"
	"nova/internal/output"
	"nova/internal/terminal"
	"nova/internal/theme"
)

func newTestContext() (*command.Context, *bytes.Buffer, *bytes.Buffer) {
	var in, stdout, stderr bytes.Buffer
	cfg := config.Default()
	caps := terminal.Capabilities{
		IsTTY:        false,
		ColorProfile: terminal.ColorNone,
	}
	th := theme.Get("default")
	logger := logging.New(&stderr, false)
	ctx := command.NewContext(&in, &stdout, &stderr, cfg, caps, th, output.ModePlain, logger)
	return ctx, &stdout, &stderr
}

func TestBashCompletion(t *testing.T) {
	ctx, stdout, _ := newTestContext()
	err := Run(ctx, []string{"bash"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "_nova_completions") || !strings.Contains(out, "complete -o default -F _nova_completions nova") {
		t.Errorf("expected bash completion script, got: %s", out)
	}
	if !strings.Contains(out, "cyberpunk") || !strings.Contains(out, "diff") {
		t.Errorf("expected new themes and diff command in completions")
	}
}

func TestZshCompletion(t *testing.T) {
	ctx, stdout, _ := newTestContext()
	err := Run(ctx, []string{"zsh"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "#compdef nova") || !strings.Contains(out, "_nova()") {
		t.Errorf("expected zsh completion script, got: %s", out)
	}
}

func TestFishCompletion(t *testing.T) {
	ctx, stdout, _ := newTestContext()
	err := Run(ctx, []string{"fish"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "complete -c nova") {
		t.Errorf("expected fish completion script, got: %s", out)
	}
}

func TestUnsupportedShell(t *testing.T) {
	ctx, _, _ := newTestContext()
	err := Run(ctx, []string{"powershell"})
	if err == nil {
		t.Fatalf("expected error for unsupported shell")
	}
}
