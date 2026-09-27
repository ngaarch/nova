# Contributing to Nova

Thank you for your interest in contributing to `nova`! `nova` is an ambitious, modern terminal utility suite built with strict systems programming rigor, clean architectural boundaries, and standard-library-first principles.

---

## 1. Core Principles & Architecture Rules

1. **Standard Library First**: Zero unvetted external dependencies. Standard library packages are utilized exclusively unless explicitly approved.
2. **Subprocess Isolation**: Never execute shell wrappers (`sh`, `bash`, `cmd.exe`, or `-c`). Any subprocess execution must invoke binaries directly with structured argument slices and strict context timeouts.
3. **Tri-Mode Contract**: Every inspection and discovery command must support:
   - **Human Mode**: Responsive layouts, semantic colors, and icons when outputting to an interactive TTY.
   - **Plain Mode (`--plain`)**: Tab- or newline-separated deterministic text with zero ANSI escapes for pipelines and scripting.
   - **JSON Mode (`--json`)**: Valid, indented JSON serialization for machine consumption.
4. **Command Independence**: Subpackages under `internal/commands/*` must never import each other. All shared functionality belongs in `internal/filesystem`, `internal/renderer`, `internal/theme`, or `internal/git`.
5. **Defensive Filesystem Operations**: All mutating commands (`cp`, `mv`, `rm`, `mkdir`) must enforce `ProtectRoot` and destination recursion checks (`ErrDestInsideSource`).

---

## 2. Development Setup

### Prerequisites
- Go 1.27+
- Git

### Building Locally
```bash
git clone https://github.com/izzdev/nova.git
cd nova
go build -o nova ./cmd/nova
```

---

## 3. Testing & Quality Standards

Before opening a pull request or submitting code, ensure that all tests, linters, and race condition audits pass cleanly:

```bash
# 1. Run all unit and integration tests
go test -v ./...

# 2. Run race condition detection across all packages
go test -race ./...

# 3. Verify static analysis
go vet ./...

# 4. Run benchmarks
go test -bench=. ./internal/cli/... ./internal/git/...
```

---

## 4. Submitting a Pull Request

1. Fork the repository and create your feature branch:
   ```bash
   git checkout -b feat/my-new-feature
   ```
2. Commit your changes with descriptive Conventional Commit messages:
   - `feat(ls): add sort by extension support`
   - `fix(tree): correct symlink arrow alignment in plain mode`
   - `test(security): add test for deep symlink loop`
3. Ensure test coverage is added for any new feature or bugfix.
4. Push to your fork and submit a Pull Request.

---

## 5. Code Style & Formatting

- All Go code must be formatted using standard `gofmt` or `goimports`.
- Preserve existing comments and docstrings.
- Provide clear error diagnostics via `command.NewOpError` or `command.NewUsageError` with helpful resolution advice.
