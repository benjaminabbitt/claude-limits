package format

import (
	"testing"
)

func TestFormatKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"five_hour_utilization", "Five Hour Utilization"},
		{"simple", "Simple"},
		{"", ""},
		{"already_Title_Case", "Already Title Case"},
		{"with_numbers_123", "With Numbers 123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := FormatKey(tt.input)
			if result != tt.expected {
				t.Errorf("FormatKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetUtilizationColor(t *testing.T) {
	colors := Colors{
		Green:  "green",
		Yellow: "yellow",
		Red:    "red",
	}

	tests := []struct {
		value    float64
		expected string
	}{
		{0, "green"},
		{50, "green"},
		{79.9, "green"},
		{80, "yellow"},
		{90, "yellow"},
		{94.9, "yellow"},
		{95, "red"},
		{100, "red"},
	}

	for _, tt := range tests {
		result := GetUtilizationColor(tt.value, colors)
		if result != tt.expected {
			t.Errorf("GetUtilizationColor(%v) = %q, want %q", tt.value, result, tt.expected)
		}
	}
}

func TestFormatNumber(t *testing.T) {
	colors := Colors{
		Green: "\033[32m",
		Reset: "\033[0m",
	}
	noColors := Colors{}

	tests := []struct {
		value       float64
		key         string
		colors      Colors
		expectColor bool
	}{
		{75.5, "utilization", colors, true},
		{75.5, "usage_percent", colors, true},
		{75.5, "count", colors, false},
		{75.5, "utilization", noColors, false},
		{100, "limit", colors, false},
	}

	for _, tt := range tests {
		result := FormatNumber(tt.value, tt.key, tt.colors)
		hasColor := len(result) > 10 // Color codes add length
		if tt.expectColor && !hasColor {
			t.Errorf("FormatNumber(%v, %q) expected color, got %q", tt.value, tt.key, result)
		}
		if !tt.expectColor && hasColor {
			t.Errorf("FormatNumber(%v, %q) unexpected color in %q", tt.value, tt.key, result)
		}
	}
}

func TestFormatString(t *testing.T) {
	tests := []struct {
		value       string
		key         string
		shouldParse bool // whether it should be parsed as datetime
	}{
		// Non-datetime fields pass through
		{"hello", "name", false},
		{"world", "description", false},
		// Datetime fields get formatted (check it's different from input)
		{"2024-01-15T10:30:00Z", "created_at", true},
		// Invalid datetime passes through
		{"not-a-date", "reset_at", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := FormatString(tt.value, tt.key)
			if tt.shouldParse {
				// Should be formatted differently from input
				if result == tt.value {
					t.Errorf("FormatString(%q, %q) should parse datetime, got %q", tt.value, tt.key, result)
				}
				// Should contain human-readable elements
				if !containsAny(result, []string{"Jan", "2024", "AM", "PM"}) {
					t.Errorf("FormatString(%q, %q) = %q, doesn't look like formatted date", tt.value, tt.key, result)
				}
			} else {
				// Should pass through unchanged
				if result != tt.value {
					t.Errorf("FormatString(%q, %q) = %q, want %q (unchanged)", tt.value, tt.key, result, tt.value)
				}
			}
		})
	}
}

func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

func TestIsTerminal(t *testing.T) {
	// Just verify it doesn't panic
	_ = IsTerminal()
}

func TestNewColors(t *testing.T) {
	// With color
	c := NewColors(false)
	// Can't test actual terminal detection, but verify structure
	_ = c

	// Without color
	c = NewColors(true)
	if c.Bold != "" || c.Reset != "" {
		t.Error("NewColors(true) should return empty colors")
	}
}
