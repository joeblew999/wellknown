package server

import (
	"net/http"

	applecalendar "github.com/joeblew999/wellknown/pkg/apple/calendar"
)

// AppleCalendar handles Apple Calendar custom event creation
func AppleCalendar(w http.ResponseWriter, r *http.Request) {
	// For now, just render the form - POST support can be added later
	renderCustomPage(w, r, "apple", "calendar", applecalendar.Examples)
}

// AppleCalendarShowcase handles Apple Calendar showcase page
func AppleCalendarShowcase(w http.ResponseWriter, r *http.Request) {
	renderShowcase(w, r, "apple", "calendar", applecalendar.Examples)
}
