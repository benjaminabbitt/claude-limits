package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/benjaminabbitt/claude-limits/internal/api"
	"github.com/benjaminabbitt/claude-limits/internal/auth"
	"github.com/benjaminabbitt/claude-limits/internal/cache"
	"github.com/benjaminabbitt/claude-limits/internal/format"
	"github.com/benjaminabbitt/claude-limits/internal/fuzzy"
	"github.com/benjaminabbitt/claude-limits/internal/models"

	"github.com/spf13/cobra"
)

var limitsCmd = &cobra.Command{
	Use:   "limits [query]",
	Short: "Display current usage",
	Long: `Fetch and display your current Claude.ai usage.

If a query is provided, fuzzy matches against field names and returns just the value.
Example: claude-limits limits five  →  returns value for "Five Hour" field

Authentication uses OAuth credentials from Claude Code (~/.claude/.credentials.json).
Make sure you have authenticated with Claude Code first.`,
	RunE: runLimits,
	Args: cobra.MaximumNArgs(1),
}

func runLimits(cmd *cobra.Command, args []string) error {
	usage, err := getUsageWithCache()
	if err != nil {
		return err
	}

	// If a query argument is provided, do fuzzy match
	if len(args) > 0 {
		return printMatchedValue(usage, args[0])
	}

	if GetOutputFormat() == "json" {
		return printJSON(usage)
	}
	return printTable(usage)
}

func getUsageWithCache() (*models.Usage, error) {
	ttl := GetCacheTTL()
	c := cache.New(IsVerbose())

	// Try to read from cache if TTL > 0
	if ttl > 0 {
		if cached, err := c.Read(ttl); err == nil {
			if IsVerbose() {
				fmt.Fprintln(os.Stderr, "Using cached data")
			}
			return cached, nil
		}
	}

	// Fetch fresh data
	creds, err := auth.Load("")
	if err != nil {
		return nil, err
	}

	if IsVerbose() {
		fmt.Fprintf(os.Stderr, "Using Claude Code credentials (subscription: %s)\n", creds.SubscriptionType)
		if creds.IsExpired() {
			fmt.Fprintln(os.Stderr, "Warning: access token may be expired")
		}
	}

	client := api.NewClient(creds.AccessToken)
	usage, err := client.GetUsage()
	if err != nil {
		return nil, err
	}

	// Save to cache
	if ttl > 0 {
		if err := c.Write(usage); err != nil && IsVerbose() {
			fmt.Fprintf(os.Stderr, "Failed to write cache: %v\n", err)
		}
	}

	return usage, nil
}

func printMatchedValue(usage *models.Usage, query string) error {
	var data map[string]interface{}
	if err := json.Unmarshal(usage.Raw, &data); err != nil {
		return fmt.Errorf("failed to parse usage data: %w", err)
	}

	pairs := fuzzy.FlattenData(data, "")
	match, err := fuzzy.FindBestMatch(pairs, query)
	if err != nil {
		return err
	}

	colors := format.NewColors(NoColor())

	switch v := match.Value.(type) {
	case float64:
		fmt.Println(format.FormatNumber(v, match.Key, colors))
	case string:
		fmt.Println(v)
	case bool:
		fmt.Println(v)
	default:
		fmt.Printf("%v\n", v)
	}

	return nil
}

func printJSON(usage *models.Usage) error {
	j, err := format.JSON(usage)
	if err != nil {
		return err
	}
	fmt.Println(j)
	return nil
}

func printTable(usage *models.Usage) error {
	colors := format.NewColors(NoColor())
	fmts := GetFormats()
	formats := format.Formats{
		Datetime: fmts.Datetime,
		Date:     fmts.Date,
		Time:     fmts.Time,
	}
	return format.Table(usage, colors, formats)
}
