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
PHASE 13 Content Intelligence & Search (`grep`)             [COMPLETED]
PHASE 14 System Intelligence & Benchmarking (`sysinfo`)     [COMPLETED]
PHASE 15 Workspace Hygiene & Archive Packaging (`clean`)    [COMPLETED]
PHASE 16 Live Monitoring & Task Automation (`watch`)        [COMPLETED]
PHASE 17 Integrity & Binary Inspection (`hash`, `hex`)          [COMPLETED]
PHASE 18 Environment Telemetry & Next-Gen TUI (`env`, TUI v1.5) [COMPLETED]
PHASE 19 Network Telemetry & HTTP Inspection (`http`)           [COMPLETED]
PHASE 20 Git Productivity Dashboard & Terminal QR (`git`, `qr`) [COMPLETED]
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

---

## Phase 13 — Content Intelligence & Search (`grep`)

**Goal:** Provide concurrent, high-performance content searching across directories and text files.

**Tasks:**
- Implement `nova grep`: Concurrent worker pool content searcher supporting regex and string literals.
- Binary detection: automatically skip binary files using magic bytes detection.
- Options: case-insensitive (`-i`), line numbering (`-n`), count-only (`-c`), files with matches (`-l`), context lines (`-C`, `-B`, `-A`), extension filtering (`--ext`), hidden file search (`--hidden`).
- Output formats: rich human dashboard with neon highlights and match summary badge, newline-delimited `--plain`, and structured `--json`.

**Exit Gate:**
- Unit test coverage passes across all flags and output formats.
- Verified binary skipping, regex compilation, and concurrency safety.

---

## Phase 14 — System Intelligence & Performance Benchmarking (`sysinfo`, `bench`)

**Goal:** Equip developers with system diagnostic telemetry and storage I/O benchmarking directly inside Nova.

**Tasks:**
- Implement `nova sysinfo`: Hardware, OS, Go runtime, memory allocation progress bar, disk usage, terminal dimensions, color profile, Git repository status, and PATH integrity analysis.
- Implement `nova bench`: Sequential write and read throughput test (MB/s), stat metadata latency percentiles (min, p50, p95, p99, max in µs), directory walk rate (files/sec), with graphical bar charts and sparklines.
- Interactive TUI Supercharging:
  - Theme cycling in real-time (`t` key).
  - Sorting mode cycling (`s` key): Name, Size, ModTime, Extension.
  - Hex dump file inspection mode (`x` key).
- Machine modes: full `--plain` and `--json` support across commands.

**Exit Gate:**
- Benchmark and diagnostics execute safely in isolated scratch environments.
- 100% test pass across all packages with zero race conditions or vet warnings.

---

## Phase 15 — Workspace Hygiene & Archive Packaging (`clean`, `archive`)

**Goal:** Automate workspace cleanup and provide portable, pure Go archiving and extraction capabilities.

**Tasks:**
- Implement `nova clean`: Detect OS artifacts (`.DS_Store`, `Thumbs.db`), editor junk (`*.swp`, `*~`), test binaries (`*.test`), broken symlinks, and empty directories (`--empty-dirs`).
- Safe execution: defaults to `--dry-run` preview, requiring `-f` / `--force` for actual deletion. Protects system roots.
- Implement `nova archive`: Pure Go ZIP and TAR.GZ compression and extraction suite (`archive/zip`, `archive/tar`, `compress/gzip`).
- Safe extraction: zip-slip path traversal guards built-in.
- Inspection mode: `nova archive list` reports uncompressed/compressed sizes and savings ratio.

**Exit Gate:**
- Deletion guards verified against accidental system root sweeps.
- Zip-slip path traversal protection confirmed.
- 100% test pass across all unit test suites.

---

## Phase 16 — Live Monitoring & Task Automation (`watch`)

**Goal:** Provide zero-dependency filesystem monitoring and automated task triggering.

**Tasks:**
- Implement `nova watch`: Periodic fingerprint-based change detection across files and directories.
- Detect event categories: `CREATE`, `MODIFY`, `DELETE`.
- Options: debouncing (`-d`), interval tuning (`-i`), screen clearing (`-c`), task command execution (`-e`), iteration limiting (`-n`), extension filtering (`--ext`).
- Interactive TUI enhancements:
  - Command Palette (`:` key): `:theme`, `:sort`, `:mkdir`, `:touch`, `:q`.
  - Bookmark navigation (`b` to add bookmark, `B` to cycle jumps).
- Full `--plain` and `--json` streaming events support.

