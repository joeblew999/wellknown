package main

import (
	"fmt"
)

func main() {
	fmt.Println("=== wellknown WebView Example ===\n")
	fmt.Println("This example demonstrates web fallbacks for deep links")
	fmt.Println("when native apps are not available.\n")

	// Calendar Web Fallback
	fmt.Println("1. Google Calendar (Web Fallback):")
	webCalendarURL := "https://calendar.google.com/calendar/render?action=TEMPLATE&text=Team%20Meeting&dates=20251026T140000Z/20251026T150000Z&details=Discuss%20Q4%20roadmap&location=Conference%20Room%20A"
	fmt.Printf("   URL: %s\n\n", webCalendarURL)

	// Maps Web Fallback
	fmt.Println("2. Google Maps (Web Fallback):")
	webMapsURL := "https://www.google.com/maps/search/?api=1&query=Space+Needle,Seattle+WA"
	fmt.Printf("   URL: %s\n\n", webMapsURL)

	// Drive Web Fallback
	fmt.Println("3. Google Drive (Web Fallback):")
	webDriveURL := "https://drive.google.com/file/d/1a2b3c4d5e6f7g8h9i0j/view"
	fmt.Printf("   URL: %s\n\n", webDriveURL)

	// iCloud Web Fallback
	fmt.Println("4. iCloud (Web Fallback):")
	webICloudURL := "https://www.icloud.com/iclouddrive/"
	fmt.Printf("   URL: %s\n\n", webICloudURL)

	fmt.Println("=== End of WebView Examples ===")
	fmt.Println("\nNote: Web fallbacks are useful when:")
	fmt.Println("  - Native apps are not installed")
	fmt.Println("  - Running in a browser context")
	fmt.Println("  - Cross-platform compatibility is needed")
}
