package server

import (
	"flag"
	"log"

	"github.com/joeblew999/wellknown/pkg/server"
)

// Main is the entry point for the server service
// args contains the command-line arguments after the service name
func Main(args []string) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	port := fs.String("port", "8080", "Port to run the server on")
	fs.Parse(args)

	srv, err := server.New(*port)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
