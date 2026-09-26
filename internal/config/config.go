package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config encapsulates persistent and runtime configuration preferences for nova.
type Config struct {
	Theme      string `json:"theme"`
	ColorMode  string `json:"color_mode"`  // "auto", "always", "never"
	IconMode   string `json:"icon_mode"`   // "auto", "always", "never"
	OutputMode string `json:"output_mode"` // "auto", "human", "plain", "json"
	Debug      bool   `json:"debug"`
}

// Default returns a Config populated with sensible default settings.
func Default() Config {
	return Config{
		Theme:      "default",
		ColorMode:  "auto",
		IconMode:   "auto",
		OutputMode: "auto",
		Debug:      false,
	}
}

// Load resolves configuration by combining built-in defaults, config file settings,
// and environment variable overrides.
func Load() (Config, error) {
	return LoadWithEnv(os.Getenv)
}

// LoadWithEnv resolves configuration using an injected environment variable lookup.
func LoadWithEnv(getenv func(string) string) (Config, error) {
	cfg := Default()

	// 1. Locate config file
	configPath := getenv("NOVA_CONFIG")
	if configPath == "" {
		if xdg := getenv("XDG_CONFIG_HOME"); xdg != "" {
			configPath = filepath.Join(xdg, "nova", "config.json")
		} else if home := getenv("HOME"); home != "" {
			configPath = filepath.Join(home, ".config", "nova", "config.json")
		}
	}

	// 2. Load from file if present
	if configPath != "" {
		if data, err := os.ReadFile(configPath); err == nil {
			var fileCfg Config
			if err := json.Unmarshal(data, &fileCfg); err != nil {
				return cfg, fmt.Errorf("parsing config file %q: %w", configPath, err)
			}
			cfg.merge(fileCfg)
		} else if !os.IsNotExist(err) && getenv("NOVA_CONFIG") != "" {
			return cfg, fmt.Errorf("reading specified config file %q: %w", configPath, err)
		}
	}

	// 3. Environment variable overrides
	if theme := getenv("NOVA_THEME"); theme != "" {
		cfg.Theme = theme
	}
	if color := getenv("NOVA_COLOR"); color != "" {
		cfg.ColorMode = strings.ToLower(color)
	}
	if icons := getenv("NOVA_ICONS"); icons != "" {
		cfg.IconMode = strings.ToLower(icons)
	}
	if out := getenv("NOVA_OUTPUT"); out != "" {
		cfg.OutputMode = strings.ToLower(out)
	}
	if dbg := getenv("NOVA_DEBUG"); dbg == "1" || strings.EqualFold(dbg, "true") {
		cfg.Debug = true
	}
	if noColor := getenv("NO_COLOR"); noColor != "" {
		cfg.ColorMode = "never"
	}

	return cfg, nil
}

// merge updates non-zero fields from incoming config.
func (c *Config) merge(other Config) {
	if other.Theme != "" {
		c.Theme = other.Theme
	}
	if other.ColorMode != "" {
		c.ColorMode = other.ColorMode
	}
	if other.IconMode != "" {
		c.IconMode = other.IconMode
	}
	if other.OutputMode != "" {
		c.OutputMode = other.OutputMode
	}
	if other.Debug {
		c.Debug = true
	}
}
