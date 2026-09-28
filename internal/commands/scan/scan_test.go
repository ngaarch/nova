package scancmd

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func newTestContext(stdin string, stdout, stderr *bytes.Buffer) *command.Context {
	cfg := config.Config{
		Theme:     "default",
		ColorMode: "never",
	}
	caps := terminal.Capabilities{
		Width:        80,
		Height:       24,
		ColorProfile: terminal.ColorNone,
		IsTTY:        false,
	}
	th := theme.Get("default")
	logger := logging.New(stderr, false)
	return command.NewContext(
		strings.NewReader(stdin),
		stdout,
		stderr,
		cfg,
		caps,
		th,
		output.ModePlain,
		logger,
	)
}

func TestParseFlags(t *testing.T) {
	opts := ParseFlags([]string{"./src", "-p", "aws-key", "-e", "4.8", "-i", "vendor,tmp"})
	if opts.Target != "./src" {
		t.Errorf("expected Target ./src, got %q", opts.Target)
	}
	if opts.Rule != "aws-key" {
		t.Errorf("expected Rule aws-key, got %q", opts.Rule)
	}
	if opts.MinEntropy != 4.8 {
		t.Errorf("expected MinEntropy 4.8, got %f", opts.MinEntropy)
	}
}

func TestCalculateShannonEntropy(t *testing.T) {
	if CalculateShannonEntropy("") != 0 {
		t.Errorf("expected 0 entropy for empty string")
	}
	// "aaaaaaaa" has 0 entropy
	if ent := CalculateShannonEntropy("aaaaaaaa"); ent != 0 {
		t.Errorf("expected 0 entropy for repetitive string, got %f", ent)
	}
	// High entropy random string
	entHigh := CalculateShannonEntropy("aB9#xL2$mQ8*zP1@wK7&")
	if entHigh < 4.0 {
		t.Errorf("expected high entropy > 4.0, got %f", entHigh)
	}
}

func TestMaskSecret(t *testing.T) {
	token := "ghp_" + "TestDummyTokenForScannerTesting12345"
	masked := MaskSecret(token)
	if !strings.HasPrefix(masked, "ghp_Te") || !strings.HasSuffix(masked, "2345") || !strings.Contains(masked, "****...") {
		t.Errorf("unexpected masked output: %q", masked)
	}
	if strings.Contains(masked, "ScannerTesting") {
		t.Errorf("masked secret should not reveal middle content: %s", masked)
	}
}

func TestExecuteScan_DetectsSecrets(t *testing.T) {
	tmpDir := t.TempDir()

	// File with secrets
	secretFile := filepath.Join(tmpDir, "config.go")
	dummyGH := "ghp_" + "TestDummyTokenForScannerTesting12345"
	content := fmt.Sprintf(`package config

const (
    GitHubToken = %q
    AWSAccess   = "AKIAIOSFODNN7EXAMPLE"
    DummyPass   = "password = \"super_secret_master_key_123\""
)
`, dummyGH)
	if err := os.WriteFile(secretFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Clean file
	cleanFile := filepath.Join(tmpDir, "clean.go")
	cleanContent := `package clean

func Hello() string {
    return "Hello, world!"
}
`
	if err := os.WriteFile(cleanFile, []byte(cleanContent), 0644); err != nil {
		t.Fatal(err)
	}

	opts := Options{
		Target:      tmpDir,
		MinEntropy:  4.5,
		IgnorePaths: []string{".git"},
	}

	res, err := ExecuteScan(opts)
	if err != nil {
		t.Fatalf("ExecuteScan failed: %v", err)
	}

	if res.FilesScanned < 2 {
		t.Errorf("expected at least 2 files scanned, got %d", res.FilesScanned)
	}

	if res.TotalFindings < 2 {
		t.Fatalf("expected at least 2 findings, got %d", res.TotalFindings)
	}

	hasGitHub := false
	hasAWS := false
	for _, f := range res.Findings {
		if f.RuleID == "github-pat" {
			hasGitHub = true
		}
		if f.RuleID == "aws-key" {
			hasAWS = true
		}
	}

	if !hasGitHub || !hasAWS {
		t.Errorf("expected github-pat and aws-key to be found: %+v", res.Findings)
	}
}

func TestRunCommand_PlainAndJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "keys.env")
	content := "AWS_KEY=AKIAIOSFODNN7EXAMPLE\n"
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	ctx := newTestContext("", &stdout, &stderr)

	// 1. Plain Mode
	err := Run(ctx, []string{filePath, "--plain"})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "aws-key") || !strings.Contains(out, "AKIA") {
		t.Errorf("expected aws-key in plain output: %s", out)
	}

	// 2. JSON Mode
	stdout.Reset()
	stderr.Reset()
	err = Run(ctx, []string{filePath, "--json"})
	if err != nil {
		t.Fatalf("Run failed for JSON: %v", err)
	}
	var res ScanResult
	if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode JSON: %v; raw: %s", err, stdout.String())
	}
	if res.TotalFindings != 1 || res.Findings[0].RuleID != "aws-key" {
		t.Errorf("unexpected JSON scan result: %+v", res)
	}

	// 3. Clean target dashboard
	stdout.Reset()
	stderr.Reset()
	cleanDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(cleanDir, "ok.txt"), []byte("All good!"), 0644)
	ctx.Printer.Mode = output.ModeHuman
	err = Run(ctx, []string{cleanDir})
	if err != nil {
		t.Fatalf("Run failed on clean target: %v", err)
	}
	if !strings.Contains(stdout.String(), "All clean") {
		t.Errorf("expected All clean in dashboard output: %s", stdout.String())
	}
}
