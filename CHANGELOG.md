# Changelog

All notable changes to `nova` are documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
