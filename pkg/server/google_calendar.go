package server

import (
	"net/http"

	googlecalendar "github.com/joeblew999/wellknown/pkg/google/calendar"
)

// GoogleCalendar handles Google Calendar event creation with UI Schema form and validation
// Uses the generic calendar handler with map-based generator (NO Event structs!)
var GoogleCalendar = GenericCalendarHandler(CalendarConfig{
	Platform:     "google",
	AppType:      "calendar",
	SuccessLabel: "URL",
	GenerateURL:  googlecalendar.GenerateURL, // Takes map[string]interface{}, returns URL
})

// GoogleCalendarShowcase handles Google Calendar showcase page
// TODO: Re-implement with map-based examples instead of Event structs
func GoogleCalendarShowcase(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Showcase page temporarily disabled during migration", http.StatusServiceUnavailable)
}

// RegisterGoogleCalendarRoutes registers all Google Calendar routes with the given mux
func RegisterGoogleCalendarRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/google/calendar", GoogleCalendar)
	mux.HandleFunc("/google/calendar/showcase", GoogleCalendarShowcase)
	registerService(ServiceConfig{
		Platform:    "google",
		AppType:     "calendar",
		Title:       "Google Calendar",
		HasCustom:   true,
		HasShowcase: false, // Temporarily disabled
	})
}
