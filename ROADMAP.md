# nova — Roadmap & Phased Execution Plan

A phased delivery strategy adhering to a **STRICT STOP** protocol after every phase.
Each phase concludes with its specific exit gate validated, a formal phase report delivered, and an explicit user approval (`CONTINUE` or `NEXT`) before any work on the subsequent phase begins. Phases are strictly sequential and must never be skipped or combined.

```text
PHASE 0  Discovery & Project Contract                       [COMPLETED]
PHASE 1  Foundation (CLI Router, Config, Modes, Detection)   [COMPLETED]
PHASE 2  Terminal Engine & Design System                    [COMPLETED]
PHASE 3  `ls` (Flagship Directory Listing)                  [COMPLETED]
PHASE 4  `cat` / File Viewer                                [COMPLETED]
PHASE 5  `tree`, `find`, `stat`, `du`                       [COMPLETED]
PHASE 6  `cp`, `mv`, `rm`, `mkdir` (Safe File Operations)   [COMPLETED]
PHASE 7  Interactive UX (`nova` TUI)                        [COMPLETED]
PHASE 8  Git Awareness & Smart Features                     [COMPLETED]
PHASE 9  Performance, Security, Accessibility & Hardening   [COMPLETED]
PHASE 10 Release Engineering & Production Readiness         [COMPLETED]
PHASE 11 Advanced Tooling (`which`, `touch`, `diff`, compl) [COMPLETED]
PHASE 12 Next-Gen UX & Visual FX (Themes, Animations, TUI)  [COMPLETED]
```

---

## Phase 0 — Discovery & Project Contract

**Goal:** Establish the project foundation, assess environment, and formalize project governance before writing production code.

**Tasks:**
- Inspect system environment (OS, Go version, architecture, repository status, existing files).
- Establish project naming (`nova`), module configuration (`go.mod`), license, and git repository.
- Formulate comprehensive requirements (`PRD.md`), roadmap (`ROADMAP.md`), and developer overview (`README.md`).
- Define scope, non-goals, architectural principles, safety rules, and quality gates.

**Deliverables:**
- `README.md`
- `PRD.md`
- `ROADMAP.md`
- `LICENSE`
- `.gitignore`
- `go.mod`

**Exit Gate:**
- Documentation is comprehensive and coherent.
- Project scope and non-goals are unambiguously defined.
- Initial files exist and are clean of unformatted code.
- No accidental destructive changes occurred.
- Strict Stop report presented; waiting for explicit user approval.

---

## Phase 1 — Foundation

**Goal:** Implement the minimal production-grade Go CLI skeleton that builds and executes.

**Tasks:**
- Establish application entry point (`cmd/nova/main.go`).
- Build robust CLI command routing, flag parser, and context lifecycle (`internal/cli`).
- Establish configuration management (`internal/config`).
- Establish error models and structured debug logging (`internal/cli/errors.go`).
- Implement tri-mode output abstraction: Human, Plain, JSON (`internal/output`).
- Implement basic terminal capability detection (TTY vs Pipe).
- Establish test infrastructure, unit test harness, and quality scripts.
- Implement `nova --help` and `nova --version`. Unimplemented subcommands return explicit notices, never fake stubs.

**Exit Gate:**
- `nova --help` and `nova --version` execute cleanly with exit code 0.
- All quality gates pass: `gofmt -l .` clean, `go vet ./...` clean, `go test ./...` passes, `go test -race ./...` passes, `go build ./...` succeeds.
- Strict Stop report presented; waiting for explicit user approval.

---

## Phase 2 — Terminal Engine & Design System

**Goal:** Build the centralized visual foundation and rendering engine.

**Tasks:**
- Advanced terminal capability detection: TTY detection, dynamic terminal width (`TIOCGWINSZ`), color profile (TrueColor, 256-color, 16-color, Mono), Unicode/Nerd Font detection (`internal/terminal`).
- Semantic theme system: Color roles, palette definitions (`default`, `minimal`, `mono`, `nord`, `dracula`, `neon`) (`internal/theme`).
- Icon system: File type mapping with reliable ASCII fallback.
- Layout engine: Text truncation, visual width calculation (accounting for wide runes/emojis), padding, alignment, column calculation, grid/table formatters (`internal/renderer`).
- Pipeline safety: Rendering engine automatically strips formatting and emojis when piped or in Plain mode.
- Golden/snapshot tests covering all themes, widths, and fallback scenarios.

**Exit Gate:**
- Visual layout engine passes snapshot tests across narrow (40 cols), standard (80 cols), and wide (160 cols) terminals.
- Automatic fallback from TrueColor to ASCII verified.
- Pipeline safety verified with redirection tests.
- All quality gates pass.
- Strict Stop report presented; waiting for explicit user approval.

---

## Phase 3 — `ls` (Flagship Directory Listing)

**Goal:** Create the signature directory listing command that balances beauty, speed, and composability.

