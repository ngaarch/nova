package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestAppHelp(t *testing.T) {
	app := NewApp()
	var stdout, stderr bytes.Buffer
	code := app.Run([]string{"--help"}, strings.NewReader(""), &stdout, &stderr)

	if code != ExitSuccess {
		t.Fatalf("expected ExitSuccess (0), got %d; stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "nova — modern, fast, beautiful") {
		t.Errorf("help missing banner: %s", out)
	}
	if !strings.Contains(out, "ls") || !strings.Contains(out, "cat") {
		t.Errorf("help missing command list: %s", out)
	}
}

func TestAppNoArgsShowsHelp(t *testing.T) {
	app := NewApp()
	var stdout, stderr bytes.Buffer
	code := app.Run([]string{}, strings.NewReader(""), &stdout, &stderr)

	if code != ExitSuccess {
		t.Fatalf("expected ExitSuccess (0) for no args, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Errorf("expected Usage in output, got: %s", stdout.String())
	}
}

func TestAppVersion(t *testing.T) {
	app := NewApp()
	var stdout, stderr bytes.Buffer
	code := app.Run([]string{"--version"}, strings.NewReader(""), &stdout, &stderr)

	if code != ExitSuccess {
		t.Fatalf("expected ExitSuccess (0), got %d", code)
	}
	out := stdout.String()
	if !strings.Contains(out, "nova 0.1.0-dev") {
		t.Errorf("expected version 0.1.0-dev, got: %s", out)
	}
}

func TestAppVersionJSON(t *testing.T) {
	app := NewApp()
	var stdout, stderr bytes.Buffer
	code := app.Run([]string{"--version", "--json"}, strings.NewReader(""), &stdout, &stderr)

	if code != ExitSuccess {
		t.Fatalf("expected ExitSuccess (0), got %d", code)
	}

	var info VersionInfo
	if err := json.Unmarshal(stdout.Bytes(), &info); err != nil {
		t.Fatalf("failed to parse JSON version: %v; raw: %s", err, stdout.String())
	}
	if info.Version != "0.1.0-dev" {
		t.Errorf("expected version 0.1.0-dev in JSON, got %q", info.Version)
	}
	if info.Platform == "" {
		t.Errorf("expected non-empty platform in JSON")
	}
}

func TestAppUnknownCommand(t *testing.T) {
	app := NewApp()
	var stdout, stderr bytes.Buffer
	code := app.Run([]string{"nonexistent_cmd"}, strings.NewReader(""), &stdout, &stderr)

	if code != ExitUsage {
		t.Fatalf("expected ExitUsage (2), got %d", code)
	}
	errStr := stdout.String() + stderr.String()
	if !strings.Contains(errStr, "unknown command \"nonexistent_cmd\"") {
		t.Errorf("expected unknown command error, got: %s", errStr)
	}
}

func TestAppUnimplementedCommand(t *testing.T) {
	app := NewApp()
	var stdout, stderr bytes.Buffer
	code := app.Run([]string{"ls"}, strings.NewReader(""), &stdout, &stderr)

	if code != ExitFailure {
		t.Fatalf("expected ExitFailure (1) for unimplemented command, got %d", code)
	}
	errStr := stdout.String() + stderr.String()
	if !strings.Contains(errStr, "command \"ls\" is not implemented yet") {
		t.Errorf("expected explicit unimplemented message, got: %s", errStr)
	}
	if !strings.Contains(errStr, "Phase 3") {
		t.Errorf("expected roadmap Phase 3 mention, got: %s", errStr)
	}
}

func TestAppExecuteRegisteredCommand(t *testing.T) {
	app := NewApp()
	executed := false
	app.RegisterCommand(&Command{
		Name:    "testcmd",
		Summary: "Test custom command",
		Run: func(ctx *Context, args []string) error {
			executed = true
			ctx.Printer.Println("executed testcmd")
			return nil
		},
	})

	var stdout, stderr bytes.Buffer
	code := app.Run([]string{"testcmd"}, strings.NewReader(""), &stdout, &stderr)

	if code != ExitSuccess {
		t.Fatalf("expected ExitSuccess (0), got %d", code)
	}
	if !executed {
		t.Errorf("expected testcmd to be executed")
	}
	if !strings.Contains(stdout.String(), "executed testcmd") {
		t.Errorf("expected output from testcmd, got: %s", stdout.String())
	}
}

func TestAppExitErrorPropagation(t *testing.T) {
	app := NewApp()
	app.RegisterCommand(&Command{
		Name:    "failcmd",
		Summary: "Fails with custom exit error",
		Run: func(ctx *Context, args []string) error {
			return NewOpError("permission denied on test path", "/tmp/forbidden", nil, "Check user permissions.")
		},
	})

	var stdout, stderr bytes.Buffer
	code := app.Run([]string{"failcmd"}, strings.NewReader(""), &stdout, &stderr)

	if code != ExitFailure {
		t.Fatalf("expected ExitFailure (1), got %d", code)
	}
	errStr := stdout.String() + stderr.String()
	if !strings.Contains(errStr, "permission denied on test path") {
		t.Errorf("expected custom error message, got: %s", errStr)
	}
	if !strings.Contains(errStr, "/tmp/forbidden") {
		t.Errorf("expected path in error message, got: %s", errStr)
	}
	if !strings.Contains(errStr, "Check user permissions.") {
		t.Errorf("expected resolution in error message, got: %s", errStr)
	}
}
