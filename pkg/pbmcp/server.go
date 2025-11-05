package pbmcp

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pocketbase/pocketbase/core"
)

// Server wraps the MCP server with PocketBase app instance
type Server struct {
	server *mcp.Server
	app    core.App
}

// NewServer creates a new MCP server for PocketBase
func NewServer(app core.App) *Server {
	impl := &mcp.Implementation{
		Name:    "pocketbase-mcp",
		Version: "1.0.0",
	}

	opts := &mcp.ServerOptions{
		Instructions: "PocketBase MCP Server - Access and manage PocketBase collections, records, and data",
	}

	mcpServer := mcp.NewServer(impl, opts)

	s := &Server{
		server: mcpServer,
		app:    app,
	}

	// Register tools
	s.registerTools()

	// Register resources
	s.registerResources()

	return s
}

// Run starts the MCP server with stdio transport (for Claude Desktop)
func (s *Server) Run(ctx context.Context) error {
	log.Println("🤖 Starting PocketBase MCP server...")
	log.Println("📡 Listening on stdio for MCP requests")
	log.Println("💡 Configure in Claude Desktop to enable integration")

	return s.server.Run(ctx, &mcp.StdioTransport{})
}
