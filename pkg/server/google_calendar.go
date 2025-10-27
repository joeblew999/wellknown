package server

import (
	"log"
	"net/http"
	"time"

	googlecalendar "github.com/joeblew999/wellknown/pkg/google/calendar"
)

// GoogleCalendar handles Google Calendar custom event creation with GET form and POST support
func GoogleCalendar(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Use hardcoded form (schema-based forms are WIP in Phase 7)
		renderCustomPage(w, r, "google", "calendar", googlecalendar.Examples)
		return
	}

	if r.Method == "POST" {
		handleGoogleCalendarPost(w, r)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handleGoogleCalendarPost handles POST requests to create Google Calendar events
func handleGoogleCalendarPost(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Parse start and end times
	startTime, err := time.Parse("2006-01-02T15:04", r.FormValue("start"))
	if err != nil {
		http.Error(w, "Invalid start time: "+err.Error(), http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse("2006-01-02T15:04", r.FormValue("end"))
	if err != nil {
		http.Error(w, "Invalid end time: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Create Google Calendar event
	event := googlecalendar.Event{
		Title:       r.FormValue("title"),
		StartTime:   startTime,
		EndTime:     endTime,
		Location:    r.FormValue("location"),
		Description: r.FormValue("description"),
	}

	// Generate URL
	url, err := event.GenerateURL()
	if err != nil {
		http.Error(w, "Failed to generate URL: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("SUCCESS! Generated URL: %s", url)
	renderSuccess(w, "google", "calendar", url)
}

// GoogleCalendarShowcase handles Google Calendar showcase page
func GoogleCalendarShowcase(w http.ResponseWriter, r *http.Request) {
	renderShowcase(w, r, "google", "calendar", googlecalendar.Examples)
}

// GoogleCalendarSchema handles Google Calendar with JSON Schema dynamic forms (WIP - Phase 7)
func GoogleCalendarSchema(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Use schema-based form generation
		renderSchemaBasedForm(w, r, "google", "calendar", googlecalendar.Schema)
		return
	}

	if r.Method == "POST" {
		// Use same POST handler as regular calendar
		handleGoogleCalendarPost(w, r)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
