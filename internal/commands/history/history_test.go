package historycmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nova/internal/command"
	"nova/internal/config"
	"nova/internal/output"
	"nova/internal/terminal"
	"nova/internal/theme"
)

func TestFlagsParse(t *testing.T) {
	opts := ParseFlags([]string{"-n", "5", "--filter=git", "--file=/tmp/test_hist", "--plain"})
	if opts.Count != 5 {
		t.Errorf("expected Count 5, got %d", opts.Count)
	}
	if opts.Filter != "git" {
		t.Errorf("expected Filter git, got %s", opts.Filter)
	}
	if opts.File != "/tmp/test_hist" {
		t.Errorf("expected File /tmp/test_hist, got %s", opts.File)
	}
	if !opts.Plain {
		t.Errorf("expected Plain true")
	}
}

func TestParseHistoryFormats(t *testing.T) {
	// 1. Bash / plain format
	bashData := []byte("git status\nls -la\ngit commit\n")
	cmdsBash := ParseHistory(bashData)
	if len(cmdsBash) != 3 {
		t.Errorf("expected 3 commands from bash, got %d", len(cmdsBash))
	}

	// 2. Zsh extended format
	zshData := []byte(": 1680000000:0;git status\n: 1680000001:0;go test ./...\n")
	cmdsZsh := ParseHistory(zshData)
	if len(cmdsZsh) != 2 || cmdsZsh[0] != "git status" || cmdsZsh[1] != "go test ./..." {
		t.Errorf("unexpected zsh parsed commands: %v", cmdsZsh)
	}

	// 3. Fish format
	fishData := []byte("- cmd: git log\n  when: 1680000\n- cmd: nova ls\n  when: 1680001\n")
	cmdsFish := ParseHistory(fishData)
	if len(cmdsFish) != 2 || cmdsFish[0] != "git log" || cmdsFish[1] != "nova ls" {
		t.Errorf("unexpected fish parsed commands: %v", cmdsFish)
	}
}

func TestAnalyzeHistory(t *testing.T) {
	cmds := []string{
		"git status",
		"git commit -m 'test'",
		"git push",
		"go test ./...",
		"go build",
		"ls -la",
	}

	rep := AnalyzeHistory(cmds, 5, "", "mock_hist")
	if rep.TotalCommands != 6 {
		t.Errorf("expected TotalCommands 6, got %d", rep.TotalCommands)
	}
	if len(rep.TopCommands) == 0 {
		t.Fatalf("expected top commands")
	}
	if rep.TopCommands[0].Command != "git" || rep.TopCommands[0].Count != 3 {
		t.Errorf("expected #1 git with 3 counts, got %s with %d", rep.TopCommands[0].Command, rep.TopCommands[0].Count)
	}
	if rep.TopCommands[0].Category != "Git" {
		t.Errorf("expected Git category, got %s", rep.TopCommands[0].Category)
	}
}

func TestRunCommand(t *testing.T) {
	tmpDir := t.TempDir()
	histFile := filepath.Join(tmpDir, "history.txt")
	content := "git status\ngit commit\nnova top\nnova ls\n"
	_ = os.WriteFile(histFile, []byte(content), 0644)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	caps := terminal.Capabilities{
		ColorProfile: terminal.ColorNone,
		Width:        80,
	}
	ctx := command.NewContext(nil, stdout, stderr, config.Config{}, caps, theme.Get("default"), output.ModePlain, nil)

	cmd := Command()
	if cmd.Name != "history" {
		t.Errorf("expected name history, got %s", cmd.Name)
	}

	// Plain execution
	err := cmd.Run(ctx, []string{"--file=" + histFile, "--plain"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !strings.Contains(stdout.String(), "RANK\tCOMMAND\tCATEGORY") {
		t.Errorf("expected plain header, got: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "git") {
		t.Errorf("expected git in plain output, got: %s", stdout.String())
	}

	// JSON execution
	stdout.Reset()
	ctx = command.NewContext(nil, stdout, stderr, config.Config{}, caps, theme.Get("default"), output.ModeJSON, nil)
	err = cmd.Run(ctx, []string{"--file=" + histFile, "--json"})
	if err != nil {
		t.Fatalf("expected nil error for JSON, got %v", err)
	}

	var rep Report
	if err := json.Unmarshal(stdout.Bytes(), &rep); err != nil {
		t.Fatalf("failed to parse JSON: %v, raw: %s", err, stdout.String())
	}
	if rep.TotalCommands != 4 {
		t.Errorf("expected 4 total commands in JSON, got %d", rep.TotalCommands)
	}
}
