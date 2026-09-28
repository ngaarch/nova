package httpcmd

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"os"
	"strings"
	"time"

	"nova/internal/command"
	"nova/internal/output"
)

// Telemetry captures network request latency metrics.
type Telemetry struct {
	DNSLookup        time.Duration `json:"dns_lookup_ms"`
	TCPConnect       time.Duration `json:"tcp_connect_ms"`
	TLSHandshake     time.Duration `json:"tls_handshake_ms"`
	ServerProcessing time.Duration `json:"server_processing_ms"` // TTFB
	ContentTransfer  time.Duration `json:"content_transfer_ms"`
	TotalLatency     time.Duration `json:"total_latency_ms"`
}

// Response holds the complete request outcome and metadata.
type Response struct {
	URL        string              `json:"url"`
	Method     string              `json:"method"`
	StatusCode int                 `json:"status_code"`
	StatusText string              `json:"status_text"`
	Proto      string              `json:"protocol"`
	Headers    map[string][]string `json:"headers"`
	Size       int64               `json:"size"`
	Telemetry  Telemetry           `json:"telemetry"`
	Body       string              `json:"body"`
	SavedTo    string              `json:"saved_to,omitempty"`
}

// Command returns the registered Command instance for http.
func Command() *command.Command {
	return &command.Command{
		Name:        "http",
		Aliases:     []string{"fetch", "curl", "request"},
		Summary:     "Execute HTTP requests with detailed latency waterfall and syntax highlighting",
		Usage:       "nova http [flags] <url>",
		Description: "Perform HTTP/REST API testing with latency telemetry (DNS, TCP, TLS, TTFB) and colored response preview.",
		Phase:       19,
		Run:         Run,
	}
}

// Run executes the http command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)
	if opts.URL == "" {
		return fmt.Errorf("missing target URL (usage: nova http [flags] <url>)")
	}

	// Ensure scheme
	targetURL := opts.URL
	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		targetURL = "http://" + targetURL
	}

	// Prepare body
	var bodyReader io.Reader
	if opts.Data != "" {
		bodyReader = strings.NewReader(opts.Data)
	} else if opts.DataFile != "" {
		f, err := os.Open(opts.DataFile)
		if err != nil {
			return fmt.Errorf("read data-file %s: %w", opts.DataFile, err)
		}
		defer f.Close()
		bodyReader = f
	}

	req, err := http.NewRequest(opts.Method, targetURL, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// Apply headers
	for _, h := range opts.Headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			req.Header.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "nova/1.6.0 (Terminal Suite)")
	}

	// Trace timings
	var dnsStart, dnsDone time.Time
	var connStart, connDone time.Time
	var tlsStart, tlsDone time.Time
	var gotFirstByte time.Time

	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) { dnsStart = time.Now() },
		DNSDone:  func(_ httptrace.DNSDoneInfo) { dnsDone = time.Now() },
		ConnectStart: func(_, _ string) {
			if connStart.IsZero() {
				connStart = time.Now()
			}
		},
		ConnectDone: func(_, _ string, _ error) { connDone = time.Now() },
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			tlsDone = time.Now()
		},
		GotFirstResponseByte: func() {
			gotFirstByte = time.Now()
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	// Configure transport & client
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: opts.Insecure, //nolint:gosec
		},
		Proxy: http.ProxyFromEnvironment,
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   opts.Timeout,
	}
	if !opts.FollowRedirects {
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	startTotal := time.Now()
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	bodyBuf := &bytes.Buffer{}
	var readStart = time.Now()
	if !gotFirstByte.IsZero() {
		readStart = gotFirstByte
	}
	n, err := io.Copy(bodyBuf, res.Body)
	transferDone := time.Now()
	totalDuration := time.Since(startTotal)

	if err != nil && err != io.EOF {
		return fmt.Errorf("read response body: %w", err)
	}

	// Compute telemetry
	telemetry := Telemetry{
		TotalLatency: totalDuration,
	}
	if !dnsStart.IsZero() && !dnsDone.IsZero() {
		telemetry.DNSLookup = dnsDone.Sub(dnsStart)
	}
	if !connStart.IsZero() && !connDone.IsZero() {
		telemetry.TCPConnect = connDone.Sub(connStart)
	}
	if !tlsStart.IsZero() && !tlsDone.IsZero() {
		telemetry.TLSHandshake = tlsDone.Sub(tlsStart)
	}
	if !gotFirstByte.IsZero() {
		connectFinish := connDone
		if !tlsDone.IsZero() {
			connectFinish = tlsDone
		}
		if !connectFinish.IsZero() {
			telemetry.ServerProcessing = gotFirstByte.Sub(connectFinish)
		} else {
			telemetry.ServerProcessing = gotFirstByte.Sub(startTotal)
		}
	}
	telemetry.ContentTransfer = transferDone.Sub(readStart)

	bodyStr := bodyBuf.String()
	savedPath := ""
	if opts.OutputFile != "" {
		if err := os.WriteFile(opts.OutputFile, bodyBuf.Bytes(), 0o644); err != nil {
			return fmt.Errorf("save output file: %w", err)
		}
		savedPath = opts.OutputFile
	}

	response := Response{
		URL:        targetURL,
		Method:     opts.Method,
		StatusCode: res.StatusCode,
		StatusText: res.Status,
		Proto:      res.Proto,
		Headers:    res.Header,
		Size:       n,
		Telemetry:  telemetry,
		Body:       bodyStr,
		SavedTo:    savedPath,
	}

	return Render(ctx, response, opts)
}
