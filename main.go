package main

// Wellknown - Single binary with multiple services
// Each service receives clean args without the service selector

import (
	"fmt"
	"log"
	"os"

	pocketbase "github.com/joeblew999/wellknown/pkg/cmd/pocketbase"
	"github.com/joeblew999/wellknown/pkg/cmd/server"
	testdatagen "github.com/joeblew999/wellknown/pkg/cmd/testdata-gen"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	service := os.Args[1]
	// Pass remaining args to the service (without the service name)
	serviceArgs := os.Args[2:]

	log.Println("🚀 Wellknown Service Orchestrator")

	switch service {
	case "pb":
		log.Println("Loading PocketBase service...")
		pocketbase.Main(serviceArgs)
	case "server":
		log.Println("Loading Well Known server...")
		server.Main(serviceArgs)
	case "gen-testdata":
		log.Println("Loading Well Known testdata generator...")
		testdatagen.Main(serviceArgs)
	default:
		fmt.Printf("Unknown service: %s\n\n", service)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: wellknown <service> [args...]")
	fmt.Println("")
	fmt.Println("Available services:")
	fmt.Println("  pb         - Start PocketBase server (port 8090)")
	fmt.Println("  server        - Start HTTP server (port 8080)")
	fmt.Println("  gen-testdata  - Generate test data")
}
