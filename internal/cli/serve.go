package cli

import (
	"github.com/benjaminabbitt/claude-limits/internal/mcp"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start MCP server",
	Long: `Start an MCP (Model Context Protocol) server that exposes usage tools.

Authentication priority:
1. --cookie and --org-id flags
2. CLAUDE_SESSION_COOKIE and CLAUDE_ORG_ID environment variables
3. Automatic extraction from browser cookies (Chrome, Firefox)`,
	RunE: runServe,
}

func runServe(cmd *cobra.Command, args []string) error {
	// Always verbose for server startup - users need to see auth status
	cookie, orgID, err := ResolveAuth(true)
	if err != nil {
		return err
	}

	return mcp.Serve(cookie, orgID)
}
