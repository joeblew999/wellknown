package main

import (
	"fmt"
	"log"

	pdfform "github.com/joeblew999/wellknown/pkg/pdf"
)

// This example shows how to use the PDF form library programmatically
func main() {
	// 1. Configure the library
	// Use FindDataDir() to auto-discover the data directory
	dataDir := pdfform.FindDataDir()
	config := pdfform.NewConfig(dataDir)
	pdfform.SetDefaultConfig(config)
	config.EnsureDirectories()

	fmt.Println("🔍 PDF Form Library - Basic Usage Example\n")

	// 2. Browse forms
	fmt.Println("Step 1: Browse Forms")
	result, err := pdfform.Browse(pdfform.BrowseOptions{
		CatalogPath: config.CatalogFilePath(),
		State:       "VIC", // Victoria
	})
	if err != nil {
		log.Fatalf("Failed to browse: %v", err)
	}

	fmt.Printf("Found %d forms for VIC\n", len(result.Forms))
	if len(result.Forms) > 0 {
		fmt.Printf("First form: %s (%s)\n\n", result.Forms[0].FormName, result.Forms[0].FormCode)
	}

	// 3. Download a form
	if len(result.Forms) > 0 {
		form := result.Forms[0]
		fmt.Printf("Step 2: Download %s\n", form.FormCode)

		downloadResult, err := pdfform.Download(pdfform.DownloadOptions{
			CatalogPath: config.CatalogFilePath(),
			FormCode:    form.FormCode,
			OutputDir:   config.DownloadsPath(),
		})
		if err != nil {
			log.Fatalf("Failed to download: %v", err)
		}

		fmt.Printf("Downloaded to: %s\n\n", downloadResult.PDFPath)

		// 4. Inspect the form
		fmt.Println("Step 3: Inspect Form Fields")
		inspectResult, err := pdfform.Inspect(pdfform.InspectOptions{
			PDFPath:   downloadResult.PDFPath,
			OutputDir: config.TemplatesPath(),
		})
		if err != nil {
			log.Fatalf("Failed to inspect: %v", err)
		}

		fmt.Printf("Found %d fields\n", inspectResult.FieldCount)
		fmt.Printf("Template saved to: %s\n\n", inspectResult.TemplatePath)

		// 5. Fill the form (you would edit the template JSON first)
		// This is just an example - in real use, edit the template
		fmt.Println("Step 4: Fill Form")
		fmt.Println("(Edit the template JSON with your data, then use Fill command)")
		fmt.Printf("Template: %s\n", inspectResult.TemplatePath)
	}

	fmt.Println("\n✅ Example complete!")
	fmt.Println("\nNext steps:")
	fmt.Println("1. Edit the template JSON with your data")
	fmt.Println("2. Use pdfform.Fill() to fill the PDF")
	fmt.Println("3. Or use the CLI: pdfform 4-fill template.json")
}
