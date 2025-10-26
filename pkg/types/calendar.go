// Package types provides shared data structures used across all platform packages.
package types

import "time"

// CalendarEvent represents a calendar event with standard fields
// that can be converted to platform-specific deep links.
//
// This struct is used by all platform implementations (Google, Apple, Web)
// to generate calendar deep links.
type CalendarEvent struct {
	// Title is the event name/subject (required)
	Title string

	// StartTime is when the event begins (required)
	StartTime time.Time

	// EndTime is when the event ends (required)
	EndTime time.Time

	// Location is the event location (optional)
	// Example: "Conference Room A" or "123 Main St, City"
	Location string

	// Description provides additional details about the event (optional)
	Description string
}

// Validate checks if the CalendarEvent has all required fields
// and returns an error if validation fails.
func (e CalendarEvent) Validate() error {
	if e.Title == "" {
		return ErrMissingTitle
	}
	if e.StartTime.IsZero() {
		return ErrMissingStartTime
	}
	if e.EndTime.IsZero() {
		return ErrMissingEndTime
	}
	if e.EndTime.Before(e.StartTime) {
		return ErrInvalidTimeRange
	}
	return nil
}
