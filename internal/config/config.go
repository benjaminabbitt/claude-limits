// Package config handles loading and parsing of the claude-limits configuration file.
package config

import (
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// Default format layouts
const (
	DefaultDatetimeFormat = "Mon, Jan 2 2006 at 3:04 PM MST"
	DefaultDateFormat     = "Mon, Jan 2 2006"
	DefaultTimeFormat     = "3:04 PM"
)

// FormatPreset contains the format strings for a named preset
type FormatPreset struct {
	Datetime string
	Date     string
	Time     string
}

// Presets maps preset names to their format configurations
var Presets = map[string]FormatPreset{
	"12hour": {
		Datetime: "Mon, Jan 2 2006 at 3:04 PM MST",
		Date:     "Mon, Jan 2 2006",
		Time:     "3:04 PM",
	},
	"24hour": {
		Datetime: "Mon, Jan 2 2006 at 15:04 MST",
		Date:     "Mon, Jan 2 2006",
		Time:     "15:04",
	},
	"iso8601": {
		Datetime: "2006-01-02T15:04:05Z07:00",
		Date:     "2006-01-02",
		Time:     "15:04:05",
	},
	"us": {
		Datetime: "Jan 2, 2006 3:04 PM MST",
		Date:     "Jan 2, 2006",
		Time:     "3:04 PM",
	},
	"eu": {
		Datetime: "2 Jan 2006 15:04 MST",
		Date:     "2 Jan 2006",
		Time:     "15:04",
	},
}

// Auth contains authentication configuration
type Auth struct {
	SessionCookie string `yaml:"session_cookie"`
	OrgID         string `yaml:"org_id"`
}

// Formats contains display format configuration
type Formats struct {
	Preset   string `yaml:"preset"`
	Datetime string `yaml:"datetime"`
	Date     string `yaml:"date"`
	Time     string `yaml:"time"`
}

// Config represents the full configuration file
type Config struct {
	Auth    Auth    `yaml:"auth"`
	Formats Formats `yaml:"formats"`
}

// ResolvedFormats returns the effective format strings, applying preset then overrides
func (c *Config) ResolvedFormats() FormatPreset {
	result := FormatPreset{
		Datetime: DefaultDatetimeFormat,
		Date:     DefaultDateFormat,
		Time:     DefaultTimeFormat,
	}

	// Apply preset if specified
	if c.Formats.Preset != "" {
		if preset, ok := Presets[c.Formats.Preset]; ok {
			result = preset
		}
	}

	// Apply individual overrides
	if c.Formats.Datetime != "" {
		result.Datetime = c.Formats.Datetime
	}
	if c.Formats.Date != "" {
		result.Date = c.Formats.Date
	}
	if c.Formats.Time != "" {
		result.Time = c.Formats.Time
	}

	return result
}

// DefaultPath returns the default configuration file path for the current OS
func DefaultPath() string {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		configDir = os.Getenv("APPDATA")
		if configDir == "" {
			configDir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
	default:
		// Linux, macOS, and others use XDG
		configDir = os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return ""
			}
			configDir = filepath.Join(home, ".config")
		}
	}

	return filepath.Join(configDir, "claude-limits", "config.yaml")
}

// Load reads and parses the configuration file from the given path.
// If path is empty, it uses the default path.
// Returns an empty config (not an error) if the file doesn't exist.
func Load(path string) (*Config, error) {
	if path == "" {
		// Check environment variable first
		path = os.Getenv("CLAUDE_LIMITS_CONFIG")
	}
	if path == "" {
		path = DefaultPath()
	}

	cfg := &Config{}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No config file is not an error
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadOrDefault loads config, returning default config on any error
func LoadOrDefault(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		return &Config{}
	}
	return cfg
}
