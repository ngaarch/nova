package servecmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"nova/internal/command"
	"nova/internal/commands/qr"
	"nova/internal/output"
)

// ServerInfo contains runtime metadata for the static web server.
type ServerInfo struct {
	URL      string `json:"url"`
	LocalURL string `json:"local_url"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Dir      string `json:"dir"`
	SPA      bool   `json:"spa"`
	CORS     bool   `json:"cors"`
}

// RequestLog holds telemetry for one processed HTTP request.
type RequestLog struct {
	Status   int           `json:"status"`
	Method   string        `json:"method"`
	Path     string        `json:"path"`
	Duration time.Duration `json:"duration"`
	Bytes    int64         `json:"bytes"`
	RemoteIP string        `json:"remote_ip"`
}

// Command returns the registered Command instance for serve.
func Command() *command.Command {
	return &command.Command{
		Name:        "serve",
		Aliases:     []string{"server", "httpd"},
		Summary:     "Zero-config static web server with QR code, SPA mode, and live logger",
		Usage:       "nova serve [dir] [flags]",
		Description: "Serve directory files over HTTP with automatic port failover, CORS, and mobile QR pairing.",
		Phase:       26,
		Run:         Run,
	}
}

// Run executes the serve command.
func Run(ctx *command.Context, args []string) error {
	for _, arg := range args {
		if arg == "--plain" {
			ctx.Printer.Mode = output.ModePlain
		} else if arg == "--json" {
			ctx.Printer.Mode = output.ModeJSON
		}
	}

	opts := ParseFlags(args)

	absDir, err := filepath.Abs(opts.Dir)
	if err != nil {
		return fmt.Errorf("resolve dir %s: %w", opts.Dir, err)
	}

	fi, err := os.Stat(absDir)
	if err != nil {
		return fmt.Errorf("stat directory %s: %w", absDir, err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("path %s is not a directory", absDir)
	}

	// Bind listener with port failover
	ln, boundPort, err := bindListener(opts.Host, opts.Port)
	if err != nil {
		return fmt.Errorf("bind server: %w", err)
	}

	localURL := fmt.Sprintf("http://localhost:%d", boundPort)
	serverURL := fmt.Sprintf("http://%s:%d", opts.Host, boundPort)

	info := ServerInfo{
		URL:      serverURL,
		LocalURL: localURL,
		Host:     opts.Host,
		Port:     boundPort,
		Dir:      absDir,
		SPA:      opts.SPA,
		CORS:     opts.CORS,
	}

	if ctx.Printer.Mode == output.ModeJSON {
		_ = ctx.Printer.PrintJSON(info)
	}

	// Render startup banner in Human/Plain mode
	var qrLines []string
	if opts.QR && ctx.Printer.Mode != output.ModePlain && ctx.Printer.Mode != output.ModeJSON {
		if qrCode, err := qrcmd.Encode(localURL); err == nil {
			qrLines = renderQRHalfBlocks(qrCode)
		}
	}

	if ctx.Printer.Mode != output.ModeJSON {
		RenderStartupBanner(ctx.Stdout, &info, qrLines, opts, ctx)
	}

	// Build handler
	fs := http.FileServer(http.Dir(absDir))
	var reqCount atomic.Int64

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		if opts.CORS {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, HEAD")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		rec := &responseRecorder{ResponseWriter: w, status: 200}

		if opts.SPA && r.Method == http.MethodGet {
			// Check if file exists; if not, serve index.html
			relPath := strings.TrimPrefix(r.URL.Path, "/")
			checkPath := filepath.Join(absDir, relPath)
			if _, err := os.Stat(checkPath); err != nil {
				indexPath := filepath.Join(absDir, "index.html")
				if _, err := os.Stat(indexPath); err == nil {
					http.ServeFile(rec, r, indexPath)
					logRequest(ctx, rec, r, start)
					reqCount.Add(1)
					return
				}
			}
		}

		fs.ServeHTTP(rec, r)
		logRequest(ctx, rec, r, start)
		reqCount.Add(1)
	})

	srv := &http.Server{
		Handler: handler,
	}

	// If Once is true (test mode), return immediately after verification
	if opts.Once {
		_ = ln.Close()
		return nil
	}

	// Graceful shutdown on SIGINT/SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case sig := <-stop:
		if ctx.Printer.Mode != output.ModePlain && ctx.Printer.Mode != output.ModeJSON {
			fmt.Fprintf(ctx.Stdout, "\n  Shutting down server on signal (%v)...\n", sig)
		}
		ctxTimeout, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctxTimeout)
		return nil
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	}
}

func bindListener(host string, startPort int) (net.Listener, int, error) {
	for p := startPort; p < startPort+25; p++ {
		addr := net.JoinHostPort(host, strconv.Itoa(p))
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			return ln, p, nil
		}
	}
	return nil, 0, fmt.Errorf("failed to bind port starting at %d (ports occupied)", startPort)
}

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int64
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.bytes += int64(n)
	return n, err
}

func logRequest(ctx *command.Context, rec *responseRecorder, r *http.Request, start time.Time) {
	dur := time.Since(start)
	log := RequestLog{
		Status:   rec.status,
		Method:   r.Method,
		Path:     r.URL.Path,
		Duration: dur,
		Bytes:    rec.bytes,
		RemoteIP: r.RemoteAddr,
	}

	if ctx.Printer.Mode != output.ModeJSON {
		RenderRequestLog(ctx.Stdout, log, ctx)
	}
}

func renderQRHalfBlocks(code qrcmd.Code) []string {
	var lines []string
	qPad := "  "
	for r := 0; r < code.Size; r += 2 {
		var row strings.Builder
		row.WriteString(qPad)
		for c := 0; c < code.Size; c++ {
			top := code.Matrix[r][c]
			bot := false
			if r+1 < code.Size {
				bot = code.Matrix[r+1][c]
			}
			if top && bot {
				row.WriteString("█")
			} else if top && !bot {
				row.WriteString("▀")
			} else if !top && bot {
				row.WriteString("▄")
			} else {
				row.WriteString(" ")
			}
		}
		lines = append(lines, row.String())
	}
	return lines
}
