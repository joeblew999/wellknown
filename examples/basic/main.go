package main

import (
	"fmt"
	"log"
	"time"

	"github.com/joeblew999/wellknown/pkg/google"
	"github.com/joeblew999/wellknown/pkg/types"
)

func main() {
	fmt.Println("=== wellknown Basic Examples ===\n")

	// Google Calendar Example - Using the real library!
	fmt.Println("1. Google Calendar Event (using wellknown library):")

	event := types.CalendarEvent{
		Title:       "Team Meeting",
		StartTime:   time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
		EndTime:     time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
		Location:    "Conference Room A",
		Description: "Discuss Q4 roadmap",
	}

	googleCalendarURL, err := google.Calendar(event)
	if err != nil {
		log.Fatalf("Failed to generate Google Calendar URL: %v", err)
	}
	fmt.Printf("   URL: %s\n", googleCalendarURL)
	fmt.Printf("   ✓ Generated with wellknown library\n\n")

	// Another example with minimal fields
	fmt.Println("2. Quick Meeting (minimal fields):")
	quickMeeting := types.CalendarEvent{
		Title:     "Quick Sync",
		StartTime: time.Now().Add(24 * time.Hour), // Tomorrow
		EndTime:   time.Now().Add(24*time.Hour + 30*time.Minute), // 30 min duration
	}

	quickMeetingURL, err := google.Calendar(quickMeeting)
	if err != nil {
		log.Fatalf("Failed to generate URL: %v", err)
	}
	fmt.Printf("   URL: %s\n", quickMeetingURL)
	fmt.Printf("   ✓ No location or description needed\n\n")

	// Placeholder for future implementations
	fmt.Println("=== Coming Soon ===\n")
	fmt.Println("3. Apple Calendar - Not yet implemented")
	fmt.Println("4. Google Maps - Not yet implemented")
	fmt.Println("5. Apple Maps - Not yet implemented")
	fmt.Println("6. Google Drive - Not yet implemented")
	fmt.Println("7. Apple Files/iCloud - Not yet implemented")

	fmt.Println("\n=== End of Examples ===")
	fmt.Println("\nℹ️  Google Calendar deep links are now generated using the wellknown library!")
	fmt.Println("ℹ️  More platforms coming soon...")
}
