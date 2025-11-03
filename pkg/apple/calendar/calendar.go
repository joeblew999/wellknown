// Package calendar provides Apple Calendar ICS file generation from validated form data.
package calendar

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"time"

	cal "github.com/joeblew999/wellknown/pkg/calendar"
)

// Apple Calendar ICS constants (exported for tests and code generation)
const (
	ProductID      = "-//wellknown//Calendar//EN"
	CalendarScale  = "GREGORIAN"
	ICSVersion     = "2.0"
	DateFormat     = "20060102"
	DateTimeFormat = "20060102T150405Z"
)

// ICS format keywords (RFC 5545) - exported for test data generation
const (
	ICSBeginCalendar = "BEGIN:VCALENDAR"
	ICSEndCalendar   = "END:VCALENDAR"
	ICSBeginEvent    = "BEGIN:VEVENT"
	ICSEndEvent      = "END:VEVENT"
	ICSSummary       = "SUMMARY"
	ICSRule          = "RRULE"
	ICSAttendee      = "ATTENDEE"
)

// Re-export shared field names from pkg/calendar for backwards compatibility
const (
	FieldTitle       = cal.FieldTitle
	FieldStart       = cal.FieldStart
	FieldEnd         = cal.FieldEnd
	FieldLocation    = cal.FieldLocation
	FieldDescription = cal.FieldDescription
	FieldAllDay      = cal.FieldAllDay
	FieldAttendees   = cal.FieldAttendees
	FieldRecurrence  = cal.FieldRecurrence
	FieldReminders   = cal.FieldReminders
	FieldOrganizer   = cal.FieldOrganizer
	FieldAlarm       = cal.FieldAlarm
)

// AdvancedFeatures lists field names that indicate advanced calendar functionality
// Used for automatic test categorization (basic vs advanced)
// Re-exported from pkg/calendar for backwards compatibility
var AdvancedFeatures = cal.AdvancedFields

// GenerateDownloadURL creates a download URL for an Apple Calendar .ics file.
// This is the CORRECT approach for iOS/macOS - Safari cannot handle data:text/calendar URIs.
//
// Returns a URL like: /apple/calendar/download?event=<base64_encoded_ics>
func GenerateDownloadURL(data map[string]interface{}) (string, error) {
	icsBytes, err := GenerateICS(data)
	if err != nil {
		return "", err
	}
	encoded := base64.URLEncoding.EncodeToString(icsBytes)
	return "/apple/calendar/download?event=" + encoded, nil
}

// GenerateDataURI creates an Apple Calendar data URI from validated form data.
// NOTE: This doesn't work on Safari/iOS! Use GenerateDownloadURL() instead.
//
// Returns a data URI: data:text/calendar;base64,<base64-encoded ICS>
func GenerateDataURI(data map[string]interface{}) (string, error) {
	icsBytes, err := GenerateICS(data)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(icsBytes)
	return fmt.Sprintf("data:text/calendar;base64,%s", encoded), nil
}

