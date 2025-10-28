package calendar

import (
	"strings"
	"testing"
)

func TestGenerateICS_WithAttendees(t *testing.T) {
	data := map[string]interface{}{
		"title":       "Team Meeting",
		"start":       "2025-11-15T10:00",
		"end":         "2025-11-15T11:00",
		"description": "Weekly sync",
		"attendees": []interface{}{
			map[string]interface{}{
				"email":    "john@example.com",
				"name":     "John Doe",
				"required": true,
			},
			map[string]interface{}{
				"email":    "jane@example.com",
				"name":     "Jane Smith",
				"required": false,
			},
		},
	}

	icsBytes, err := GenerateICS(data)
	if err != nil {
		t.Fatalf("GenerateICS failed: %v", err)
	}

	ics := string(icsBytes)

	// Verify attendee lines are present
	if !strings.Contains(ics, "ATTENDEE;CN=John Doe;ROLE=REQ-PARTICIPANT;RSVP=TRUE:mailto:john@example.com") {
		t.Error("Missing required attendee John Doe")
	}
	if !strings.Contains(ics, "ATTENDEE;CN=Jane Smith;ROLE=OPT-PARTICIPANT;RSVP=TRUE:mailto:jane@example.com") {
		t.Error("Missing optional attendee Jane Smith")
	}
}

func TestGenerateICS_WithRecurrence(t *testing.T) {
	data := map[string]interface{}{
		"title": "Weekly Standup",
		"start": "2025-11-15T10:00",
		"end":   "2025-11-15T10:30",
		"recurrence": map[string]interface{}{
			"frequency": "WEEKLY",
			"interval":  float64(1),
			"count":     float64(10),
		},
	}

	icsBytes, err := GenerateICS(data)
	if err != nil {
		t.Fatalf("GenerateICS failed: %v", err)
	}

	ics := string(icsBytes)

	// Verify RRULE is present (INTERVAL=1 is optional, defaults to 1)
	if !strings.Contains(ics, "RRULE:FREQ=WEEKLY") || !strings.Contains(ics, "COUNT=10") {
		t.Errorf("Missing RRULE with weekly recurrence. Got:\n%s", ics)
	}
}

func TestGenerateICS_WithRecurrence_UntilDate(t *testing.T) {
	data := map[string]interface{}{
		"title": "Daily Reminder",
		"start": "2025-11-15T09:00",
		"end":   "2025-11-15T09:15",
		"recurrence": map[string]interface{}{
			"frequency": "DAILY",
			"until":     "2025-12-31",
		},
	}

	icsBytes, err := GenerateICS(data)
	if err != nil {
		t.Fatalf("GenerateICS failed: %v", err)
	}

	ics := string(icsBytes)

	// Verify RRULE with UNTIL
	if !strings.Contains(ics, "RRULE:FREQ=DAILY;UNTIL=20251231") {
		t.Error("Missing RRULE with UNTIL date")
	}
}

func TestGenerateICS_WithReminders(t *testing.T) {
	data := map[string]interface{}{
		"title": "Important Meeting",
		"start": "2025-11-15T14:00",
		"end":   "2025-11-15T15:00",
		"reminders": []interface{}{
			map[string]interface{}{
				"minutesBefore": float64(15),
			},
			map[string]interface{}{
				"minutesBefore": float64(60),
			},
		},
	}

	icsBytes, err := GenerateICS(data)
	if err != nil {
		t.Fatalf("GenerateICS failed: %v", err)
	}

	ics := string(icsBytes)

	// Verify VALARM blocks are present
	if !strings.Contains(ics, "BEGIN:VALARM") {
		t.Error("Missing VALARM block")
	}
	if !strings.Contains(ics, "TRIGGER:-PT15M") {
		t.Error("Missing 15-minute reminder")
	}
	if !strings.Contains(ics, "TRIGGER:-PT60M") {
		t.Error("Missing 60-minute reminder")
	}
	if !strings.Contains(ics, "ACTION:DISPLAY") {
		t.Error("Missing ACTION:DISPLAY")
	}
	if !strings.Contains(ics, "END:VALARM") {
		t.Error("Missing END:VALARM")
	}
}

func TestGenerateICS_WithAllAdvancedFeatures(t *testing.T) {
	data := map[string]interface{}{
		"title":       "Comprehensive Test Event",
		"start":       "2025-11-15T10:00",
		"end":         "2025-11-15T11:00",
		"location":    "Conference Room A",
		"description": "Testing all advanced features",
		"attendees": []interface{}{
			map[string]interface{}{
				"email":    "alice@example.com",
				"name":     "Alice Johnson",
				"required": true,
			},
		},
		"recurrence": map[string]interface{}{
			"frequency": "WEEKLY",
			"interval":  float64(2),
			"count":     float64(5),
		},
		"reminders": []interface{}{
			map[string]interface{}{
				"minutesBefore": float64(30),
			},
		},
	}

	icsBytes, err := GenerateICS(data)
	if err != nil {
		t.Fatalf("GenerateICS failed: %v", err)
	}

	ics := string(icsBytes)

	// Verify all features are present
	tests := []struct {
		name     string
		contains string
	}{
		{"Basic fields", "SUMMARY:Comprehensive Test Event"},
		{"Location", "LOCATION:Conference Room A"},
		{"Description", "DESCRIPTION:Testing all advanced features"},
		{"Attendee", "ATTENDEE;CN=Alice Johnson;ROLE=REQ-PARTICIPANT"},
		{"Recurrence", "RRULE:FREQ=WEEKLY;INTERVAL=2;COUNT=5"},
		{"Reminder", "TRIGGER:-PT30M"},
	}

	for _, tt := range tests {
		if !strings.Contains(ics, tt.contains) {
			t.Errorf("%s: missing expected content: %s", tt.name, tt.contains)
		}
	}
}
