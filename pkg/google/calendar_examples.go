// Package google provides examples for demonstration purposes.
// These examples are used in the web demo showcase to show realistic use cases.
package google

import (
	"time"

	"github.com/joeblew999/wellknown/pkg/types"
)

// CalendarExample represents a user-friendly example for the web demo
type CalendarExample struct {
	Name        string
	Description string
	Event       types.CalendarEvent
	ExpectedURL string // Generated URL for display
}

// CalendarExamples provides beautiful, realistic examples for the web demo showcase
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
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251026T140000Z%2F20251026T150000Z&details=Quarterly+planning+meeting+-+please+bring+your+Q4+goals&location=Conference+Room+A&text=Team+Meeting",
	},
	{
		Name:        "Quick Sync",
		Description: "Short 15-minute check-in call",
		Event: types.CalendarEvent{
			Title:     "Quick Sync",
			StartTime: time.Date(2025, 10, 27, 10, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 27, 10, 15, 0, 0, time.UTC),
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251027T100000Z%2F20251027T101500Z&text=Quick+Sync",
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
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251101T090000Z%2F20251101T100000Z&details=Q4+business+review+and+2026+planning+discussion&location=123+Main+St%2C+San+Francisco%2C+CA&text=Client+Visit+-+Acme+Corp",
	},
	{
		Name:        "Lunch Break",
		Description: "Block out time for lunch",
		Event: types.CalendarEvent{
			Title:     "Lunch Break",
			StartTime: time.Date(2025, 10, 26, 12, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 26, 13, 0, 0, 0, time.UTC),
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251026T120000Z%2F20251026T130000Z&text=Lunch+Break",
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
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251105T090000Z%2F20251105T130000Z&details=Topics%3A+Goroutines%2C+Channels%2C+Context%2C+and+Best+Practices.+Bring+your+laptop%21&location=Training+Room+B&text=Go+Programming+Workshop",
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
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251028T153000Z%2F20251028T160000Z&details=Catch+up+and+discuss+career+goals&location=Starbucks+Downtown&text=Coffee+Chat",
	},
}
