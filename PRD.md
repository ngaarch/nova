# nova — Product Requirements Document (PRD)

Status: Living document, authoritative for scope and engineering contracts during phased development.
Product Name: `nova` (provisional name; renaming before release remains low-cost).

---

## 1. Problem Statement

Classic Unix utilities (`ls`, `cat`, `cp`, `mv`, `rm`, `find`, `tree`, `stat`, `du`) were invented in the 1970s for teletypes and early terminals. While their composability and minimalism remain gold standards, modern computing demands:
1. **Adaptive terminal UX**: dynamic width adjustment, semantic color coding, Nerd Font / Unicode icon support with robust ASCII fallbacks, and intuitive responsive tabular layouts.
2. **Deterministic machine piping**: many "modern" command line replacements break Unix pipelines by dumping ANSI escapes, non-standard tables, or unpredictable column wraps into stdout when redirected.
3. **Safety by default**: traditional destructive commands (`rm`, `mv`, `cp`) make catastrophic mistakes trivial, lacking built-in dry-run awareness, path traversal protections, or clear interactive confirmations.
4. **Performance without bloat**: compiled tools frequently suffer from slow startup, memory bloat on large directories or huge files, or unnecessary subprocess forks.

`nova` resolves this tension: delivering 2026-era terminal ergonomics and safety while preserving Unix composability, sub-millisecond startup discipline, and rock-solid deterministic machine modes.

---

## 2. Product Vision & Core Principles

> A modern, fast, beautiful, interactive, and professional terminal utility suite that preserves the familiarity and composability of Unix commands while dramatically improving usability and visual experience.

### Core Principles:
- **FAST**: Sub-10ms startup overhead, streaming I/O without slurping large files, bounded concurrency, minimal memory allocations.
- **BEAUTIFUL**: Clean typography, semantic palettes, icon support, thoughtful information density, responsive layouts.
- **SIMPLE**: Sensible defaults, zero required configuration to be productive, straightforward flags.
- **POWERFUL**: Advanced inspection, depth filtering, structured queries, git-aware status.
- **INTERACTIVE**: Rich, responsive terminal interface for file browsing, fuzzy navigation, and contextual inspection.
- **COMPOSABLE**: Human mode when connected to an interactive TTY; clean, unformatted, deterministic plain text or JSON when piped or requested.
- **RELIABLE**: Defensive filesystem manipulation, atomic moves where possible, comprehensive automated test suites, race detection, cross-platform portability.

---

## 3. Scope & Command Architecture

### 3.1 In Scope (v1 Core Suite)

| Command | Category | Core Responsibilities |
|---|---|---|
| `nova ls` | Inspection | Directory & file listing; compact & long grid layouts; sorting by name, size, mtime; semantic colors; icons; human sizes; symlink targets; `--json` & `--plain` |
| `nova cat` | Viewer | Streaming file viewer; syntax highlighting; line numbers; paging; search; binary detection; invalid UTF-8 tolerance; memory bounded |
| `nova tree` | Inspection | Recursive visual directory hierarchy; depth limits; counts; size aggregation; icon trees; git-status decorators |
| `nova find` | Discovery | Fast filesystem search by name glob, file type, file size, modification timestamp; path predicates; `--json` & `--plain` |
| `nova stat` | Inspection | Structured metadata inspection (permissions, uid/gid, size, timestamps, inode/device, extended attributes where available) |
| `nova du` | Inspection | Disk usage aggregation; human-readable units; sort by size; cycle & symlink safe traversal |
| `nova cp` | Operation | File & directory copying; metadata preservation; recursive mode; overwrite controls (`-i`, `-n`, `-f`); streaming I/O; `--dry-run` |
| `nova mv` | Operation | File & directory rename/move; atomic operations where possible; overwrite protections; cross-device fallback; `--dry-run` |
| `nova rm` | Operation | Safe deletion; confirmation prompt on recursive or multiple targets; symlink safety; broad deletion safeguards (`/`, `~`, parent traps); `--dry-run` |
| `nova mkdir` | Operation | Directory creation with parent creation semantics (`-p`); permissions handling |
| `nova` / `nova interactive` | TUI | Signature interactive file navigator; keyboard-driven; fuzzy search; file preview pane; contextual shortcuts; non-blocking event loop |

