package server

import (
	"log"
	"net/http"
	"time"

	googlecalendar "github.com/joeblew999/wellknown/pkg/google/calendar"
	"github.com/joeblew999/wellknown/pkg/schema"
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

// GoogleCalendarSchema handles Google Calendar with JSON Schema dynamic forms (Phase 7)
func GoogleCalendarSchema(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Load schema from external JSON file
		schemaJSON, err := loadSchemaFromFile("google", "calendar", "schema")
		if err != nil {
			http.Error(w, "Failed to load schema: "+err.Error(), http.StatusInternalServerError)
			return
		}
		renderSchemaBasedForm(w, r, "google", "calendar", schemaJSON)
		return
	}

	if r.Method == "POST" {
		// Use same POST handler as regular calendar
		handleGoogleCalendarPost(w, r)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// GoogleCalendarUISchema handles Google Calendar with UI Schema + JSON Schema (Phase 8 + 9)
func GoogleCalendarUISchema(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Load schemas from external JSON files
		schemaJSON, err := loadSchemaFromFile("google", "calendar", "schema")
		if err != nil {
			http.Error(w, "Failed to load schema: "+err.Error(), http.StatusInternalServerError)
			return
		}
		uiSchemaJSON, err := loadSchemaFromFile("google", "calendar", "uischema")
		if err != nil {
			http.Error(w, "Failed to load UI schema: "+err.Error(), http.StatusInternalServerError)
			return
		}
		renderUISchemaBasedForm(w, r, "google", "calendar", schemaJSON, uiSchemaJSON)
		return
	}

	if r.Method == "POST" {
		// Handle POST with validation (Phase 9)
		handleGoogleCalendarUISchemaPost(w, r)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handleGoogleCalendarUISchemaPost handles POST with validation for UI Schema forms
func handleGoogleCalendarUISchemaPost(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Load schemas from external JSON files
	schemaJSON, err := loadSchemaFromFile("google", "calendar", "schema")
	if err != nil {
		http.Error(w, "Failed to load schema: "+err.Error(), http.StatusInternalServerError)
		return
	}
	uiSchemaJSON, err := loadSchemaFromFile("google", "calendar", "uischema")
	if err != nil {
		http.Error(w, "Failed to load UI schema: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse JSON Schema for validation
	jsonSchema, err := schema.ParseSchema(schemaJSON)
	if err != nil {
		log.Printf("Schema parse error: %v", err)
		http.Error(w, "Failed to parse schema: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert form data to map
	formData := schema.FormDataToMap(r.Form)

	// Validate against schema
	validationErrors := schema.ValidateAgainstSchema(formData, jsonSchema)

	// If there are validation errors, re-render form with errors
	if len(validationErrors) > 0 {
		log.Printf("Validation errors: %v", validationErrors)
		renderUISchemaBasedFormWithErrors(w, r, "google", "calendar", schemaJSON, uiSchemaJSON, formData, validationErrors)
		return
	}

	// Validation passed - proceed with generating URL
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
