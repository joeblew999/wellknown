package handlers

import (
	"github.com/joeblew999/wellknown/pkg/apple"
	"github.com/joeblew999/wellknown/pkg/examples"
)

// AppleCalendarService is the registered Apple Calendar service
var AppleCalendarService = RegisterService(ServiceConfig{
	Platform:  "apple",
	AppType:   "calendar",
	Examples:  examples.CalendarExamples,
	Generator: apple.Calendar,
})

// AppleCalendar handles Apple Calendar custom event creation
var AppleCalendar = AppleCalendarService.CustomHandler

// AppleCalendarShowcase handles Apple Calendar showcase page
var AppleCalendarShowcase = AppleCalendarService.ShowcaseHandler
