// Package calendar provides Apple Calendar ICS file generation from validated form data.
package calendar

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"time"
)

// GenerateDataURI creates an Apple Calendar data URI from validated form data.
//
// Expected data fields (validated by schema.json):
//   - title: string (required)
//   - start: string in datetime-local format "2006-01-02T15:04" (required)
//   - end: string in datetime-local format "2006-01-02T15:04" (required)
//   - location: string (optional)
//   - description: string (optional)
//   - allDay: boolean (optional, default false)
//   - TODO: attendees, recurrence, reminders (future enhancement)
//
// This function assumes data has already been validated against schema.json.
// It does NOT perform validation - that's the JSON Schema's job!
//
// Returns a data URI: data:text/calendar;base64,<base64-encoded ICS>
func GenerateDataURI(data map[string]interface{}) (string, error) {
	// Generate ICS file content
	icsBytes, err := GenerateICS(data)
	if err != nil {
		return "", err
	}

	// Encode as base64 data URI
	encoded := base64.StdEncoding.EncodeToString(icsBytes)
	return fmt.Sprintf("data:text/calendar;base64,%s", encoded), nil
}

// GenerateICS creates an ICS file from validated form data.
// Returns raw ICS bytes that can be served as a download or encoded as data URI.
func GenerateICS(data map[string]interface{}) ([]byte, error) {
	// Extract required fields from validated data
	title, ok := data["title"].(string)
	if !ok || title == "" {
		return nil, fmt.Errorf("missing or invalid title field")
	}

	startStr, ok := data["start"].(string)
	if !ok || startStr == "" {
		return nil, fmt.Errorf("missing or invalid start field")
	}

	endStr, ok := data["end"].(string)
	if !ok || endStr == "" {
		return nil, fmt.Errorf("missing or invalid end field")
	}

	// Parse datetime-local format: "2006-01-02T15:04"
	startTime, err := time.Parse("2006-01-02T15:04", startStr)
	if err != nil {
		return nil, fmt.Errorf("invalid start time format: %w", err)
	}

	endTime, err := time.Parse("2006-01-02T15:04", endStr)
	if err != nil {
		return nil, fmt.Errorf("invalid end time format: %w", err)
	}

	// Check for all-day flag
	allDay := false
	if allDayVal, ok := data["allDay"].(bool); ok {
		allDay = allDayVal
	}

	// Build ICS file
	var buf bytes.Buffer

	// ICS file header
	buf.WriteString("BEGIN:VCALENDAR\r\n")
	buf.WriteString("VERSION:2.0\r\n")
	buf.WriteString("PRODID:-//wellknown//Calendar//EN\r\n")
	buf.WriteString("CALSCALE:GREGORIAN\r\n")
	buf.WriteString("METHOD:PUBLISH\r\n")

	// Event
	buf.WriteString("BEGIN:VEVENT\r\n")
	buf.WriteString(fmt.Sprintf("UID:%d@wellknown\r\n", time.Now().Unix()))
	buf.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", formatICSTime(time.Now())))

	// Date/time fields
	if allDay {
		// All-day events use VALUE=DATE format: YYYYMMDD
		buf.WriteString(fmt.Sprintf("DTSTART;VALUE=DATE:%s\r\n", formatICSDate(startTime)))
		buf.WriteString(fmt.Sprintf("DTEND;VALUE=DATE:%s\r\n", formatICSDate(endTime)))
	} else {
		// Regular events use UTC format: YYYYMMDDTHHMMSSZ
		buf.WriteString(fmt.Sprintf("DTSTART:%s\r\n", formatICSTime(startTime)))
		buf.WriteString(fmt.Sprintf("DTEND:%s\r\n", formatICSTime(endTime)))
	}

	// Summary (title)
	buf.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeICS(title)))

	// Optional fields
	if location, ok := data["location"].(string); ok && location != "" {
		buf.WriteString(fmt.Sprintf("LOCATION:%s\r\n", escapeICS(location)))
	}

	if description, ok := data["description"].(string); ok && description != "" {
		buf.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeICS(description)))
	}

	// TODO: Handle complex fields when form supports them
	// - attendees (array of objects)
	// - recurrence (object)
	// - reminders (array)

	// ICS file footer
	buf.WriteString("END:VEVENT\r\n")
	buf.WriteString("END:VCALENDAR\r\n")

	return buf.Bytes(), nil
}

// formatICSDate converts a time.Time to ICS DATE format: YYYYMMDD
// Used for all-day events
func formatICSDate(t time.Time) string {
	return t.Format("20060102")
}

// Note: formatICSTime() and escapeICS() are defined in event.go
// When event.go is deleted, move those functions here

// formatICSTime converts a time.Time to ICS format: YYYYMMDDTHHMMSSZ (UTC)
func formatICSTime(t time.Time) string {
	return t.UTC().Format("20060102T150405Z")
}

// escapeICS escapes special characters in ICS text fields per RFC 5545
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
