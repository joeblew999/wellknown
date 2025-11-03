// Package calendar provides Google Calendar URL generation from validated form data.
package calendar

import (
	"fmt"
	"net/url"
	"time"

	cal "github.com/joeblew999/wellknown/pkg/calendar"
)

// Google Calendar URL constants (exported for tests)
const (
	BaseURL          = "https://calendar.google.com/calendar/render"
	ActionParam      = "TEMPLATE"
	TimeFormat       = "20060102T150405Z"
	QueryParamAction = "action"
	QueryParamDates  = "dates"
)

// Re-export shared field names from pkg/calendar for backwards compatibility
const (
	FieldTitle       = cal.FieldTitle
	FieldStart       = cal.FieldStart
	FieldEnd         = cal.FieldEnd
	FieldLocation    = cal.FieldLocation
	FieldDescription = cal.FieldDescription
)

// FieldMapping maps schema fields to Google Calendar URL parameters (exported for tests)
var FieldMapping = map[string]string{
	cal.FieldTitle:       "text",
	cal.FieldLocation:    "location",
	cal.FieldDescription: "details",
}

// GenerateURL creates a Google Calendar web URL from validated form data.
//
// Expected data fields (validated by schema.json):
//   - title: string (required)
//   - start: string in datetime-local format "2006-01-02T15:04" (required)
//   - end: string in datetime-local format "2006-01-02T15:04" (required)
//   - location: string (optional)
//   - description: string (optional)
//
// This function assumes data has already been validated against schema.json.
// It does NOT perform validation - that's the JSON Schema's job!
func GenerateURL(data map[string]interface{}) (string, error) {
	// Extract required fields from validated data
	title, ok := data[FieldTitle].(string)
	if !ok || title == "" {
		return "", fmt.Errorf("missing or invalid title field")
	}

	startStr, ok := data[FieldStart].(string)
	if !ok || startStr == "" {
		return "", fmt.Errorf("missing or invalid start field")
	}

	endStr, ok := data[FieldEnd].(string)
	if !ok || endStr == "" {
		return "", fmt.Errorf("missing or invalid end field")
	}

	// Parse datetime-local format: "2006-01-02T15:04"
	// This is the HTML5 datetime-local input format
	startTime, err := time.Parse("2006-01-02T15:04", startStr)
	if err != nil {
		return "", fmt.Errorf("invalid start time format: %w", err)
	}

	endTime, err := time.Parse("2006-01-02T15:04", endStr)
	if err != nil {
		return "", fmt.Errorf("invalid end time format: %w", err)
	}

	// Format times in Google Calendar format (UTC, ISO 8601: 20060102T150405Z)
	formattedStart := formatTime(startTime)
	formattedEnd := formatTime(endTime)

	// Build URL with parameters
	params := url.Values{}
	params.Set(QueryParamAction, ActionParam)
	params.Set(FieldMapping[FieldTitle], title)
	params.Set(QueryParamDates, fmt.Sprintf("%s/%s", formattedStart, formattedEnd))

	// Add optional fields if present
	if location, ok := data[FieldLocation].(string); ok && location != "" {
		params.Set(FieldMapping[FieldLocation], location)
	}

	if description, ok := data[FieldDescription].(string); ok && description != "" {
		params.Set(FieldMapping[FieldDescription], description)
	}

	return BaseURL + "?" + params.Encode(), nil
}


// formatTime converts a time.Time to Google Calendar format: 20060102T150405Z
// Google Calendar requires UTC time in this specific format
func formatTime(t time.Time) string {
	return t.UTC().Format(TimeFormat)
}
