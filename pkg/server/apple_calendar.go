package server

import (
	"encoding/base64"
	"fmt"
	"net/http"

	applecalendar "github.com/joeblew999/wellknown/pkg/apple/calendar"
)

// AppleCalendar handles Apple Calendar event creation with UI Schema form and validation
// Uses the generic calendar handler to eliminate code duplication
var AppleCalendar = GenericCalendarHandler(CalendarConfig{
	Platform:     "apple",
	AppType:      "calendar",
	SuccessLabel: "Download Link",
	BuildEvent: func(r *http.Request) (interface{}, error) {
		startTime, err := parseFormTime(r.FormValue("start"))
		if err != nil {
			return nil, err
		}
		endTime, err := parseFormTime(r.FormValue("end"))
		if err != nil {
			return nil, err
		}

		return applecalendar.Event{
			Title:       r.FormValue("title"),
			StartTime:   startTime,
			EndTime:     endTime,
			Location:    r.FormValue("location"),
			Description: r.FormValue("description"),
			AllDay:      r.FormValue("allDay") == "true",
		}, nil
	},
	GenerateURL: func(event interface{}) (string, error) {
		// Generate ICS content
		icsContent, err := event.(applecalendar.Event).GenerateICS()
		if err != nil {
			return "", err
		}
		// Encode as base64 for URL parameter
		encoded := base64.URLEncoding.EncodeToString(icsContent)
		// Return download endpoint URL
		return "/apple/calendar/download?event=" + encoded, nil
	},
})

// AppleCalendarShowcase handles Apple Calendar showcase page
// Uses ValidTestCases from testdata.go - comprehensive examples validated by JSON Schema
func AppleCalendarShowcase(w http.ResponseWriter, r *http.Request) {
	renderShowcase(w, r, "apple", "calendar", applecalendar.ValidTestCases)
}

// AppleCalendarDownload serves .ics file for download
// This is the CORRECT way to handle Apple Calendar on iOS/macOS
// Safari cannot handle data:text/calendar URIs - it requires actual file downloads
func AppleCalendarDownload(w http.ResponseWriter, r *http.Request) {
	// Get base64-encoded ICS content from query parameter
	eventParam := r.URL.Query().Get("event")
	if eventParam == "" {
		http.Error(w, "Missing event parameter", http.StatusBadRequest)
		return
	}

	// Decode base64 ICS content
	icsContent, err := base64.URLEncoding.DecodeString(eventParam)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid event data: %v", err), http.StatusBadRequest)
		return
	}

	// Set proper headers for .ics file
	// Using 'inline' instead of 'attachment' allows Safari to open Calendar.app automatically
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", "inline; filename=\"event.ics\"")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(icsContent)))

	// Write ICS content
	w.Write(icsContent)
}

// RegisterAppleCalendarRoutes registers all Apple Calendar routes with the given mux
func RegisterAppleCalendarRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/apple/calendar", AppleCalendar)
	mux.HandleFunc("/apple/calendar/showcase", AppleCalendarShowcase)
	mux.HandleFunc("/apple/calendar/download", AppleCalendarDownload)
	registerService(ServiceConfig{
		Platform:    "apple",
		AppType:     "calendar",
		Title:       "Apple Calendar",
		HasCustom:   true,
		HasShowcase: true,
	})
}