**Exit Gate:**
- Live watch loops terminate cleanly on SIGINT / SIGTERM with zero goroutine leaks.
- 100% test pass across all 34 packages.

---

## Phase 17 — Integrity, Cryptography & Binary Inspection (`hash`, `hex`)

**Goal:** Provide native, pure Go cryptographic hashing and colorized hex dump inspection.

**Tasks:**
- Implement `nova hash` (aliases: `checksum`, `digest`, `sha256`):
  - Support algorithms: `sha256`, `sha512`, `sha1`, `md5`, `crc32`.
  - Verification mode (`-c` / `--check <file>`): validates standard checksum files with PASS/FAILED/MISSING status badges.
  - Recursive directory tree hashing (`-r` / `--recursive`).
  - Streaming constant-memory pipeline for gigabyte-sized files.
- Implement `nova hex` (aliases: `hexdump`, `dump`, `xxd`):
  - Modern canonical hex dump layout with colorized byte categorizations (null, printable ASCII, control bytes, high bytes).
  - Configurable byte grouping (`-g`), columns (`-c`), byte offset skip (`-s`), and length limits (`-n`).
  - Integrated ASCII preview panel with dot fallback.
  - Pipe & stdin friendly: works seamlessly in command pipelines.

**Exit Gate:**
- 100% test pass across checksum verification, recursive hashing, and hex formats.
- Safe execution with deterministic `--plain` and structured `--json` outputs.

---

## Phase 18 — Environment Telemetry & Advanced TUI Navigation (`env`, TUI v1.5)

**Goal:** Provide environment variable auditing, secret protection, and rich modal TUI dialogs.

**Tasks:**
- Implement `nova env` (aliases: `environ`, `envinfo`):
  - Detects and masks sensitive credentials (`*KEY*`, `*TOKEN*`, `*PASSWORD*`, `*SECRET*`, `*AUTH*`, `*PRIVATE*`).
  - Pattern search and filtering (`-f` / `--filter`).
  - Categorization grouping: Runtimes, System & Shell, Cloud & DevOps, General.
  - Shell export generators: `--export=sh` and `--export=fish`.
- Interactive TUI Supercharging:
  - File Metadata Inspector modal (`i` key): floating modal card detailing full path, exact & human size, octal mode & permissions, timestamps, and symlink targets.
  - Quick SHA-256 Checksum (`#` key): instant hash calculation of focused file in interactive view.
  - Palette integration: `:inspect`, `:hash`, `:hex` commands.

**Exit Gate:**
- Environment variable secrets verified masked by default.
- 100% unit test coverage across all new packages.
- Zero race conditions or vet warnings.

---

## Phase 19 — Network Telemetry & HTTP Inspection (`http`)

**Goal:** Provide native HTTP request execution with network latency breakdown and response preview.

**Tasks:**
- Implement `nova http` (aliases: `fetch`, `curl`, `request`):
  - Request methods: GET, POST, PUT, DELETE, HEAD, PATCH (`-X`).
  - Network latency tracing waterfall: DNS lookup, TCP connect, TLS handshake, TTFB, and Transfer latency.
  - Formatted and syntax-highlighted responses with pretty JSON formatting.
  - Custom headers (`-H`), request body payload (`-d`), file upload (`--data-file`), output to file (`-o`).
  - Headless automation with `--plain` and structured `--json`.

**Exit Gate:**
- Verified against test HTTP servers across GET, POST, headers, and output files.
- 100% unit test coverage and clean vet.

---

## Phase 20 — Git Productivity Dashboard & Terminal QR (`git`, `qr`, TUI v1.6)

**Goal:** Provide repository health visibility, instant terminal QR sharing, and interactive Git diff navigation.

**Tasks:**
- Implement `nova git` (aliases: `gstatus`, `glog`, `repo`):
  - Repository status dashboard (branch, tracking status, ahead/behind counters, staged, unstaged, untracked).
  - Compact commit history graph (`nova git log`).
  - Branch listing (`nova git branch`).
- Implement `nova qr` (aliases: `qrcode`):
  - Pure Go standard library QR matrix generation (Levels L/M, Versions 1-7).
  - High-resolution terminal half-block rendering (`▀`, `▄`, `█`, ` `) with quiet zone border.
  - Inversion mode (`-i`) and file input (`-f`).
- Interactive TUI Supercharging:
  - Inline Git diff viewer (`D` key) rendering colorized unified diff directly in preview pane.
  - Real-time Git branch and dirty badges in header.
  - Palette integration: `:diff` command.