**Tasks:**
- Implement `nova ls` supporting compact and long formats (`-l`).
- Attributes: permissions, human-readable file sizes (`-h`), modification timestamps, file owners/groups, symlink targets.
- Sorting options: by name, size (`-S`), modification time (`-t`), extension, with reverse support (`-r`).
- Filtering: show hidden/dotfiles (`-a`), directory-first sorting, glob filtering.
- Visual integration: Semantic coloring, icons, responsive columnar grid adapting to terminal width.
- Machine modes: `--plain` (newline-delimited) and `--json` (structured metadata stream).
- Recursive listing (`-R`) with depth control.

**Exit Gate:**
- Performance benchmarks on empty, small (100 files), medium (1,000 files), and large (50,000 files) directories.
- Edge case test coverage: Unicode names, spaces, leading dashes, broken symlinks, permission-denied directories.
- TTY rendering vs piped output verified.
- All quality gates pass.
- Strict Stop report presented; waiting for explicit user approval.

---

## Phase 4 — `cat` / File Viewer

**Goal:** Create a modern, memory-efficient file viewing and streaming experience.

**Tasks:**
- Implement `nova cat [files...]` with buffered streaming I/O (constant memory footprint).
- Syntax highlighting with automatic language detection based on extension and content shebang.
- Features: Line numbers (`-n`), non-printable character rendering, binary file detection with safe inspection warnings.
- Paging support: Integrated lightweight pager for interactive viewing (`↑`/`↓`, `PageUp`/`PageDown`, search `/`, `q` to quit).
- Stdin streaming support: Seamlessly acts as a pager or highlighter in pipelines.
- Formats: Pretty-printed JSON view when requested (`--json`).

**Exit Gate:**
- Verified on empty files, tiny files, huge files (1GB+ streaming test with flat memory usage), binary files, invalid UTF-8 files, very long single lines.
- Pipe input (`echo "test" | nova cat`) and redirected output verified.
- All quality gates pass.
- Strict Stop report presented; waiting for explicit user approval.

---

## Phase 5 — `tree`, `find`, `stat`, `du` (Filesystem Inspection)

**Goal:** Build advanced filesystem inspection utilities on top of a single unified traversal engine.

**Tasks:**
- Shared traversal engine in `internal/filesystem`: cycle detection, permission handling, bounded concurrency, cancellation context.
- `nova tree`: Recursive visual directory hierarchy, depth limiting (`-L`), file counts, aggregate sizes, icon tree connectors.
- `nova find`: Fast query engine supporting predicates (`--name`, `--type`, `--size`, `--mtime`, `--max-depth`), with `--plain` and `--json` streams.
- `nova stat`: Deep structured metadata display (permissions, timestamps, inode, block count, device ID, file type).
- `nova du`: Disk usage accumulator with directory sorting, human units, depth limits, and cycle prevention.

**Exit Gate:**
- Traversal validated against deep nesting (100+ levels), circular symlinks, permission traps, and special files (sockets, named pipes).
- Zero code duplication between traversal logic of `tree`, `find`, and `du`.
- All quality gates pass.
- Strict Stop report presented; waiting for explicit user approval.

---

## Phase 6 — `cp`, `mv`, `rm`, `mkdir` (Safe File Operations)

**Goal:** Implement fast, defensive file manipulation commands where safety outranks visual polish.

**Tasks:**
- `nova cp`: High-speed streaming copy, recursive directory copy (`-r`), metadata preservation (`-p`), overwrite safety (`-i`, `-n`, `-f`), progress indicator for large files, `--dry-run`.
- `nova mv`: Fast file/directory rename, atomic moves where supported, cross-device fallback, overwrite safeguards, `--dry-run`.
- `nova rm`: Safe deletion, interactive confirmation on multiple or recursive targets, symlink safety (never delete symlink targets), broad deletion guards (`/`, `~`, `..`), `--dry-run`.
- `nova mkdir`: Directory hierarchy creation with standard `-p` semantics.

**Exit Gate:**
- 100% of destructive operation tests executed strictly in isolated temp directories.
- Verified overwrite prevention, broad deletion blocking, and atomic move semantics.
- All quality gates pass.
- Strict Stop report presented; waiting for explicit user approval.

---

## Phase 7 — Interactive UX (`nova` TUI)

**Goal:** Build the signature interactive terminal interface for exploratory filesystem navigation.

**Tasks:**
- Full-screen TUI launched via `nova` or `nova interactive`.
- Non-blocking event loop using terminal raw mode and escape code parsers.
- Dual-pane or multi-pane layout: Directory tree / item list on the left, file preview / metadata on the right.
- Keyboard navigation (`↑`/`k`, `↓`/`j`, `Enter` to open, `Backspace` to parent, `q` to quit, `?` for help).
- Real-time fuzzy search / filtering.
- Terminal resize handling (`SIGWINCH`) with zero visual corruption.
- Asynchronous file loading and previews to prevent freezing the UI.

