# Nova (`nova`)

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.27+-00ADD8?logo=go)](go.mod)
[![Status](https://img.shields.io/badge/Release-v1.4.0-success.svg)](CHANGELOG.md)
[![Dependencies](https://img.shields.io/badge/Dependencies-Zero-brightgreen.svg)](go.mod)
[![Architecture](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey.svg)](ROADMAP.md)

**Nova** is a modern, fast, beautiful, and interactive terminal utility suite in Go. It rebuilds the classic Unix core utilities — `ls`, `cat`, `tree`, `find`, `stat`, `du`, `cp`, `mv`, `rm`, `mkdir`, `grep`, alongside workspace maintenance (`clean`), compression (`archive`), live monitor (`watch`), system diagnostics (`sysinfo`), I/O benchmarks (`bench`), and an interactive dual-pane file explorer — into a unified, zero-dependency binary engineered for 2026 terminal workflows.

```text
FAST · BEAUTIFUL · SIMPLE · POWERFUL · INTERACTIVE · COMPOSABLE · RELIABLE
```

---

## Highlights

- ⚡ **Blazing Fast**: Single-digit millisecond startup, streaming I/O with bounded allocations, and up to **18.6 GB/s** cat throughput.
- 🎨 **Terminal-Aware Design**: Automatic adaptation to TrueColor, 256-color, 16-color, Unicode icons, and automatic clean ASCII fallback.
- 🔄 **Tri-Mode Output Contract**:
  - **Human Mode**: Colored, responsive layouts with icons for interactive terminal sessions.
  - **Plain Mode (`--plain`)**: Unformatted, deterministic, tab/newline-separated text for pipelines.
  - **JSON Mode (`--json`)**: Structured, schema-valid JSON for machine processing.
- 🔍 **Concurrent Search (`grep`)**: Multi-threaded regex/literal search engine with binary detection and neon highlights.
- 🧹 **Workspace Clean (`clean`)**: Detect and sweep OS junk (`.DS_Store`), editor backups (`*.swp`), broken symlinks, and empty dirs.
- 📦 **Pure Go Archiver (`archive`)**: Pack, unpack, and list ZIP and TAR.GZ archives with zip-slip security guards.
- 👀 **Live File Monitor (`watch`)**: Periodic change detector with event streaming and automatic task execution (`-e`).
- 📊 **Telemetry & Benchmarks**: Real-time system intelligence (`sysinfo`) and isolated disk I/O throughput tests (`bench`).
- 🌿 **Lightweight Git Awareness**: Instant `.git/HEAD` branch inspection and non-blocking status badges (`M`, `A`, `?`, `D`, `R`, `!`) with a strict 50ms timeout guard.
- 🛡️ **Safety by Default**: Destructive operations (`rm`, `mv`, `cp`) enforce root protection (`/`, volume roots), prevent self-descendant recursion, and support `-n` / `--dry-run`.
- 🖥️ **Interactive TUI**: Dual-pane file navigator with syntax-highlighted previews, command palette (`:`), directory bookmarks (`b`/`B`), real-time theme cycling (`t`), sorting modes (`s`), and hex dump inspector (`x`).
- 📦 **Zero External Runtime Dependencies**: Standard library only.

---

## Command Reference

| Command | Category | Description |
|---|---|---|
| `nova` / `nova interactive` | Explorer | Interactive dual-pane terminal file browser with live syntax previews, git diffs (`D`), command palette, metadata inspector, bookmarks, and fuzzy search |
| `nova ls` | Inspection | Modern directory listing with compact responsive grid, table (`-l`), sorting, and git badges |
| `nova cat` | Viewer | Streaming file viewer with syntax highlighting, line numbers, pagination, and hex dumps (`--hex`) |
| `nova grep` | Search | Concurrent multi-threaded regex/text content search with binary skipping and neon match highlights |
| `nova http` | Network | Full HTTP testing client with latency tracing waterfall (DNS, TCP, TLS, TTFB), headers, and syntax preview |
| `nova qr` | Sharing | Pure Go high-resolution terminal QR code generator using Unicode half-blocks for instant camera scanning |
| `nova git` | Productivity | Git repository health dashboard, staged/unstaged changes, ahead/behind counters, and commit graph |
| `nova hash` | Cryptography | Multi-algorithm checksum engine (SHA-256, SHA-512, SHA-1, MD5, CRC32) with verification (`-c`) and recursive tree hashing |
| `nova hex` | Inspection | Colorized canonical hex dump viewer with byte categorization, configurable grouping, offsets, and ASCII panel |
| `nova env` | Telemetry | Developer environment & secret auditor with automatic credential masking, filtering, and shell export generators |
| `nova clean` | Maintenance | Safe workspace hygiene tool detecting and removing OS junk, editor backups, test binaries, and empty dirs |
| `nova archive` | Utility | Pure Go ZIP and TAR.GZ compressor and extractor with zip-slip protection and compression ratio metrics |
| `nova watch` | Automation | Live filesystem monitor tracking file creations, modifications, and deletions with automated task execution |
| `nova sysinfo` | Diagnostics | System, hardware, Go memory, disk usage, terminal capabilities, and PATH diagnostic cards |
| `nova bench` | Benchmark | High-precision sequential read/write throughput, directory walking, and stat latency percentiles |
| `nova tree` | Inspection | Visual directory hierarchy with depth controls (`-L`), size rollups, icons, and git status |
| `nova find` | Discovery | Fast filesystem search by glob name (`--name`), type (`--type`), size, and modification timestamp |
| `nova stat` | Inspection | Rich structured file metadata cards, octal permissions, inode/device numbers, and timestamps |
| `nova du` | Inspection | Visual disk usage summaries with proportional capacity bars and depth limits |
| `nova cp` | Operation | Safe file/directory copying with metadata preservation, recursion guards, and dry-run mode |
| `nova mv` | Operation | Atomic moves (`os.Rename`) with cross-device fallback (`EXDEV`) and overwrite protections |
| `nova rm` | Operation | Defensive deletion with root protection (`ProtectRoot`) and safe symlink removal |
| `nova mkdir` | Operation | Directory creation with parent creation (`-p`) and custom permission modes (`-m`) |
| `nova which` | Discovery | Locate binary executables in `$PATH`, resolve symlinks, show file size and permissions |
| `nova touch` | Operation | Create empty files or update timestamps with parent directory auto-creation (`-p`) |
| `nova diff` | Inspection | Colorized unified file comparator with additions/deletions stats and change summaries |
| `nova completion` | Shell | Native autocompletion script generator for `bash`, `zsh`, and `fish` |

---

## Installation

### Pre-Compiled Binaries
Download the latest release tarball or zip for your operating system and architecture from the [Releases](https://github.com/ngaarch/nova/releases) page:

- **Linux**: `nova_1.6.0_linux_amd64.tar.gz` | `nova_1.6.0_linux_arm64.tar.gz`
- **macOS**: `nova_1.6.0_darwin_amd64.tar.gz` | `nova_1.6.0_darwin_arm64.tar.gz`
- **Windows**: `nova_1.6.0_windows_amd64.zip` | `nova_1.6.0_windows_arm64.zip`

Extract and place the `nova` binary into your system `PATH` (e.g., `/usr/local/bin`).

### Build from Source
```bash
git clone https://github.com/ngaarch/nova.git
cd nova
go build -o nova ./cmd/nova
sudo mv nova /usr/local/bin/
```

### Go Install
```bash
go install github.com/ngaarch/nova/cmd/nova@latest
```

---

## Usage Guide & Examples

### 1. Directory Listing (`nova ls`)
```bash
# Compact responsive grid
nova ls

# Long listing with git status column, permissions, and human sizes
nova ls -lh

# Sort by size or modification time
nova ls -l -S      # Sort by file size (descending)
nova ls -l -t -r   # Sort by time (oldest first)

# Machine-readable output for scripts
nova ls --plain
nova ls --json
```

### 2. File Viewer (`nova cat`)
```bash
# Syntax-highlighted file viewing
nova cat main.go

# Numbered lines with pagination
nova cat -n --pager README.md

# Binary inspection with hex dump preview
nova cat --hex /bin/ls
```

### 3. Visual Hierarchy (`nova tree`)
```bash
# Indented directory tree limited to 2 levels
nova tree -L 2

# Include file permissions and sizes
nova tree -p -h -L 2

# Output tree hierarchy as structured JSON
nova tree --json -L 2
```

### 4. Fast Discovery (`nova find`)
```bash
# Search by name pattern
nova find . --name "*.go"

# Filter by file type and size
nova find . --type f --size "+1M"

# Export search results as JSON
nova find . --name "*.md" --json
```

### 5. Detailed Status (`nova stat`)
```bash
nova stat README.md
nova stat --plain README.md
nova stat --json README.md
```

### 6. Disk Usage (`nova du`)
```bash
# Visual bar breakdown of directories
nova du -L 1 -h

# Sort by size (largest first)
nova du -S
```

### 7. Interactive Explorer (`nova` / `nova interactive`)
Run `nova` without arguments in an interactive terminal to enter the dual-pane navigator:

| Keybinding | Action |
|---|---|
| `j` / `↓` | Move selection down |
| `k` / `↑` | Move selection up |
| `Enter` / `l` | Open selected directory or preview |
| `Backspace` / `h` | Go up to parent directory |
| `/` | Real-time fuzzy filter query |
| `.` | Toggle visibility of hidden files |
| `g` / `G` | Jump to top / bottom |
| `PgUp` / `PgDn` | Scroll preview / page |
| `?` | Toggle modal help overlay |
| `q` / `Ctrl+C` | Exit interactive mode |

---

## Configuration & Themes

Nova reads optional configuration from `~/.config/nova/config.toml`:

```toml
theme = "dracula"
icons = "auto"
dirs_first = true

[ls]
all = false
human_readable = true

[cat]
syntax = true
line_numbers = false
```

### Available Themes
Specify `--theme=<name>` or set in configuration:
- `default`: One Dark inspired balanced modern palette.
- `nord`: Arctic, elegant north-bluish palette.
- `dracula`: Vibrant dark theme for contrast and readability.
- `neon`: High-energy cyber/fluorescent theme.
- `minimal`: Pure typography formatting without color escapes (attributes only).
- `mono`: Plain unstyled monochrome text.

Nova strictly obeys the [`NO_COLOR`](https://no-color.org) specification: setting `NO_COLOR=1` strips all ANSI escape codes.

---

## Benchmarks & Performance

Measured on baseline low-power hardware (Intel Celeron N3060 @ 1.60GHz, Go 1.27):

| Benchmark | Measurement | Performance Result |
|---|---|---|
| **Cat Streaming Throughput** | 1 MB File Stream | **18,649 MB/s** (~18.6 GB/s) |
| **Git Root Traversal** | Inside Git Repository | **0.039 ms** (39,825 ns/op) |
| **Git Root Traversal** | Outside Git Repository | **0.033 ms** (32,858 ns/op) |
| **Porcelain Status Parser** | 10,000 files stream | **32.6 ms** |
| **Tree Traversal** | Deep nested hierarchy | **3.8 ms** |
| **Ls Large Directory** | 1,000 files | **45.8 ms** |

---

## Contributing

Contributions are welcome! Please review [CONTRIBUTING.md](CONTRIBUTING.md) for architectural rules, testing requirements, and development guidelines.

---

## License

MIT License — see [LICENSE](LICENSE) for details.
