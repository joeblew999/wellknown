// Package calendar provides Apple Calendar ICS generation with full iCalendar spec support.
//
// Unlike Google Calendar URLs, ICS format supports advanced features:
// - Recurring events (daily, weekly, monthly, yearly)
// - Multiple attendees with RSVP
// - Multiple reminders
// - Event status (confirmed, tentative, cancelled)
// - Priority levels
// - And much more!
package calendar

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Event represents a full-featured Apple Calendar event with ICS capabilities.
// This supports the complete iCalendar (RFC 5545) specification.
type Event struct {
	// Basic fields (required/common)
	Title       string    // Event title/summary (required)
	StartTime   time.Time // Event start time (required)
	EndTime     time.Time // Event end time (required)
	Location    string    // Event location (optional)
	Description string    // Event description (optional)

	// Advanced ICS-only fields
	Organizer   *Organizer      // Event organizer (optional)
	Attendees   []Attendee      // List of attendees (optional)
	Recurrence  *RecurrenceRule // Recurring event rule (optional)
	Reminders   []Reminder      // Event reminders/alarms (optional)
	Status      EventStatus     // Event status (default: CONFIRMED)
	Priority    int             // Priority 0-9 (0=undefined, 1=highest, 9=lowest)
	AllDay      bool            // All-day event flag
	URL         string          // Related URL (optional)
	Categories  []string        // Event categories/tags (optional)
}

// Validation errors
var (
	ErrMissingTitle     = errors.New("apple calendar: event title is required")
	ErrMissingStartTime = errors.New("apple calendar: event start time is required")
	ErrMissingEndTime   = errors.New("apple calendar: event end time is required")
	ErrInvalidTimeRange = errors.New("apple calendar: end time must be after start time")
	ErrInvalidPriority  = errors.New("apple calendar: priority must be 0-9")
)

// Validate checks if the Event has all required fields and valid values.
func (e Event) Validate() error {
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
	if e.Priority < 0 || e.Priority > 9 {
		return ErrInvalidPriority
	}
	return nil
}

// GenerateDataURI creates a data URI with ICS content for this event.
// Format: data:text/calendar;base64,<encoded ICS>
//
// This works on iOS and macOS and will prompt to open in the Calendar app.
func (e Event) GenerateDataURI() (string, error) {
	icsContent, err := e.GenerateICS()
	if err != nil {
		return "", err
	}

	// Encode as base64
	encoded := base64.StdEncoding.EncodeToString(icsContent)

	return fmt.Sprintf("data:text/calendar;base64,%s", encoded), nil
}

