package env

import (
	"bytes"
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

func TestEnv_SecretMasking(t *testing.T) {
	rawEnv := []string{
		"USER=testuser",
		"GITHUB_TOKEN=ghp_secret1234567890",
		"AWS_SECRET_ACCESS_KEY=verysecretkey",
		"GOPATH=/home/test/go",
	}

	// Default: mask secrets
	opts := ParseFlags([]string{})
	res := inspectEnvironment(rawEnv, opts)

	if res.Total != 4 {
		t.Fatalf("expected 4 variables, got %d", res.Total)
	}
	if res.SecretsCount != 2 {
		t.Fatalf("expected 2 secrets detected, got %d", res.SecretsCount)
	}

	for _, v := range res.Variables {
		if v.Key == "GITHUB_TOKEN" {
			if v.DisplayValue != "********" {
				t.Errorf("expected GITHUB_TOKEN to be masked, got: %s", v.DisplayValue)
			}
			if !v.IsSecret {
				t.Errorf("expected IsSecret to be true")
			}
		}
		if v.Key == "USER" && v.IsSecret {
			t.Errorf("USER should not be detected as secret")
		}
	}

	// Show secrets
	optsShow := ParseFlags([]string{"-s"})
	resShow := inspectEnvironment(rawEnv, optsShow)
	for _, v := range resShow.Variables {
		if v.Key == "GITHUB_TOKEN" && v.DisplayValue != "ghp_secret1234567890" {
			t.Errorf("expected revealed secret with -s, got: %s", v.DisplayValue)
		}
	}
}

func TestEnv_Filter(t *testing.T) {
	rawEnv := []string{
		"GO111MODULE=on",
		"GOROOT=/usr/local/go",
		"PATH=/usr/bin",
		"SHELL=/bin/bash",
	}

	opts := ParseFlags([]string{"-f", "GO"})
	res := inspectEnvironment(rawEnv, opts)

	if res.Total != 2 {
		t.Fatalf("expected 2 filtered variables, got %d", res.Total)
	}
	for _, v := range res.Variables {
		if !strings.HasPrefix(v.Key, "GO") {
			t.Errorf("unexpected key in filtered output: %s", v.Key)
		}
	}
}

func TestEnv_ExportShell(t *testing.T) {
	ctx, stdout, _ := newTestContext(output.ModeHuman)
	err := Run(ctx, []string{"--export=sh", "-f", "PATH"})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "export PATH=") {
		t.Errorf("expected 'export PATH=' in export output, got: %s", out)
	}

	// Test fish export
	ctxFish, stdoutFish, _ := newTestContext(output.ModeHuman)
	errFish := Run(ctxFish, []string{"--export=fish", "-f", "PATH"})
	if errFish != nil {
		t.Fatalf("fish export failed: %v", errFish)
	}
	outFish := stdoutFish.String()
	if !strings.Contains(outFish, "set -gx PATH ") {
		t.Errorf("expected 'set -gx PATH ' in fish export, got: %s", outFish)
	}
}

func TestEnv_PlainAndJSON(t *testing.T) {
	ctxPlain, stdoutPlain, _ := newTestContext(output.ModePlain)
	errPlain := Run(ctxPlain, []string{"--plain", "-f", "PATH"})
	if errPlain != nil {
		t.Fatalf("plain Run failed: %v", errPlain)
	}
	if !strings.Contains(stdoutPlain.String(), "PATH=") {
		t.Errorf("expected PATH= in plain output")
	}

	ctxJSON, stdoutJSON, _ := newTestContext(output.ModeJSON)
	errJSON := Run(ctxJSON, []string{"--json", "-f", "PATH"})
	if errJSON != nil {
		t.Fatalf("json Run failed: %v", errJSON)
	}
	if !strings.Contains(stdoutJSON.String(), `"variables":`) {
		t.Errorf("expected valid JSON structure")
	}
}
