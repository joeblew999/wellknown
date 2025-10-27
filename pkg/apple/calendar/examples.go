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

// ExpectedURL generates and returns the expected Apple Calendar data URI for this example
func (e Example) ExpectedURL() string {
	ics, _ := e.Event.GenerateICS()
	return string(ics)
}

// Helper function to create a pointer to an int
func ptrInt(i int) *int {
	return &i
}

// Examples provides realistic Apple Calendar event examples for the web demo showcase.
// These examples showcase both basic events AND advanced ICS features like
// recurring events, attendees, and reminders.
var Examples = []Example{
	{
		Name:        "Team Meeting",
		Description: "Simple 1-hour meeting (basic fields only)",
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
		Name:        "Weekly Standup",
		Description: "Recurring meeting every Monday with reminders",
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
		Name:        "Product Launch",
		Description: "Important event with multiple attendees and high priority",
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
		Name:        "Daily Exercise",
		Description: "Recurring daily reminder",
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
		Name:        "Monthly Board Meeting",
		Description: "First Monday of every month",
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
				Count:      ptrInt(12), // 12 months
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
		Name:        "Coffee Chat",
		Description: "Casual 1:1 conversation",
		Event: Event{
			Title:       "Coffee Chat",
			StartTime:   time.Date(2025, 10, 28, 15, 30, 0, 0, time.UTC),
			EndTime:     time.Date(2025, 10, 28, 16, 0, 0, 0, time.UTC),
			Location:    "Starbucks Downtown",
			Description: "Catch up and discuss career goals",
			Status:      StatusConfirmed,
		},
	},
}
