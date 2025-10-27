package handlers

import (
	"github.com/joeblew999/wellknown/pkg/examples"
	"github.com/joeblew999/wellknown/pkg/google"
)

// GoogleCalendarService is the registered Google Calendar service
var GoogleCalendarService = RegisterService(ServiceConfig{
	Platform:  "google",
	AppType:   "calendar",
	Examples:  examples.CalendarExamples,
	Generator: google.Calendar,
})

// GoogleCalendar handles Google Calendar custom event creation
var GoogleCalendar = GoogleCalendarService.CustomHandler

// GoogleCalendarShowcase handles Google Calendar showcase page
var GoogleCalendarShowcase = GoogleCalendarService.ShowcaseHandler