**Exit Gate:**
- Tested in narrow, wide, and resized terminal dimensions.
- Verified responsive interaction on large directories (10k+ items) without UI thread blocking.
- Clean terminal teardown on normal quit, `SIGINT` (Ctrl+C), or terminal close.
- All quality gates pass.
- Strict Stop report presented; waiting for explicit user approval.

---

## Phase 8 — Git Awareness & Smart Features

**Goal:** Provide contextual repository status without turning the suite into a Git client or slowing basic commands.

**Tasks:**
- Lightweight Git inspection in `internal/git`: read Git repository state directly or via minimal cached query.
- Git status indicators in `ls`, `tree`, and interactive mode: modified (`M`), added (`A`), untracked (`?`), deleted (`D`), ignored (`!`).
- Smart heuristics: file type classification, contextual badges, directory health summaries.
- Performance boundary: If Git status query exceeds timeout (e.g. 50ms on massive monorepos), gracefully omit Git status without blocking the command.

**Exit Gate:**
- Benchmarked inside non-Git directories, normal Git repos, large monorepos, and detached HEAD states.
- Zero performance regression on standard `ls` and `tree` in non-git directories.
- All quality gates pass.
- Strict Stop report presented; waiting for explicit user approval.

---

## Phase 9 — Performance, Security, Accessibility & Hardening

**Goal:** Thoroughly harden, benchmark, and secure all codebase components.

**Tasks:**
- Comprehensive profiling: CPU, memory allocations, goroutine lifecycles using `pprof` and `benchstat`.
- Concurrency audit: Strict race condition analysis (`go test -race ./...`), worker pool bounding, goroutine leak checks.
- Security audit: Shell injection prevention (0 shell executions), path traversal validation (`..`), symlink attack resistance, TOCTOU handling.
- Accessibility audit: Full functional parity in `--plain` and `NO_COLOR=1` environments; high-contrast theme validation.
- Error UX audit: Uniform, actionable error reporting across all commands.

**Exit Gate:**
- `go test -race ./...` passes cleanly across all test suites.
- Static analysis and security checklists completed with zero critical issues.
- Benchmark comparisons recorded and validated against performance targets.
- All quality gates pass.
- Strict Stop report presented; waiting for explicit user approval.

---

## Phase 10 — Release Engineering & Production Readiness

**Goal:** Prepare the production open-source release artifacts, comprehensive documentation, and cross-platform distribution.

**Tasks:**
- Complete end-user documentation: command reference guides, installation guides, configuration guides, contributing guide (`CONTRIBUTING.md`), changelog (`CHANGELOG.md`).
- Multi-architecture build pipelines: Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64, arm64).
- Artifact generation: reproducible binary builds, SHA256 checksums, automated version injection via ldflags.
- Verification of installation pathways: direct binary, go install, packaging scripts.

**Exit Gate:**
- Reproducible cross-compiled binaries verified on target architectures.
- All documentation complete, accurate, and hyperlinked.
- Final quality gate in PRD §10 satisfied: Functional, Fast, Beautiful, Responsive, Safe, Tested, Documented, Maintainable.
- Strict Stop report presented; waiting for final release authorization.

---

## Phase 11 — Advanced Tooling (`which`, `touch`, `diff`, `completion`)

**Goal:** Broaden Nova's core utility suite with fast, high-utility commands for daily developer workflows.

**Tasks:**
- Implement `nova which`: Locate executables in `$PATH`, resolve symlink targets, format file sizes, permissions, and multi-match lookups (`-a`).
- Implement `nova touch`: Create files and update timestamps, with `--parents` / `-p` recursive directory auto-creation, custom dates (`-d`), and reference copying (`-r`).
- Implement `nova diff`: Pure Go unified diff engine with LCS algorithm, line numbers, green/red addition/deletion highlights, and change summary cards.
- Implement `nova completion`: Generate native autocompletion code for `bash`, `zsh`, and `fish`.
- Full `--plain` and `--json` support across all new commands.

**Exit Gate:**
- Unit test coverage passes across all 4 new commands.
- Benchmark and race detection verification (`go test -race ./...`).

---

## Phase 12 — Next-Gen UX & Visual FX (Themes, Animations, TUI)

**Goal:** Transform Nova into a visually stunning, responsive terminal environment.

**Tasks:**
- Visual Effects Engine (`internal/renderer`): Braille animated spinners, proportional gradient progress bars, 8-level sparklines, relative human timestamps, rounded card framing.
- Designer Themes (`internal/theme`): 5 new palettes (`cyberpunk`, `synthwave`, `tokyo-night`, `catppuccin`, `gruvbox`) with TrueColor RGB gradients.
- Interactive TUI Overhaul: Multi-select (`Space`), quick file creation (`n`), quick directory creation (`N`), preview collapse/expand toggle (`p`), inline rename (`r`), delete with confirm (`d`), copy path toast (`c`), breadcrumbs navigation, and real-time fuzzy search highlighting.

