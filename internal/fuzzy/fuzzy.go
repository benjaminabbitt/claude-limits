// Package fuzzy provides fuzzy string matching for usage field queries.
package fuzzy

import (
	"fmt"
	"sort"
	"strings"

	apierrors "github.com/benjaminabbitt/claude-limits/internal/errors"
)

// KeyValue represents a flattened key-value pair from JSON data
type KeyValue struct {
	Path  string // Full path, e.g., "five_hour_utilization"
	Key   string // Leaf key only
	Value interface{}
}

// numberWords maps arabic numerals to english words
var numberWords = map[string]string{
	"0": "zero", "1": "one", "2": "two", "3": "three", "4": "four",
	"5": "five", "6": "six", "7": "seven", "8": "eight", "9": "nine",
	"10": "ten", "11": "eleven", "12": "twelve", "13": "thirteen",
	"14": "fourteen", "15": "fifteen", "16": "sixteen", "17": "seventeen",
	"18": "eighteen", "19": "nineteen", "20": "twenty", "24": "twentyfour",
	"30": "thirty", "48": "fortyeight", "72": "seventytwo",
}

// numberWordKeys returns keys sorted by length descending (longer numbers first)
// to ensure "24" is replaced before "2" and "4"
var numberWordKeys = func() []string {
	keys := make([]string, 0, len(numberWords))
	for k := range numberWords {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return len(keys[i]) > len(keys[j])
	})
	return keys
}()

// ExpandNumbers converts arabic numerals to english words in a string
func ExpandNumbers(s string) string {
	result := s
	for _, num := range numberWordKeys {
		result = strings.ReplaceAll(result, num, numberWords[num])
	}
	return result
}

// Score returns a score for how well the query matches the target.
// Higher is better, 0 means no match.
func Score(query, target string) int {
	if query == "" {
		return 0
	}

	// Normalize both: expand numbers and remove underscores
	normalizedQuery := strings.ReplaceAll(ExpandNumbers(query), "_", "")
	normalizedTarget := strings.ReplaceAll(target, "_", "")

	// Exact match gets highest score
	if normalizedTarget == normalizedQuery {
		return 1000
	}

	// Contains the full query - bonus if it ends with the query
	if strings.Contains(normalizedTarget, normalizedQuery) {
		score := 500 + len(normalizedQuery)
		if strings.HasSuffix(normalizedTarget, normalizedQuery) {
			score += 100
		}
		return score
	}

	// Check if all query chars appear in order
	targetRunes := []rune(normalizedTarget)
	score := 0
	targetIdx := 0
	for _, qChar := range normalizedQuery {
		found := false
		for targetIdx < len(targetRunes) {
			if targetRunes[targetIdx] == qChar {
				score += 10
				// Bonus for matching at word boundaries
				if targetIdx == 0 {
					score += 5
				}
				targetIdx++
				found = true
				break
			}
			targetIdx++
		}
		if !found {
			return 0 // Query char not found, no match
		}
	}

	return score
}

// FlattenData recursively flattens nested JSON data into key-value pairs
func FlattenData(data map[string]interface{}, prefix string) []KeyValue {
	var pairs []KeyValue

	for key, value := range data {
		path := key
		if prefix != "" {
			path = prefix + "_" + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			pairs = append(pairs, FlattenData(v, path)...)
		case []interface{}:
			for i, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					indexedPath := fmt.Sprintf("%s_%d", path, i+1)
					pairs = append(pairs, FlattenData(m, indexedPath)...)
				}
			}
		case nil:
			// Skip nil values
		default:
			pairs = append(pairs, KeyValue{Path: path, Key: key, Value: value})
		}
	}

	return pairs
}

// FindBestMatch finds the best matching field for a query
func FindBestMatch(pairs []KeyValue, query string) (*KeyValue, error) {
	queryLower := strings.ToLower(query)
	var bestMatch *KeyValue
	bestScore := 0

	for i := range pairs {
		score := Score(queryLower, strings.ToLower(pairs[i].Path))
		if score > bestScore {
			bestScore = score
			bestMatch = &pairs[i]
		}
	}

	if bestMatch == nil || bestScore == 0 {
		return nil, apierrors.NewQueryError(query, apierrors.ErrNoMatch)
	}

	return bestMatch, nil
}
