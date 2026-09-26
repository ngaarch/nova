package cli

import (
	"bytes"
	"encoding/json"
	"os"
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
	code := app.Run([]string{"cp"}, strings.NewReader(""), &stdout, &stderr)

	if code != ExitFailure {
		t.Fatalf("expected ExitFailure (1) for unimplemented command, got %d", code)
	}
	errStr := stdout.String() + stderr.String()
	if !strings.Contains(errStr, "command \"cp\" is not implemented yet") {
		t.Errorf("expected explicit unimplemented message, got: %s", errStr)
	}
	if !strings.Contains(errStr, "Phase 6") {
		t.Errorf("expected roadmap Phase 6 mention, got: %s", errStr)
	}
}

func TestAppExecutePhase5Commands(t *testing.T) {
	app := NewApp()
	tmpDir := t.TempDir()
	os.WriteFile(tmpDir+"/f.txt", []byte("data"), 0644)

	// Test tree
	var stdout, stderr bytes.Buffer
	code := app.Run([]string{"tree", "--plain", tmpDir}, strings.NewReader(""), &stdout, &stderr)
	if code != ExitSuccess {
		t.Errorf("expected ExitSuccess for tree, got %d; stderr: %s", code, stderr.String())
	}

	// Test find
	stdout.Reset()
	stderr.Reset()
	code = app.Run([]string{"find", tmpDir, "--plain"}, strings.NewReader(""), &stdout, &stderr)
	if code != ExitSuccess {
		t.Errorf("expected ExitSuccess for find, got %d; stderr: %s", code, stderr.String())
	}

	// Test stat
	stdout.Reset()
	stderr.Reset()
	code = app.Run([]string{"stat", "--plain", tmpDir + "/f.txt"}, strings.NewReader(""), &stdout, &stderr)
	if code != ExitSuccess {
		t.Errorf("expected ExitSuccess for stat, got %d; stderr: %s", code, stderr.String())
	}

	// Test du
	stdout.Reset()
	stderr.Reset()
	code = app.Run([]string{"du", "--plain", tmpDir}, strings.NewReader(""), &stdout, &stderr)
	if code != ExitSuccess {
		t.Errorf("expected ExitSuccess for du, got %d; stderr: %s", code, stderr.String())
	}
}

func TestAppExecuteCat(t *testing.T) {
	app := NewApp()
	var stdout, stderr bytes.Buffer
	code := app.Run([]string{"cat", "--plain"}, strings.NewReader("hello from test\n"), &stdout, &stderr)

	if code != ExitSuccess {
		t.Fatalf("expected ExitSuccess (0) running cat, got %d; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "hello from test") {
		t.Errorf("expected cat to stream stdin to stdout, got: %q", stdout.String())
	}
}

func TestAppExecuteLS(t *testing.T) {
	app := NewApp()
	var stdout, stderr bytes.Buffer
	code := app.Run([]string{"ls", "."}, strings.NewReader(""), &stdout, &stderr)

	if code != ExitSuccess {
		t.Fatalf("expected ExitSuccess (0) for ls, got %d; stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "app.go") && !strings.Contains(out, "go.mod") {
		t.Errorf("expected directory items in ls output, got: %s", out)
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
