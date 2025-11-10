package main

import (
	"log"

	"github.com/joeblew999/wellknown/pkg/pdf/web"
)

func main() {
	log.Println("Starting PDF form web server with zero-config...")
	log.Println("This example demonstrates the easiest way to mount the web server")
	log.Println("")

	// That's it! One line to start a full-featured web server with:
	// - HTTPS auto-generated certificates
	// - Auto-discovery of .data directory
	// - LAN accessibility
	// - Full API and GUI
	if err := web.Start(8080); err != nil {
		log.Fatal(err)
	}
}
