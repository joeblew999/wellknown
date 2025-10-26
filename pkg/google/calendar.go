// Package google provides URL generators for Google apps.
//
// Note: Google Calendar does NOT support native deep links (like comgooglecalendar://)
// with event parameters. This package generates web URLs (https://calendar.google.com)
// which are the ONLY working method for creating calendar events with pre-filled data.
package google

import (
	"time"

	"fmt"
	"net/url"

	"github.com/joeblew999/wellknown/pkg/types"
)

// Calendar generates a Google Calendar web URL for creating an event.
// It validates the event and returns an error if validation fails.
//
// The generated URL uses the format:
// https://calendar.google.com/calendar/render?action=TEMPLATE&...
//
// This works on all devices and will prompt to open in the Google Calendar app if installed.
func Calendar(event types.CalendarEvent) (string, error) {
	// Validate event
	if err := event.Validate(); err != nil {
		return "", err
	}

	// Format times in Google Calendar format (UTC, ISO 8601)
	startTime := formatTime(event.StartTime)
	endTime := formatTime(event.EndTime)

	// Build URL with parameters
	baseURL := "https://calendar.google.com/calendar/render"
	params := url.Values{}
	params.Set("action", "TEMPLATE")
	params.Set("text", event.Title)
	params.Set("dates", fmt.Sprintf("%s/%s", startTime, endTime))

	// Add optional fields
	if event.Location != "" {
		params.Set("location", event.Location)
	}

	if event.Description != "" {
		params.Set("details", event.Description)
	}

	return baseURL + "?" + params.Encode(), nil
}

// formatTime converts a time.Time to Google Calendar's expected format
// Format: YYYYMMDDTHHMMSSZ (UTC)
func formatTime(t time.Time) string {
	return t.UTC().Format("20060102T150405Z")
}
