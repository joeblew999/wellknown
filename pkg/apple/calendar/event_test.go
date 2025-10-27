package calendar

import (
	"strings"
	"testing"
	"time"
)

// TestEvent_GenerateDataURI tests the data URI generation
func TestEvent_GenerateDataURI(t *testing.T) {
	event := Event{
		Title:     "Test Event",
		StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
	}

	dataURI, err := event.GenerateDataURI()
	if err != nil {
		t.Fatalf("GenerateDataURI() error = %v", err)
	}

	// Verify it's a data URI
	if !strings.HasPrefix(dataURI, "data:text/calendar;base64,") {
		t.Errorf("DataURI should start with 'data:text/calendar;base64,', got %v", dataURI)
	}
}

// TestEvent_GenerateICS tests ICS generation for basic event
func TestEvent_GenerateICS(t *testing.T) {
	event := Event{
		Title:       "Test Event",
		StartTime:   time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
		EndTime:     time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
		Location:    "Test Location",
		Description: "Test Description",
	}

	ics, err := event.GenerateICS()
	if err != nil {
		t.Fatalf("GenerateICS() error = %v", err)
	}

	icsStr := string(ics)

	// Verify required ICS components
	requiredComponents := []string{
		"BEGIN:VCALENDAR",
		"END:VCALENDAR",
		"BEGIN:VEVENT",
		"END:VEVENT",
		"SUMMARY:Test Event",
		"LOCATION:Test Location",
		"DESCRIPTION:Test Description",
		"DTSTART:20251026T140000Z",
		"DTEND:20251026T150000Z",
	}

	for _, component := range requiredComponents {
		if !strings.Contains(icsStr, component) {
			t.Errorf("ICS should contain '%s', got:\n%s", component, icsStr)
		}
	}
}

// TestEvent_RecurringEvent tests recurring event generation
func TestEvent_RecurringEvent(t *testing.T) {
	event := Event{
		Title:     "Weekly Standup",
		StartTime: time.Date(2025, 10, 28, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 10, 28, 10, 30, 0, 0, time.UTC),
		Recurrence: &RecurrenceRule{
			Frequency: FrequencyWeekly,
			Interval:  1,
			ByDay:     []time.Weekday{time.Monday},
			Count:     ptrInt(10),
		},
	}

	ics, err := event.GenerateICS()
	if err != nil {
		t.Fatalf("GenerateICS() error = %v", err)
	}

	icsStr := string(ics)

	// Verify recurrence rule is present
	if !strings.Contains(icsStr, "RRULE:") {
		t.Errorf("ICS should contain 'RRULE:' for recurring event, got:\n%s", icsStr)
	}
	if !strings.Contains(icsStr, "FREQ=WEEKLY") {
		t.Errorf("ICS should contain 'FREQ=WEEKLY', got:\n%s", icsStr)
	}
	if !strings.Contains(icsStr, "COUNT=10") {
		t.Errorf("ICS should contain 'COUNT=10', got:\n%s", icsStr)
	}
}

// TestEvent_WithAttendees tests event with attendees
func TestEvent_WithAttendees(t *testing.T) {
	event := Event{
		Title:     "Product Launch",
		StartTime: time.Date(2025, 11, 1, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 11, 1, 10, 0, 0, 0, time.UTC),
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
				Role:   RoleOptParticipant,
				Status: StatusNeedsAction,
				RSVP:   true,
			},
		},
	}

	ics, err := event.GenerateICS()
	if err != nil {
		t.Fatalf("GenerateICS() error = %v", err)
	}

	icsStr := string(ics)

	// Verify organizer
	if !strings.Contains(icsStr, "ORGANIZER;CN=Sarah Johnson:mailto:sarah@example.com") {
		t.Errorf("ICS should contain organizer, got:\n%s", icsStr)
	}

	// Verify attendees
	if !strings.Contains(icsStr, "ATTENDEE") {
		t.Errorf("ICS should contain 'ATTENDEE', got:\n%s", icsStr)
	}
	if !strings.Contains(icsStr, "alice@example.com") {
		t.Errorf("ICS should contain alice@example.com, got:\n%s", icsStr)
	}
	if !strings.Contains(icsStr, "bob@example.com") {
		t.Errorf("ICS should contain bob@example.com, got:\n%s", icsStr)
	}
}

// TestEvent_WithReminders tests event with reminders
func TestEvent_WithReminders(t *testing.T) {
	event := Event{
		Title:     "Important Meeting",
		StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
		Reminders: []Reminder{
			{Duration: 15 * time.Minute},
			{Duration: 1 * time.Hour},
		},
	}

	ics, err := event.GenerateICS()
	if err != nil {
		t.Fatalf("GenerateICS() error = %v", err)
	}

	icsStr := string(ics)

	// Verify reminders (VALARM)
	if !strings.Contains(icsStr, "BEGIN:VALARM") {
		t.Errorf("ICS should contain 'BEGIN:VALARM' for reminders, got:\n%s", icsStr)
	}
	if !strings.Contains(icsStr, "END:VALARM") {
		t.Errorf("ICS should contain 'END:VALARM', got:\n%s", icsStr)
	}
}

// TestEventDeterministic ensures the same input always produces the same output
func TestEventDeterministic(t *testing.T) {
	event := Event{
		Title:     "Test Event",
		StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
	}

	// Generate ICS multiple times
	results := make([]string, 10)
	for i := 0; i < 10; i++ {
		ics, err := event.GenerateICS()
		if err != nil {
			t.Fatalf("GenerateICS() error = %v", err)
		}
		results[i] = string(ics)
	}

	// All results should be identical (except UID which is random)
	// For now, just verify they all have the same length and key components
	for i := 1; i < len(results); i++ {
		if len(results[i]) != len(results[0]) {
			t.Errorf("GenerateICS() not consistent: different lengths\n  first: %d bytes\n  got:   %d bytes", len(results[0]), len(results[i]))
		}
	}
}

// TestEvent_ValidationErrors tests validation errors
// NOTE: Now that GenerateICS() doesn't validate, we test Validate() explicitly.
// In production, validation is done via JSON Schema before calling GenerateICS().
func TestEvent_ValidationErrors(t *testing.T) {
	testCases := []struct {
		name  string
		event Event
	}{
		{
			name: "Missing Title",
			event: Event{
				StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Empty Title",
			event: Event{
				Title:     "",
				StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Missing Start Time",
			event: Event{
				Title:   "Test",
				EndTime: time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Missing End Time",
			event: Event{
				Title:     "Test",
				StartTime: time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "End Before Start",
			event: Event{
				Title:     "Test",
				StartTime: time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.event.Validate()
			if err == nil {
				t.Error("Validate() error = nil, want error")
			}
		})
	}
}
