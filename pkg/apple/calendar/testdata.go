package calendar

import (
	"encoding/base64"
	"time"
)

// TestCase represents a comprehensive test case for Apple Calendar event generation.
// These test cases are validated by JSON Schema and used for both unit testing and showcase display.
type TestCase struct {
	Name        string
	Description string
	Event       Event
	ShouldPass  bool // Whether this test case should pass validation
}

// GetName returns the test case name (implements ServiceExample interface for showcase)
func (tc TestCase) GetName() string {
	return tc.Name
}

// GetDescription returns a description derived from the event details
func (tc TestCase) GetDescription() string {
	if tc.Description != "" {
		return tc.Description
	}
	// Auto-generate description from event
	desc := ""
	if tc.Event.Location != "" {
		desc = "Location: " + tc.Event.Location
	}
	if tc.Event.Description != "" {
		if desc != "" {
			desc += " • "
		}
		desc += tc.Event.Description
	}
	if desc == "" {
		desc = "Calendar event example"
	}
	return desc
}

// ExpectedURL generates and returns the download URL for this test case
// This returns a link to /apple/calendar/download endpoint with base64-encoded ICS
func (tc TestCase) ExpectedURL() string {
	ics, err := tc.Event.GenerateICS()
	if err != nil {
		return ""
	}
	// Encode ICS as base64 for URL parameter (using URLEncoding to match server)
	encoded := base64.URLEncoding.EncodeToString(ics)
	// Return download endpoint URL
	return "/apple/calendar/download?event=" + encoded
}

// Helper function to create a pointer to an int
func ptrInt(i int) *int {
	return &i
}

