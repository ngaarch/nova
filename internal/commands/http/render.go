package httpcmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"nova/internal/command"
	"nova/internal/output"
	"nova/internal/theme"
)

// Render outputs the HTTP response.
func Render(ctx *command.Context, res Response, opts Options) error {
	if ctx.Printer.Mode == output.ModeJSON {
		return ctx.Printer.PrintJSON(res)
	}

	if ctx.Printer.Mode == output.ModePlain {
		var b strings.Builder
		if opts.IncludeHeaders {
			b.WriteString(fmt.Sprintf("%s %s\n", res.Proto, res.StatusText))
			for k, vals := range res.Headers {
				for _, v := range vals {
					b.WriteString(fmt.Sprintf("%s: %s\n", k, v))
				}
			}
			b.WriteString("\n")
		}
		b.WriteString(res.Body)
		if !strings.HasSuffix(res.Body, "\n") {
			b.WriteString("\n")
		}
		ctx.Printer.Print(b.String())
		return nil
	}

	// Human mode
	prof := ctx.Caps.ColorProfile
	th := ctx.Theme

	title := fmt.Sprintf("🚀 HTTP %s %s", res.Method, res.URL)
	header := fmt.Sprintf("╭ %s %s╮\n", title, strings.Repeat("─", max(2, 70-len(title))))
	var body strings.Builder
	body.WriteString(header)

	// Status badge
	statusRole := theme.RoleSuccess
	if res.StatusCode >= 400 && res.StatusCode < 500 {
		statusRole = theme.RoleWarning
	} else if res.StatusCode >= 500 {
		statusRole = theme.RoleError
	}
	statusBadge := th.Format(statusRole, fmt.Sprintf("%d %s", res.StatusCode, res.StatusText), prof)
	summaryLine := fmt.Sprintf("  Status: %s  Proto: %s  Size: %s  Time: %s",
		statusBadge,
		th.Format(theme.RoleMuted, res.Proto, prof),
		th.Format(theme.RoleInfo, formatBytes(res.Size), prof),
		th.Format(theme.RoleAccent, fmt.Sprintf("%.2fms", float64(res.Telemetry.TotalLatency.Microseconds())/1000.0), prof),
	)
	body.WriteString(fmt.Sprintf("│%s\n", summaryLine))

	// Timing waterfall
	if opts.Timing {
		body.WriteString(fmt.Sprintf("│ %s\n", strings.Repeat("─", 74)))
		timingLine := fmt.Sprintf("│  DNS: %s │ TCP: %s │ TLS: %s │ TTFB: %s │ Transfer: %s\n",
			th.Format(theme.RoleMuted, formatDuration(res.Telemetry.DNSLookup), prof),
			th.Format(theme.RoleMuted, formatDuration(res.Telemetry.TCPConnect), prof),
			th.Format(theme.RoleMuted, formatDuration(res.Telemetry.TLSHandshake), prof),
			th.Format(theme.RoleAccent, formatDuration(res.Telemetry.ServerProcessing), prof),
			th.Format(theme.RoleMuted, formatDuration(res.Telemetry.ContentTransfer), prof),
		)
		body.WriteString(timingLine)
	}

	// Headers
	if opts.IncludeHeaders {
		body.WriteString(fmt.Sprintf("│ %s\n", strings.Repeat("─", 74)))
		body.WriteString(fmt.Sprintf("│  %s\n", th.Format(theme.RoleAccent, "[ Response Headers ]", prof)))
		for k, vals := range res.Headers {
			for _, v := range vals {
				body.WriteString(fmt.Sprintf("│    %s: %s\n",
					th.Format(theme.RoleExecutable, k, prof),
					th.Format(theme.RoleRegularFile, v, prof),
				))
			}
		}
	}

	if res.SavedTo != "" {
		body.WriteString(fmt.Sprintf("│ %s\n", strings.Repeat("─", 74)))
		body.WriteString(fmt.Sprintf("│  ✔ Response body saved to: %s\n", th.Format(theme.RoleSuccess, res.SavedTo, prof)))
	}

	body.WriteString(fmt.Sprintf("╰%s╯\n", strings.Repeat("─", 76)))
	ctx.Printer.Print(body.String())

	// Print formatted body if not saved to file and not empty
	if res.SavedTo == "" && res.Body != "" {
		contentType := ""
		for k, v := range res.Headers {
			if strings.EqualFold(k, "Content-Type") && len(v) > 0 {
				contentType = v[0]
				break
			}
		}

		formattedBody := res.Body
		// If JSON, pretty print
		if strings.Contains(contentType, "application/json") || (strings.HasPrefix(strings.TrimSpace(res.Body), "{") && strings.HasSuffix(strings.TrimSpace(res.Body), "}")) {
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, []byte(res.Body), "", "  "); err == nil {
				formattedBody = prettyJSON.String()
			}
		}

		if !strings.HasSuffix(formattedBody, "\n") {
			formattedBody += "\n"
		}
		ctx.Printer.Print(formattedBody)
	}

	return nil
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "0ms"
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	return fmt.Sprintf("%.1fms", float64(d.Microseconds())/1000.0)
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