### 3.2 Candidate / Deferred Commands (Post-v1 Roadmap)

The following commands are intentionally deferred to future evaluations and must not be implemented until v1 exit criteria are met:
- `nova pwd`, `nova touch`, `nova ln`
- `nova diff`, `nova grep`, `nova which`, `nova env`
- `nova inspect`, `nova preview`

### 3.3 Explicit Non-Goals

1. **Not a shell**: `nova` will not implement shell job control, scripting syntax, builtins, prompt management, or piping interpreters.
2. **Not a full Git client**: Git integration is strictly read-only contextual indicators (status badges, branch decorators) without cloning, staging, or committing.
3. **No daemon or background service**: No persistent background processes, system services, or automatic file indexing daemons.
4. **No telemetry or remote network requests**: Zero external network requests, zero telemetry, zero analytics tracking.
5. **No premature plugin engine**: No Lua/Wasm/dynamic plugin system in v1. Architectural modularity must accommodate extension later without over-engineering now.
6. **No gimmick AI features**: No LLM calls embedded inside core CLI utilities. Every feature must deliver direct, deterministic utility.

---

## 4. User Experience & Design System

### 4.1 Dual Operating Modes

1. **Human Mode** (default when stdout is a TTY):
   - Adaptive responsive grid layout based on detected terminal width (`$COLUMNS` / `TIOCGWINSZ`).
   - Semantic color rendering matching file types and statuses.
   - Nerd Font / Unicode file type icons.
   - Human-readable units (`B`, `KB`, `MB`, `GB`, `TB`) using binary (1024) or decimal (1000) flags.
2. **Machine Mode** (default when stdout is redirected/piped, or when `--plain`/`--json` is explicitly passed):
   - Plain Mode: Tab/newline-delimited, zero ANSI escape codes, zero decorative icons, deterministic sorting.
   - JSON Mode: Valid, schema-versioned, newline-delimited or structured JSON document for machine ingestion.

### 4.2 Semantic Theme System

No hardcoded ANSI escape sequences scattered throughout command packages. All styles flow from a centralized theme abstraction:
- Semantic Roles: `directory`, `regular_file`, `executable`, `symlink`, `broken_symlink`, `pipe_fifo`, `socket`, `device`, `archive`, `code`, `document`, `image`, `audio`, `video`, `hidden`, `success`, `warning`, `error`, `info`, `muted`, `selection`, `accent`.
- Built-in Themes: `default`, `minimal` (monochrome with bold/dim), `mono`, `nord`, `dracula`, `neon`.
- Automatic Fallback: Automatically downgrades truecolor (24-bit) -> ANSI 256 -> ANSI 16 -> Plain ASCII based on terminal capability detection and environment variables (`NO_COLOR`, `CLICOLOR`, `CLICOLOR_FORCE`, `TERM`).

### 4.3 Error Experience

Error output must be actionable and informative:
- Explicit indication of what failed, the exact file path involved, the OS/system reason, and a suggested resolution where applicable.
- Standard Unix exit codes:
  - `0`: Success.
  - `1`: Operational failure (e.g. file not found, permission denied, copy failed).
  - `2`: Command line usage error / invalid flags.

---

## 5. Performance Targets & Resource Budgets

| Metric | Target | Measurement Strategy |
|---|---|---|
| Startup overhead | < 10ms execution time for `nova --help` / `nova --version` | `benchstat`, hyperfine, Go benchmarks |
| Memory overhead | < 25MB peak RSS on standard directory operations | `pprof` heap profiles, runtime metrics |
| Concurrency model | Bounded worker pools (`runtime.NumCPU()`), zero goroutine leakage | Go benchmark stress tests, race detector |
| Large file handling | Streamed buffers (e.g. 32KB/64KB chunks); never slurp whole files | `nova cat` on 1GB+ files |
| Huge directories | Progressive/buffered streaming traversal for 100k+ entries | Traversal synthetic benchmarks |
| Shell execution | 0 calls to `sh -c` or `bash -c` for filesystem actions | Code audit / static analysis |

