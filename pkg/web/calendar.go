// Package web provides web fallback URL generators for when native apps aren't available.
package web

import (
	"fmt"
	"net/url"
	"time"

	"github.com/joeblew999/wellknown/pkg/types"
)

// GoogleCalendar generates a web URL for Google Calendar.
// This opens calendar.google.com in a browser instead of the native app.
//
// Use this for:
// - Testing deep links in a browser before testing on device
// - Fallback when Google Calendar app is not installed
// - Cross-platform compatibility (works everywhere)
//
// The web URL can be tested immediately by opening in a browser,
// unlike native deep links which require a mobile device.
func GoogleCalendar(event types.CalendarEvent) (string, error) {
	// Validate the event
	if err := event.Validate(); err != nil {
		return "", err
	}

	// Format times
	startTime := formatTime(event.StartTime)
	endTime := formatTime(event.EndTime)

	// Build URL
	// Note: Google Calendar web uses action=TEMPLATE (not CREATE like native app)
	baseURL := "https://calendar.google.com/calendar/render"
	params := url.Values{}
	params.Set("action", "TEMPLATE")
	params.Set("text", event.Title)
	params.Set("dates", fmt.Sprintf("%s/%s", startTime, endTime))

	if event.Location != "" {
		params.Set("location", event.Location)
	}

	if event.Description != "" {
		params.Set("details", event.Description)
	}

	return baseURL + "?" + params.Encode(), nil
}

// formatTime converts time.Time to Google Calendar web format
// Format: YYYYMMDDTHHMMSSZ (same as native app)
func formatTime(t time.Time) string {
	return t.UTC().Format("20060102T150405Z")
}
