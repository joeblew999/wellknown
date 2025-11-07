package mcp

import (
	"github.com/spf13/cobra"
)

// NewCommand creates the MCP server command
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start MCP server for Claude Desktop integration",
		Long: `Start Model Context Protocol (MCP) server for Claude Desktop.

The MCP server runs in stdio mode and provides:
  • Calendar API integration
  • Google OAuth token management
  • Structured data exchange with Claude Desktop

Configure in Claude Desktop's config file to enable this integration.`,
		Run: func(cmd *cobra.Command, args []string) {
			Main(args)
		},
	}
}
