// Package apple provides URL generators for Apple apps.
//
// Apple Calendar supports data URIs with ICS (iCalendar) format for adding events.
// This is the most reliable cross-platform method for iOS and macOS.
package apple

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/joeblew999/wellknown/pkg/types"
)

// Calendar generates a data URI with ICS content for creating an Apple Calendar event.
// It validates the event and returns an error if validation fails.
//
// The generated URL uses the format:
// data:text/calendar;base64,<base64-encoded ICS content>
//
// This works on iOS and macOS devices and will prompt to open in the Calendar app.
func Calendar(event types.CalendarEvent) (string, error) {
	// Validate event
	if err := event.Validate(); err != nil {
		return "", err
	}

	// Generate ICS content
	icsContent := generateICS(event)

	// Encode as base64
	encoded := base64.StdEncoding.EncodeToString([]byte(icsContent))

	// Return data URI
	return fmt.Sprintf("data:text/calendar;base64,%s", encoded), nil
}

// generateICS creates ICS file content from a calendar event
func generateICS(event types.CalendarEvent) string {
	var buf bytes.Buffer

	// ICS file header
	buf.WriteString("BEGIN:VCALENDAR\r\n")
	buf.WriteString("VERSION:2.0\r\n")
	buf.WriteString("PRODID:-//wellknown//Calendar Event//EN\r\n")
	buf.WriteString("BEGIN:VEVENT\r\n")

	// Event details
	buf.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeICS(event.Title)))
	buf.WriteString(fmt.Sprintf("DTSTART:%s\r\n", formatICSTime(event.StartTime)))
	buf.WriteString(fmt.Sprintf("DTEND:%s\r\n", formatICSTime(event.EndTime)))

	// Optional fields
	if event.Location != "" {
		buf.WriteString(fmt.Sprintf("LOCATION:%s\r\n", escapeICS(event.Location)))
	}

	if event.Description != "" {
		buf.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeICS(event.Description)))
	}

	// Generate a simple UID using timestamp
	buf.WriteString(fmt.Sprintf("UID:%d@wellknown\r\n", time.Now().Unix()))

	// ICS file footer
	buf.WriteString("END:VEVENT\r\n")
	buf.WriteString("END:VCALENDAR\r\n")

	return buf.String()
}

// formatICSTime converts a time.Time to ICS format
// Format: YYYYMMDDTHHMMSSZ (UTC)
func formatICSTime(t time.Time) string {
	return t.UTC().Format("20060102T150405Z")
}

// escapeICS escapes special characters in ICS text fields
func escapeICS(s string) string {
	// ICS requires escaping of backslash, semicolon, comma, and newline
	s = bytes.NewBuffer([]byte(s)).String()
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
