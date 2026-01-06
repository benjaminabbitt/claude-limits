package cli

import (
	"fmt"

	"github.com/benjaminabbitt/claude-limits/internal/auth"
	"github.com/benjaminabbitt/claude-limits/internal/mcp"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start MCP server",
	Long: `Start an MCP (Model Context Protocol) server that exposes usage tools.

Authentication uses OAuth credentials from Claude Code (~/.claude/.credentials.json).
Make sure you have authenticated with Claude Code first.`,
	RunE: runServe,
}

func runServe(cmd *cobra.Command, args []string) error {
	creds, err := auth.Load("")
	if err != nil {
		return err
	}

	fmt.Printf("Starting MCP server (subscription: %s)\n", creds.SubscriptionType)

	return mcp.Serve(creds.AccessToken)
}
