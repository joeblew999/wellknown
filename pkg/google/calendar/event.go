// Package calendar provides Google Calendar URL generation.
//
// Google Calendar does NOT support native deep links (like comgooglecalendar://)
// with event parameters. This package generates web URLs which are the ONLY
// working method for creating calendar events with pre-filled data.
//
// Limitations:
// - URL length constraints (typically 2048 characters)
// - No support for: recurring events, attendees, reminders
// - Only basic fields: title, times, location, description
package calendar

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

// Event represents a Google Calendar event with URL-compatible fields only.
// Google Calendar URLs have significant limitations compared to ICS format.
type Event struct {
	Title       string    // Required - Event title (max ~255 chars for URL safety)
	StartTime   time.Time // Required - Event start time (converted to UTC)
	EndTime     time.Time // Required - Event end time (converted to UTC)
	Location    string    // Optional - Event location (max ~255 chars for URL safety)
	Description string    // Optional - Event description (max ~1000 chars for URL safety)
}

// Validation errors
var (
	ErrMissingTitle      = errors.New("google calendar: event title is required")
	ErrMissingStartTime  = errors.New("google calendar: event start time is required")
	ErrMissingEndTime    = errors.New("google calendar: event end time is required")
	ErrInvalidTimeRange  = errors.New("google calendar: end time must be after start time")
	ErrTitleTooLong      = errors.New("google calendar: title too long for URL (max 255 characters)")
	ErrLocationTooLong   = errors.New("google calendar: location too long for URL (max 255 characters)")
	ErrDescriptionTooLong = errors.New("google calendar: description too long for URL (max 1000 characters)")
)

// Validate checks if the Event has all required fields and valid values.
// Google Calendar has stricter limits than ICS due to URL length constraints.
func (e Event) Validate() error {
	if e.Title == "" {
		return ErrMissingTitle
	}
	if len(e.Title) > 255 {
		return ErrTitleTooLong
	}

	if e.StartTime.IsZero() {
		return ErrMissingStartTime
	}
	if e.EndTime.IsZero() {
		return ErrMissingEndTime
	}
	if e.EndTime.Before(e.StartTime) || e.EndTime.Equal(e.StartTime) {
		return ErrInvalidTimeRange
	}

	if len(e.Location) > 255 {
		return ErrLocationTooLong
	}
	if len(e.Description) > 1000 {
		return ErrDescriptionTooLong
	}

	return nil
}

// GenerateURL creates a Google Calendar web URL for this event.
// The URL format: https://calendar.google.com/calendar/render?action=TEMPLATE&...
//
// This works on all devices and will prompt to open in the Google Calendar app if installed.
func (e Event) GenerateURL() (string, error) {
	// Validate event
	if err := e.Validate(); err != nil {
		return "", err
	}

	// Format times in Google Calendar format (UTC, ISO 8601)
	startTime := formatTime(e.StartTime)
	endTime := formatTime(e.EndTime)

	// Build URL with parameters
	baseURL := "https://calendar.google.com/calendar/render"
	params := url.Values{}
	params.Set("action", "TEMPLATE")
	params.Set("text", e.Title)
	params.Set("dates", fmt.Sprintf("%s/%s", startTime, endTime))

	// Add optional fields
	if e.Location != "" {
		params.Set("location", e.Location)
	}

	if e.Description != "" {
		params.Set("details", e.Description)
	}

	return baseURL + "?" + params.Encode(), nil
}

// formatTime converts a time.Time to Google Calendar's expected format.
// Format: YYYYMMDDTHHMMSSZ (UTC)
func formatTime(t time.Time) string {
	return t.UTC().Format("20060102T150405Z")
}
