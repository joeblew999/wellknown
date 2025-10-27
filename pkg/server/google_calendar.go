package server

import (
	"net/http"

	googlecalendar "github.com/joeblew999/wellknown/pkg/google/calendar"
)

// GoogleCalendar handles Google Calendar event creation with UI Schema form and validation
// Uses the generic calendar handler to eliminate code duplication
var GoogleCalendar = GenericCalendarHandler(CalendarConfig{
	Platform:     "google",
	AppType:      "calendar",
	SuccessLabel: "URL",
	BuildEvent: func(r *http.Request) (interface{}, error) {
		startTime, err := parseFormTime(r.FormValue("start"))
		if err != nil {
			return nil, err
		}
		endTime, err := parseFormTime(r.FormValue("end"))
		if err != nil {
			return nil, err
		}

		return googlecalendar.Event{
			Title:       r.FormValue("title"),
			StartTime:   startTime,
			EndTime:     endTime,
			Location:    r.FormValue("location"),
			Description: r.FormValue("description"),
		}, nil
	},
	GenerateURL: func(event interface{}) (string, error) {
		return event.(googlecalendar.Event).GenerateURL()
	},
})

// GoogleCalendarShowcase handles Google Calendar showcase page
// Uses ValidTestCases from testdata.go - comprehensive examples validated by JSON Schema
func GoogleCalendarShowcase(w http.ResponseWriter, r *http.Request) {
	renderShowcase(w, r, "google", "calendar", googlecalendar.ValidTestCases)
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
		HasShowcase: true,
	})
}
