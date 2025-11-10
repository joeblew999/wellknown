package main

import (
	"fmt"

	pdfform "github.com/joeblew999/wellknown/pkg/pdf"
)

// This example shows how to configure the library using auto-discovery
func main() {
	// Configure the library
	// Use FindDataDir() to auto-discover the data directory
	dataDir := pdfform.FindDataDir()
	config := pdfform.NewConfig(dataDir)
	pdfform.SetDefaultConfig(config)

	fmt.Println("📋 PDF Form Library - Configuration Example\n")

	// Show discovered configuration
	fmt.Printf("Data Directory: %s\n", config.DataDir)
	fmt.Printf("Catalog Path:   %s\n", config.CatalogPath())
	fmt.Printf("Downloads Path: %s\n", config.DownloadsPath())
	fmt.Printf("Templates Path: %s\n", config.TemplatesPath())
	fmt.Printf("Outputs Path:   %s\n", config.OutputsPath())
	fmt.Printf("Cases Path:     %s\n", config.CasesPath())
	fmt.Printf("Temp Path:      %s\n", config.TempPath())

	// Ensure all directories exist
	if err := config.EnsureDirectories(); err != nil {
		fmt.Printf("Error creating directories: %v\n", err)
		return
	}

	fmt.Println("\n✅ Configuration complete!")
	fmt.Println("\nThe library will use these paths for all operations.")
	fmt.Println("\nYou can override the data directory by setting:")
	fmt.Println("  export PDFFORM_DATA_DIR=/your/custom/path")
	fmt.Println("\nOr the library will auto-detect:")
	fmt.Println("  - Docker: /app/.data")
	fmt.Println("  - Local: .data (current or parent directories)")
}
