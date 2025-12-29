package fuzzy

import (
	"testing"
)

func TestExpandNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"5", "five"},
		{"24", "twentyfour"},
		{"5h", "fiveh"},
		{"24hour", "twentyfourhour"},
		{"", ""},
		{"no numbers", "no numbers"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ExpandNumbers(tt.input)
			if result != tt.expected {
				t.Errorf("ExpandNumbers(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestScore(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		target   string
		minScore int
		maxScore int
	}{
		{"exact match", "five_hour", "five_hour", 1000, 1000},
		{"exact match normalized", "fivehour", "five_hour", 1000, 1000},
		{"contains query", "hour", "five_hour_utilization", 500, 700},
		{"suffix match bonus", "utilization", "five_hour_utilization", 600, 700},
		{"partial char match", "5h", "five_hour", 500, 600}, // "5h" -> "fiveh" contains in "fivehour"
		{"no match", "xyz", "five_hour", 0, 0},
		{"empty query", "", "five_hour", 0, 0},
		{"number expansion", "5", "five_hour", 500, 700},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := Score(tt.query, tt.target)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("Score(%q, %q) = %d, want between %d and %d",
					tt.query, tt.target, score, tt.minScore, tt.maxScore)
			}
		})
	}
}

func TestFlattenData(t *testing.T) {
	data := map[string]interface{}{
		"simple": "value",
		"nested": map[string]interface{}{
			"inner": 42,
		},
		"array": []interface{}{
			map[string]interface{}{"item": "first"},
			map[string]interface{}{"item": "second"},
		},
		"nil_value": nil,
	}

	pairs := FlattenData(data, "")

	// Check we got the expected number of pairs (excluding nil)
	if len(pairs) != 4 {
		t.Errorf("FlattenData returned %d pairs, want 4", len(pairs))
	}

	// Verify specific paths exist
	paths := make(map[string]bool)
	for _, p := range pairs {
		paths[p.Path] = true
	}

	expectedPaths := []string{"simple", "nested_inner", "array_1_item", "array_2_item"}
	for _, path := range expectedPaths {
		if !paths[path] {
			t.Errorf("Expected path %q not found in flattened data", path)
		}
	}
}

func TestFindBestMatch(t *testing.T) {
	pairs := []KeyValue{
		{Path: "five_hour_utilization", Key: "utilization", Value: 75.5},
		{Path: "weekly_limit", Key: "limit", Value: 100},
		{Path: "context_window_utilization", Key: "utilization", Value: 50.0},
	}

	tests := []struct {
		query       string
		expectPath  string
		expectError bool
	}{
		{"five", "five_hour_utilization", false},
		{"5h", "five_hour_utilization", false},
		{"weekly", "weekly_limit", false},
		{"context", "context_window_utilization", false},
		{"nonexistent", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			match, err := FindBestMatch(pairs, tt.query)
			if tt.expectError {
				if err == nil {
					t.Errorf("FindBestMatch(%q) expected error, got nil", tt.query)
				}
				return
			}
			if err != nil {
				t.Errorf("FindBestMatch(%q) unexpected error: %v", tt.query, err)
				return
			}
			if match.Path != tt.expectPath {
				t.Errorf("FindBestMatch(%q) = %q, want %q", tt.query, match.Path, tt.expectPath)
			}
		})
	}
}
