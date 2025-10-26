package types

import "errors"

var (
	// ErrMissingTitle is returned when the event title is empty
	ErrMissingTitle = errors.New("event title is required")

	// ErrMissingStartTime is returned when the start time is zero
	ErrMissingStartTime = errors.New("event start time is required")

	// ErrMissingEndTime is returned when the end time is zero
	ErrMissingEndTime = errors.New("event end time is required")

	// ErrInvalidTimeRange is returned when end time is before start time
	ErrInvalidTimeRange = errors.New("end time must be after start time")
)