**Exit Gate:**
- 100% test pass across git, qr, and interactive diff tests.
- QR codes verified scannable and format compliant.

---

## Phase 21 — Real-Time Process & Resource Monitor (`top`, `proc`)

**Goal:** Provide full system resource inspection and process monitoring with zero dependencies.

**Tasks:**
- Implement `nova top` (aliases: `proc`, `ps`, `monitor`):
  - Reads `/proc` on Linux with portable fallback for CPU%, RSS, VSize, process states (R, S, D, Z), load average, and uptime.
  - Color gradient progress bars for CPU and Memory utilization.
  - Filtering (`-f`), sorting (`-s cpu|mem|pid|name`), and output limits (`-n`).
  - Process signal delivery (`-k` / `--kill`).
  - Real-time continuous mode (`-w` / `--live`) and deterministic `--plain` / `--json` streams.

**Exit Gate:**
- Verified on Linux `/proc` and portable environments.
- 100% unit test coverage across all metrics and rendering formats.

---

## Phase 22 — Network Diagnostics & Port Scanner (`net`, `ping`)

**Goal:** Provide low-latency network telemetry, TCP ping, port scanning, and DNS record auditing.

**Tasks:**
- Implement `nova net` (aliases: `ping`, `latency`, `portscan`):
  - High-precision TCP ping with microsecond resolution, visual sparkline, and jitter metrics.
  - Concurrent TCP port scanner with service name heuristics and open/closed/timeout badges.
  - DNS resolution query inspecting A, AAAA, CNAME, MX, and TXT records.
  - Fully automation-compatible with `--plain` and structured `--json`.

**Exit Gate:**
- 100% test pass against local test servers and DNS lookups.
- Zero external dependencies.

---

## Phase 23 — Smart Task Runner & TUI Fuzzy Navigation (`run`, TUI v1.7)

**Goal:** Orchestrate project tasks automatically and provide fast fuzzy navigation inside the interactive TUI.

**Tasks:**
- Implement `nova run` (aliases: `task`, `exec`):
  - Automatic task discovery from `package.json`, `Makefile`, `go.mod`, `Cargo.toml`, `Taskfile.yml`.
  - Visual card grid displaying all detected workspace tasks and commands.
  - Direct task execution with real-time timers and exit status indicators.
  - Automatic workspace root discovery and zero-shell security compliance.
- Interactive TUI Supercharging:
  - Fuzzy File Finder overlay modal (`f` or `Ctrl+P`) with real-time match highlighting.
  - Visual bookmark indicators (`🔖`) in directory tree.
  - TrueColor linear gradient title banner.
  - Palette commands: `:find`, `:bookmark`, `:jump`, `:top`, `:net`, `:run`.
- Themes Expansion:
  - Built-in `solarized` and `rose-pine` themes.

**Exit Gate:**
- 100% test pass across all packages and clean security audit.
- Release version `1.7.0` verified and deployed.

---

## Phase 24 — Scientific & Programmer Terminal Calculator (`calc`, `math`, `eval`)

**Goal:** Provide full algebraic, scientific, bitwise, and multi-base programmer arithmetic directly in the terminal.

**Tasks:**
- Implement `nova calc` (aliases: `expr`, `math`, `eval`):
  - Pure Go recursive descent parser supporting `+`, `-`, `*`, `/`, `%`, `^`, bitwise (`<<`, `>>`, `&`, `|`, `^`, `~`).
  - Mathematical functions (`sqrt`, `abs`, `round`, `floor`, `ceil`, `sin`, `cos`, `tan`, `log`, `log2`, `log10`, `pow`, `fact`) and constants (`pi`, `e`, `tau`, `phi`).
  - Unit capacity scaling (`B`, `KB`, `MB`, `GB`, `TB`, `KiB`, `MiB`, `GiB`).
  - Multi-base breakdown: Decimal, Hexadecimal, Binary, Octal, Human-readable bytes.
  - Formatted dashboard output, precision control (`-p`), and raw pipeline `--plain` / structured `--json`.

**Exit Gate:**
- Comprehensive unit tests covering syntax, precedence, division-by-zero guards, bitwise shifts, and multi-base rendering.
- 100% test pass.

---

## Phase 25 — Shell History Intelligence & Command Analytics (`history`, `hist`)

**Goal:** Parse shell history across modern shells and visualize productivity analytics with zero dependencies.

