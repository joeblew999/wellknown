// Package google provides deep link generators for Google apps.
package google

import (
	"bytes"
	_ "embed"
	"net/url"
	"text/template"
	"time"

	"github.com/joeblew999/wellknown/pkg/types"
)

//go:embed calendar.tmpl
var calendarTemplate string

// Calendar generates a Google Calendar deep link URL for the given event.
//
// ⚠️ WARNING: Google Calendar native deep links (googlecalendar://) do NOT work!
// Google does not support passing event parameters via native deep links.
// This function exists for API completeness but the URLs it generates will not work.
//
// ✅ RECOMMENDED: Use web.GoogleCalendar() instead which generates working
// https://calendar.google.com URLs that open in any browser and work reliably.
//
// Returns an error if the event fails validation.
//
// Example:
//
//	event := types.CalendarEvent{
//	    Title:       "Team Meeting",
//	    StartTime:   time.Date(2025, 10, 26, 14, 0, 0, 0, time.UTC),
//	    EndTime:     time.Date(2025, 10, 26, 15, 0, 0, 0, time.UTC),
//	    Location:    "Conference Room A",
//	    Description: "Quarterly planning meeting",
//	}
//	url, err := google.Calendar(event)
//	// Returns: googlecalendar://render?action=CREATE&text=Team%20Meeting&dates=20251026T140000Z/20251026T150000Z&location=Conference%20Room%20A&details=Quarterly%20planning%20meeting
func Calendar(event types.CalendarEvent) (string, error) {
	// Validate the event
	if err := event.Validate(); err != nil {
		return "", err
	}

	// Parse the template
	tmpl, err := template.New("calendar").Parse(calendarTemplate)
	if err != nil {
		return "", err
	}

	// Prepare template data with formatted times and URL-encoded values
	data := struct {
		Title       string
		StartTime   string
		EndTime     string
		Location    string
		Description string
	}{
		Title:       url.QueryEscape(event.Title),
		StartTime:   formatTime(event.StartTime),
		EndTime:     formatTime(event.EndTime),
		Location:    url.QueryEscape(event.Location),
		Description: url.QueryEscape(event.Description),
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// formatTime converts a time.Time to Google Calendar's required format.
// Format: YYYYMMDDTHHMMSSZ (UTC, compact, no separators)
// Example: 2025-10-26 14:00:00 UTC → 20251026T140000Z
func formatTime(t time.Time) string {
	return t.UTC().Format("20060102T150405Z")
}
