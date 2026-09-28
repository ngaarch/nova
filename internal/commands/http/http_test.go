package httpcmd

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nova/internal/command"
	"nova/internal/config"
	"nova/internal/logging"
	"nova/internal/output"
	"nova/internal/terminal"
	"nova/internal/theme"
)

func newTestContext(mode output.Mode) (*command.Context, *bytes.Buffer, *bytes.Buffer) {
	var stdout, stderr bytes.Buffer
	in := strings.NewReader("")
	cfg := config.Default()
	caps := terminal.Capabilities{
		IsTTY:            mode == output.ModeHuman,
		Width:            80,
		Height:           24,
		ColorProfile:     terminal.ColorTrueColor,
		UnicodeSupported: true,
	}
	th := theme.Get("default")
	logger := logging.New(&stderr, false)
	ctx := command.NewContext(in, &stdout, &stderr, cfg, caps, th, mode, logger)
	return ctx, &stdout, &stderr
}

func TestHTTP_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"status":"ok","message":"hello nova"}`)
	}))
	defer ts.Close()

	ctx, stdout, _ := newTestContext(output.ModePlain)
	err := Run(ctx, []string{"--plain", ts.URL})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, `"status":"ok"`) {
		t.Errorf("expected json body in output, got: %s", out)
	}
}

func TestHTTP_PostWithHeaders(t *testing.T) {
	var receivedHeader string
	var receivedBody string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-Custom-Header")
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r.Body)
		receivedBody = buf.String()
		w.WriteHeader(http.StatusCreated)
		_, _ = fmt.Fprint(w, "created")
	}))
	defer ts.Close()

	ctx, stdout, _ := newTestContext(output.ModeHuman)
	err := Run(ctx, []string{
		"-X", "POST",
		"-H", "X-Custom-Header: NovaClient",
		"-d", "sample payload",
		ts.URL,
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if receivedHeader != "NovaClient" {
		t.Errorf("expected X-Custom-Header 'NovaClient', got: %q", receivedHeader)
	}
	if receivedBody != "sample payload" {
		t.Errorf("expected body 'sample payload', got: %q", receivedBody)
	}
	if !strings.Contains(stdout.String(), "201") {
		t.Errorf("expected 201 status in output, got: %s", stdout.String())
	}
}

func TestHTTP_OutputFile(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "saved content")
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "resp.txt")

	ctx, stdout, _ := newTestContext(output.ModeHuman)
	err := Run(ctx, []string{"-o", outFile, ts.URL})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if string(content) != "saved content" {
		t.Errorf("expected 'saved content' in file, got: %q", string(content))
	}
	if !strings.Contains(stdout.String(), outFile) {
		t.Errorf("expected saved path in stdout, got: %s", stdout.String())
	}
}

func TestHTTP_JSONOutput(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "ping")
	}))
	defer ts.Close()

	ctx, stdout, _ := newTestContext(output.ModeJSON)
	err := Run(ctx, []string{"--json", ts.URL})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, `"status_code":`) || !strings.Contains(out, `"telemetry":`) {
		t.Errorf("expected JSON telemetry, got: %s", out)
	}
}

func TestHTTP_MissingURL(t *testing.T) {
	ctx, _, _ := newTestContext(output.ModePlain)
	err := Run(ctx, []string{})
	if err == nil {
		t.Fatalf("expected error for missing URL, got nil")
	}
}
