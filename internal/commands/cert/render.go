package certcmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"nova/internal/command"
	"nova/internal/renderer"
	"nova/internal/theme"
)

// RenderCertPlain prints certificate information as tab-separated values.
func RenderCertPlain(w io.Writer, res *CertResult) {
	fmt.Fprintf(w, "TARGET\tCN\tISSUER\tSTATUS\tDAYS_LEFT\tNOT_BEFORE\tNOT_AFTER\tKEY_ALGO\n")
	for _, c := range res.Certificates {
		notBeforeStr := c.NotBefore.Format("2006-01-02")
		notAfterStr := c.NotAfter.Format("2006-01-02")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\t%s\t%s\n",
			res.Target, c.CommonName, c.Issuer, c.Status, c.DaysRemaining, notBeforeStr, notAfterStr, c.KeyAlgorithm)
	}
}

// RenderCertDashboard renders an informative box-drawing card of the certificate chain.
func RenderCertDashboard(w io.Writer, res *CertResult, ctx *command.Context) {
	th := ctx.Theme
	prof := ctx.Caps.ColorProfile
	width := ctx.Caps.Width
	if width < 74 {
		width = 74
	}
	if width > 100 {
		width = 100
	}

	header := fmt.Sprintf("╭ 🔒 TLS Certificate Inspection: %s %s╮", res.Target, strings.Repeat("─", width-36-len(res.Target)))
	fmt.Fprintln(w, th.Format(theme.RoleAccent, renderer.Truncate(header, width, "─╮"), prof))

	fmt.Fprintf(w, "  Target:        %s\n", th.Format(theme.RoleAccent, res.Target, prof))
	if res.IsRemote {
		proto := fmt.Sprintf("%s (%s)", res.TLSVersion, res.CipherSuite)
		fmt.Fprintf(w, "  Protocol:      %s\n", th.Format(theme.RoleInfo, proto, prof))
	}
	fmt.Fprintf(w, "  Certificates:  %d certificate(s) in chain\n", len(res.Certificates))

	div := "  " + strings.Repeat("─", width-4)
	fmt.Fprintln(w, th.Format(theme.RoleMuted, div, prof))

	for i, c := range res.Certificates {
		badge := fmt.Sprintf("[#%d Leaf Certificate]", i+1)
		if i > 0 {
			badge = fmt.Sprintf("[#%d Intermediate / CA Certificate]", i+1)
		}
		fmt.Fprintln(w, "  "+th.Format(theme.RoleDocument, badge, prof))

		fmt.Fprintf(w, "    Common Name:   %s\n", th.Format(theme.RoleAccent, c.CommonName, prof))
		fmt.Fprintf(w, "    Subject:       %s\n", renderer.Truncate(c.Subject, width-20, "…"))
		fmt.Fprintf(w, "    Issuer:        %s\n", renderer.Truncate(c.Issuer, width-20, "…"))

		// Status with color
		statusRole := theme.RoleSuccess
		statusIcon := "✔"
		if c.Status == "Expired" {
			statusRole = theme.RoleError
			statusIcon = "✖"
		} else if c.Status == "Expiring Soon" {
			statusRole = theme.RoleWarning
			statusIcon = "⚠"
		}
		statusText := fmt.Sprintf("%s %s (%d days remaining)", statusIcon, c.Status, c.DaysRemaining)
		fmt.Fprintf(w, "    Status:        %s\n", th.Format(statusRole, statusText, prof))

		validityStr := fmt.Sprintf("%s to %s", c.NotBefore.Format("2006-01-02"), c.NotAfter.Format("2006-01-02"))
		fmt.Fprintf(w, "    Validity:      %s\n", validityStr)
		fmt.Fprintf(w, "    Algorithm:     %s • %s\n", c.KeyAlgorithm, c.SignatureAlgorithm)

		if len(c.SANs) > 0 {
			sansJoined := strings.Join(c.SANs, ", ")
			fmt.Fprintf(w, "    SANs (%d):     %s\n", len(c.SANs), renderer.Truncate(sansJoined, width-20, "…"))
		}

		if c.SerialNumber != "" {
			fmt.Fprintf(w, "    Serial:        %s\n", renderer.Truncate(c.SerialNumber, width-20, "…"))
		}

		if i < len(res.Certificates)-1 {
			fmt.Fprintln(w, th.Format(theme.RoleMuted, "  "+strings.Repeat("┈", width-4), prof))
		}
	}

	bot := fmt.Sprintf("╰%s╯", strings.Repeat("─", width-2))
	fmt.Fprintln(w, th.Format(theme.RoleAccent, bot, prof))
}

// RenderJWTPlain outputs JWT claims as tab-separated values.
func RenderJWTPlain(w io.Writer, info *JWTInfo) {
	fmt.Fprintf(w, "ALGORITHM\t%s\n", info.Algorithm)
	fmt.Fprintf(w, "TYPE\t%s\n", info.Type)
	fmt.Fprintf(w, "SUBJECT\t%s\n", info.Subject)
	fmt.Fprintf(w, "ISSUER\t%s\n", info.Issuer)
	fmt.Fprintf(w, "AUDIENCE\t%s\n", info.Audience)
	fmt.Fprintf(w, "STATUS\t%s\n", info.TimeStatus)
	fmt.Fprintf(w, "HEADER_JSON\t%s\n", info.RawHeader)
	fmt.Fprintf(w, "PAYLOAD_JSON\t%s\n", info.RawPayload)
}

// RenderJWTDashboard renders a formatted dashboard for decoded JWT tokens.
func RenderJWTDashboard(w io.Writer, info *JWTInfo, ctx *command.Context) {
	th := ctx.Theme
	prof := ctx.Caps.ColorProfile
	width := ctx.Caps.Width
	if width < 74 {
		width = 74
	}
	if width > 100 {
		width = 100
	}

	header := fmt.Sprintf("╭ 🔑 JSON Web Token (JWT) Decoder %s╮", strings.Repeat("─", width-36))
	fmt.Fprintln(w, th.Format(theme.RoleAccent, header, prof))

	fmt.Fprintf(w, "  Algorithm:    %s  │  Type: %s\n",
		th.Format(theme.RoleAccent, info.Algorithm, prof),
		th.Format(theme.RoleInfo, info.Type, prof))

	statusRole := theme.RoleSuccess
	statusIcon := "✔"
	if info.IsExpired {
		statusRole = theme.RoleError
		statusIcon = "✖"
	}
	fmt.Fprintf(w, "  Status:       %s\n", th.Format(statusRole, fmt.Sprintf("%s %s", statusIcon, info.TimeStatus), prof))

	div := "  " + strings.Repeat("─", width-4)
	fmt.Fprintln(w, th.Format(theme.RoleMuted, div, prof))

	fmt.Fprintln(w, "  "+th.Format(theme.RoleDocument, "Header:", prof))
	headerIndent, _ := json.MarshalIndent(info.Header, "    ", "  ")
	fmt.Fprintf(w, "    %s\n", string(headerIndent))

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "  "+th.Format(theme.RoleDocument, "Payload Claims:", prof))
	payloadIndent, _ := json.MarshalIndent(info.Payload, "    ", "  ")
	fmt.Fprintf(w, "    %s\n", string(payloadIndent))

	fmt.Fprintln(w, th.Format(theme.RoleMuted, div, prof))
	fmt.Fprintf(w, "  Signature:    %s\n", renderer.Truncate(info.Signature, width-18, "…"))

	bot := fmt.Sprintf("╰%s╯", strings.Repeat("─", width-2))
	fmt.Fprintln(w, th.Format(theme.RoleAccent, bot, prof))
}
