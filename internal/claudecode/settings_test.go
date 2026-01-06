package claudecode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultUserSettingsPath(t *testing.T) {
	path := DefaultUserSettingsPath()
	if path == "" {
		t.Error("DefaultUserSettingsPath returned empty string")
	}
	if filepath.Base(path) != "settings.json" {
		t.Errorf("Expected settings.json, got %s", filepath.Base(path))
	}
}

func TestDefaultProjectSettingsPath(t *testing.T) {
	path := DefaultProjectSettingsPath()
	if path != filepath.Join(".claude", "settings.json") {
		t.Errorf("Expected .claude/settings.json, got %s", path)
	}
}

func TestLoadSettings_NonExistent(t *testing.T) {
	settings, err := LoadSettings("/nonexistent/path/settings.json")
	if err != nil {
		t.Errorf("LoadSettings should not error on nonexistent file, got %v", err)
	}
	if settings == nil {
		t.Error("LoadSettings should return empty settings, not nil")
	}
	if len(settings) != 0 {
		t.Error("LoadSettings should return empty map for nonexistent file")
	}
}

func TestLoadSettings_ValidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")

	content := `{
  "mcpServers": {
    "test": {
      "command": "test-cmd"
    }
  }
}`
	if err := os.WriteFile(settingsPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test settings: %v", err)
	}

	settings, err := LoadSettings(settingsPath)
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}

	if _, ok := settings["mcpServers"]; !ok {
		t.Error("LoadSettings should preserve existing fields")
	}
}

func TestLoadSettings_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")

	if err := os.WriteFile(settingsPath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write test settings: %v", err)
	}

	_, err := LoadSettings(settingsPath)
	if err == nil {
		t.Error("LoadSettings should error on invalid JSON")
	}
}

func TestSettings_HasStatusLine(t *testing.T) {
	tests := []struct {
		name     string
		settings Settings
		expected bool
	}{
		{
			name:     "empty settings",
			settings: Settings{},
			expected: false,
		},
		{
			name: "has statusLine",
			settings: Settings{
				"statusLine": map[string]interface{}{
					"type":    "command",
					"command": "test",
				},
			},
			expected: true,
		},
		{
			name: "has other fields",
			settings: Settings{
				"mcpServers": map[string]interface{}{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.settings.HasStatusLine()
			if got != tt.expected {
				t.Errorf("HasStatusLine() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSettings_SetStatusLine(t *testing.T) {
	tests := []struct {
		name        string
		settings    Settings
		command     string
		force       bool
		expectError bool
	}{
		{
			name:        "set on empty settings",
			settings:    Settings{},
			command:     "/path/to/script.sh",
			force:       false,
			expectError: false,
		},
		{
			name: "set when statusLine exists without force",
			settings: Settings{
				"statusLine": map[string]interface{}{
					"type":    "command",
					"command": "existing",
				},
			},
			command:     "/path/to/script.sh",
			force:       false,
			expectError: true,
		},
		{
			name: "set when statusLine exists with force",
			settings: Settings{
				"statusLine": map[string]interface{}{
					"type":    "command",
					"command": "existing",
				},
			},
			command:     "/path/to/script.sh",
			force:       true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.settings.SetStatusLine(tt.command, tt.force)

			if tt.expectError {
				if err == nil {
					t.Error("SetStatusLine should return error")
				}
				if err != ErrStatusLineExists {
					t.Errorf("SetStatusLine should return ErrStatusLineExists, got %v", err)
				}
				return
			}

			if err != nil {
				t.Errorf("SetStatusLine failed: %v", err)
				return
			}

			sl, ok := tt.settings["statusLine"].(StatusLine)
			if !ok {
				t.Error("statusLine should be set as StatusLine type")
				return
			}
			if sl.Type != "command" {
				t.Errorf("statusLine.type = %q, want %q", sl.Type, "command")
			}
			if sl.Command != tt.command {
				t.Errorf("statusLine.command = %q, want %q", sl.Command, tt.command)
			}
		})
	}
}

func TestSaveSettings(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, ".claude", "settings.json")

	settings := Settings{
		"mcpServers": map[string]interface{}{
			"test": map[string]interface{}{
				"command": "test-cmd",
			},
		},
	}
	settings.SetStatusLine("/path/to/script.sh", false)

	if err := SaveSettings(settingsPath, settings); err != nil {
		t.Fatalf("SaveSettings failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(settingsPath); err != nil {
		t.Errorf("Settings file was not created: %v", err)
	}

	// Verify content is valid JSON
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Failed to read settings file: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("Saved settings is not valid JSON: %v", err)
	}

	// Verify statusLine was saved
	if _, ok := parsed["statusLine"]; !ok {
		t.Error("statusLine should be in saved settings")
	}

	// Verify existing fields preserved
	if _, ok := parsed["mcpServers"]; !ok {
		t.Error("mcpServers should be preserved in saved settings")
	}
}

func TestSaveSettings_CreatesDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "nested", "dirs", "settings.json")

	settings := Settings{}
	settings.SetStatusLine("/path/to/script.sh", false)

	if err := SaveSettings(settingsPath, settings); err != nil {
		t.Fatalf("SaveSettings should create directories: %v", err)
	}

	if _, err := os.Stat(settingsPath); err != nil {
		t.Errorf("Settings file was not created: %v", err)
	}
}
