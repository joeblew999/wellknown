package calendar

import (
	"strings"
	"testing"
)

func TestGenerateURL(t *testing.T) {
	// Test data matching what the form would send
	data := map[string]interface{}{
		"title":       "Team Meeting",
		"start":       "2025-11-01T14:00",
		"end":         "2025-11-01T15:30",
		"location":    "Conference Room A",
		"description": "Quarterly planning session",
	}

	// Generate URL
	url, err := GenerateURL(data)
	if err != nil{
		t.Fatalf("GenerateURL failed: %v", err)
	}

	// Verify URL structure
	if !strings.HasPrefix(url, "https://calendar.google.com/calendar/render?") {
		t.Errorf("URL should start with Google Calendar base URL\nGot: %s", url)
	}

	// Verify required parameters
	requiredParams := []string{
		"action=TEMPLATE",
		"text=Team+Meeting",
		"dates=20251101T140000Z",  // UTC format
		"location=Conference+Room+A",
		"details=Quarterly+planning+session",
	}

	for _, param := range requiredParams {
		if !strings.Contains(url, param) {
			t.Errorf("URL missing required parameter: %s\nGot: %s", param, url)
		}
	}

	t.Logf("✅ Generated URL (%d bytes):\n%s", len(url), url)
}

func TestGenerateURL_MinimalFields(t *testing.T) {
	// Test with only required fields
	data := map[string]interface{}{
		"title": "Quick Meeting",
		"start": "2025-11-01T10:00",
		"end":   "2025-11-01T10:30",
	}

	url, err := GenerateURL(data)
	if err != nil {
		t.Fatalf("GenerateURL failed: %v", err)
	}

	// Should NOT contain location or details
	if strings.Contains(url, "location=") {
		t.Errorf("URL should not contain location parameter")
	}
	if strings.Contains(url, "details=") {
		t.Errorf("URL should not contain details parameter")
	}

	t.Logf("✅ Generated minimal URL: %s", url)
}

func TestGenerateURL_MissingFields(t *testing.T) {
	// Test missing required field
	data := map[string]interface{}{
		"title": "Missing Times",
		// start and end are missing
	}

	_, err := GenerateURL(data)
	if err == nil {
		t.Fatal("GenerateURL should fail with missing required fields")
	}

	t.Logf("✅ Correctly rejected invalid data: %v", err)
}

func TestGenerateURL_InvalidTimeFormat(t *testing.T) {
	// Test invalid datetime format
	data := map[string]interface{}{
		"title": "Bad Time Format",
		"start": "invalid-date",
		"end":   "2025-11-01T10:30",
	}

	_, err := GenerateURL(data)
	if err == nil {
		t.Fatal("GenerateURL should fail with invalid time format")
	}

	t.Logf("✅ Correctly rejected invalid time format: %v", err)
}
