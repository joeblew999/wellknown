package main

import (
	"fmt"
	"os"

	pdfform "github.com/joeblew999/wellknown/pkg/pdf"
	"github.com/joeblew999/wellknown/pkg/pdf/cli"
)

func main() {
	// Find data directory using config's smart search
	dataDir := pdfform.FindDataDir()

	// Configure the package with our data directory
	config := pdfform.NewConfig(dataDir)
	pdfform.SetDefaultConfig(config)

	// Ensure all directories exist
	if err := config.EnsureDirectories(); err != nil {
		fmt.Printf("Error creating directories: %v\n", err)
		os.Exit(1)
	}

	if err := cli.Run(config); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
