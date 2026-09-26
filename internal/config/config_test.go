package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	if cfg.Theme != "default" {
		t.Errorf("expected default theme, got %q", cfg.Theme)
	}
	if cfg.ColorMode != "auto" {
		t.Errorf("expected auto color mode, got %q", cfg.ColorMode)
	}
	if cfg.IconMode != "auto" {
		t.Errorf("expected auto icon mode, got %q", cfg.IconMode)
	}
	if cfg.OutputMode != "auto" {
		t.Errorf("expected auto output mode, got %q", cfg.OutputMode)
	}
	if cfg.Debug {
		t.Errorf("expected debug = false by default")
	}
}

func TestLoadConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "config.json")
	content := `{
		"theme": "nord",
		"color_mode": "always",
		"icon_mode": "never",
		"output_mode": "plain",
		"debug": true
	}`
	if err := os.WriteFile(confPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	getenv := func(key string) string {
		if key == "NOVA_CONFIG" {
			return confPath
		}
		return ""
	}

	cfg, err := LoadWithEnv(getenv)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.Theme != "nord" {
		t.Errorf("expected theme nord, got %q", cfg.Theme)
	}
	if cfg.ColorMode != "always" {
		t.Errorf("expected color_mode always, got %q", cfg.ColorMode)
	}
	if cfg.IconMode != "never" {
		t.Errorf("expected icon_mode never, got %q", cfg.IconMode)
	}
	if cfg.OutputMode != "plain" {
		t.Errorf("expected output_mode plain, got %q", cfg.OutputMode)
	}
	if !cfg.Debug {
		t.Errorf("expected debug = true")
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	env := map[string]string{
		"NOVA_THEME":  "dracula",
		"NOVA_COLOR":  "never",
		"NOVA_ICONS":  "always",
		"NOVA_OUTPUT": "json",
		"NOVA_DEBUG":  "1",
	}
	getenv := func(key string) string {
		return env[key]
	}

	cfg, err := LoadWithEnv(getenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Theme != "dracula" {
		t.Errorf("expected theme dracula, got %q", cfg.Theme)
	}
	if cfg.ColorMode != "never" {
		t.Errorf("expected color_mode never, got %q", cfg.ColorMode)
	}
	if cfg.IconMode != "always" {
		t.Errorf("expected icon_mode always, got %q", cfg.IconMode)
	}
	if cfg.OutputMode != "json" {
		t.Errorf("expected output_mode json, got %q", cfg.OutputMode)
	}
	if !cfg.Debug {
		t.Errorf("expected debug = true")
	}
}

func TestNoColorEnvOverride(t *testing.T) {
	env := map[string]string{
		"NO_COLOR": "1",
	}
	getenv := func(key string) string {
		return env[key]
	}

	cfg, err := LoadWithEnv(getenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ColorMode != "never" {
		t.Errorf("expected color_mode never when NO_COLOR set, got %q", cfg.ColorMode)
	}
}

func TestInvalidConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "bad_config.json")
	if err := os.WriteFile(confPath, []byte("invalid-json"), 0o600); err != nil {
		t.Fatalf("failed to write bad config: %v", err)
	}

	getenv := func(key string) string {
		if key == "NOVA_CONFIG" {
			return confPath
		}
		return ""
	}

	_, err := LoadWithEnv(getenv)
	if err == nil {
		t.Fatalf("expected error loading invalid JSON, got nil")
	}
}
