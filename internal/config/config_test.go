package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultPath(t *testing.T) {
	path := DefaultPath()
	if path == "" {
		t.Error("DefaultPath returned empty string")
	}
	// Should end with config.yaml
	if filepath.Base(path) != "config.yaml" {
		t.Errorf("DefaultPath should end with config.yaml, got %s", path)
	}
}

func TestLoadNonExistent(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Errorf("Load should not error on nonexistent file, got %v", err)
	}
	if cfg == nil {
		t.Error("Load should return empty config, not nil")
	}
}

func TestLoadValidConfig(t *testing.T) {
	// Create a temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `
formats:
  preset: "24hour"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Formats.Preset != "24hour" {
		t.Errorf("Expected preset '24hour', got '%s'", cfg.Formats.Preset)
	}
}

func TestResolvedFormatsDefault(t *testing.T) {
	cfg := &Config{}
	fmts := cfg.ResolvedFormats()

	if fmts.Datetime != DefaultDatetimeFormat {
		t.Errorf("Expected default datetime format, got '%s'", fmts.Datetime)
	}
	if fmts.Date != DefaultDateFormat {
		t.Errorf("Expected default date format, got '%s'", fmts.Date)
	}
	if fmts.Time != DefaultTimeFormat {
		t.Errorf("Expected default time format, got '%s'", fmts.Time)
	}
}

func TestResolvedFormatsPreset(t *testing.T) {
	tests := []struct {
		preset   string
		datetime string
		date     string
		time     string
	}{
		{"12hour", "Mon, Jan 2 2006 at 3:04 PM MST", "Mon, Jan 2 2006", "3:04 PM"},
		{"24hour", "Mon, Jan 2 2006 at 15:04 MST", "Mon, Jan 2 2006", "15:04"},
		{"iso8601", "2006-01-02T15:04:05Z07:00", "2006-01-02", "15:04:05"},
		{"us", "Jan 2, 2006 3:04 PM MST", "Jan 2, 2006", "3:04 PM"},
		{"eu", "2 Jan 2006 15:04 MST", "2 Jan 2006", "15:04"},
	}

	for _, tt := range tests {
		t.Run(tt.preset, func(t *testing.T) {
			cfg := &Config{Formats: Formats{Preset: tt.preset}}
			fmts := cfg.ResolvedFormats()

			if fmts.Datetime != tt.datetime {
				t.Errorf("Expected datetime '%s', got '%s'", tt.datetime, fmts.Datetime)
			}
			if fmts.Date != tt.date {
				t.Errorf("Expected date '%s', got '%s'", tt.date, fmts.Date)
			}
			if fmts.Time != tt.time {
				t.Errorf("Expected time '%s', got '%s'", tt.time, fmts.Time)
			}
		})
	}
}

func TestResolvedFormatsCustomOverride(t *testing.T) {
	cfg := &Config{
		Formats: Formats{
			Preset:   "24hour",
			Datetime: "custom-datetime",
			// Date and Time not set, should use preset values
		},
	}
	fmts := cfg.ResolvedFormats()

	if fmts.Datetime != "custom-datetime" {
		t.Errorf("Custom datetime should override preset, got '%s'", fmts.Datetime)
	}
	// Date and Time should come from 24hour preset
	if fmts.Date != "Mon, Jan 2 2006" {
		t.Errorf("Date should use preset value, got '%s'", fmts.Date)
	}
	if fmts.Time != "15:04" {
		t.Errorf("Time should use preset value, got '%s'", fmts.Time)
	}
}

func TestResolvedFormatsUnknownPreset(t *testing.T) {
	cfg := &Config{Formats: Formats{Preset: "unknown"}}
	fmts := cfg.ResolvedFormats()

	// Should fall back to defaults
	if fmts.Datetime != DefaultDatetimeFormat {
		t.Errorf("Unknown preset should use default datetime, got '%s'", fmts.Datetime)
	}
}

func TestLoadOrDefault(t *testing.T) {
	cfg := LoadOrDefault("/nonexistent/path.yaml")
	if cfg == nil {
		t.Error("LoadOrDefault should never return nil")
	}
}

func TestLoadFromEnvVar(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `
formats:
  preset: "eu"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Set env var
	os.Setenv("CLAUDE_LIMITS_CONFIG", configPath)
	defer os.Unsetenv("CLAUDE_LIMITS_CONFIG")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Formats.Preset != "eu" {
		t.Errorf("Expected preset 'eu', got '%s'", cfg.Formats.Preset)
	}
}
