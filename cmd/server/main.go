package main

import (
	"flag"
	"log"

	"github.com/joeblew999/wellknown/pkg/server"
)

func main() {
	port := flag.String("port", "8080", "Port to run the server on")
	flag.Parse()

	srv, err := server.New(*port)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
