package types

import "time"

// CalendarEvent represents a calendar event with all necessary fields
type CalendarEvent struct {
	Title       string
	StartTime   time.Time
	EndTime     time.Time
	Location    string
	Description string
}

// Validate checks if the CalendarEvent has all required fields and valid values
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
	if e.EndTime.Before(e.StartTime) || e.EndTime.Equal(e.StartTime) {
		return ErrInvalidTimeRange
	}
	return nil
}
