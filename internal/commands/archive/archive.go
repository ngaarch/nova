package archive

import (
	"errors"
	"strings"

	"nova/internal/command"
	"nova/internal/output"
)

// Command returns the registered Command instance for the archive subcommand.
func Command() *command.Command {
	return &command.Command{
		Name:        "archive",
		Aliases:     []string{"pack", "zip", "tar"},
		Summary:     "Create, extract, and inspect ZIP and TAR archives with pure Go engine",
		Usage:       "nova archive <pack|unpack|list> [flags] <archive> [files...]",
		Description: "High-performance compression and extraction suite with zip-slip safety.",
		Phase:       15,
		Run:         Run,
	}
}

// Run executes the archive command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)
	if opts.Plain {
		ctx.Printer.Mode = output.ModePlain
	} else if opts.JSON {
		ctx.Printer.Mode = output.ModeJSON
	}

	switch opts.Mode {
	case ModePack:
		if opts.Output == "" {
			return command.NewUsageError("missing archive destination (-o)", "Usage: nova archive pack -o <archive.zip> <files...>")
		}
		if len(opts.Targets) == 0 {
			return command.NewUsageError("no input files specified to archive", "Usage: nova archive pack -o <archive.zip> <files...>")
		}

		var count int
		var size int64
		var err error

		if strings.HasSuffix(opts.Output, ".tar.gz") || strings.HasSuffix(opts.Output, ".tgz") {
			count, size, err = PackTarGz(opts.Output, opts.Targets)
		} else {
			count, size, err = PackZip(opts.Output, opts.Targets)
		}
		if err != nil {
			return err
		}

		return RenderPackResult(ctx, opts.Output, count, size)

	case ModeUnpack:
		if opts.ArchiveFile == "" {
			return command.NewUsageError("missing archive to unpack", "Usage: nova archive unpack <archive.zip> [-C <dest>]")
		}

		count, err := UnpackZip(opts.ArchiveFile, opts.Destination)
		if err != nil {
			return err
		}

		return RenderUnpackResult(ctx, opts.ArchiveFile, opts.Destination, count)

	case ModeList:
		if opts.ArchiveFile == "" {
			return command.NewUsageError("missing archive to inspect", "Usage: nova archive list <archive.zip>")
		}

		entries, err := ListZip(opts.ArchiveFile)
		if err != nil {
			return err
		}

		return RenderListResult(ctx, opts.ArchiveFile, entries)

	default:
		return errors.New("unknown archive mode: must be pack, unpack, or list")
	}
}
