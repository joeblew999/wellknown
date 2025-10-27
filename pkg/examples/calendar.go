// Package examples provides shared test data and examples for all platforms.
// This ensures consistency across Google Calendar, Apple Calendar, and future services.
package examples

import (
	"time"

	"github.com/joeblew999/wellknown/pkg/types"
)

// CalendarExample represents a user-friendly example for the web demo showcase.
// It wraps a CalendarEvent with metadata like name and description.
type CalendarExample struct {
	Name        string
	Description string
	Event       types.CalendarEvent
}

// GetName returns the example name (implements ServiceExample interface)
func (e CalendarExample) GetName() string {
	return e.Name
}

// GetDescription returns the example description (implements ServiceExample interface)
func (e CalendarExample) GetDescription() string {
	return e.Description
}

// CalendarExamples provides beautiful, realistic examples for the web demo showcase.
// These examples are shared across all calendar platforms (Google, Apple, etc.)
// to ensure consistent test data and user experience.
var CalendarExamples = []CalendarExample{
	{
		Name:        "Team Meeting",
		Description: "Weekly team sync with location and agenda",
		Event: types.CalendarEvent{
			Title:       "Team Meeting",
			StartTime:   time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			EndTime:     time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
			Location:    "Conference Room A",
			Description: "Quarterly planning meeting - please bring your Q4 goals",
		},
	},
	{
		Name:        "Quick Sync",
		Description: "Short 15-minute check-in call",
		Event: types.CalendarEvent{
			Title:     "Quick Sync",
			StartTime: time.Date(2025, 10, 27, 10, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 27, 10, 15, 0, 0, time.UTC),
		},
	},
	{
		Name:        "Client Visit",
		Description: "In-person client meeting at their office",
		Event: types.CalendarEvent{
			Title:       "Client Visit - Acme Corp",
			StartTime:   time.Date(2025, 11, 1, 9, 0, 0, 0, time.UTC),
			EndTime:     time.Date(2025, 11, 1, 10, 0, 0, 0, time.UTC),
			Location:    "123 Main St, San Francisco, CA",
			Description: "Q4 business review and 2026 planning discussion",
		},
	},
	{
		Name:        "Lunch Break",
		Description: "Block out time for lunch",
		Event: types.CalendarEvent{
			Title:     "Lunch Break",
			StartTime: time.Date(2025, 10, 26, 12, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 26, 13, 0, 0, 0, time.UTC),
		},
	},
	{
		Name:        "Workshop",
		Description: "Half-day training session with detailed agenda",
		Event: types.CalendarEvent{
			Title:       "Go Programming Workshop",
			StartTime:   time.Date(2025, 11, 5, 9, 0, 0, 0, time.UTC),
			EndTime:     time.Date(2025, 11, 5, 13, 0, 0, 0, time.UTC),
			Location:    "Training Room B",
			Description: "Topics: Goroutines, Channels, Context, and Best Practices. Bring your laptop!",
		},
	},
	{
		Name:        "Coffee Chat",
		Description: "Casual 1:1 conversation",
		Event: types.CalendarEvent{
			Title:       "Coffee Chat",
			StartTime:   time.Date(2025, 10, 28, 15, 30, 0, 0, time.UTC),
			EndTime:     time.Date(2025, 10, 28, 16, 0, 0, 0, time.UTC),
			Location:    "Starbucks Downtown",
			Description: "Catch up and discuss career goals",
		},
	},
}
