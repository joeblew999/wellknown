package types

import "errors"

// Common validation errors
var (
	ErrMissingTitle     = errors.New("calendar event title is required")
	ErrMissingStartTime = errors.New("calendar event start time is required")
	ErrMissingEndTime   = errors.New("calendar event end time is required")
	ErrInvalidTimeRange = errors.New("calendar event end time must be after start time")
)