// ValidTestCases contains comprehensive test cases that should pass validation.
// These are used for:
// 1. Unit testing (validates ICS generation works correctly)
// 2. JSON Schema validation (ensures schema.json is correct)
// 3. Web showcase (displays comprehensive examples to users)
var ValidTestCases = []TestCase{
	{
		Name:        "Complete Event - All Basic Fields",
		Description: "Simple 1-hour meeting (basic fields only)",
		ShouldPass:  true,
		Event: Event{
			Title:       "Team Meeting",
			StartTime:   time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			EndTime:     time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
			Location:    "Conference Room A",
			Description: "Quarterly planning meeting - please bring your Q4 goals",
			Status:      StatusConfirmed,
		},
	},
	{
		Name:        "Minimal Event - Only Required Fields",
		Description: "Calendar event example",
		ShouldPass:  true,
		Event: Event{
			Title:     "Quick Sync",
			StartTime: time.Date(2025, 10, 27, 10, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 27, 10, 30, 0, 0, time.UTC),
		},
	},
	{
		Name:        "Event with Location Only",
		Description: "Location: 123 Main St",
		ShouldPass:  true,
		Event: Event{
			Title:     "Client Visit",
			StartTime: time.Date(2025, 11, 1, 9, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 11, 1, 10, 0, 0, 0, time.UTC),
			Location:  "123 Main St",
		},
	},
	{
		Name:        "Event with Description Only",
		Description: "Discuss roadmap for next quarter",
		ShouldPass:  true,
		Event: Event{
			Title:       "Planning Session",
			StartTime:   time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			EndTime:     time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
			Description: "Discuss roadmap for next quarter",
		},
	},
	{
		Name:        "All Day Event",
		Description: "Full day conference",
		ShouldPass:  true,
		Event: Event{
			Title:     "Company Conference",
			StartTime: time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 11, 15, 23, 59, 59, 0, time.UTC),
			Location:  "Convention Center",
			AllDay:    true,
		},
	},
	{
		Name:        "Weekly Recurring Event",
		Description: "Recurring meeting every Monday with reminders",
		ShouldPass:  true,
		Event: Event{
			Title:       "Weekly Standup",
			StartTime:   time.Date(2025, 10, 28, 10, 0, 0, 0, time.UTC),
			EndTime:     time.Date(2025, 10, 28, 10, 30, 0, 0, time.UTC),
			Location:    "Zoom",
			Description: "Weekly team sync - review progress and blockers",
			Recurrence: &RecurrenceRule{
				Frequency: FrequencyWeekly,
				Interval:  1,
				ByDay:     []time.Weekday{time.Monday},
				Count:     ptrInt(10), // 10 occurrences
			},
			Reminders: []Reminder{
				{Duration: 15 * time.Minute},
			},
			Status: StatusConfirmed,
		},
	},
	{
		Name:        "Daily Recurring Event",
		Description: "Recurring daily reminder",
		ShouldPass:  true,
		Event: Event{
			Title:     "Morning Exercise",
			StartTime: time.Date(2025, 10, 27, 7, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 27, 8, 0, 0, 0, time.UTC),
			Recurrence: &RecurrenceRule{
				Frequency: FrequencyDaily,
				Interval:  1,
				Count:     ptrInt(30), // 30 days
			},
			Reminders: []Reminder{
				{Duration: 10 * time.Minute},
			},
			Status:     StatusConfirmed,
			Categories: []string{"Health", "Personal"},
		},
	},
	{
		Name:        "Monthly Recurring Event",
		Description: "First Monday of every month",
		ShouldPass:  true,
		Event: Event{
			Title:       "Board Meeting",
			StartTime:   time.Date(2025, 11, 3, 9, 0, 0, 0, time.UTC),
			EndTime:     time.Date(2025, 11, 3, 11, 0, 0, 0, time.UTC),
			Location:    "Boardroom",
			Description: "Monthly board meeting - review financials and strategy",
			Recurrence: &RecurrenceRule{
				Frequency:  FrequencyMonthly,
				Interval:   1,
				ByDay:      []time.Weekday{time.Monday},
				ByMonthDay: []int{1, 2, 3, 4, 5, 6, 7}, // First week of month
				Count:      ptrInt(12),                  // 12 months
			},
			Organizer: &Organizer{
				Name:  "CEO",
				Email: "ceo@example.com",
			},
			Reminders: []Reminder{
				{Duration: 24 * time.Hour},
			},
			Priority:   1,
			Status:     StatusConfirmed,
			Categories: []string{"Executive", "Board"},
		},
	},
	{
		Name:        "Event with Multiple Attendees",
		Description: "Important event with multiple attendees and high priority",
		ShouldPass:  true,
		Event: Event{
			Title:       "Product Launch Meeting",
			StartTime:   time.Date(2025, 11, 15, 14, 0, 0, 0, time.UTC),
			EndTime:     time.Date(2025, 11, 15, 16, 0, 0, 0, time.UTC),
			Location:    "Main Conference Room",
			Description: "Final review before product launch - all hands required",
			Organizer: &Organizer{
				Name:  "Sarah Johnson",
				Email: "sarah@example.com",
			},
			Attendees: []Attendee{
				{
					Name:   "Alice Smith",
					Email:  "alice@example.com",
					Role:   RoleReqParticipant,
					Status: StatusAccepted,
					RSVP:   true,
				},
				{
					Name:   "Bob Williams",
					Email:  "bob@example.com",
					Role:   RoleReqParticipant,
					Status: StatusNeedsAction,
					RSVP:   true,
				},
				{
					Name:   "Charlie Davis",
					Email:  "charlie@example.com",
					Role:   RoleOptParticipant,
					Status: StatusTentative,
					RSVP:   true,
				},
			},
			Reminders: []Reminder{
				{Duration: 1 * time.Hour},
				{Duration: 24 * time.Hour},
			},
			Priority:   1, // Highest priority
			Status:     StatusConfirmed,
			Categories: []string{"Product", "Launch", "Important"},
		},
	},
	{
		Name:        "Event with Multiple Reminders",
		Description: "Event with 15min and 1-day advance reminders",
		ShouldPass:  true,
		Event: Event{
			Title:     "Important Presentation",
			StartTime: time.Date(2025, 11, 10, 14, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 11, 10, 15, 0, 0, 0, time.UTC),
			Reminders: []Reminder{
				{Duration: 15 * time.Minute},
				{Duration: 24 * time.Hour},
			},
		},
	},
	{
		Name:        "Event with Categories",
		Description: "Event tagged with multiple categories",
		ShouldPass:  true,
		Event: Event{
			Title:      "Marketing Review",
			StartTime:  time.Date(2025, 11, 5, 10, 0, 0, 0, time.UTC),
			EndTime:    time.Date(2025, 11, 5, 11, 0, 0, 0, time.UTC),
			Categories: []string{"Marketing", "Review", "Q4"},
		},
	},
	{
		Name:        "High Priority Event",
		Description: "Urgent meeting with priority 1 (highest)",
		ShouldPass:  true,
		Event: Event{
			Title:     "Emergency Planning",
			StartTime: time.Date(2025, 10, 30, 9, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 30, 10, 0, 0, 0, time.UTC),
			Priority:  1, // Highest priority
		},
	},
	{
		Name:        "Event at Midnight",
		Description: "Event starting at midnight",
		ShouldPass:  true,
		Event: Event{
			Title:     "Midnight Launch",
			StartTime: time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 12, 1, 1, 0, 0, 0, time.UTC),
		},
	},
	{
		Name:        "Multi-Day Event",
		Description: "Event spanning multiple days",
		ShouldPass:  true,
		Event: Event{
			Title:     "Training Week",
			StartTime: time.Date(2025, 11, 10, 9, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 11, 14, 17, 0, 0, 0, time.UTC),
			Location:  "Training Center",
		},
	},
	{
		Name:        "Event with Unicode in Title",
		Description: "Calendar event example",
		ShouldPass:  true,
		Event: Event{
			Title:     "Meeting 会議 встреча",
			StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
		},
	},
	{
		Name:        "Event with Special Characters in Location",
		Description: "Location: Joe's Café & Bistro",
		ShouldPass:  true,
		Event: Event{
			Title:     "Coffee Meeting",
			StartTime: time.Date(2025, 10, 28, 15, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 28, 16, 0, 0, 0, time.UTC),
			Location:  "Joe's Café & Bistro",
		},
	},
	{
		Name:        "Event with Newlines in Description",
		Description: "Multi-line description with agenda items",
		ShouldPass:  true,
		Event: Event{
			Title:       "Project Kickoff",
			StartTime:   time.Date(2025, 11, 1, 10, 0, 0, 0, time.UTC),
			EndTime:     time.Date(2025, 11, 1, 11, 0, 0, 0, time.UTC),
			Description: "Agenda:\n1. Introductions\n2. Project scope\n3. Timeline discussion",
		},
	},
	{
		Name:        "Event with URL in Description",
		Description: "Join: https://meet.google.com/abc-defg-hij",
		ShouldPass:  true,
		Event: Event{
			Title:       "Virtual Meeting",
			StartTime:   time.Date(2025, 11, 2, 14, 0, 0, 0, time.UTC),
			EndTime:     time.Date(2025, 11, 2, 15, 0, 0, 0, time.UTC),
			Description: "Join: https://meet.google.com/abc-defg-hij",
		},
	},
}

// InvalidTestCases contains test cases that should fail validation.
// These are used for testing error handling and validation logic.
var InvalidTestCases = []TestCase{
	{
		Name:       "Missing Title",
		ShouldPass: false,
		Event: Event{
			StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
		},
	},
	{
		Name:       "Missing Start Time",
		ShouldPass: false,
		Event: Event{
			Title:   "Test Event",
			EndTime: time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
		},
	},
	{
		Name:       "Missing End Time",
		ShouldPass: false,
		Event: Event{
			Title:     "Test Event",
			StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
		},
	},
	{
		Name:       "End Time Before Start Time",
		ShouldPass: false,
		Event: Event{
			Title:     "Invalid Event",
			StartTime: time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
		},
	},
}
