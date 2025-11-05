package mcp

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joeblew999/wellknown/pkg/pb"
	"github.com/joeblew999/wellknown/pkg/pbmcp"
)

// Main is the entry point for the MCP server command
func Main(args []string) {
	// Create PocketBase Wellknown instance
	wk, err := wellknown.New()
	if err != nil {
		log.Fatalf("Failed to create Wellknown: %v", err)
	}

	// Create MCP server
	server := pbmcp.NewServer(wk.App)

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\n🛑 Shutting down MCP server...")
		cancel()
	}()

	// Run the MCP server
	if err := server.Run(ctx); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}
