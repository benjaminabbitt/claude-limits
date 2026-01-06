package claudecode

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ErrStatusLineExists indicates the statusLine field already exists in settings
var ErrStatusLineExists = errors.New("statusLine already configured in settings")

// StatusLine represents the statusLine configuration in Claude Code settings
type StatusLine struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// Settings represents the Claude Code settings.json structure
// Uses map to preserve unknown fields
type Settings map[string]interface{}

// DefaultUserSettingsPath returns the default path to user-level Claude Code settings
func DefaultUserSettingsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "settings.json")
}

// DefaultProjectSettingsPath returns the path to project-level Claude Code settings
func DefaultProjectSettingsPath() string {
	return filepath.Join(".claude", "settings.json")
}

// LoadSettings reads Claude Code settings from the given path
// Returns empty Settings if file doesn't exist
func LoadSettings(path string) (Settings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(Settings), nil
		}
		return nil, fmt.Errorf("failed to read settings: %w", err)
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings: %w", err)
	}

	return settings, nil
}

// HasStatusLine checks if the settings already contain a statusLine configuration
func (s Settings) HasStatusLine() bool {
	_, exists := s["statusLine"]
	return exists
}

// SetStatusLine sets the statusLine configuration
// Returns ErrStatusLineExists if statusLine already exists and force is false
func (s Settings) SetStatusLine(command string, force bool) error {
	if s.HasStatusLine() && !force {
		return ErrStatusLineExists
	}

	s["statusLine"] = StatusLine{
		Type:    "command",
		Command: command,
	}
	return nil
}

// SaveSettings writes the settings to the given path
// Creates parent directories if they don't exist
func SaveSettings(path string, settings Settings) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create settings directory: %w", err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Add trailing newline
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings: %w", err)
	}

	return nil
}