---

## 6. Safety & Security Requirements

1. **Destructive Operation Protection**:
   - `nova rm` refuses recursive deletion of root directory (`/`), home directory (`~`), or parent traversal traps (`../..`) without an explicit override flag (`--no-preserve-root`).
   - Prompt confirmation by default on destructive commands when executed in an interactive TTY, bypassed only with `-f`/`--force`.
   - Dry-run mode (`-n` / `--dry-run`) supported for `cp`, `mv`, `rm`.
2. **Path Sanitization & Injection Prevention**:
   - Absolute and relative paths cleanly resolved using standard filepath cleaners.
   - Filenames with terminal escape sequences, control characters, or leading dashes sanitized before rendering to prevent terminal hijacking.
3. **Symlink Safety**:
   - Avoid infinite loops on circular symlinks through cycle detection.
   - Distinguish broken symlinks clearly in visual outputs.
   - Operations that alter targets must never inadvertently follow untrusted symlinks unless explicitly flagged (`-L` / `--dereference`).
4. **Permission & TOCTOU Awareness**:
   - Atomic file operations where possible (`os.Rename`).
   - Handle permission denials gracefully without crashing or corrupting adjacent operations.

---

## 7. Supported Platforms

| Platform | Tier | Target Scope |
|---|---|---|
| Linux x86_64 (`amd64`) | Tier 1 | Fully supported & continuously validated across all phases |
| Linux aarch64 (`arm64`) | Tier 1 | Target for Phase 10 cross-compilation & test validation |
| macOS x86_64 & Apple Silicon (`arm64`) | Tier 2 | Best-effort POSIX parity until Phase 10 validation |
| Windows x86_64 & arm64 | Tier 2 | Best-effort; POSIX permission/symlink differences documented |

---

## 8. Package & System Architecture

```text
nova/
├── cmd/
│   └── nova/                 # Main application entry point
├── internal/
│   ├── cli/                  # Command router, flag parsing, context, root command
│   ├── commands/             # Individual command implementations
│   │   ├── ls/               # nova ls
│   │   ├── cat/              # nova cat
│   │   ├── tree/             # nova tree
│   │   ├── find/             # nova find
│   │   ├── stat/             # nova stat
│   │   ├── du/               # nova du
│   │   ├── cp/               # nova cp
│   │   ├── mv/               # nova mv
│   │   ├── rm/               # nova rm
│   │   └── mkdir/            # nova mkdir
│   ├── filesystem/           # Unified traversal engine, metadata extraction, safe ops
│   ├── terminal/             # TTY detection, width query, color level, Unicode check
│   ├── theme/                # Semantic roles, palette definitions, icon registry
│   ├── renderer/             # Columnar layout, text truncation, grid/table formatters
│   ├── output/               # Output mode controller (Human, Plain, JSON)
│   ├── interactive/          # TUI event loop, keyboard navigator, preview pane
│   ├── git/                  # Read-only Git status integration
│   └── platform/             # OS-specific syscall abstractions (attributes, permissions)
```

Architectural Invariants:
- Packages in `internal/commands/*` must never import each other; all shared logic belongs in `internal/filesystem`, `internal/renderer`, etc.
- Heavy dependencies must be strictly justified; standard library is prioritized.

---

## 9. Quality Gates & Exit Criteria

Every implementation phase must pass all mandatory quality gates:
1. Formatting: `gofmt -l .` reports no unformatted files.
2. Static Analysis: `go vet ./...` reports zero issues.
3. Test Suite: `go test ./...` passes 100% of unit and integration tests.
4. Race Detector: `go test -race ./...` passes without race conditions.
5. Compilation: `go build ./...` produces clean builds without warnings or errors.
6. Phase Exit Gate: Specific criteria outlined in `ROADMAP.md` must be satisfied.
7. Strict Stop: A formal phase report must be generated and awaiting user authorization before moving to the next phase.
