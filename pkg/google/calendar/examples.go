package calendar

import "time"

// Example represents a user-friendly example for the web demo showcase.
type Example struct {
	Name        string
	Description string
	Event       Event
}

// GetName returns the example name (implements ServiceExample interface)
func (e Example) GetName() string {
	return e.Name
}

// GetDescription returns the example description (implements ServiceExample interface)
func (e Example) GetDescription() string {
	return e.Description
}

// ExpectedURL generates and returns the expected Google Calendar URL for this example
func (e Example) ExpectedURL() string {
	url, _ := e.Event.GenerateURL()
	return url
}

// Examples provides realistic Google Calendar event examples for the web demo showcase.
// These examples use only the 5 basic fields supported by Google Calendar URLs.
var Examples = []Example{
	{
		Name:        "Team Meeting",
		Description: "Weekly team sync with location and agenda",
		Event: Event{
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
		Event: Event{
			Title:     "Quick Sync",
			StartTime: time.Date(2025, 10, 27, 10, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 27, 10, 15, 0, 0, time.UTC),
		},
	},
	{
		Name:        "Client Visit",
		Description: "In-person client meeting at their office",
		Event: Event{
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
		Event: Event{
			Title:     "Lunch Break",
			StartTime: time.Date(2025, 10, 26, 12, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 26, 13, 0, 0, 0, time.UTC),
		},
	},
	{
		Name:        "Workshop",
		Description: "Half-day training session with detailed agenda",
		Event: Event{
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
		Event: Event{
			Title:       "Coffee Chat",
			StartTime:   time.Date(2025, 10, 28, 15, 30, 0, 0, time.UTC),
			EndTime:     time.Date(2025, 10, 28, 16, 0, 0, 0, time.UTC),
			Location:    "Starbucks Downtown",
			Description: "Catch up and discuss career goals",
		},
	},
}
