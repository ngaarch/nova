# Changelog

All notable changes to `nova` are documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.4.0] - 2026-09-27

### Added
- **Workspace Hygiene & Junk Cleaner (`nova clean` / `nova tidy` / `nova sweep`)**:
  - Automatically detect and sweep OS junk (`.DS_Store`, `Thumbs.db`, `desktop.ini`), editor backup files (`*.swp`, `*~`, `*.tmp`), test binaries (`*.test`), broken symlinks, and empty directories (`--empty-dirs`).
  - Safe by default: runs in `--dry-run` mode by default, requiring `-f` / `--force` for actual removal.
  - Reclaimed space tracking: displays exact byte savings and percentage.
  - Machine formats: `--plain` and structured `--json`.

- **Pure Go Compression & Extraction Suite (`nova archive` / `nova pack` / `nova zip`)**:
  - Pure Go standard library ZIP and TAR.GZ compression and extraction without external tools (`archive/zip`, `archive/tar`, `compress/gzip`).
  - Actions: `pack` (`-o <archive>`), `unpack` (`-C <dest>`), `list`.
  - Built-in Zip-Slip security: rigorously validates relative paths against directory traversal attacks.
  - Compression metrics: `nova archive list` reports uncompressed size, compressed size, and compression ratio per file.

- **Live Filesystem Monitor & Auto-Reloader (`nova watch` / `nova monitor`)**:
  - Cross-platform periodic polling file monitor with modification, creation, and deletion detection.
  - Automated command runner (`-e "<cmd>"`), debouncing (`-d`), screen clearing (`-c`), and iteration limits (`-n`).
  - Glowing live status banner with colored event badges (`[CREATE]`, `[MODIFY]`, `[DELETE]`).
  - Full machine streaming with newline-delimited `--plain` and JSON lines `--json`.

- **Interactive TUI Supercharging (`nova interactive`)**:
  - **Command Palette (`:` key)**: Quick command bar for executing `:theme <name>`, `:sort <mode>`, `:mkdir <name>`, `:touch <name>`, `:reload`, `:q`.
  - **Bookmarks & Fast Jump (`b` and `B` keys)**: Bookmark favourite directories (`b`) and cycle jump (`B`) directly across projects!
  - **Bookmark Status Counter**: Real-time `[N bm]` badge in status bar.

---

## [1.3.0] - 2026-09-27

### Added
- **Concurrent Content Searcher (`nova grep` / `nova search`)**:
  - High-performance multi-threaded search engine with worker pools and regex support.
  - Automatic binary file skipping with null-byte detection.
  - Flags: `-i` (case-insensitive), `-n` (line numbers), `-c` (count only), `-l` (files with matches), `-C`/`-B`/`-A` (context lines), `--ext` (extension filtering), `--hidden` (include dotfiles), `-m` (max matches).
  - Vibrant human presentation: neon matching substring highlights, line gutters, file header icons, and scan timing/file summary badges.
  - Machine-ready `--plain` (`file:line:content`) and structured `--json` streaming.

- **System Diagnostics & Environment Intelligence (`nova sysinfo` / `nova sys` / `nova info`)**:
  - Hardware, Host, and Architecture breakdown (Hostname, OS, Arch, CPU cores, Go runtime version).
  - Runtime Memory & Storage metrics: Active Goroutines, Alloc/Sys memory with gradient progress bars, GC statistics, and filesystem disk usage.
  - Workspace Context: Git repository status (branch, clean/dirty badge) and PATH integrity analysis (identifying missing directories).
  - Nova Environment: Terminal dimensions, TTY status, Color Profile, active theme, and icon modes.
  - Full machine output with `--plain` key=value and `--json` serialization.

- **Storage & Filesystem I/O Benchmarking (`nova bench` / `nova benchmark`)**:
  - Sequential write and read throughput measurement (MB/s) using isolated scratch buffers.
  - Microsecond-precision metadata stat latency distribution (min, p50, p90, p95, p99, max) with Unicode sparkline graphs.
  - Directory walk rate testing (files/sec).
  - Completely safe execution: runs exclusively in auto-cleaned scratch directories.
  - Structured `--json` and tab-separated `--plain` outputs for automated CI/CD performance tracking.

