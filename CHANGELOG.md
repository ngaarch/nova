# Changelog

All notable changes to `nova` are documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2026-09-28

### Milestone 2.0 Generation Release

### Added
- **Developer Secret Scanner & Shannon Entropy Detector (`nova scan` / `nova audit` / `nova secret`) [Phase 30]**:
  - High-performance security scanner and credential auditor detecting leaked secrets, private keys, API credentials, and high-entropy strings across codebases.
  - Multi-pattern signature engine detecting AWS access keys, GitHub personal access tokens, OpenAI/Anthropic/HuggingFace API keys, Google Cloud & Slack tokens, Stripe keys, and RSA/OpenSSH/PGP private keys.
  - Shannon Entropy calculator ($H = -\sum p_i \log_2 p_i$) identifying suspicious random password/token hashes with configurable entropy threshold (`-e` / `--entropy`, default 4.5 bits/byte).
  - Automatic credential masking (`AKIA****************EXAMPLE`) ensuring raw credentials are never leaked in terminal logs or stdout.
  - Rule filtering (`-p` / `--rule`), directory ignoring (`-i` / `--ignore` for `.git`, `node_modules`, `vendor`), deterministic tab-delimited `--plain`, and structured audit `--json`.

- **Multi-Core Hardware Stress & Benchmark Suite (`nova stress` / `nova cpu` / `nova burn` / `nova benchmark`) [Phase 31]**:
  - Multi-threaded CPU, memory, and hashing stress suite testing hardware stability, thread contention, and maximum throughput.
  - Three specialized stress engines:
    - **SHA-256**: High-throughput cryptographic hashing pipeline measuring hashes/sec and data processing speed (MB/s).
    - **Floating Math**: Heavy algebraic and trigonometric floating-point operations benchmark measuring MFLOPS.
    - **Memory Bandwidth**: Multi-threaded 64KB block allocation and memory transfer benchmark measuring RAM throughput (GB/s).
  - Configurable execution duration (`-d` / `--duration`, default 3s) and thread concurrency (`-t` / `--threads`, default runtime.NumCPU()).
  - Pipeline-friendly tab-separated `--plain` and structured `--json` metrics.

- **Terminal Color Palette, Contrast Studio & WCAG Inspector (`nova color` / `nova palette` / `nova colors` / `nova contrast`) [Phase 32]**:
  - Comprehensive terminal color visualization and accessibility suite.
  - 16 ANSI Standard color palette with standard and high-intensity bright variant blocks.
  - 256-color xterm grid rendering the 6x6x6 RGB color cube and 24-step grayscale ramp.
  - 24-bit TrueColor spectrum generator producing smooth full-spectrum gradient bars.
  - WCAG 2.1 Contrast Ratio Inspector computing relative luminance ($L$) and contrast ratios ($(L_1 + 0.05)/(L_2 + 0.05)$) with compliance levels (`AAA`, `AA`, `AA Large`, `Fail`) and live sample text previews.
  - Custom foreground and background color testing (`--fg`, `--bg`) supporting 6-digit and 3-digit hex strings (`#RRGGBB`).
  - Pipeline-friendly `--plain` and structured `--json` reports.

- **UI/UX 2.0 Supercharging & Modernization**:
  - **Categorized Help Catalog (`nova --help`)**: Grouped subcommands into clear, functional sections: *Files & Navigation*, *Network & Security*, *System & Diagnostics*, and *Developer & Utilities*, with visual icons and descriptions.
  - **Fluorescent Neon Filter Highlighting**: Interactive TUI `/filter` search now highlights matched query characters in high-contrast fluorescent neon (cyan/magenta) against standard filenames.
  - **Modernized Status Bar Pill Design**: Redesigned TUI bottom status bar with sleek status pills (`● READY`, `🔍 FILTER`), clean column dividers, and enhanced selection/bookmark badges.

---

## [1.9.0] - 2026-09-28

