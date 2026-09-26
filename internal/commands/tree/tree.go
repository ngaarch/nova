package tree

import (
	"nova/internal/command"
	"nova/internal/filesystem"
	"nova/internal/output"
)

// Command returns the registered Command instance for the tree subcommand.
func Command() *command.Command {
	return &command.Command{
		Name:        "tree",
		Summary:     "Display directory hierarchy as a visual tree with metrics",
		Usage:       "nova tree [flags] [path]",
		Description: "Display directory hierarchy as an indented tree with icons, colors, and metrics.",
		Phase:       5,
		Run:         Run,
	}
}

// Run executes the tree command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)
	if ctx.Printer.Mode == output.ModePlain {
		opts.Plain = true
	} else if ctx.Printer.Mode == output.ModeJSON {
		opts.JSON = true
	}

	walkOpts := filesystem.DefaultWalkOptions()
	walkOpts.MaxDepth = opts.MaxDepth
	walkOpts.IncludeHidden = opts.All
	walkOpts.DirsOnly = opts.DirsOnly
	walkOpts.NeedDetails = opts.Permissions

	rootNode, err := filesystem.BuildTree(opts.Path, walkOpts)
	if err != nil {
		return command.NewOpError("failed to inspect directory tree", opts.Path, err, "Check file permissions and directory path.")
	}

	if opts.JSON {
		return RenderJSON(ctx, rootNode)
	}

	showIcons := false
	if opts.Icons == "always" {
		showIcons = true
	} else if opts.Icons == "never" {
		showIcons = false
	} else {
		// auto: human mode with color and unicode
		showIcons = (ctx.Printer.Mode == output.ModeHuman && !opts.Plain)
	}

	RenderTree(ctx, rootNode, opts, showIcons)
	return nil
}