- **Interactive TUI Supercharging (`nova interactive`)**:
  - **Live Theme Switcher (`t` key)**: Cycle through all 11 themes (`default`, `nord`, `dracula`, `neon`, `cyberpunk`, `synthwave`, `tokyo-night`, `catppuccin`, `gruvbox`, `minimal`, `mono`) on the fly with instant live recoloring!
  - **Sorting Mode Toggle (`s` key)**: Cycle sorting orders instantly: Name (A-Z) -> Size (largest first) -> ModTime (newest first) -> Extension!
  - **Hex Dump Inspector (`x` key)**: Toggle raw hexadecimal inspection with offset, byte columns, and ASCII preview.
  - Enhanced status bar and header badges displaying current sorting order and active theme.

---

## [1.2.0] - 2026-09-27

### Added
- **Executable Locator (`nova which`)**:
  - Locate binaries in `$PATH` with symlink target resolution, file sizes, and permission strings.
  - Multi-match support (`-a` / `--all`) to find all instances across `$PATH` directories.
  - Silent check mode (`-s` / `--silent`) with standard exit codes for script testing.
  - Structured `--json` and clean `--plain` pipeline output.

- **File Creation & Timestamp Manager (`nova touch`)**:
  - Create empty files and update access/modification timestamps.
  - Parent directory auto-creation (`-p` / `--parents`) for deep nested file creation.
  - Custom date parsing (`-d` / `--date`) supporting RFC3339, YYYY-MM-DD, and HH:MM:SS.
  - Reference file timestamp copying (`-r` / `--reference`).
  - No-create safeguard flag (`-c` / `--no-create`).

- **Colorized File Comparator (`nova diff`)**:
  - Pure Go unified diff engine with LCS difference calculation and configurable context lines (`-u`).
  - Human mode with syntax-aware line numbers, green addition (`+`) and red deletion (`-`) indicators, and change summary cards.
  - Brief difference reporting (`-q` / `--brief`) and whitespace/case ignore flags (`-w`, `-i`).
  - Safe binary difference detection and structured JSON output.

- **Shell Autocompletion Generator (`nova completion`)**:
  - Native autocompletion scripts for `bash`, `zsh`, and `fish`.
  - Autocomplete support for all 15 subcommands, global flags, and 11 designer themes.

- **Interactive TUI Navigator Supercharge (`nova interactive`)**:
  - New file creation shortcut (`n`) with inline filename entry and immediate reload.
  - New folder creation shortcut (`N`) with parent directory handling.
  - Preview pane toggle (`p`) to collapse/expand into full-width file browser mode.
  - Multi-select (`Space`), inline rename (`r`), safe delete confirmation (`d`), and copy path toasts (`c`).
  - Breadcrumb path navigation and real-time fuzzy search highlighting.

- **5 Designer Themes & TrueColor Gradients**:
  - `cyberpunk`, `synthwave`, `tokyo-night`, `catppuccin`, and `gruvbox`.
  - TrueColor RGB gradient generator (`InterpolateRGB`, `FormatGradient`).

- **Visual Effects & Widgets Engine (`internal/renderer`)**:
  - Braille animated spinners (`SpinnerFrames`, `SpinnerFrame`).
  - Dynamic Unicode progress bars (`RenderProgressBar`).
  - 8-level sparklines generator (`RenderSparkline`).
  - Relative human time converter (`FormatRelativeTime`).
  - Rounded-corner card borders (`╭─╮`, `╰─╯`).

---

## [1.0.0] - 2026-09-27

### Added
- **Core CLI Engine & Tri-Mode Output (`internal/cli`, `internal/output`)**:
  - Unified command router with standard exit codes (`0` success, `1` operational error, `2` usage error).
  - Tri-mode output architecture: Human mode (responsive, colors, icons), Plain mode (`--plain` for scripts and pipe safety), and JSON mode (`--json` for structured data pipelines).
  - Pipe safety: automatic ANSI escape stripping when stdout is redirected or piped.
  - Configuration loader (`~/.config/nova/config.toml`) with deterministic flag precedence.
  - Standardized, actionable error diagnostics (`ExitError`) with causes, paths, and resolution hints.