// GenerateICS creates an ICS file from validated form data.
// Returns raw ICS bytes that can be served as a download or encoded as data URI.
func GenerateICS(data map[string]interface{}) ([]byte, error) {
	// Extract required fields from validated data
	title, ok := data[FieldTitle].(string)
	if !ok || title == "" {
		return nil, fmt.Errorf("missing or invalid title field")
	}

	startStr, ok := data[FieldStart].(string)
	if !ok || startStr == "" {
		return nil, fmt.Errorf("missing or invalid start field")
	}

	endStr, ok := data[FieldEnd].(string)
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
	if allDayVal, ok := data[FieldAllDay].(bool); ok {
		allDay = allDayVal
	}

	// Build ICS file
	var buf bytes.Buffer

	// ICS file header
	buf.WriteString(ICSBeginCalendar + "\r\n")
	buf.WriteString("VERSION:" + ICSVersion + "\r\n")
	buf.WriteString("PRODID:" + ProductID + "\r\n")
	buf.WriteString("CALSCALE:" + CalendarScale + "\r\n")
	buf.WriteString("METHOD:PUBLISH\r\n")

	// Event
	buf.WriteString(ICSBeginEvent + "\r\n")
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
	if location, ok := data[FieldLocation].(string); ok && location != "" {
		buf.WriteString(fmt.Sprintf("LOCATION:%s\r\n", escapeICS(location)))
	}

	if description, ok := data[FieldDescription].(string); ok && description != "" {
		buf.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeICS(description)))
	}

	// Handle attendees (array of objects)
	if attendeesRaw, ok := data[FieldAttendees]; ok {
		if attendees, ok := attendeesRaw.([]interface{}); ok {
			for _, attendeeRaw := range attendees {
				if attendee, ok := attendeeRaw.(map[string]interface{}); ok {
					email, _ := attendee["email"].(string)
					name, _ := attendee["name"].(string)
					required, _ := attendee["required"].(bool)

					if email != "" {
						// Build ATTENDEE line
						var attendeeLine string
						if name != "" {
							attendeeLine = fmt.Sprintf("%s;CN=%s", ICSAttendee, escapeICS(name))
						} else {
							attendeeLine = ICSAttendee
						}

						// Add role (REQ-PARTICIPANT or OPT-PARTICIPANT)
						if required {
							attendeeLine += ";ROLE=REQ-PARTICIPANT"
						} else {
							attendeeLine += ";ROLE=OPT-PARTICIPANT"
						}

						// Add RSVP and email
						attendeeLine += fmt.Sprintf(";RSVP=TRUE:mailto:%s\r\n", email)
						buf.WriteString(attendeeLine)
					}
				}
			}
		}
	}

	// Handle recurrence (object)
	if recurrenceRaw, ok := data[FieldRecurrence]; ok {
		if recurrence, ok := recurrenceRaw.(map[string]interface{}); ok {
			frequency, _ := recurrence["frequency"].(string)
			if frequency != "" {
				// Start RRULE
				rrule := fmt.Sprintf("%s:FREQ=%s", ICSRule, frequency)

				// Add interval if specified
				if interval, ok := recurrence["interval"].(float64); ok && interval > 1 {
					rrule += fmt.Sprintf(";INTERVAL=%d", int(interval))
				}

				// Add count (number of occurrences)
				if count, ok := recurrence["count"].(float64); ok && count > 0 {
					rrule += fmt.Sprintf(";COUNT=%d", int(count))
				}

				// Add until date
				if until, ok := recurrence["until"].(string); ok && until != "" {
					// Parse date format YYYY-MM-DD
					if untilTime, err := time.Parse("2006-01-02", until); err == nil {
						rrule += fmt.Sprintf(";UNTIL=%s", formatICSDate(untilTime))
					}
				}

				buf.WriteString(rrule + "\r\n")
			}
		}
	}

	// Handle reminders (array of objects)
	if remindersRaw, ok := data[FieldReminders]; ok {
		if reminders, ok := remindersRaw.([]interface{}); ok {
			for _, reminderRaw := range reminders {
				if reminder, ok := reminderRaw.(map[string]interface{}); ok {
					minutesBefore, ok := reminder["minutesBefore"].(float64)
					if ok && minutesBefore >= 0 {
						// Convert minutes to duration format (e.g., PT15M)
						buf.WriteString("BEGIN:VALARM\r\n")
						buf.WriteString("ACTION:DISPLAY\r\n")
						buf.WriteString(fmt.Sprintf("TRIGGER:-PT%dM\r\n", int(minutesBefore)))
						buf.WriteString("END:VALARM\r\n")
					}
				}
			}
		}
	}

	// ICS file footer
	buf.WriteString(ICSEndEvent + "\r\n")
	buf.WriteString(ICSEndCalendar + "\r\n")

	return buf.Bytes(), nil
}

// formatICSDate converts a time.Time to ICS DATE format: YYYYMMDD
// Used for all-day events
func formatICSDate(t time.Time) string {
	return t.Format(DateFormat)
}

// Note: formatICSTime() and escapeICS() are defined in event.go
// When event.go is deleted, move those functions here

// formatICSTime converts a time.Time to ICS format: YYYYMMDDTHHMMSSZ (UTC)
func formatICSTime(t time.Time) string {
	return t.UTC().Format(DateTimeFormat)
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