### Added
- **Terminal Markdown Document Renderer & Reader (`nova md` / `nova doc` / `nova view`)**:
  - Full terminal Markdown rendering engine supporting styled headings (`# H1` banner, `## H2` accent, `### H3`), blockquotes (`│ `), bulleted/numbered lists, and interactive todo checkboxes (`[ ]` / `[✓]`).
  - Table formatter with clean Unicode box borders (`┌─┬─┐`) and column width auto-calculation.
  - Code blocks with syntax highlighting via `cat/syntax` and framed border boxes.
  - Table of Contents generator (`-t` / `--toc`) extracting document headings and hierarchy tree.
  - Text reflow line-wrapping (`-w` / `--width`), interactive terminal pager (`-p` / `--pager`), raw unstyled `--plain`, and structured AST `--json`.

- **Developer Notes, Scratchpad & TODO Checklist Tracker (`nova note` / `nova notes` / `nova todo`)**:
  - Fast developer scratchpad and markdown notes manager storing notes locally (`~/.config/nova/notes/` or `.nova/notes/`).
  - Automatic task extraction from `- [ ]` and `- [x]` markdown checklists with overall progress calculation and progress bars (`[██████░░░░] 60%`).
  - Interactive task toggler (`nova note toggle <id> <task#>`) allowing check/uncheck directly from terminal without opening an editor.
  - Note search (`nova note search <query>`), tags filtering, full note viewer, and pipeline `--plain` / structured `--json`.

- **SSL/TLS Certificate Validator & JWT Token Decoder (`nova cert` / `nova ssl` / `nova tls` / `nova jwt`)**:
  - Remote TLS host inspector connecting and extracting peer certificate chains, SNI, negotiated TLS version (TLS 1.2, TLS 1.3), cipher suites, Subject Alternative Names (SANs), serial numbers, and key algorithms (RSA, ECDSA).
  - Expiration countdown with status badges (`Valid`, `Expiring Soon < 30d`, `Expired`).
  - Local certificate inspector decoding `.crt`, `.pem`, `.cer` files with PEM block unwrapping and DER fallback.
  - High-performance JSON Web Token (JWT) decoder (`nova cert jwt <token>`) splitting header, payload claims, and signature, with human-readable timestamps for `exp`, `iat`, `nbf`.
  - Machine-friendly `--plain` tab-delimited streams and schema-valid `--json`.

- **Interactive TUI Supercharging (TUI v1.9)**:
  - **Formatted Markdown Previews**: Inspecting `.md` and `.markdown` files now renders rich typography, styled headings, list markers, and status checkboxes directly in the dual-pane preview.
  - **Quick Bookmark Navigation (`1-9`)**: Pressing numeric keys `1` through `9` in normal navigation mode instantly jumps to corresponding bookmarked directories.
  - **Command Palette Expansion**: Added palette tips for `:md`, `:note`, `:todo`, `:cert`.

---

## [1.8.0] - 2026-09-28

### Added
- **Scientific, Programmer & Byte-Unit Terminal Calculator (`nova calc` / `nova expr` / `nova math` / `nova eval`)**:
  - Full algebraic, scientific, bitwise, and byte-unit expression evaluator with strict recursive descent precedence parser.
  - Multi-base numeric conversions: Decimal, Hexadecimal (`0x...`), Binary (`0b...`), Octal (`0o...`), and human byte capacities (`B`, `KB`, `MB`, `GB`, `TB`, `PB`, `KiB`, `MiB`, `GiB`, `TiB`).
  - Bitwise operations (`<<`, `>>`, `&`, `|`, `^`, `~`) and math functions (`sqrt`, `abs`, `round`, `floor`, `ceil`, `sin`, `cos`, `tan`, `log`, `log2`, `log10`, `pow`, `fact`) with mathematical constants (`pi`, `e`, `tau`, `phi`).
  - Visual dashboard with structured cards, precision formatting (`-p` / `--precision`), and raw numeric pipeline `--plain` / structured `--json`.

- **Shell History Intelligence & Command Analytics (`nova history` / `nova hist` / `nova analytics`)**:
  - Multi-shell parser engine supporting Bash (`~/.bash_history`), Zsh extended history (`: <timestamp>:0;<cmd>` in `~/.zsh_history`), and Fish shell (`~/.local/share/fish/fish_history`).
  - Intelligent command distribution analytics with horizontal ANSI bar charts, percentages, and execution counters.
  - Productivity categorization engine (Git, Development, System, Containers, Network, Editors, Shell utilities).
  - Search and filter queries (`-q` / `--query`), top command limits (`-n` / `--top`), custom history path (`-f` / `--file`), and pipeline `--plain` / structured `--json`.