- **Terminal Engine & Design System (`internal/terminal`, `internal/theme`, `internal/renderer`)**:
  - Detection of terminal dimensions, ANSI color levels (TrueColor, 256-color, 16-color, none), Unicode support, and `NO_COLOR=1` / `CLICOLOR_FORCE`.
  - 6 built-in themes: `default`, `minimal`, `mono`, `nord`, `dracula`, and `neon`.
  - Semantic role styling system (`RoleDirectory`, `RoleExecutable`, `RoleSymlink`, `RoleError`, etc.).
  - Icon registry with automatic ASCII fallback for non-Unicode environments.
  - Visual cell width computation (`VisibleWidth`, `RuneWidth`) and ANSI-aware string truncation (`Truncate`).
  - Responsive `Grid` and `Table` layout engines.

- **`nova ls` (Flagship Directory Listing)**:
  - POSIX flag compatibility (`-a`, `-A`, `-l`, `-h`, `-r`, `-t`, `-S`, `-R`, `-d`).
  - Compact responsive columnar grid layout.
  - Long listing table format (`-l`) with permissions, owner, group, humanized file sizes, timestamps, and symlink targets.
  - Natural sort with directory prioritization (`--dirs-first`).

- **`nova cat` (File Viewer & Hex Dump Preview)**:
  - Streaming line-by-line processor with bounded buffer allocations.
  - Syntax highlighter for Go, Python, JavaScript/TypeScript, JSON, YAML, TOML, Markdown, Shell, and C/C++.
  - Binary detection with automatic hex dump mode (`--hex`).
  - Terminal height pagination support (`--pager`).
  - Line numbering (`-n`, `-b`), non-printable byte formatting (`-v`, `-A`, `-E`, `-T`), and blank line squeezing (`-s`).

- **Filesystem Inspection Suite (`nova tree`, `nova find`, `nova stat`, `nova du`)**:
  - `nova tree`: Indented visual hierarchy, depth limits (`-L`), dir/file counts, size aggregations, and cycle/symlink recursion detection.
  - `nova find`: Fast directory search with predicates (`--name`, `--type`, `--size`, `--mtime`, `--empty`, `--max-depth`).
  - `nova stat`: Structured metadata cards, octal permissions, UID/GID, block allocation counts, timestamps, inode/device numbers.
  - `nova du`: Visual disk usage aggregation with proportional terminal capacity bars and depth limiting.

- **Safe Filesystem Operations (`nova cp`, `nova mv`, `nova rm`, `nova mkdir`)**:
  - Atomic moves (`os.Rename`) with graceful `EXDEV` cross-device fallback.
  - Root protection (`ProtectRoot`) preventing deletion of `/`, volume roots, or current directories.
  - Recursion guards (`ErrDestInsideSource`) preventing copying or moving directories into themselves.
  - Interactive confirmation prompts (`-i`), no-clobber controls (`-n`), and dry-run mode (`--dry-run`).
  - Preserved timestamps and POSIX file permissions during copy.

- **Interactive UX & TUI (`nova interactive` / auto-launch)**:
  - Dual-pane responsive file navigator with instant preview pane.
  - Real-time fuzzy filtering (`/`) and modal help (`?`).
  - Full keyboard navigation (`j`/`k`, `h`/`l`, `Enter`, `g`/`G`, `PgUp`/`PgDn`).
  - Dynamic `SIGWINCH` resize handling with thread-safe `sync.RWMutex` synchronization.
  - Syntax-highlighted and hex-dumped previews for active files.

- **Git Awareness (`internal/git`)**:
  - Lightweight `.git/HEAD` reader with zero subprocess execution for branch identification.
  - Bounded-execution query (`git status --porcelain=v1 -z`) with strict 50ms timeout guard.
  - Contextual status badges (`M`, `A`, `?`, `D`, `R`, `!`) in `nova ls -l`, `nova tree`, `nova stat`, and interactive TUI.
  - Automatic directory status aggregation from untracked/modified children.

- **Release Engineering & Distribution**:
  - Multi-architecture reproducible builds: Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64, arm64).
  - SHA-256 artifact verification manifests (`SHA256SUMS.txt`).
  - Compile-time version, commit hash, and build timestamp injection via ldflags.
