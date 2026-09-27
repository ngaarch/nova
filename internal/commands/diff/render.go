package diff

import (
	"encoding/json"
	"fmt"
	"strings"

	"nova/internal/command"
	"nova/internal/theme"
)

// RenderHuman prints a rich, color-coded visual diff with line numbers and summaries.
func RenderHuman(ctx *command.Context, res DiffResult, opts Options) {
	th := ctx.Theme
	prof := ctx.Caps.ColorProfile
	unicode := ctx.Caps.UnicodeSupported

	if opts.Brief {
		if res.Identical {
			ctx.Printer.Println(fmt.Sprintf("Files %s and %s are identical", res.File1, res.File2))
		} else {
			ctx.Printer.Println(fmt.Sprintf("Files %s and %s differ", res.File1, res.File2))
		}
		return
	}

	if res.IsBinary {
		if res.Identical {
			msg := fmt.Sprintf("✔ Binary files '%s' and '%s' are identical", res.File1, res.File2)
			ctx.Printer.Println(th.Format(theme.RoleSuccess, msg, prof))
		} else {
			msg := fmt.Sprintf("✖ Binary files '%s' and '%s' differ", res.File1, res.File2)
			ctx.Printer.Println(th.Format(theme.RoleWarning, msg, prof))
		}
		return
	}

	if res.Identical {
		icon := "✔ "
		if !unicode {
			icon = "= "
		}
		msg := fmt.Sprintf("%sFiles '%s' and '%s' are identical", icon, res.File1, res.File2)
		ctx.Printer.Println(th.Format(theme.RoleSuccess, msg, prof))
		return
	}

	// File headers
	tFmt := "2006-01-02 15:04:05"
	delFileHeader := fmt.Sprintf("--- a/%s\t%s", res.File1, res.Time1.Format(tFmt))
	addFileHeader := fmt.Sprintf("+++ b/%s\t%s", res.File2, res.Time2.Format(tFmt))
	ctx.Printer.Println(th.Format(theme.RoleError, delFileHeader, prof))
	ctx.Printer.Println(th.Format(theme.RoleSuccess, addFileHeader, prof))

	// Hunks
	for _, hunk := range res.Hunks {
		hunkHdr := fmt.Sprintf("@@ -%d,%d +%d,%d @@", hunk.OldStart, hunk.OldCount, hunk.NewStart, hunk.NewCount)
		ctx.Printer.Println(th.Format(theme.RoleAccent, hunkHdr, prof))

		for _, l := range hunk.Lines {
			switch l.Type {
			case LineContext:
				prefix := fmt.Sprintf("%4d │   ", l.NewNum)
				lineStr := th.Format(theme.RoleMuted, prefix, prof) + l.Content
				ctx.Printer.Println(lineStr)
			case LineDelete:
				prefix := fmt.Sprintf("%4d │ - ", l.OldNum)
				lineStr := th.Format(theme.RoleError, prefix+l.Content, prof)
				ctx.Printer.Println(lineStr)
			case LineInsert:
				prefix := fmt.Sprintf("%4d │ + ", l.NewNum)
				lineStr := th.Format(theme.RoleSuccess, prefix+l.Content, prof)
				ctx.Printer.Println(lineStr)
			}
		}
	}

	// Change summary card
	bTopL, bTopR, bBotL, bBotR := "╭", "╮", "╰", "╯"
	bHor := "─"
	bVert := "│"
	if !unicode {
		bTopL, bTopR, bBotL, bBotR = "+", "+", "+", "+"
		bHor = "-"
		bVert = "|"
	}

	cardW := 52
	addStr := th.Format(theme.RoleSuccess, fmt.Sprintf("+%d", res.Additions), prof)
	delStr := th.Format(theme.RoleError, fmt.Sprintf("-%d", res.Deletions), prof)
	hunkWord := "hunks"
	if len(res.Hunks) == 1 {
		hunkWord = "hunk"
	}
	summaryInner := fmt.Sprintf("  %s additions, %s deletions across %d %s", addStr, delStr, len(res.Hunks), hunkWord)

	bFmt := func(s string) string {
		return th.Format(theme.RoleMuted, s, prof)
	}

	ctx.Printer.Println(bFmt(bTopL + strings.Repeat(bHor, cardW) + bTopR))
	ctx.Printer.Println(bFmt(bVert) + summaryInner)
	ctx.Printer.Println(bFmt(bBotL + strings.Repeat(bHor, cardW) + bBotR))
}

// RenderPlain prints standard POSIX unified diff.
func RenderPlain(ctx *command.Context, res DiffResult, opts Options) {
	if opts.Brief {
		if res.Identical {
			ctx.Printer.Println(fmt.Sprintf("Files %s and %s are identical", res.File1, res.File2))
		} else {
			ctx.Printer.Println(fmt.Sprintf("Files %s and %s differ", res.File1, res.File2))
		}
		return
	}

	if res.Identical {
		return
	}

	tFmt := "2006-01-02 15:04:05"
	ctx.Printer.Println(fmt.Sprintf("--- %s\t%s", res.File1, res.Time1.Format(tFmt)))
	ctx.Printer.Println(fmt.Sprintf("+++ %s\t%s", res.File2, res.Time2.Format(tFmt)))

	for _, hunk := range res.Hunks {
		ctx.Printer.Println(fmt.Sprintf("@@ -%d,%d +%d,%d @@", hunk.OldStart, hunk.OldCount, hunk.NewStart, hunk.NewCount))
		for _, l := range hunk.Lines {
			switch l.Type {
			case LineContext:
				ctx.Printer.Println(" " + l.Content)
			case LineDelete:
				ctx.Printer.Println("-" + l.Content)
			case LineInsert:
				ctx.Printer.Println("+" + l.Content)
			}
		}
	}
}

// RenderJSON serializes the DiffResult to structured JSON.
func RenderJSON(ctx *command.Context, res DiffResult) error {
	data, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	ctx.Printer.Println(string(data))
	return nil
}
