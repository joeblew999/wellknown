// Package calendar provides Google Calendar URL generation from validated form data.
package calendar

import (
	"fmt"
	"net/url"
	"time"
)

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
	title, ok := data["title"].(string)
	if !ok || title == "" {
		return "", fmt.Errorf("missing or invalid title field")
	}

	startStr, ok := data["start"].(string)
	if !ok || startStr == "" {
		return "", fmt.Errorf("missing or invalid start field")
	}

	endStr, ok := data["end"].(string)
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
	baseURL := "https://calendar.google.com/calendar/render"
	params := url.Values{}
	params.Set("action", "TEMPLATE")
	params.Set("text", title)
	params.Set("dates", fmt.Sprintf("%s/%s", formattedStart, formattedEnd))

	// Add optional fields if present
	if location, ok := data["location"].(string); ok && location != "" {
		params.Set("location", location)
	}

	if description, ok := data["description"].(string); ok && description != "" {
		params.Set("details", description)
	}

	return baseURL + "?" + params.Encode(), nil
}

