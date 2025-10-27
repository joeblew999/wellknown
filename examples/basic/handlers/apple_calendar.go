package handlers

import (
	"github.com/joeblew999/wellknown/pkg/apple"
)

// AppleCalendarService is the registered Apple Calendar service
var AppleCalendarService = RegisterService(ServiceConfig{
	Platform:  "apple",
	AppType:   "calendar",
	Examples:  apple.CalendarExamples,
	Generator: apple.Calendar,
})

// AppleCalendar handles Apple Calendar custom event creation
var AppleCalendar = AppleCalendarService.CustomHandler

// AppleCalendarShowcase handles Apple Calendar showcase page
var AppleCalendarShowcase = AppleCalendarService.ShowcaseHandler
