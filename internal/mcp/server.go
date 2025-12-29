package mcp

import (
	"context"
	"fmt"

	"github.com/benjaminabbitt/claude-limits/internal/api"
	"github.com/benjaminabbitt/claude-limits/internal/version"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Serve starts the MCP server on stdio.
// The mcp-go library handles SIGTERM/SIGINT for graceful shutdown.
func Serve(sessionCookie, orgID string) error {
	s := server.NewMCPServer(
		"claude-limits",
		version.Version,
		server.WithToolCapabilities(true),
	)

	// Define the get_usage tool
	usageTool := mcp.NewTool("get_usage",
		mcp.WithDescription("Get current Claude.ai usage for your Pro/Max subscription"),
	)

	// Create API client
	client := api.NewClient(sessionCookie, orgID)

	// Add the tool with its handler
	s.AddTool(usageTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		usage, err := client.GetUsage()
		if err != nil {
			return nil, fmt.Errorf("failed to get usage: %w", err)
		}

		json, err := usage.ToJSON()
		if err != nil {
			return nil, fmt.Errorf("failed to serialize usage: %w", err)
		}

		return mcp.NewToolResultText(json), nil
	})

	// Start the server on stdio (library handles signal-based shutdown)
	return server.ServeStdio(s)
}
