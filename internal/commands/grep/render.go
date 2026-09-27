package grep

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"nova/internal/command"
	"nova/internal/renderer"
	"nova/internal/terminal"
	"nova/internal/theme"
)

// RenderJSON serializes search results to JSON format.
func RenderJSON(ctx *command.Context, results []FileResult, summary Summary) error {
	payload := struct {
		Summary Summary      `json:"summary"`
		Results []FileResult `json:"results"`
	}{
		Summary: summary,
		Results: results,
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	ctx.Printer.Println(string(data))
	return nil
}

// RenderPlain outputs search results in clean, machine-parsable format.
func RenderPlain(ctx *command.Context, results []FileResult, opts Options) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].Path < results[j].Path
	})

	for _, fr := range results {
		if opts.FilesWithMatch {
			ctx.Printer.Println(fr.Path)
			continue
		}
		if opts.CountOnly {
			ctx.Printer.Printf("%s:%d\n", fr.Path, fr.MatchCount)
			continue
		}

		for _, m := range fr.Matches {
			if !m.IsMatch && opts.ContextBefore == 0 && opts.ContextAfter == 0 {
				continue
			}
			sep := ":"
			if !m.IsMatch {
				sep = "-"
			}
			if opts.LineNumbers {
				ctx.Printer.Printf("%s%s%d%s%s\n", fr.Path, sep, m.LineNumber, sep, m.Content)
			} else {
				ctx.Printer.Printf("%s%s%s\n", fr.Path, sep, m.Content)
			}
		}
	}
}

// RenderHuman renders a vibrant, interactive dashboard of search results.
func RenderHuman(ctx *command.Context, results []FileResult, summary Summary, opts Options, re *regexp.Regexp) {
	th := ctx.Theme
	prof := ctx.Caps.ColorProfile
	useIcons := ctx.Caps.UnicodeSupported && ctx.Config.IconMode != "never"

	if opts.FilesWithMatch {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Path < results[j].Path
		})
		for _, fr := range results {
			icon := ""
			if useIcons {
				icon = theme.LookupIcon(fr.Path, theme.TypeRegular, true) + " "
			}
			ctx.Printer.Printf("  %s%s\n", icon, th.Format(theme.RoleInfo, fr.Path, prof))
		}
		renderSummaryBadge(ctx, summary)
		return
	}

	if opts.CountOnly {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Path < results[j].Path
		})
		for _, fr := range results {
			icon := ""
			if useIcons {
				icon = theme.LookupIcon(fr.Path, theme.TypeRegular, true) + " "
			}
			pathStr := th.Format(theme.RoleInfo, fr.Path, prof)
			countStr := th.Format(theme.RoleAccent, fmt.Sprintf("%d matches", fr.MatchCount), prof)
			ctx.Printer.Printf("  %s%-40s %s\n", icon, pathStr, countStr)
		}
		renderSummaryBadge(ctx, summary)
		return
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Path < results[j].Path
	})

	for _, fr := range results {
		icon := ""
		if useIcons {
			icon = theme.LookupIcon(fr.Path, theme.TypeRegular, true) + " "
		}

		header := fmt.Sprintf(" %s%s (%d matches)", icon, fr.Path, fr.MatchCount)
		ctx.Printer.Println(th.Format(theme.RoleAccent, header, prof))

		for _, m := range fr.Matches {
			lineNumStr := ""
			if opts.LineNumbers {
				if m.IsMatch {
					lineNumStr = th.Format(theme.RoleSuccess, fmt.Sprintf("%4d │ ", m.LineNumber), prof)
				} else {
					lineNumStr = th.Format(theme.RoleMuted, fmt.Sprintf("%4d │ ", m.LineNumber), prof)
				}
			}

			content := m.Content
			if m.IsMatch {
				content = highlightRegexMatches(content, re, th, prof)
			} else {
				content = th.Format(theme.RoleMuted, content, prof)
			}

			ctx.Printer.Printf("  %s%s\n", lineNumStr, content)
		}
		ctx.Printer.Println()
	}

	renderSummaryBadge(ctx, summary)
}

func highlightRegexMatches(text string, re *regexp.Regexp, th *theme.Theme, prof terminal.ColorProfile) string {
	matches := re.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return text
	}

	var sb strings.Builder
	lastIdx := 0

	fg := theme.HexRGB(0xFFFFFF)
	bg := theme.HexRGB(0xE0218A) // Neon magenta
	matchStyle := theme.Style{
		Fg:   &fg,
		Bg:   &bg,
		Bold: true,
	}

	for _, m := range matches {
		start, end := m[0], m[1]
		if start > lastIdx {
			sb.WriteString(text[lastIdx:start])
		}
		matchedSub := text[start:end]
		sb.WriteString(matchStyle.Format(matchedSub, prof))
		lastIdx = end
	}

	if lastIdx < len(text) {
		sb.WriteString(text[lastIdx:])
	}

	return sb.String()
}

func renderSummaryBadge(ctx *command.Context, summary Summary) {
	th := ctx.Theme
	prof := ctx.Caps.ColorProfile

	ms := float64(summary.Duration.Microseconds()) / 1000.0
	timing := fmt.Sprintf("%.2fms", ms)

	summaryText := fmt.Sprintf("✦ Found %s in %s (scanned %s files in %s)",
		th.Format(theme.RoleSuccess, fmt.Sprintf("%d matches", summary.TotalMatches), prof),
		th.Format(theme.RoleAccent, fmt.Sprintf("%d files", summary.TotalFiles), prof),
		th.Format(theme.RoleInfo, fmt.Sprintf("%d", summary.FilesScanned), prof),
		th.Format(theme.RoleMuted, timing, prof),
	)

	border := strings.Repeat("─", renderer.VisibleWidth(summaryText)+2)
	ctx.Printer.Println(th.Format(theme.RoleMuted, "  "+border, prof))
	ctx.Printer.Printf("  %s\n", summaryText)
}
