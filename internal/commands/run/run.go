package runcmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"nova/internal/command"
	"nova/internal/output"
)

// Task represents a discovered runnable project task or script.
type Task struct {
	Name        string `json:"name"`
	Command     string `json:"command"`
	Source      string `json:"source"`
	Runtime     string `json:"runtime"`
	Description string `json:"description"`
}

// ExecutionResult encapsulates the outcome of a task execution.
type ExecutionResult struct {
	Task      Task          `json:"task"`
	ExitCode  int           `json:"exit_code"`
	Duration  time.Duration `json:"duration"`
	Success   bool          `json:"success"`
	Timestamp time.Time     `json:"timestamp"`
}

// Command returns the registered Command instance for run / task.
func Command() *command.Command {
	return &command.Command{
		Name:        "run",
		Aliases:     []string{"task", "exec"},
		Summary:     "Smart task runner and project workspace script orchestrator",
		Usage:       "nova run [task-name] [task-args...] [flags]",
		Description: "Automatically discover and execute tasks from package.json, Makefile, go.mod, Cargo.toml, and more.",
		Phase:       23,
		Run:         Run,
	}
}

// Run executes the run / task command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)

	tasks, err := DiscoverTasks(opts.Dir)
	if err != nil {
		return fmt.Errorf("discover tasks: %w", err)
	}

	// If no task specified or list requested, show tasks
	if opts.TaskName == "" || opts.List {
		if ctx.Printer.Mode == output.ModeJSON {
			return ctx.Printer.PrintJSON(tasks)
		}
		if ctx.Printer.Mode == output.ModePlain || opts.Plain {
			RenderTasksPlain(ctx.Stdout, tasks)
		} else {
			RenderTasksDashboard(ctx.Stdout, tasks, ctx)
		}
		return nil
	}

	// Find requested task
	var targetTask *Task
	for _, t := range tasks {
		if strings.EqualFold(t.Name, opts.TaskName) {
			taskCopy := t
			targetTask = &taskCopy
			break
		}
	}

	if targetTask == nil {
		return fmt.Errorf("task %q not found (run 'nova run' to list available tasks)", opts.TaskName)
	}

	return executeTask(targetTask, opts.TaskArgs, ctx)
}

func executeTask(task *Task, extraArgs []string, ctx *command.Context) error {
	start := time.Now()

	if ctx.Printer.Mode != output.ModePlain && ctx.Printer.Mode != output.ModeJSON {
		RenderTaskStartBanner(ctx.Stdout, task, extraArgs, ctx)
	}

	cmd, err := buildCommand(task, extraArgs)
	if err != nil {
		return err
	}

	cmd.Stdin = ctx.Stdin
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr

	runErr := cmd.Run()
	dur := time.Since(start)

	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	res := ExecutionResult{
		Task:      *task,
		ExitCode:  exitCode,
		Duration:  dur,
		Success:   exitCode == 0,
		Timestamp: time.Now(),
	}

	if ctx.Printer.Mode == output.ModeJSON {
		return ctx.Printer.PrintJSON(res)
	}

	if ctx.Printer.Mode != output.ModePlain {
		RenderTaskEndBanner(ctx.Stdout, &res, ctx)
	}

	if exitCode != 0 {
		return fmt.Errorf("task %s exited with status %d", task.Name, exitCode)
	}
	return nil
}

func buildCommand(task *Task, extraArgs []string) (*exec.Cmd, error) {
	switch task.Runtime {
	case "npm", "pnpm", "yarn", "bun":
		args := []string{"run", task.Name}
		if len(extraArgs) > 0 {
			args = append(args, "--")
			args = append(args, extraArgs...)
		}
		return exec.Command(task.Runtime, args...), nil
	case "make":
		args := []string{task.Name}
		args = append(args, extraArgs...)
		return exec.Command("make", args...), nil
	case "cargo":
		parts := strings.Fields(task.Command)
		if len(parts) > 1 {
			args := parts[1:]
			args = append(args, extraArgs...)
			return exec.Command(parts[0], args...), nil
		}
		return exec.Command("cargo", task.Name), nil
	case "go":
		parts := strings.Fields(task.Command)
		if len(parts) > 1 {
			args := parts[1:]
			args = append(args, extraArgs...)
			return exec.Command(parts[0], args...), nil
		}
		return exec.Command("go", task.Name), nil
	default:
		// Shell execution
		fullCmd := task.Command
		if len(extraArgs) > 0 {
			fullCmd += " " + strings.Join(extraArgs, " ")
		}
		return exec.Command("sh", "-c", fullCmd), nil
	}
}