**Tasks:**
- Implement `nova history` (aliases: `hist`, `analytics`):
  - Multi-shell parser: Bash (`~/.bash_history`), Zsh extended history (`~/.zsh_history`), and Fish (`~/.local/share/fish/fish_history`).
  - Command distribution breakdown with horizontal ANSI bar charts and percentage calculations.
  - Productivity categorization (Git, Development, System, Containers, Network, Editors, Shell utilities).
  - Search queries (`-q`), top limits (`-n`), custom file overrides (`-f`), and pipeline `--plain` / structured `--json`.

**Exit Gate:**
- Unit tests covering Bash, Zsh extended timestamp formats, and Fish history files.
- 100% test pass.

---

## Phase 26 — Zero-Config Web Server & TUI Archive Previews (`serve`, TUI v1.8)

**Goal:** Deliver instant static file serving with QR code pairing and supercharge the interactive TUI with archive inspection.

**Tasks:**
- Implement `nova serve` (aliases: `server`, `httpd`):
  - Zero-config static HTTP file server with auto-failover port binding (increments up to 100 ports).
  - Single Page Application (SPA) fallback routing mode (`--spa`).
  - Terminal QR code generation rendering half-block QR code for instant mobile device pairing.
  - Real-time colorized HTTP request logging table with status codes, badges, and latency metrics.
  - CORS header injection (`--cors`) and one-shot probe testing (`--once`).
- Interactive TUI Supercharging:
  - Archive (`.zip`, `.jar`) content preview with file listing and compression savings ratio.
  - Inline command palette calculator (`:calc <expr>`) with multi-base results.
  - Palette tips for `:history` and `:serve`.

**Exit Gate:**
- 100% test pass across all 46 packages and clean security audit.
- Release version `1.8.0` verified and deployed.

---

## Phase 27 — Terminal Markdown Document Reader (`md`, `doc`, `view`)

**Goal:** Provide an ergonomic, beautiful terminal Markdown reader with styled headings, syntax code blocks, and table auto-formatting.

**Tasks:**
- Implement `nova md` (aliases: `doc`, `view`, `markdown`):
  - Pure Go Markdown parser for headings, bulleted/numbered lists, task checkboxes (`[ ]` / `[✓]`), blockquotes (`│ `), and horizontal rules.
  - Table auto-formatting with Unicode box borders and column width calculations.
  - Code block rendering with syntax highlighting via `cat/syntax`.
  - Table of Contents generator (`-t` / `--toc`).
  - Text line-wrapping (`-w`), interactive pager (`-p`), pipeline `--plain`, and structured `--json`.

**Exit Gate:**
- 100% test pass across parser, tables, code blocks, and TOC generators.

---

## Phase 28 — Developer Notes & TODO Checklist Tracker (`note`, `notes`, `todo`)

**Goal:** Build a lightweight terminal scratchpad and markdown task checklist tracker.

**Tasks:**
- Implement `nova note` (aliases: `notes`, `todo`, `memo`):
  - Local storage in `~/.config/nova/notes/` or workspace `.nova/notes/`.
  - Automatic task checklist extraction from `- [ ]` and `- [x]` markdown items with progress bars.
  - In-place task toggling (`nova note toggle <id> <task#>`) without external editor.
  - Fast search, tags filtering, and pipeline `--plain` / structured `--json`.

**Exit Gate:**
- 100% test pass across lifecycle (add, show, todo, toggle, search, rm).

---

## Phase 29 — SSL/TLS Certificate Validator & JWT Decoder (`cert`, `ssl`, `jwt`)

**Goal:** Deliver low-latency SSL/TLS certificate inspection and JSON Web Token decoding with zero dependencies.

**Tasks:**
- Implement `nova cert` (aliases: `ssl`, `tls`, `jwt`):
  - Remote TLS host connection, SNI, TLS version, cipher suite, certificate chain extraction.
  - Expiration countdowns, SANs, serial numbers, RSA/ECDSA key algorithms.
  - Local certificate file (`.crt`, `.pem`, `.cer`) parsing.
  - Pure Go JWT token decoder splitting header, payload, and signature with claim evaluation (`exp`, `iat`, `nbf`, `sub`).
  - Pipeline `--plain` and structured `--json`.
- Interactive TUI Supercharging:
  - Formatted Markdown previews in preview pane.
  - Numeric bookmark jump shortcuts (`1-9`).
  - Command palette tips for `:md`, `:note`, `:cert`.

**Exit Gate:**
- 100% test pass across all 49 packages.
- Release version `1.9.0` verified and deployed.