- **Zero-Config Static Web Server & Mobile QR Pairing (`nova serve` / `nova server` / `nova httpd`)**:
  - Instant zero-configuration local HTTP static file server with automatic port conflict failover (`--port 8080`, auto-increments up to 100 ports if occupied).
  - Single Page Application (SPA) fallback routing mode (`--spa`) forwarding unhandled requests to `index.html`.
  - Terminal QR code generation rendering a high-contrast half-block QR code for instant mobile device camera pairing over local Wi-Fi / LAN.
  - Real-time colorized HTTP request logging table with status codes (2xx, 3xx, 4xx, 5xx), method badges, latency timing, client IP, and transfer sizes.
  - CORS header injection (`--cors`) and one-shot testing probe mode (`--once`).

- **Interactive TUI Supercharging (TUI v1.8)**:
  - **Archive Content Preview**: Automatic inspection of `.zip` and `.jar` archives in the file preview pane, showing file listings, compressed/uncompressed sizes, and compression savings ratio.
  - **Inline Command Palette Calculator (`:calc <expr>`)**: Compute equations, hex, binary, and byte conversions directly within the TUI without exiting.
  - **Command Palette Tips**: Added `:history` and `:serve` shortcut guidance in command palette.

---

## [1.7.0] - 2026-09-28

### Added
- **System Resource Telemetry & Process Monitor (`nova top` / `nova proc` / `nova ps` / `nova monitor`)**:
  - Real-time CPU, memory, load average (1m, 5m, 15m), and system uptime inspection.
  - CPU & Memory progress bars with graduated color ramps (green -> yellow -> red).
  - Detailed process table: PID, User, State (Running `R`, Sleeping `S`, Disk `D`, Zombie `Z`), CPU%, Mem%, RSS, Threads, Command.
  - Process filtering (`-f` / `--filter`), sorting (`-s cpu|mem|pid|name`), and output limits (`-n`).
  - Process termination & signal delivery (`-k` / `--kill <pid>`, `--signal <num>`).
  - Continuous live refresh mode (`-w` / `--live`) and deterministic `--plain` / `--json` streams.

- **Network Latency Telemetry & Port Scanner (`nova net` / `nova ping` / `nova latency` / `nova portscan`)**:
  - High-precision TCP ping with microsecond resolution, visual latency sparkline (` ▂▃▅▆▇█`), packet loss percentage, and min/avg/max/jitter metrics.
  - Fast concurrent port scanner (`nova net scan <host> -p 22,80,443,...`) with service name heuristics and open/closed/timeout status.
  - DNS resolution inspector (`nova net dns <domain>`) querying A, AAAA, CNAME, MX, and TXT records.
  - Automation friendly with `--plain` and structured `--json`.

- **Smart Task Runner & Workspace Automator (`nova run` / `nova task` / `nova exec`)**:
  - Automatic workspace task discovery across `package.json` (npm/pnpm/yarn/bun scripts), `Makefile` (targets), `go.mod` (Go toolchain tasks), `Cargo.toml` (Cargo tasks), and `Taskfile.yml`.
  - Visual task card grid displaying available tasks, detected runtimes, commands, and descriptions.
  - Direct execution with live task launch banner, execution duration timers, and exit status indicators.
  - Automatic parent workspace root traversal and zero-shell security compliance.

- **Interactive TUI Supercharging (TUI v1.7)**:
  - **Fuzzy File Finder Overlay (`f` or `Ctrl+P`)**: Fast recursive fuzzy search modal with instant character match highlighting and direct jump to matching items.
  - **Visual Bookmark Markers (`🔖`)**: Bookmarked files and directories now display prominent visual bookmark flags in the explorer tree.
  - **TrueColor Linear Gradient Header**: Dynamic 24-bit RGB linear gradient title banner.
  - **Command Palette Expansion**: Added `:find`, `:bookmark`, `:jump`, `:top`, `:net`, `:run`.

- **Theme Engine Expansion**:
  - Added built-in `solarized` (Solarized Dark) and `rose-pine` (Rosé Pine) themes.
  - Added RGB linear gradient text interpolation (`RenderGradientText`, `InterpolateRGB`).

---

## [1.6.0] - 2026-09-28