// GenerateICS creates ICS file content (RFC 5545 format) for this event.
//
// NOTE: This method does NOT validate the event. Validation should be done via JSON Schema
// before calling this method. The Validate() method is kept for backward compatibility
// with existing unit tests but should not be called in production code.
func (e Event) GenerateICS() ([]byte, error) {
	var buf bytes.Buffer

	// ICS file header
	buf.WriteString("BEGIN:VCALENDAR\r\n")
	buf.WriteString("VERSION:2.0\r\n")
	buf.WriteString("PRODID:-//wellknown//Apple Calendar//EN\r\n")
	buf.WriteString("CALSCALE:GREGORIAN\r\n")
	buf.WriteString("METHOD:PUBLISH\r\n")
	buf.WriteString("BEGIN:VEVENT\r\n")

	// Required fields
	// Generate deterministic UID based on event content
	// This ensures the same event always gets the same UID (deterministic)
	// But different events get different UIDs (unique)
	uidContent := fmt.Sprintf("%s|%s|%s", e.Title, formatICSTime(e.StartTime), formatICSTime(e.EndTime))
	uidHash := fmt.Sprintf("%x", []byte(uidContent)) // Simple hash
	buf.WriteString(fmt.Sprintf("UID:%s@wellknown\r\n", uidHash))
	buf.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", formatICSTime(time.Now())))
	buf.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeICS(e.Title)))
	buf.WriteString(fmt.Sprintf("DTSTART:%s\r\n", formatICSTime(e.StartTime)))
	buf.WriteString(fmt.Sprintf("DTEND:%s\r\n", formatICSTime(e.EndTime)))

	// Optional basic fields
	if e.Location != "" {
		buf.WriteString(fmt.Sprintf("LOCATION:%s\r\n", escapeICS(e.Location)))
	}
	if e.Description != "" {
		buf.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeICS(e.Description)))
	}
	if e.URL != "" {
		buf.WriteString(fmt.Sprintf("URL:%s\r\n", e.URL))
	}

	// Status
	if e.Status != "" {
		buf.WriteString(fmt.Sprintf("STATUS:%s\r\n", e.Status))
	} else {
		buf.WriteString("STATUS:CONFIRMED\r\n")
	}

	// Priority
	if e.Priority > 0 {
		buf.WriteString(fmt.Sprintf("PRIORITY:%d\r\n", e.Priority))
	}

	// Categories
	if len(e.Categories) > 0 {
		buf.WriteString(fmt.Sprintf("CATEGORIES:%s\r\n", strings.Join(e.Categories, ",")))
	}

	// Organizer
	if e.Organizer != nil {
		if e.Organizer.Name != "" {
			buf.WriteString(fmt.Sprintf("ORGANIZER;CN=%s:mailto:%s\r\n", e.Organizer.Name, e.Organizer.Email))
		} else {
			buf.WriteString(fmt.Sprintf("ORGANIZER:mailto:%s\r\n", e.Organizer.Email))
		}
	}

	// Attendees
	for _, att := range e.Attendees {
		line := "ATTENDEE"
		if att.Name != "" {
			line += fmt.Sprintf(";CN=%s", att.Name)
		}
		if att.Role != "" {
			line += fmt.Sprintf(";ROLE=%s", att.Role)
		}
		if att.Status != "" {
			line += fmt.Sprintf(";PARTSTAT=%s", att.Status)
		}
		if att.RSVP {
			line += ";RSVP=TRUE"
		}
		line += fmt.Sprintf(":mailto:%s\r\n", att.Email)
		buf.WriteString(line)
	}

	// Recurrence rule
	if e.Recurrence != nil {
		buf.WriteString(fmt.Sprintf("RRULE:%s\r\n", formatRRule(e.Recurrence)))
	}

	// Reminders (VALARM)
	for _, reminder := range e.Reminders {
		buf.WriteString("BEGIN:VALARM\r\n")
		buf.WriteString("ACTION:DISPLAY\r\n")
		buf.WriteString("DESCRIPTION:Reminder\r\n")
		buf.WriteString(fmt.Sprintf("TRIGGER:-PT%dM\r\n", int(reminder.Duration.Minutes())))
		buf.WriteString("END:VALARM\r\n")
	}

	// ICS file footer
	buf.WriteString("END:VEVENT\r\n")
	buf.WriteString("END:VCALENDAR\r\n")

	return buf.Bytes(), nil
}

// formatICSTime converts a time.Time to ICS format: YYYYMMDDTHHMMSSZ (UTC)
func formatICSTime(t time.Time) string {
	return t.UTC().Format("20060102T150405Z")
}

// escapeICS escapes special characters in ICS text fields
func escapeICS(s string) string {
	var buf bytes.Buffer
	for _, r := range s {
		switch r {
		case '\\':
			buf.WriteString("\\\\")
		case ';':
			buf.WriteString("\\;")
		case ',':
			buf.WriteString("\\,")
		case '\n':
			buf.WriteString("\\n")
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

// formatRRule formats a RecurrenceRule as an RRULE string
func formatRRule(r *RecurrenceRule) string {
	parts := []string{fmt.Sprintf("FREQ=%s", r.Frequency)}

	if r.Interval > 1 {
		parts = append(parts, fmt.Sprintf("INTERVAL=%d", r.Interval))
	}

	if r.Count != nil {
		parts = append(parts, fmt.Sprintf("COUNT=%d", *r.Count))
	}

	if r.Until != nil {
		parts = append(parts, fmt.Sprintf("UNTIL=%s", formatICSTime(*r.Until)))
	}

	if len(r.ByDay) > 0 {
		days := make([]string, len(r.ByDay))
		dayMap := map[time.Weekday]string{
			time.Sunday:    "SU",
			time.Monday:    "MO",
			time.Tuesday:   "TU",
			time.Wednesday: "WE",
			time.Thursday:  "TH",
			time.Friday:    "FR",
			time.Saturday:  "SA",
		}
		for i, day := range r.ByDay {
			days[i] = dayMap[day]
		}
		parts = append(parts, fmt.Sprintf("BYDAY=%s", strings.Join(days, ",")))
	}

	if len(r.ByMonthDay) > 0 {
		days := make([]string, len(r.ByMonthDay))
		for i, day := range r.ByMonthDay {
			days[i] = fmt.Sprintf("%d", day)
		}
		parts = append(parts, fmt.Sprintf("BYMONTHDAY=%s", strings.Join(days, ",")))
	}

	return strings.Join(parts, ";")
}