// DiscoverTasks scans the directory for build tools, package managers, and configuration files.
func DiscoverTasks(dir string) ([]Task, error) {
	if dir == "" {
		dir = "."
	}

	var tasks []Task

	// 1. package.json
	pkgPath := filepath.Join(dir, "package.json")
	if data, err := os.ReadFile(pkgPath); err == nil {
		var pkg struct {
			Scripts map[string]string `json:"scripts"`
		}
		if err := json.Unmarshal(data, &pkg); err == nil {
			runtime := detectNodeRuntime(dir)
			for name, cmd := range pkg.Scripts {
				tasks = append(tasks, Task{
					Name:        name,
					Command:     cmd,
					Source:      "package.json",
					Runtime:     runtime,
					Description: fmt.Sprintf("Run %s script '%s'", runtime, name),
				})
			}
		}
	}

	// 2. Makefile
	makePath := filepath.Join(dir, "Makefile")
	if data, err := os.ReadFile(makePath); err == nil {
		targets := parseMakefileTargets(string(data))
		for _, target := range targets {
			tasks = append(tasks, Task{
				Name:        target,
				Command:     "make " + target,
				Source:      "Makefile",
				Runtime:     "make",
				Description: fmt.Sprintf("Execute Makefile target '%s'", target),
			})
		}
	}

	// 3. go.mod
	goModPath := filepath.Join(dir, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		tasks = append(tasks, []Task{
			{Name: "test", Command: "go test -v ./...", Source: "go.mod", Runtime: "go", Description: "Run all package unit tests"},
			{Name: "build", Command: "go build ./...", Source: "go.mod", Runtime: "go", Description: "Compile packages and dependencies"},
			{Name: "vet", Command: "go vet ./...", Source: "go.mod", Runtime: "go", Description: "Run standard Go static analysis"},
			{Name: "tidy", Command: "go mod tidy", Source: "go.mod", Runtime: "go", Description: "Prune and verify go.mod dependencies"},
			{Name: "bench", Command: "go test -bench=. -benchmem ./...", Source: "go.mod", Runtime: "go", Description: "Run benchmark test suites"},
		}...)
	}

	// 4. Cargo.toml
	cargoPath := filepath.Join(dir, "Cargo.toml")
	if _, err := os.Stat(cargoPath); err == nil {
		tasks = append(tasks, []Task{
			{Name: "build", Command: "cargo build", Source: "Cargo.toml", Runtime: "cargo", Description: "Build Rust project"},
			{Name: "test", Command: "cargo test", Source: "Cargo.toml", Runtime: "cargo", Description: "Execute Rust test suite"},
			{Name: "check", Command: "cargo check", Source: "Cargo.toml", Runtime: "cargo", Description: "Analyze Rust project without code gen"},
			{Name: "clippy", Command: "cargo clippy", Source: "Cargo.toml", Runtime: "cargo", Description: "Run Rust linter"},
		}...)
	}

	// 5. Taskfile.yml
	for _, tfName := range []string{"Taskfile.yml", "Taskfile.yaml"} {
		tfPath := filepath.Join(dir, tfName)
		if data, err := os.ReadFile(tfPath); err == nil {
			tasks = append(tasks, parseTaskfile(string(data), tfName)...)
			break
		}
	}

	// Sort tasks by name
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Source != tasks[j].Source {
			return tasks[i].Source < tasks[j].Source
		}
		return tasks[i].Name < tasks[j].Name
	})

	return tasks, nil
}

func detectNodeRuntime(dir string) string {
	if _, err := os.Stat(filepath.Join(dir, "pnpm-lock.yaml")); err == nil {
		return "pnpm"
	}
	if _, err := os.Stat(filepath.Join(dir, "yarn.lock")); err == nil {
		return "yarn"
	}
	if _, err := os.Stat(filepath.Join(dir, "bun.lockb")); err == nil {
		return "bun"
	}
	return "npm"
}

func parseMakefileTargets(content string) []string {
	var targets []string
	targetRegex := regexp.MustCompile(`^([a-zA-Z0-9_\-\.]+):`)
	scanner := bufio.NewScanner(strings.NewReader(content))
	seen := make(map[string]bool)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ".PHONY") || strings.HasPrefix(line, ".") {
			continue
		}
		if match := targetRegex.FindStringSubmatch(line); len(match) > 1 {
			target := match[1]
			if !seen[target] {
				seen[target] = true
				targets = append(targets, target)
			}
		}
	}
	return targets
}

func parseTaskfile(content, source string) []Task {
	var tasks []Task
	lines := strings.Split(content, "\n")
	taskRegex := regexp.MustCompile(`^\s+([a-zA-Z0-9_\-]+):`)
	inTasks := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "tasks:" {
			inTasks = true
			continue
		}
		if inTasks && strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") {
			if match := taskRegex.FindStringSubmatch(line); len(match) > 1 {
				name := match[1]
				tasks = append(tasks, Task{
					Name:        name,
					Command:     "task " + name,
					Source:      source,
					Runtime:     "task",
					Description: fmt.Sprintf("Execute Taskfile task '%s'", name),
				})
			}
		}
	}
	return tasks
}
