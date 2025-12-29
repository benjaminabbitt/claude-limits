// Package format provides output formatting for usage data.
package format

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/benjaminabbitt/claude-limits/internal/models"
)

// ANSI color codes
const (
	Bold   = "\033[1m"
	Cyan   = "\033[36m"
	Yellow = "\033[33m"
	Green  = "\033[32m"
	Red    = "\033[31m"
	Reset  = "\033[0m"
)

// Colors holds the color configuration for output
type Colors struct {
	Bold   string
	Cyan   string
	Yellow string
	Green  string
	Red    string
	Reset  string
}

// Formats holds the configurable date/time format strings
type Formats struct {
	Datetime string // Format for full datetime (e.g., "Mon, Jan 2 2006 at 3:04 PM MST")
	Date     string // Format for date only (e.g., "Mon, Jan 2 2006")
	Time     string // Format for time only (e.g., "3:04 PM")
}

// DefaultFormats returns the default format configuration
func DefaultFormats() Formats {
	return Formats{
		Datetime: "Mon, Jan 2 2006 at 3:04 PM MST",
		Date:     "Mon, Jan 2 2006",
		Time:     "3:04 PM",
	}
}

// NewColors creates a Colors configuration based on terminal and user preferences
func NewColors(noColor bool) Colors {
	if !IsTerminal() || noColor {
		return Colors{}
	}
	return Colors{
		Bold:   Bold,
		Cyan:   Cyan,
		Yellow: Yellow,
		Green:  Green,
		Red:    Red,
		Reset:  Reset,
	}
}

// IsTerminal returns true if stdout is a terminal
func IsTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// JSON formats usage data as indented JSON
func JSON(usage *models.Usage) (string, error) {
	return usage.ToJSON()
}

// Table formats usage data as a human-readable table
func Table(usage *models.Usage, colors Colors, formats Formats) error {
	var data map[string]interface{}
	if err := json.Unmarshal(usage.Raw, &data); err != nil {
		// Fall back to JSON output on parse error
		j, err := usage.ToJSON()
		if err != nil {
			return err
		}
		fmt.Println(j)
		return nil
	}

	fmt.Println()
	fmt.Printf("%s%sClaude.ai Usage%s\n", colors.Bold, colors.Cyan, colors.Reset)
	fmt.Println(strings.Repeat("═", 50))

	printDataRecursive(data, "", colors, formats)

	fmt.Println()
	return nil
}

func printDataRecursive(data map[string]interface{}, indent string, colors Colors, formats Formats) {
	// Sort keys for deterministic output
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := data[key]
		displayKey := FormatKey(key)

		switch v := value.(type) {
		case map[string]interface{}:
			fmt.Printf("%s%s%s:%s\n", indent, colors.Bold, displayKey, colors.Reset)
			printDataRecursive(v, indent+"  ", colors, formats)
		case []interface{}:
			fmt.Printf("%s%s%s:%s\n", indent, colors.Bold, displayKey, colors.Reset)
			for i, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					fmt.Printf("%s  %s[%d]%s\n", indent, colors.Cyan, i+1, colors.Reset)
					printDataRecursive(m, indent+"    ", colors, formats)
				} else {
					fmt.Printf("%s  • %v\n", indent, item)
				}
			}
		case float64:
			valueStr := FormatNumber(v, key, colors)
			fmt.Printf("%s%-22s %s\n", indent, displayKey+":", valueStr)
		case string:
			if v == "" {
				continue // Skip empty strings
			}
			formatted := FormatStringWithFormats(v, key, formats)
			fmt.Printf("%s%-22s %s\n", indent, displayKey+":", formatted)
		case bool:
			fmt.Printf("%s%-22s %t\n", indent, displayKey+":", v)
		case nil:
			// Skip nil values
		default:
			fmt.Printf("%s%-22s %v\n", indent, displayKey+":", v)
		}
	}
}

// FormatKey converts snake_case to Title Case
func FormatKey(key string) string {
	parts := strings.Split(key, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

// FormatNumber formats a numeric value with optional colorization for utilization fields
func FormatNumber(v float64, key string, colors Colors) string {
	keyLower := strings.ToLower(key)
	isUtilization := strings.Contains(keyLower, "utilization") ||
		strings.Contains(keyLower, "percent") ||
		strings.Contains(keyLower, "usage") ||
		strings.Contains(keyLower, "ratio")

	var numStr string
	if v == float64(int64(v)) {
		numStr = fmt.Sprintf("%d", int64(v))
	} else {
		numStr = fmt.Sprintf("%.2f", v)
	}

	if isUtilization && colors.Reset != "" {
		color := GetUtilizationColor(v, colors)
		return fmt.Sprintf("%s%s%s", color, numStr, colors.Reset)
	}

	return numStr
}

// GetUtilizationColor returns the appropriate color based on utilization percentage
func GetUtilizationColor(value float64, colors Colors) string {
	switch {
	case value >= 95:
		return colors.Red
	case value >= 80:
		return colors.Yellow
	default:
		return colors.Green
	}
}

// FormatString formats a string value, converting ISO datetimes to local format.
// Uses default format settings.
func FormatString(v, key string) string {
	return FormatStringWithFormats(v, key, DefaultFormats())
}

// FormatStringWithFormats formats a string value using the provided format settings.
func FormatStringWithFormats(v, key string, fmts Formats) string {
	if !isDatetimeField(key) {
		return v
	}

	inputFormats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, inputFmt := range inputFormats {
		if t, err := time.Parse(inputFmt, v); err == nil {
			local := t.Local()
			if inputFmt == "2006-01-02" {
				return local.Format(fmts.Date)
			}
			return local.Format(fmts.Datetime)
		}
	}

	return v
}

// isDatetimeField returns true if the field name suggests it contains a datetime
func isDatetimeField(key string) bool {
	keyLower := strings.ToLower(key)

	suffixes := []string{
		"_at", "_date", "_time", "_reset", "_start", "_end",
		"_expires", "_created", "_updated", "_timestamp",
	}
	for _, suffix := range suffixes {
		if strings.HasSuffix(keyLower, suffix) {
			return true
		}
	}

	exactMatches := []string{
		"date", "time", "timestamp", "reset", "start", "end",
		"created", "updated", "expires", "datetime",
	}
	for _, match := range exactMatches {
		if keyLower == match {
			return true
		}
	}

	return false
}
