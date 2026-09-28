package runcmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"nova/internal/command"
	"nova/internal/config"
	"nova/internal/output"
	"nova/internal/terminal"
	"nova/internal/theme"
)

func TestFlagsParse(t *testing.T) {
	opts := ParseFlags([]string{"build", "--dir=/tmp", "--plain", "arg1", "arg2"})
	if opts.TaskName != "build" {
		t.Errorf("expected TaskName build, got %s", opts.TaskName)
	}
	if opts.Dir != "/tmp" {
		t.Errorf("expected Dir /tmp, got %s", opts.Dir)
	}
	if !opts.Plain {
		t.Errorf("expected Plain true")
	}
	if len(opts.TaskArgs) != 2 || opts.TaskArgs[0] != "arg1" || opts.TaskArgs[1] != "arg2" {
		t.Errorf("expected TaskArgs [arg1, arg2], got %v", opts.TaskArgs)
	}
}

func TestDiscoverTasksGoMod(t *testing.T) {
	tmpDir := t.TempDir()
	goModContent := "module testmod\n\ngo 1.22\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to create mock go.mod: %v", err)
	}

	tasks, err := DiscoverTasks(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverTasks failed: %v", err)
	}

	foundTest := false
	foundBuild := false
	for _, tsk := range tasks {
		if tsk.Name == "test" && tsk.Runtime == "go" {
			foundTest = true
		}
		if tsk.Name == "build" && tsk.Runtime == "go" {
			foundBuild = true
		}
	}

	if !foundTest || !foundBuild {
		t.Errorf("expected Go test and build tasks, got %+v", tasks)
	}
}

func TestDiscoverTasksPackageJSON(t *testing.T) {
	tmpDir := t.TempDir()
	pkgJSON := `{
		"name": "mock-app",
		"scripts": {
			"lint": "eslint .",
			"compile": "tsc"
		}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatalf("failed to create mock package.json: %v", err)
	}

	tasks, err := DiscoverTasks(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverTasks failed: %v", err)
	}

	foundLint := false
	foundCompile := false
	for _, tsk := range tasks {
		if tsk.Name == "lint" && tsk.Source == "package.json" {
			foundLint = true
		}
		if tsk.Name == "compile" && tsk.Source == "package.json" {
			foundCompile = true
		}
	}

	if !foundLint || !foundCompile {
		t.Errorf("expected lint and compile tasks from package.json, got %+v", tasks)
	}
}

func TestDiscoverTasksMakefile(t *testing.T) {
	tmpDir := t.TempDir()
	makeContent := ".PHONY: all clean\nall:\n\t@echo all\nclean:\n\t@echo clean\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "Makefile"), []byte(makeContent), 0644); err != nil {
		t.Fatalf("failed to create mock Makefile: %v", err)
	}

	tasks, err := DiscoverTasks(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverTasks failed: %v", err)
	}

	foundClean := false
	for _, tsk := range tasks {
		if tsk.Name == "clean" && tsk.Runtime == "make" {
			foundClean = true
			break
		}
	}

	if !foundClean {
		t.Errorf("expected clean target from Makefile, got %+v", tasks)
	}
}

func TestRunCommandList(t *testing.T) {
	tmpDir := t.TempDir()
	goModContent := "module testmod\n\ngo 1.22\n"
	_ = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	caps := terminal.Capabilities{
		ColorProfile: terminal.ColorNone,
		Width:        80,
	}
	ctx := command.NewContext(nil, stdout, stderr, config.Config{}, caps, theme.Get("default"), output.ModePlain, nil)

	cmd := Command()
	if cmd.Name != "run" {
		t.Errorf("expected command name run, got %s", cmd.Name)
	}

	// Plain list
	err := cmd.Run(ctx, []string{"--dir=" + tmpDir, "--plain"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("test\tgo\tgo.mod")) {
		t.Errorf("expected plain test task in output, got: %s", stdout.String())
	}

	// JSON list
	stdout.Reset()
	ctx = command.NewContext(nil, stdout, stderr, config.Config{}, caps, theme.Get("default"), output.ModeJSON, nil)
	err = cmd.Run(ctx, []string{"--dir=" + tmpDir, "--json"})
	if err != nil {
		t.Fatalf("expected nil error for JSON, got %v", err)
	}

	var parsed []Task
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v, raw: %s", err, stdout.String())
	}
	if len(parsed) == 0 {
		t.Errorf("expected tasks in JSON output")
	}
}
