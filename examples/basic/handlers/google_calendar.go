package handlers

import (
	"github.com/joeblew999/wellknown/pkg/google"
)

// GoogleCalendarConfig is the configuration for Google Calendar handlers
var GoogleCalendarConfig = ServiceConfig{
	Platform:  "google",
	AppType:   "calendar",
	Examples:  google.CalendarExamples,
	Generator: google.Calendar,
}

// GoogleCalendar handles Google Calendar custom event creation
var GoogleCalendar = CalendarHandler(GoogleCalendarConfig)

// GoogleCalendarShowcase handles Google Calendar showcase page
var GoogleCalendarShowcase = ShowcaseHandler(GoogleCalendarConfig)