### Added
- **HTTP Client & Latency Telemetry (`nova http` / `nova fetch` / `nova curl`)**:
  - Full HTTP request suite supporting `GET`, `POST`, `PUT`, `DELETE`, `HEAD`, `PATCH` (`-X`).
  - Network latency tracing waterfall: DNS lookup, TCP connect, TLS handshake, TTFB (server processing), and content transfer times.
  - Formatted and syntax-highlighted responses with pretty JSON formatting.
  - Custom headers (`-H`), request body payload (`-d`), file upload (`--data-file`), output to file (`-o`).
  - Headless automation with `--plain` and structured `--json`.

- **Pure Go Terminal QR Code Generator (`nova qr` / `nova qrcode`)**:
  - Encodes URLs, text, Wi-Fi configuration strings into high-resolution Unicode half-blocks (`▀`, `▄`, `█`, ` `) for instant phone camera scanning off the screen.
  - Multi-version automatic matrix expansion with Reed-Solomon Error Correction Level M.
  - File input (`-f`), inversion mode (`-i`), quiet zone control (`-q`), ASCII `--plain`, and boolean 2D array `--json`.

- **Git Productivity Dashboard (`nova git` / `nova gstatus` / `nova glog`)**:
  - Repository health dashboard: branch name, tracking branch, ahead/behind commit counter.
  - Clear categorization cards: Staged (`+`), Unstaged (`~`), Untracked (`?`), and clean state indicator.
  - Compact commit history graph log (`nova git log -n 10`).
  - Branch overview (`nova git branch`).

- **Interactive TUI Supercharging (`nova interactive`)**:
  - **Inline Git Diff Preview (`D` key)**: Instantly preview uncommitted Git changes with colorized unified diff directly in the right preview pane!
  - **Real-Time Git Status in Header**: Displays current branch and modified file count dynamically.
  - **Command Palette Expansion**: `:diff` command added to palette.

---

## [1.5.0] - 2026-09-28

### Added
- **Cryptographic & Integrity Checksum Suite (`nova hash` / `nova checksum` / `nova digest`)**:
  - Multi-algorithm support: SHA-256 (default), SHA-512, SHA-1, MD5, CRC32.
  - Streaming I/O with constant memory footprint for gigabyte-scale files.
  - Verification mode (`-c` / `--check <file>`): validates standard checksum files (compatible with `sha256sum -c`), reporting status badges (`[PASS]`, `[FAILED]`, `[MISSING]`).
  - Recursive directory tree hashing (`-r` / `--recursive`).
  - Formats: Card table with algorithm pill, standard UNIX shasum `--plain`, and structured `--json`.

- **Modern Colorized Hex Dump Inspector (`nova hex` / `nova hexdump` / `nova dump`)**:
  - Byte categorization in terminal: null bytes (dim gray), printable ASCII (bright green/cyan), control characters (magenta), high bytes (amber).
  - Offset column in hex, configurable byte grouping (`-g`), columns per row (`-c`), byte offsets (`-s` / `--skip`), length limit (`-n` / `--length`).
  - Integrated ASCII preview panel with dot fallback for unprintable characters.
  - Stream-friendly: reads seamlessly from stdin or files.
  - Formats: canonical plain hex dump and structured `--json`.

- **Developer Environment & Secret Auditor (`nova env` / `nova environ`)**:
  - Audits environment variables with intelligent secret detection & masking (`API_KEY`, `TOKEN`, `PASSWORD`, `SECRET`, `AUTH`, `PRIVATE`, `CREDENTIAL`).
  - Filter & search by pattern (`-f` / `--filter`).
  - Categorization grouping: Development Runtimes, System & Shell, Cloud & DevOps, General.
  - Reveal mode (`-s` / `--show-secrets`).
  - Shell export generator: `--export=sh` and `--export=fish`.
  - Formats: Human dashboard, `KEY=VAL` `--plain`, and structured `--json`.

- **Interactive TUI Supercharging (`nova interactive`)**:
  - **File Metadata Inspector (`i` key)**: Floating modal window displaying deep file metadata (full path, size in bytes & human units, octal permissions, mode string, timestamps, symlink target).
  - **Quick SHA-256 Checksum (`#` key)**: Instantly computes the cryptographic hash of the selected file directly in the interactive navigator.
  - **Command Palette Expansion**: `:inspect`, `:hash`, `:hex` commands now accessible via the palette.

---

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
