package servecmd

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nova/internal/command"
	"nova/internal/config"
	"nova/internal/output"
	"nova/internal/terminal"
	"nova/internal/theme"
)

func TestFlagsParse(t *testing.T) {
	opts := ParseFlags([]string{"-p", "3000", "-d", "/tmp/site", "--spa", "--cors", "--no-qr", "--once"})
	if opts.Port != 3000 {
		t.Errorf("expected Port 3000, got %d", opts.Port)
	}
	if opts.Dir != "/tmp/site" {
		t.Errorf("expected Dir /tmp/site, got %s", opts.Dir)
	}
	if !opts.SPA {
		t.Errorf("expected SPA true")
	}
	if !opts.CORS {
		t.Errorf("expected CORS true")
	}
	if opts.QR {
		t.Errorf("expected QR false")
	}
	if !opts.Once {
		t.Errorf("expected Once true")
	}
}

func TestServeStartupAndRequest(t *testing.T) {
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "index.html")
	_ = os.WriteFile(indexPath, []byte("<h1>Nova Server Test</h1>"), 0644)

	ln, _, err := bindListener("127.0.0.1", 18000)
	if err != nil {
		t.Fatalf("bindListener failed: %v", err)
	}
	defer ln.Close()

	fs := http.FileServer(http.Dir(tmpDir))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fs.ServeHTTP(w, r)
	})

	srv := &http.Server{Handler: handler}
	go func() {
		_ = srv.Serve(ln)
	}()
	defer srv.Close()

	// Wait briefly for server startup
	time.Sleep(20 * time.Millisecond)

	url := "http://" + ln.Addr().String() + "/index.html"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("HTTP GET failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Nova Server Test") {
		t.Errorf("expected body to contain test string, got: %s", string(body))
	}
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected CORS header *")
	}
}

func TestRunCommandOnce(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte("<h1>Live</h1>"), 0644)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	caps := terminal.Capabilities{
		ColorProfile: terminal.ColorNone,
		Width:        80,
	}
	ctx := command.NewContext(nil, stdout, stderr, config.Config{}, caps, theme.Get("default"), output.ModePlain, nil)

	cmd := Command()
	if cmd.Name != "serve" {
		t.Errorf("expected command name serve, got %s", cmd.Name)
	}

	err := cmd.Run(ctx, []string{"--once", "--dir=" + tmpDir, "--no-qr"})
	if err != nil {
		t.Fatalf("expected nil error with --once, got %v", err)
	}
	if !strings.Contains(stdout.String(), "Serving:") {
		t.Errorf("expected Serving in startup banner, got: %s", stdout.String())
	}
}
