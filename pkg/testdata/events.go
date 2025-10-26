// Package testdata provides shared test cases for both unit tests and examples.
// This ensures consistency between what we test and what we demonstrate.
package testdata

import (
	"time"

	"github.com/joeblew999/wellknown/pkg/types"
)

// CalendarTestCase represents a test case with an event and expected URL
type CalendarTestCase struct {
	Name        string
	Event       types.CalendarEvent
	ExpectedURL string
	ShouldError bool
}

// CalendarEvents provides test cases that can be used by both unit tests and examples
var CalendarEvents = []CalendarTestCase{
	{
		Name: "Team Meeting - Complete Event",
		Event: types.CalendarEvent{
			Title:       "Team Meeting",
			StartTime:   time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			EndTime:     time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
			Location:    "Conference Room A",
			Description: "Quarterly planning meeting",
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251026T140000Z%2F20251026T150000Z&details=Quarterly+planning+meeting&location=Conference+Room+A&text=Team+Meeting",
		ShouldError: false,
	},
	{
		Name: "Quick Sync - Minimal Event",
		Event: types.CalendarEvent{
			Title:     "Quick Sync",
			StartTime: time.Date(2025, 10, 27, 10, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 27, 10, 30, 0, 0, time.UTC),
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251027T100000Z%2F20251027T103000Z&text=Quick+Sync",
		ShouldError: false,
	},
	{
		Name: "Client Visit - With Location",
		Event: types.CalendarEvent{
			Title:     "Client Visit",
			StartTime: time.Date(2025, 11, 1, 9, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 11, 1, 10, 0, 0, 0, time.UTC),
			Location:  "123 Main St",
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251101T090000Z%2F20251101T100000Z&location=123+Main+St&text=Client+Visit",
		ShouldError: false,
	},
	{
		Name: "Project Review - Special Characters",
		Event: types.CalendarEvent{
			Title:     "Project Review & Planning",
			StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
		},
		ExpectedURL: "https://calendar.google.com/calendar/render?action=TEMPLATE&dates=20251026T140000Z%2F20251026T150000Z&text=Project+Review+%26+Planning",
		ShouldError: false,
	},
}

// ErrorTestCases provides test cases that should produce validation errors
var ErrorTestCases = []CalendarTestCase{
	{
		Name: "Missing Title",
		Event: types.CalendarEvent{
			StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
		},
		ShouldError: true,
	},
	{
		Name: "Missing Start Time",
		Event: types.CalendarEvent{
			Title:   "Test",
			EndTime: time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
		},
		ShouldError: true,
	},
	{
		Name: "Missing End Time",
		Event: types.CalendarEvent{
			Title:     "Test",
			StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
		},
		ShouldError: true,
	},
	{
		Name: "End Before Start",
		Event: types.CalendarEvent{
			Title:     "Test",
			StartTime: time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
		},
		ShouldError: true,
	},
}
