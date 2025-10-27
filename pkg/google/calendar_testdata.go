// Package google provides comprehensive test cases for unit testing.
// These test cases are separate from examples to allow thorough testing
// without cluttering the user-facing demo.
package google

import (
	"time"

	"github.com/joeblew999/wellknown/pkg/types"
)

// CalendarTestCase represents a test case with an event and expected URL
type CalendarTestCase struct {
	Name        string
	Event       types.CalendarEvent
	ExpectedURL string
}

// CalendarErrorCase represents a test case that should produce a validation error
type CalendarErrorCase struct {
	Name  string
	Event types.CalendarEvent
}

// CalendarValidTestCases provides comprehensive test cases for valid calendar events
var CalendarValidTestCases = []CalendarTestCase{
	// Basic valid cases
	{
		Name: "Complete Event - All Fields",
		Event: types.CalendarEvent{
			Title:       "Team Meeting",
			StartTime:   time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			EndTime:     time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
			Location:    "Conference Room A",
			Description: "Quarterly planning meeting",
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251026T140000Z%2F20251026T150000Z&details=Quarterly+planning+meeting&location=Conference+Room+A&text=Team+Meeting",
	},
	{
		Name: "Minimal Event - Only Required Fields",
		Event: types.CalendarEvent{
			Title:     "Quick Sync",
			StartTime: time.Date(2025, 10, 27, 10, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 27, 10, 30, 0, 0, time.UTC),
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251027T100000Z%2F20251027T103000Z&text=Quick+Sync",
	},
	{
		Name: "Event with Location Only",
		Event: types.CalendarEvent{
			Title:     "Client Visit",
			StartTime: time.Date(2025, 11, 1, 9, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 11, 1, 10, 0, 0, 0, time.UTC),
			Location:  "123 Main St",
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251101T090000Z%2F20251101T100000Z&location=123+Main+St&text=Client+Visit",
	},
	{
		Name: "Event with Description Only",
		Event: types.CalendarEvent{
			Title:       "Planning Session",
			StartTime:   time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			EndTime:     time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
			Description: "Discuss roadmap for next quarter",
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251026T140000Z%2F20251026T150000Z&details=Discuss+roadmap+for+next+quarter&text=Planning+Session",
	},

	// Special characters and encoding
	{
		Name: "Title with Ampersand",
		Event: types.CalendarEvent{
			Title:     "Project Review & Planning",
			StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251026T140000Z%2F20251026T150000Z&text=Project+Review+%26+Planning",
	},
	{
		Name: "Location with Special Characters",
		Event: types.CalendarEvent{
			Title:     "Meeting",
			StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
			Location:  "Joe's Café & Bistro",
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251026T140000Z%2F20251026T150000Z&location=Joe%27s+Caf%C3%A9+%26+Bistro&text=Meeting",
	},
	{
		Name: "Description with Newlines",
		Event: types.CalendarEvent{
			Title:       "Multi-line Event",
			StartTime:   time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			EndTime:     time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
			Description: "Line 1\nLine 2\nLine 3",
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251026T140000Z%2F20251026T150000Z&details=Line+1%0ALine+2%0ALine+3&text=Multi-line+Event",
	},
	{
		Name: "URL in Description",
		Event: types.CalendarEvent{
			Title:       "Join Meeting",
			StartTime:   time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			EndTime:     time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
			Description: "Join: https://meet.google.com/abc-defg-hij",
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251026T140000Z%2F20251026T150000Z&details=Join%3A+https%3A%2F%2Fmeet.google.com%2Fabc-defg-hij&text=Join+Meeting",
	},
	{
		Name: "Unicode Characters in Title",
		Event: types.CalendarEvent{
			Title:     "Meeting 会議 встреча",
			StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251026T140000Z%2F20251026T150000Z&text=Meeting+%E4%BC%9A%E8%AD%B0+%D0%B2%D1%81%D1%82%D1%80%D0%B5%D1%87%D0%B0",
	},

	// Time boundary conditions
	{
		Name: "Event at Midnight",
		Event: types.CalendarEvent{
			Title:     "Midnight Event",
			StartTime: time.Date(2025, 10, 26, 0, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 26, 1, 0, 0, 0, time.UTC),
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251026T000000Z%2F20251026T010000Z&text=Midnight+Event",
	},
	{
		Name: "Event Crossing Day Boundary",
		Event: types.CalendarEvent{
			Title:     "Late Night Event",
			StartTime: time.Date(2025, 10, 26, 23, 30, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 27, 0, 30, 0, 0, time.UTC),
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251026T233000Z%2F20251027T003000Z&text=Late+Night+Event",
	},
	{
		Name: "Event on New Year's Eve",
		Event: types.CalendarEvent{
			Title:     "New Year Celebration",
			StartTime: time.Date(2025, 12, 31, 23, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC),
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251231T230000Z%2F20260101T010000Z&text=New+Year+Celebration",
	},
	{
		Name: "Event on Leap Day",
		Event: types.CalendarEvent{
			Title:     "Leap Day Event",
			StartTime: time.Date(2024, 2, 29, 10, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2024, 2, 29, 11, 0, 0, 0, time.UTC),
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20240229T100000Z%2F20240229T110000Z&text=Leap+Day+Event",
	},

	// Duration variations
	{
		Name: "1 Minute Event",
		Event: types.CalendarEvent{
			Title:     "Quick Check",
			StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 26, 14, 1, 0, 0, time.UTC),
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251026T140000Z%2F20251026T140100Z&text=Quick+Check",
	},
	{
		Name: "All Day Event (8 hours)",
		Event: types.CalendarEvent{
			Title:     "Conference",
			StartTime: time.Date(2025, 10, 26, 9, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 26, 17, 0, 0, 0, time.UTC),
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251026T090000Z%2F20251026T170000Z&text=Conference",
	},
	{
		Name: "Multi-Day Event",
		Event: types.CalendarEvent{
			Title:     "Training Week",
			StartTime: time.Date(2025, 10, 26, 9, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 30, 17, 0, 0, 0, time.UTC),
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251026T090000Z%2F20251030T170000Z&text=Training+Week",
	},

	// Edge cases for field lengths
	{
		Name: "Very Long Title",
		Event: types.CalendarEvent{
			Title:     "This is a very long event title that contains many words and should still be properly URL encoded without any issues when passed to Google Calendar",
			StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251026T140000Z%2F20251026T150000Z&text=This+is+a+very+long+event+title+that+contains+many+words+and+should+still+be+properly+URL+encoded+without+any+issues+when+passed+to+Google+Calendar",
	},
	{
		Name: "Very Long Description",
		Event: types.CalendarEvent{
			Title:     "Event",
			StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
			Description: "This is a very long description that contains multiple sentences and should be properly encoded. " +
				"It includes various details about the meeting agenda. " +
				"We will discuss project timelines, resource allocation, and budget considerations.",
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251026T140000Z%2F20251026T150000Z&details=This+is+a+very+long+description+that+contains+multiple+sentences+and+should+be+properly+encoded.+It+includes+various+details+about+the+meeting+agenda.+We+will+discuss+project+timelines%2C+resource+allocation%2C+and+budget+considerations.&text=Event",
	},
}

// CalendarErrorTestCases provides test cases that should produce validation errors
var CalendarErrorTestCases = []CalendarErrorCase{
	{
		Name: "Missing Title",
		Event: types.CalendarEvent{
			StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
		},
	},
	{
		Name: "Empty Title",
		Event: types.CalendarEvent{
			Title:     "",
			StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
		},
	},
	{
		Name: "Missing Start Time",
		Event: types.CalendarEvent{
			Title:   "Test",
			EndTime: time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
		},
	},
	{
		Name: "Missing End Time",
		Event: types.CalendarEvent{
			Title:     "Test",
			StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
		},
	},
	{
		Name: "End Before Start",
		Event: types.CalendarEvent{
			Title:     "Test",
			StartTime: time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
		},
	},
	{
		Name: "End Same as Start",
		Event: types.CalendarEvent{
			Title:     "Test",
			StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
		},
	},
}
