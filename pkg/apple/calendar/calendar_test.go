package calendar

import (
	"strings"
	"testing"
)

func TestGenerateICS(t *testing.T) {
	// Test data matching what the form would send
	data := map[string]interface{}{
		"title":       "Test Meeting",
		"start":       "2025-11-01T10:00",
		"end":         "2025-11-01T11:00",
		"location":    "Conference Room A",
		"description": "Discuss project roadmap",
	}

	// Generate ICS content
	icsBytes, err := GenerateICS(data)
	if err != nil {
		t.Fatalf("GenerateICS failed: %v", err)
	}

	ics := string(icsBytes)

	// Verify ICS content has required fields
	requiredFields := []string{
		"BEGIN:" + "VCALENDAR",
		"VERSION:" + ICSVersion,
		"BEGIN:VEVENT",
		"SUMMARY:Test Meeting",
		"DTSTART:20251101T100000Z",
		"DTEND:20251101T110000Z",
		"LOCATION:Conference Room A",
		"DESCRIPTION:Discuss project roadmap",
		"END:VEVENT",
		"END:VCALENDAR",
	}

	for _, field := range requiredFields {
		if !strings.Contains(ics, field) {
			t.Errorf("ICS content missing required field: %s\nGot:\n%s", field, ics)
		}
	}

	t.Logf("✅ Generated ICS content (%d bytes):\n%s", len(icsBytes), ics)
}

func TestGenerateICS_AllDay(t *testing.T) {
	// Test all-day event
	data := map[string]interface{}{
		"title":  "All Day Event",
		"start":  "2025-11-01T00:00",
		"end":    "2025-11-02T00:00",
		"allDay": true,
	}

	icsBytes, err := GenerateICS(data)
	if err != nil {
		t.Fatalf("GenerateICS failed: %v", err)
	}

	ics := string(icsBytes)

	// All-day events should use VALUE=DATE format
	if !strings.Contains(ics, "DTSTART;VALUE=DATE:20251101") {
		t.Errorf("All-day event should use VALUE=DATE format\nGot:\n%s", ics)
	}

	t.Logf("✅ Generated all-day ICS content:\n%s", ics)
}

func TestGenerateDataURI(t *testing.T) {
	data := map[string]interface{}{
		"title": "URI Test",
		"start": "2025-11-01T10:00",
		"end":   "2025-11-01T11:00",
	}

	dataURI, err := GenerateDataURI(data)
	if err != nil {
		t.Fatalf("GenerateDataURI failed: %v", err)
	}

	// Should be base64 encoded data URI
	if !strings.HasPrefix(dataURI, "data:text/calendar;base64,") {
		t.Errorf("Data URI should start with 'data:text/calendar;base64,'\nGot: %s", dataURI[:50])
	}

	t.Logf("✅ Generated data URI (%d bytes): %s...", len(dataURI), dataURI[:50])
}

func TestGenerateICS_MissingFields(t *testing.T) {
	// Test missing required field
	data := map[string]interface{}{
		"title": "Missing Times",
		// start and end are missing
	}

	_, err := GenerateICS(data)
	if err == nil {
		t.Fatal("GenerateICS should fail with missing required fields")
	}

	t.Logf("✅ Correctly rejected invalid data: %v", err)
}
