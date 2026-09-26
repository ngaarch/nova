package du

import (
	"sort"
	"strings"

	"nova/internal/command"
	"nova/internal/filesystem"
	"nova/internal/output"
)

// Command returns the registered Command instance for the du subcommand.
func Command() *command.Command {
	return &command.Command{
		Name:        "du",
		Summary:     "Estimate file and directory disk space usage with aggregated visual metrics",
		Usage:       "nova du [flags] [paths...]",
		Description: "Estimate file and directory space consumption with responsive visual usage bars.",
		Phase:       5,
		Run:         Run,
	}
}

// Run executes the du command.
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
	walkOpts.IncludeHidden = true
	walkOpts.NeedDetails = true

	var allItems []Item
	var grandTotal int64

	for _, target := range opts.Paths {
		tree, err := filesystem.BuildTree(target, walkOpts)
		if err != nil {
			return command.NewOpError("cannot calculate disk usage", target, err, "Check file read permissions.")
		}

		var items []Item
		collectItems(tree, opts.MaxDepth, opts.All, &items)

		// Sort items
		if opts.SortBy == "size" {
			sort.Slice(items, func(i, j int) bool {
				return items[i].Bytes > items[j].Bytes
			})
		} else {
			sort.Slice(items, func(i, j int) bool {
				return strings.ToLower(items[i].Path) < strings.ToLower(items[j].Path)
			})
		}

		allItems = append(allItems, items...)
		grandTotal += tree.TotalSize
	}

	if opts.Total {
		allItems = append(allItems, Item{
			Path:      "total",
			Bytes:     grandTotal,
			HumanSize: formatDUSize(grandTotal, opts.HumanReadable),
			IsDir:     true,
			Percent:   100.0,
		})
	}

	if opts.JSON {
		return RenderJSON(ctx, allItems)
	} else if opts.Plain {
		RenderPlain(ctx, allItems, opts.HumanReadable)
	} else {
		RenderHuman(ctx, allItems, grandTotal)
	}

	return nil
}

func collectItems(node *filesystem.TreeNode, maxDepth int, includeFiles bool, out *[]Item) {
	if node == nil {
		return
	}

	shouldInclude := node.Entry.IsDir || includeFiles
	depthOk := (maxDepth < 0) || (node.Depth <= maxDepth)

	if shouldInclude && depthOk {
		pct := 0.0
		*out = append(*out, Item{
			Path:      node.Entry.Path,
			Bytes:     node.TotalSize,
			HumanSize: formatDUSize(node.TotalSize, true),
			IsDir:     node.Entry.IsDir,
			Files:     node.FileCount,
			Dirs:      node.DirCount,
			Percent:   pct,
		})
	}

	for _, child := range node.Children {
		collectItems(child, maxDepth, includeFiles, out)
	}
}
