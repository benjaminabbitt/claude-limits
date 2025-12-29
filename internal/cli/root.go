package cli

import (
	"os"

	"github.com/benjaminabbitt/claude-limits/internal/version"
	"github.com/spf13/cobra"
)

var (
	sessionCookie string
	orgID         string
	outputFormat  string
	verbose       bool
	noColor       bool
	cacheTTL      int
)

// RootCmd is the root command for the CLI
var RootCmd = &cobra.Command{
	Use:     "claude-limits [query]",
	Short:   "Check Claude.ai usage limits",
	Long:    `A CLI tool to check your Claude.ai usage and limits for Pro/Max subscriptions.`,
	Version: version.Version,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default to running limits command
		return limitsCmd.RunE(cmd, args)
	},
}

func init() {
	RootCmd.PersistentFlags().StringVar(&sessionCookie, "cookie", "", "Claude.ai session cookie (or set CLAUDE_SESSION_COOKIE)")
	RootCmd.PersistentFlags().StringVar(&orgID, "org-id", "", "Claude.ai organization ID (or set CLAUDE_ORG_ID)")
	RootCmd.PersistentFlags().StringVar(&outputFormat, "format", "table", "Output format: table or json")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	RootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	RootCmd.PersistentFlags().IntVar(&cacheTTL, "cache", 30, "Cache TTL in seconds (0 to disable)")

	RootCmd.AddCommand(limitsCmd)
	RootCmd.AddCommand(serveCmd)
	RootCmd.AddCommand(installScriptCmd)
}

// GetSessionCookie returns the session cookie from flag or environment variable
func GetSessionCookie() string {
	if sessionCookie != "" {
		return sessionCookie
	}
	return os.Getenv("CLAUDE_SESSION_COOKIE")
}

// GetOrgID returns the org ID from flag or environment variable
func GetOrgID() string {
	if orgID != "" {
		return orgID
	}
	return os.Getenv("CLAUDE_ORG_ID")
}

// GetOutputFormat returns the output format setting
func GetOutputFormat() string {
	return outputFormat
}

// IsVerbose returns true if verbose output is enabled
func IsVerbose() bool {
	return verbose
}

// NoColor returns true if colored output should be disabled
func NoColor() bool {
	return noColor
}

// GetCacheTTL returns the cache TTL in seconds
func GetCacheTTL() int {
	return cacheTTL
}
