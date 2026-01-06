package cli

import (
	"github.com/benjaminabbitt/claude-limits/internal/config"
	"github.com/benjaminabbitt/claude-limits/internal/version"
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	verbose      bool
	noColor      bool
	cacheTTL     int
	configPath   string
	cfg          *config.Config
)

// RootCmd is the root command for the CLI
var RootCmd = &cobra.Command{
	Use:     "claude-limits [query]",
	Short:   "Check Claude.ai usage limits",
	Long:    `A CLI tool to check your Claude.ai usage and limits for Pro/Max subscriptions.`,
	Version: version.Version,
	Args:    cobra.MaximumNArgs(1),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration file
		cfg = config.LoadOrDefault(configPath)
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default to running limits command
		return limitsCmd.RunE(cmd, args)
	},
}

func init() {
	RootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Config file path (default: ~/.config/claude-limits/config.yaml)")
	RootCmd.PersistentFlags().StringVar(&outputFormat, "format", "table", "Output format: table or json")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	RootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	RootCmd.PersistentFlags().IntVar(&cacheTTL, "cache", 30, "Cache TTL in seconds (0 to disable)")

	RootCmd.AddCommand(limitsCmd)
	RootCmd.AddCommand(serveCmd)
	RootCmd.AddCommand(installScriptCmd)
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

// GetFormats returns the resolved format settings from config
func GetFormats() config.FormatPreset {
	if cfg != nil {
		return cfg.ResolvedFormats()
	}
	return config.FormatPreset{
		Datetime: config.DefaultDatetimeFormat,
		Date:     config.DefaultDateFormat,
		Time:     config.DefaultTimeFormat,
	}
}
