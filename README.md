# nova

A modern, fast, beautiful, and interactive terminal utility suite written in Go.

`nova` rebuilds the classic Unix file utilities — `ls`, `cat`, `tree`, `find`, `stat`, `du`, `cp`, `mv`, `rm`, `mkdir`, and interactive directory exploration — around four core principles modern terminal tools demand:

- **Terminal-aware rendering.** Output adapts intelligently to terminal width, color profile (truecolor, 256-color, 16-color), Unicode capability, and whether standard output is connected to an interactive TTY or a pipeline/redirection.
- **Two honest output modes.** Human mode is readable, beautifully styled, and responsive. Machine mode (`--json`, `--plain`) is deterministic, unformatted, and pipeline-safe. Visual polish never breaks a pipeline.
- **Safety by default.** Destructive operations (`rm`, `mv`, `cp`) protect against accidental broad deletion, handle symlinks deliberately, and provide clear confirmations and dry-run modes (`-n`/`--dry-run`).
- **High-performance systems engineering.** Fast startup (single-digit millisecond baseline), streaming I/O for huge files, bounded concurrency, minimal allocations, and zero shell spawning.

Status: **pre-alpha, under active phased development.** See [ROADMAP.md](ROADMAP.md).

## Requirements

- Go 1.27+ (build from source)
- Linux (x86_64, arm64) is the primary tier-1 target, with macOS and Windows supported as best-effort targets until Phase 10 validation.

## Build

```bash
go build -o nova ./cmd/nova
```

## Quick Start

```bash
nova --help
nova --version
```

Individual commands ship phase by phase according to [ROADMAP.md](ROADMAP.md). Unimplemented commands fail loudly with explicit status messages; no functionality is silently stubbed or faked.

## Design Principles

```text
FAST · BEAUTIFUL · SIMPLE · POWERFUL · INTERACTIVE · COMPOSABLE · RELIABLE
```

- **Unix philosophy.** Each utility does one thing well, composes seamlessly with standard tools via pipes, and obeys standard exit codes (`0` success, `1` operational error, `2` syntax error).
- **Performance as a feature.** Fast startup, bounded concurrency, streaming I/O, no unnecessary subprocesses or syscalls. All performance claims are grounded in repeatable benchmarks.
- **Graceful degradation.** No blind assumption of truecolor, UTF-8/Nerd Font icons, mouse support, or interactive TTY. Unsupported terminal features fall back to clean ASCII and plain text automatically.
- **Accessible by construction.** Color is never the sole information carrier; every visual indicator has a plain-text equivalent.

## Documentation

| Document | Purpose |
|---|---|
| [PRD.md](PRD.md) | Product requirements, scope, non-goals, architecture, quality gates |
| [ROADMAP.md](ROADMAP.md) | Sequential phased roadmap with strict stop exit gates |
| [LICENSE](LICENSE) | MIT License |

## License

MIT — see [LICENSE](LICENSE).
